package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogManager(t *testing.T) {
	m := newLogManager("")
	var streamed []logEntry
	m.emit = func(event string, data interface{}) {
		if event != eventLogEntry {
			t.Fatalf("unexpected event %q", event)
		}
		streamed = append(streamed, data.(logEntry))
	}

	m.log(levelInfo, "test", "hello %s", "world")
	m.log(levelError, "test", "boom")

	entries := m.list()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Message != "hello world" || entries[0].Level != levelInfo {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Level != levelError || entries[1].Source != "test" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
	// every entry is streamed as it happens, with unique sequential IDs
	if len(streamed) != 2 || streamed[0].ID == streamed[1].ID {
		t.Fatalf("streamed entries wrong: %+v", streamed)
	}

	m.clear()
	if got := m.list(); len(got) != 0 {
		t.Fatalf("expected empty log after clear, got %d entries", len(got))
	}
}

func TestLogManagerCap(t *testing.T) {
	m := newLogManager("")
	for i := range maxLogEntries + 25 {
		m.log(levelDebug, "test", "entry %d", i)
	}
	entries := m.list()
	if len(entries) != maxLogEntries {
		t.Fatalf("got %d entries, want cap %d", len(entries), maxLogEntries)
	}
	// the oldest entries were dropped; the tail survives
	last := entries[len(entries)-1]
	if want := fmt.Sprintf("entry %d", maxLogEntries+24); last.Message != want {
		t.Fatalf("last entry = %q, want %q", last.Message, want)
	}
}

func TestLogManagerPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "croc-desktop.log")

	m := newLogManager(path)
	m.log(levelInfo, "app", "started")
	m.log(levelError, "transfer", "could not connect to 198.18.0.53:9009: EOF")

	// file has two JSON lines
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	_ = f.Close()
	if len(lines) != 2 {
		t.Fatalf("disk lines: got %d, want 2", len(lines))
	}
	var e logEntry
	if err := json.Unmarshal([]byte(lines[1]), &e); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(e.Message, "198.18.0.53") || e.Level != levelError {
		t.Fatalf("disk entry wrong: %+v", e)
	}

	// reload survives process restart
	m2 := newLogManager(path)
	got := m2.list()
	if len(got) != 2 {
		t.Fatalf("reloaded %d entries, want 2", len(got))
	}
	if got[1].Message != e.Message {
		t.Fatalf("reload mismatch: %q vs %q", got[1].Message, e.Message)
	}
	// seq continues past loaded max
	m2.log(levelInfo, "app", "next")
	if got2 := m2.list(); len(got2) != 3 || got2[2].ID == "1" {
		t.Fatalf("seq did not advance past loaded IDs: %+v", got2)
	}
}

func TestLogManagerRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "croc-desktop.log")
	m := newLogManager(path)

	// pre-fill the active log past the rotation threshold
	pad := strings.Repeat("x", 400)
	var content strings.Builder
	for content.Len() < logMaxBytes {
		e := logEntry{
			ID:      "0",
			Time:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Level:   levelInfo,
			Source:  "pad",
			Message: pad,
		}
		b, _ := json.Marshal(e)
		content.Write(b)
		content.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(content.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	m.log(levelWarn, "transfer", "after rotate")

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("expected rotated file %s.1: %v", path, err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() >= logMaxBytes {
		t.Fatalf("active log still huge after rotate: %d", st.Size())
	}

	m3 := newLogManager(path)
	found := false
	for _, e := range m3.list() {
		if e.Message == "after rotate" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("rotated log did not retain the post-rotate entry")
	}
}

func TestLogManagerClearRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "croc-desktop.log")
	m := newLogManager(path)
	m.log(levelInfo, "app", "hi")
	// fake a rotated sibling
	if err := os.WriteFile(path+".1", []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m.clear()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("active log should be gone, err=%v", err)
	}
	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Fatalf("rotated log should be gone, err=%v", err)
	}
	if len(m.list()) != 0 {
		t.Fatal("memory buffer not cleared")
	}
}
