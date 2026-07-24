package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Mirrors the real-app failure: default settings (local relay enabled) with
// the configured relay UNREACHABLE, then Cancel while "Waiting for
// recipient…". This is the path a sender takes when the public relay can't
// be reached (offline, blocked, wrong address).
func TestCancelWhileWaitingRelayUnreachable(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "cancel-dead-relay.txt")
	if err := os.WriteFile(src, []byte("cancel regression payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestAppDefault()
	sender.settings.RelayAddress = "127.0.0.1:1" // nothing listening
	sender.settings.RelayAddress6 = ""

	if _, err := sender.StartSend([]string{src}); err != nil {
		t.Fatalf("StartSend: %v", err)
	}
	time.Sleep(8 * time.Second) // let the dead-relay dial time out first
	sender.CancelTransfer()

	waitState(t, senderLog, "cancelled")
}
