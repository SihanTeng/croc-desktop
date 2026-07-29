package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildHistoryItem(t *testing.T) {
	accept := &acceptPayload{
		TotalSize: 12,
		Files: []fileEntry{
			{Name: "a.txt", Size: 5},
			{Name: "b.png", Folder: "sub", Size: 7},
		},
	}

	// sender, files
	it := buildHistoryItem(true, "completed", "", nil, donePayload{}, "",
		[]historyFile{{Name: "a.txt", Size: 5}, {Name: "sub/b.png", Size: 7}}, 2, 12, "")
	if it.Direction != "send" || it.IsText || len(it.Files) != 2 || it.TotalSize != 12 {
		t.Errorf("sender file entry wrong: %+v", it)
	}

	// sender, text (files must be suppressed)
	it = buildHistoryItem(true, "completed", "", nil, donePayload{}, "/tmp",
		[]historyFile{{Name: "croc-text-123", Size: 5}}, 1, 5, "hello")
	if !it.IsText || it.Text != "hello" || len(it.Files) != 0 {
		t.Errorf("sender text entry wrong: %+v", it)
	}

	// receiver, files — folder joined into the display name
	it = buildHistoryItem(false, "completed", "", accept, donePayload{}, "/dl", nil, 0, 0, "")
	if it.Direction != "receive" || it.Dir != "/dl" || len(it.Files) != 2 || it.Files[1].Name != "sub/b.png" {
		t.Errorf("receiver file entry wrong: %+v", it)
	}

	// receiver, text — wrapper file suppressed, snippet kept
	it = buildHistoryItem(false, "completed", "", accept,
		donePayload{IsText: true, Text: "hi there"}, "/dl", nil, 0, 0, "")
	if !it.IsText || it.Text != "hi there" || len(it.Files) != 0 || it.TotalFiles != 0 {
		t.Errorf("receiver text entry wrong: %+v", it)
	}

	// error keeps the message and direction
	it = buildHistoryItem(true, "error", "boom", nil, donePayload{}, "", nil, 0, 0, "")
	if it.Status != "error" || it.Error != "boom" || it.Direction != "send" {
		t.Errorf("error entry wrong: %+v", it)
	}
}

func TestBuildHistoryItemFileCap(t *testing.T) {
	var accept acceptPayload
	for range maxHistoryFiles + 10 {
		accept.Files = append(accept.Files, fileEntry{Name: "f.txt", Size: 1})
	}
	it := buildHistoryItem(false, "completed", "", &accept, donePayload{}, "/dl", nil, 0, 0, "")
	if len(it.Files) != maxHistoryFiles {
		t.Fatalf("files not capped: got %d, want %d", len(it.Files), maxHistoryFiles)
	}
	if it.TotalFiles != maxHistoryFiles+10 {
		t.Fatalf("total lost: got %d, want %d", it.TotalFiles, maxHistoryFiles+10)
	}
}

func TestGetFileDataURLSniffing(t *testing.T) {
	// an unknown extension must fall back to content sniffing
	dir := t.TempDir()
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0, 0, 0, 0, 0}
	path := filepath.Join(dir, "no-extension")
	if err := os.WriteFile(path, png, 0o644); err != nil {
		t.Fatal(err)
	}
	a := NewApp()
	u, err := a.GetFileDataURL(path)
	if err != nil {
		t.Fatalf("GetFileDataURL: %v", err)
	}
	if !strings.HasPrefix(u, "data:image/png;base64,") {
		t.Fatalf("sniffing failed, got prefix %q", u[:30])
	}
}
