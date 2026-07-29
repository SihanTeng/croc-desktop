package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// maxLogEntries bounds the in-memory log; the frontend mirrors it.
const maxLogEntries = 500

type logEntry struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"` // debug | info | warn | error
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

// logManager is the app's centralized logger: every subsystem records
// leveled entries here, kept in a bounded in-memory buffer (newest last) and
// streamed to the frontend as they happen.
type logManager struct {
	mu      sync.Mutex
	emit    func(event string, data interface{})
	entries []logEntry
	seq     int
}

func newLogManager() *logManager {
	return &logManager{}
}

func (m *logManager) setWailsCtx(ctx context.Context) {
	m.mu.Lock()
	m.emit = func(event string, data interface{}) {
		wailsruntime.EventsEmit(ctx, event, data)
	}
	m.mu.Unlock()
}

// log records one entry and streams it to the frontend.
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

func (m *logManager) clear() {
	m.mu.Lock()
	m.entries = nil
	m.mu.Unlock()
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
