package main

import (
	"fmt"
	"testing"
)

func TestLogManager(t *testing.T) {
	m := newLogManager()
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
	m := newLogManager()
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
