package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/schollz/croc/v10/src/utils"
)

// caps keep one noisy transfer from flooding the history file
const (
	maxHistoryItems = 500 // entries kept on disk
	maxHistoryFiles = 50  // file names listed per entry
	maxHistoryText  = 500 // chars of a text transfer kept per entry
)

type historyFile struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"` // absolute, for previews
	Size int64  `json:"size"`
}

type historyItem struct {
	ID         string        `json:"id"`
	Time       time.Time     `json:"time"`
	Direction  string        `json:"direction"` // send | receive
	Status     string        `json:"status"`    // completed | cancelled | error
	Error      string        `json:"error,omitempty"`
	IsText     bool          `json:"isText,omitempty"`
	Text       string        `json:"text,omitempty"`
	Files      []historyFile `json:"files,omitempty"`
	TotalFiles int           `json:"totalFiles,omitempty"`
	TotalSize  int64         `json:"totalSize,omitempty"`
	Dir        string        `json:"dir,omitempty"` // receive destination
}

// historyManager keeps the transfer history (newest first) and persists it
// as JSON next to the settings file.
type historyManager struct {
	mu    sync.Mutex
	path  string // empty means in-memory only
	items []historyItem
}

func historyPath() string {
	dir, err := utils.GetConfigDir(false)
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "croc-desktop-history.json")
}

func newHistoryManager(path string) *historyManager {
	h := &historyManager{path: path}
	if path != "" {
		b, err := os.ReadFile(path)
		fromLegacy := false
		if err != nil {
			// fall back to the pre-rename history file (croc-gui → croc-desktop)
			legacy := filepath.Join(filepath.Dir(path), "croc-gui-history.json")
			b, err = os.ReadFile(legacy)
			fromLegacy = err == nil
		}
		if err == nil {
			_ = json.Unmarshal(b, &h.items)
			// migrate onto the current filename so later loads don't need the legacy path
			if fromLegacy {
				h.saveLocked()
			}
		}
	}
	return h
}

func (h *historyManager) add(item historyItem) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if item.ID == "" {
		item.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	h.items = append([]historyItem{item}, h.items...)
	if len(h.items) > maxHistoryItems {
		h.items = h.items[:maxHistoryItems]
	}
	h.saveLocked()
}

func (h *historyManager) list() []historyItem {
	h.mu.Lock()
	defer h.mu.Unlock()
	items := make([]historyItem, len(h.items))
	copy(items, h.items)
	return items
}

func (h *historyManager) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = nil
	h.saveLocked()
}

func (h *historyManager) saveLocked() {
	if h.path == "" {
		return
	}
	b, err := json.MarshalIndent(h.items, "", "  ")
	if err != nil {
		return
	}
	// the config dir is only created lazily (mirrors saveSettings)
	if err := os.MkdirAll(filepath.Dir(h.path), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(h.path, b, 0o644)
}

func truncateText(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
