package main

import (
	"strings"
	"testing"
)

// fast validation paths that need no relay or network

func TestStartSendValidation(t *testing.T) {
	a := NewApp()
	if _, err := a.StartSend(nil); err == nil || !strings.Contains(err.Error(), "no files") {
		t.Fatalf("expected 'no files' error, got %v", err)
	}
	if _, err := a.StartSend([]string{"/nonexistent/path/file.txt"}); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestStartSendTextValidation(t *testing.T) {
	a := NewApp()
	if _, err := a.StartSendText("   \n "); err == nil || !strings.Contains(err.Error(), "no text") {
		t.Fatalf("expected 'no text' error, got %v", err)
	}
}

func TestStartReceiveValidation(t *testing.T) {
	a := NewApp()
	if err := a.StartReceive("abc", t.TempDir()); err == nil ||
		!strings.Contains(err.Error(), "too short") {
		t.Fatalf("expected 'too short' error, got %v", err)
	}
}

func TestGetLogsEmpty(t *testing.T) {
	// Use isolated app so we don't load the real on-disk log from other tests
	// or a previous desktop session.
	a, _ := newTestApp()
	// must be a non-nil empty slice so the frontend gets [] rather than null
	logs := a.GetLogs()
	if logs == nil || len(logs) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", logs)
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1000, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{2500000, "2.5 MB"},
		{1073741824, "1.1 GB"},
	}
	for _, c := range cases {
		if got := humanBytes(c.in); got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
