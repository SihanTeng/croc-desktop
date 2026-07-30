package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryPersistence(t *testing.T) {
	// nested dir does not exist yet: add() must create it
	path := filepath.Join(t.TempDir(), "sub", "history.json")

	h := newHistoryManager(path)
	h.add(historyItem{ID: "1", Direction: "send", Status: "completed",
		Files: []historyFile{{Name: "a.txt", Size: 3}}})
	h.add(historyItem{ID: "2", Direction: "receive", Status: "error", Error: "boom"})

	// a fresh manager on the same file sees the saved entries, newest first
	h2 := newHistoryManager(path)
	items := h2.list()
	if len(items) != 2 {
		t.Fatalf("loaded %d entries, want 2", len(items))
	}
	if items[0].ID != "2" || items[1].ID != "1" {
		t.Fatalf("wrong order: %+v", items)
	}
	if items[1].Files[0].Name != "a.txt" || items[1].Files[0].Size != 3 {
		t.Fatalf("files did not round-trip: %+v", items[1].Files)
	}

	// clear removes entries and persists the removal
	h2.clear()
	if items := newHistoryManager(path).list(); len(items) != 0 {
		t.Fatalf("expected empty history after clear, got %+v", items)
	}
}

func TestHistoryCap(t *testing.T) {
	h := newHistoryManager("")
	for range maxHistoryItems + 10 {
		h.add(historyItem{Direction: "send", Status: "completed"})
	}
	if got := len(h.list()); got != maxHistoryItems {
		t.Fatalf("history length: got %d, want capped %d", got, maxHistoryItems)
	}
}

func TestHistoryIgnoresCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if items := newHistoryManager(path).list(); len(items) != 0 {
		t.Fatalf("corrupt file should load as empty, got %+v", items)
	}
}

// Pre-rename history files must load and be rewritten under the new name.
func TestHistoryLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "croc-gui-history.json")
	current := filepath.Join(dir, "croc-desktop-history.json")
	payload := `[{"id":"old","direction":"send","status":"completed"}]`
	if err := os.WriteFile(legacy, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}

	h := newHistoryManager(current)
	items := h.list()
	if len(items) != 1 || items[0].ID != "old" {
		t.Fatalf("legacy history not loaded: %+v", items)
	}
	if _, err := os.Stat(current); err != nil {
		t.Fatalf("expected migration write to %s: %v", current, err)
	}
	_ = os.Remove(legacy)
	h2 := newHistoryManager(current)
	if got := h2.list(); len(got) != 1 || got[0].ID != "old" {
		t.Fatalf("post-migration load failed: %+v", got)
	}
}

func TestTruncateText(t *testing.T) {
	short := "hello"
	if got := truncateText(short, 10); got != short {
		t.Fatalf("short text changed: %q", got)
	}
	long := strings.Repeat("x", maxHistoryText*2)
	got := truncateText(long, maxHistoryText)
	if len([]rune(got)) != maxHistoryText+1 || !strings.HasSuffix(got, "…") {
		t.Fatalf("truncation wrong: len=%d", len([]rune(got)))
	}
	// multibyte runes must not be split
	mb := strings.Repeat("🐊", maxHistoryText+1)
	if got := truncateText(mb, maxHistoryText); len([]rune(got)) != maxHistoryText+1 {
		t.Fatalf("multibyte truncation split runes: len=%d", len([]rune(got)))
	}
}
