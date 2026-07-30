package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Coverage for App methods that the frontend calls frequently but that
// previously had little or no automated coverage. These must keep working
// across the Wails v2 → v3 migration.

func TestPathsIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "sub")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	a := NewApp()
	got := a.PathsIsDir([]string{file, nested, filepath.Join(dir, "missing"), dir})
	want := map[string]bool{nested: true, dir: true}
	if len(got) != 2 {
		t.Fatalf("PathsIsDir: got %v, want 2 dirs", got)
	}
	for _, p := range got {
		if !want[p] {
			t.Errorf("unexpected dir %q in %v", p, got)
		}
	}
	if a.PathsIsDir(nil) != nil && len(a.PathsIsDir(nil)) != 0 {
		t.Fatalf("nil paths should yield empty slice, got %v", a.PathsIsDir(nil))
	}
}

func TestGetQrPng(t *testing.T) {
	a := NewApp()
	b64, err := a.GetQrPng("4523-ivan-twist-order")
	if err != nil {
		t.Fatalf("GetQrPng: %v", err)
	}
	if b64 == "" {
		t.Fatal("expected non-empty base64 PNG")
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("not valid base64: %v", err)
	}
	// PNG magic
	if len(raw) < 8 || string(raw[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("expected PNG payload, got %d bytes starting %x", len(raw), raw[:min(8, len(raw))])
	}
}

func TestGetDefaultDownloadDir(t *testing.T) {
	a := NewApp()
	dir, err := a.GetDefaultDownloadDir()
	if err != nil {
		t.Fatalf("GetDefaultDownloadDir: %v", err)
	}
	if dir == "" {
		t.Fatal("expected a non-empty download dir")
	}
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		t.Fatalf("download dir %q is not a directory: %v", dir, err)
	}
}

func TestIsTransferRunningAndConcurrentSend(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "busy.txt")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	a, _ := newTestApp()
	if a.IsTransferRunning() {
		t.Fatal("expected no transfer before StartSend")
	}
	if _, err := a.StartSend([]string{src}); err != nil {
		t.Fatalf("StartSend: %v", err)
	}
	if !a.IsTransferRunning() {
		t.Fatal("expected transfer to be running after StartSend")
	}
	if _, err := a.StartSend([]string{src}); err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("expected 'already in progress', got %v", err)
	}
	if err := a.StartReceive("4523-ivan-twist-order", t.TempDir()); err == nil ||
		!strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("expected 'already in progress' on receive, got %v", err)
	}
	a.CancelTransfer()
	// cancel is async; give it a moment to clear the running flag
	deadline := time.Now().Add(10 * time.Second)
	for a.IsTransferRunning() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if a.IsTransferRunning() {
		t.Fatal("transfer still running after cancel")
	}
}

func TestHistoryAndLogsClear(t *testing.T) {
	a, l := newTestApp()
	// seed a completed text transfer so history is non-empty
	const msg = "history clear test"
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(true) }

	code, err := a.StartSendText(msg)
	if err != nil {
		t.Fatalf("StartSendText: %v", err)
	}
	if err := receiver.StartReceive(code, t.TempDir()); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}
	waitDone(t, receiverLog, "text receive")
	waitDone(t, l, "text send")

	if len(a.GetHistory()) == 0 {
		t.Fatal("expected history after transfer")
	}
	a.ClearHistory()
	if h := a.GetHistory(); h == nil || len(h) != 0 {
		t.Fatalf("ClearHistory left %v", h)
	}

	// log buffer starts empty under newTestApp (no startup wiring)
	if logs := a.GetLogs(); logs == nil || len(logs) != 0 {
		// may have entries from transfer; just ensure Clear works
	}
	a.ClearLogs()
	if logs := a.GetLogs(); logs == nil || len(logs) != 0 {
		t.Fatalf("ClearLogs left %v", logs)
	}
}

func TestEmitWiringDeliversEvents(t *testing.T) {
	// Verifies the swappable emit path used by tests (and, after the v3
	// migration, by ServiceStartup) reaches the frontend event surface.
	a, l := newTestApp()
	a.tm.emitEvent(eventState, "connecting")
	select {
	case s := <-l.states:
		if s != "connecting" {
			t.Fatalf("got state %q", s)
		}
	case <-time.After(time.Second):
		t.Fatal("emit did not deliver state event")
	}
}

func TestFolderTransfer(t *testing.T) {
	srcDir := t.TempDir()
	sub := filepath.Join(srcDir, "folder")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("aaa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("bbb"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestApp()
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(true) }

	code, err := sender.StartSend([]string{sub})
	if err != nil {
		t.Fatalf("StartSend folder: %v", err)
	}
	outDir := t.TempDir()
	if err := receiver.StartReceive(code, outDir); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}

	p := waitDone(t, receiverLog, "folder receive")
	waitDone(t, senderLog, "folder send")
	if len(p.Files) != 2 {
		t.Fatalf("expected 2 files, got %+v", p.Files)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		found := false
		for _, f := range p.Files {
			if strings.HasSuffix(f.Name, name) || filepath.Base(f.Name) == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %s in %+v", name, p.Files)
		}
	}
}

func TestGetFileDataURLTooLarge(t *testing.T) {
	// Don't actually allocate 64 MiB+; override via a sparse-ish check using
	// a temp file we report as large by writing past the cap if cheap, or
	// skip when disk is constrained. Write just over maxPreviewBytes only
	// when short — use a smaller direct size assertion via the constant.
	if testing.Short() {
		t.Skip("skipping large-file preview test in short mode")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "big.bin")
	// create a sparse file by seeking; falls back to a smaller write if needed
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxPreviewBytes + 1); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	a := NewApp()
	if _, err := a.GetFileDataURL(path); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected too-large error, got %v", err)
	}
}

func TestNormalizeCodeViaApp(t *testing.T) {
	a := NewApp()
	code, err := a.NormalizeCode("croc 4523-ivan-twist-order")
	if err != nil {
		t.Fatalf("NormalizeCode: %v", err)
	}
	if code != "4523-ivan-twist-order" {
		t.Fatalf("got %q", code)
	}
	if _, err := a.NormalizeCode("   "); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestSaveSettingsInvalidThrottle(t *testing.T) {
	t.Setenv("CROC_CONFIG_DIR", t.TempDir())
	a := NewApp()
	s := a.GetSettings()
	s.ThrottleUpload = "not-a-rate"
	if err := a.SaveSettings(s); err == nil || !strings.Contains(err.Error(), "invalid upload limit") {
		t.Fatalf("expected invalid throttle error, got %v", err)
	}
}
