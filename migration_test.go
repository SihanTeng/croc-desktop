package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/schollz/croc/v11/src/croc"
)

func TestTextLimitCountsUTF8Bytes(t *testing.T) {
	a, _ := newTestApp()
	for _, text := range []string{strings.Repeat("x", maxTextTransferBytes+1), strings.Repeat("🐊", maxTextTransferBytes/4+1)} {
		if _, err := a.StartSendText(text); err == nil || !strings.Contains(err.Error(), "1 MiB") {
			t.Fatalf("expected text limit error, got %v", err)
		}
		if a.IsTransferRunning() {
			t.Fatal("oversized text started a transfer")
		}
	}
}

func TestTextLimitBoundaryAndCleanup(t *testing.T) {
	text := strings.Repeat("🐊", maxTextTransferBytes/4)
	sender, sent := newTestApp()
	receiver, received := newTestApp()
	received.onAccept = func() { receiver.RespondAccept(true) }
	code, err := sender.StartSendText(text)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := receiver.StartReceive(code, dir); err != nil {
		t.Fatal(err)
	}
	got := waitDone(t, received, "text at limit")
	waitDone(t, sent, "text send at limit")
	if !got.IsText || got.Text != text {
		t.Fatal("text at byte limit did not arrive intact")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("croc must clean up its temporary text file: %v, %v", entries, err)
	}
}

func TestV11OverwritePrompts(t *testing.T) {
	for _, empty := range []bool{false, true} {
		for _, accept := range []bool{false, true} {
			name := "regular"
			if empty {
				name = "empty"
			}
			if accept {
				name += "-accept"
			} else {
				name += "-decline"
			}
			t.Run(name, func(t *testing.T) {
				src := filepath.Join(t.TempDir(), "overwrite.txt")
				content := "new file contents"
				if empty {
					content = ""
				}
				if err := os.WriteFile(src, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				dest := filepath.Join(dir, "overwrite.txt")
				if err := os.WriteFile(dest, []byte("keep existing contents"), 0600); err != nil {
					t.Fatal(err)
				}
				sender, sent := newTestApp()
				receiver, received := newTestApp()
				received.onAccept = func() { receiver.RespondAccept(true) }
				prompted := false
				emit := receiver.tm.emit
				receiver.tm.emit = func(event string, data interface{}) {
					emit(event, data)
					if event == eventOverwrite {
						prompted = true
						receiver.RespondOverwrite(accept)
					}
				}
				code, err := sender.StartSend([]string{src})
				if err != nil {
					t.Fatal(err)
				}
				if err := receiver.StartReceive(code, dir); err != nil {
					t.Fatal(err)
				}
				waitDone(t, received, "overwrite receive")
				waitDone(t, sent, "overwrite send")
				if !prompted {
					t.Fatal("existing destination bypassed the desktop overwrite prompt")
				}
				want := "keep existing contents"
				if accept {
					want = content
				}
				got, err := os.ReadFile(dest)
				if err != nil || string(got) != want {
					t.Fatalf("destination = %q, %v; want %q", got, err, want)
				}
			})
		}
	}
}

func TestCancelWhileAcceptPromptOpen(t *testing.T) {
	src := filepath.Join(t.TempDir(), "pending.txt")
	if err := os.WriteFile(src, []byte("never accepted"), 0600); err != nil {
		t.Fatal(err)
	}
	sender, _ := newTestApp()
	receiver, received := newTestApp()
	received.onAccept = func() { receiver.CancelTransfer() }
	code, err := sender.StartSend([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := receiver.StartReceive(code, dir); err != nil {
		t.Fatal(err)
	}
	waitState(t, received, "cancelled")
	sender.CancelTransfer()
	deadline := time.Now().Add(5 * time.Second)
	for sender.IsTransferRunning() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if sender.IsTransferRunning() {
		t.Fatal("sender did not stop after receiver cancellation")
	}
	if _, err := os.Stat(filepath.Join(dir, "pending.txt")); !os.IsNotExist(err) {
		t.Fatalf("unaccepted file was written: %v", err)
	}
}

func TestDesktopKeepsRelayAndLANTransport(t *testing.T) {
	for _, sender := range []bool{false, true} {
		for _, local := range []bool{false, true} {
			settings := defaultSettings()
			settings.OnlyLocal = local
			opts := buildCrocOptions(settings, sender)
			if !opts.DisableTailcat {
				t.Fatal("desktop must not silently change configured network routing")
			}
			opts.SharedSecret = "1234-test-code"
			client, err := croc.New(opts)
			if err != nil {
				t.Fatalf("desktop options rejected for sender=%v local=%v: %v", sender, local, err)
			}
			client.Cancel()
		}
	}
}
