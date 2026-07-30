package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/schollz/croc/v10/src/utils"
)

// log levels
const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

// eventLogEntry is emitted to the frontend for every recorded entry.
const eventLogEntry = "log:entry"

// maxLogEntries bounds the in-memory log mirrored by the frontend.
const maxLogEntries = 500

// rotation defaults for the on-disk log (JSON lines).
const (
	logMaxBytes        = 2 << 20 // 2 MiB per file
	logMaxRotatedFiles = 3       // keep .log + .log.1 .. .log.N
)

type logEntry struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"` // debug | info | warn | error
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

// logManager is the app's centralized logger: every subsystem records
// leveled entries here. Entries are kept in a bounded in-memory buffer
// (newest last), streamed to the frontend, and appended to a rotating
// log file on disk when a path is configured.
type logManager struct {
	mu      sync.Mutex
	emit    func(event string, data interface{})
	entries []logEntry
	seq     int
	path    string // empty = memory only (tests)
}

func logPath() string {
	dir, err := utils.GetConfigDir(false)
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "croc-desktop.log")
}

// newLogManager creates a logger. When path is non-empty, existing entries
// are loaded from disk (tail of the current + rotated files) and new lines
// are appended with size-based rotation.
func newLogManager(path string) *logManager {
	m := &logManager{path: path}
	if path != "" {
		m.loadFromDisk()
	}
	return m
}

// setEmit wires the manager to a frontend event publisher.
func (m *logManager) setEmit(emit func(event string, data interface{})) {
	m.mu.Lock()
	m.emit = emit
	m.mu.Unlock()
}

// log records one entry, persists it, and streams it to the frontend.
func (m *logManager) log(level, source, format string, args ...interface{}) {
	m.mu.Lock()
	m.seq++
	e := logEntry{
		ID:      fmt.Sprintf("%d", m.seq),
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: fmt.Sprintf(format, args...),
	}
	m.entries = append(m.entries, e)
	if len(m.entries) > maxLogEntries {
		m.entries = m.entries[len(m.entries)-maxLogEntries:]
	}
	if m.path != "" {
		m.appendLocked(e)
	}
	emit := m.emit
	m.mu.Unlock()
	if emit != nil {
		emit(eventLogEntry, e)
	}
}

func (m *logManager) list() []logEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries := make([]logEntry, len(m.entries))
	copy(entries, m.entries)
	return entries
}

// clear empties the in-memory buffer and truncates the on-disk log (including
// rotated files).
func (m *logManager) clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
	if m.path == "" {
		return
	}
	_ = os.Remove(m.path)
	for i := 1; i <= logMaxRotatedFiles; i++ {
		_ = os.Remove(fmt.Sprintf("%s.%d", m.path, i))
	}
}

// appendLocked writes one JSON line and rotates if the file exceeds the size
// budget. Caller must hold m.mu.
func (m *logManager) appendLocked(e logEntry) {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return
	}
	// rotate before write if the current file is already at/over the limit
	if st, err := os.Stat(m.path); err == nil && st.Size() >= logMaxBytes {
		m.rotateLocked()
	}
	f, err := os.OpenFile(m.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
}

// rotateLocked renames croc-desktop.log → .1 → .2 … and drops the oldest.
// Caller must hold m.mu.
func (m *logManager) rotateLocked() {
	oldest := fmt.Sprintf("%s.%d", m.path, logMaxRotatedFiles)
	_ = os.Remove(oldest)
	for i := logMaxRotatedFiles - 1; i >= 1; i-- {
		from := fmt.Sprintf("%s.%d", m.path, i)
		to := fmt.Sprintf("%s.%d", m.path, i+1)
		_ = os.Rename(from, to)
	}
	_ = os.Rename(m.path, m.path+".1")
}

// loadFromDisk rebuilds the in-memory buffer from the current log and any
// rotated files (oldest → newest), keeping only the last maxLogEntries.
// Called from the constructor only (no concurrent access yet).
func (m *logManager) loadFromDisk() {
	var paths []string
	for i := logMaxRotatedFiles; i >= 1; i-- {
		paths = append(paths, fmt.Sprintf("%s.%d", m.path, i))
	}
	paths = append(paths, m.path)

	var loaded []logEntry
	var maxSeq int
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		// allow long log lines
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Bytes()
			if len(line) == 0 {
				continue
			}
			var e logEntry
			if err := json.Unmarshal(line, &e); err != nil {
				continue
			}
			loaded = append(loaded, e)
			var n int
			if _, err := fmt.Sscanf(e.ID, "%d", &n); err == nil && n > maxSeq {
				maxSeq = n
			}
		}
		_ = f.Close()
	}
	if len(loaded) > maxLogEntries {
		loaded = loaded[len(loaded)-maxLogEntries:]
	}
	m.entries = loaded
	m.seq = maxSeq
}

// convenience helpers on App so call sites read naturally
func (a *App) logDebug(source, format string, args ...interface{}) {
	a.lm.log(levelDebug, source, format, args...)
}
func (a *App) logInfo(source, format string, args ...interface{}) {
	a.lm.log(levelInfo, source, format, args...)
}
func (a *App) logWarn(source, format string, args ...interface{}) {
	a.lm.log(levelWarn, source, format, args...)
}
func (a *App) logError(source, format string, args ...interface{}) {
	a.lm.log(levelError, source, format, args...)
}
