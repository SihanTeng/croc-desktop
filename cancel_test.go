package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Regression: clicking Cancel on the Send page while waiting for a recipient
// must end the transfer promptly and emit the "cancelled" state.
func TestCancelWhileWaiting(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "cancel-me.txt")
	if err := os.WriteFile(src, []byte("cancel regression payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestApp()
	if _, err := sender.StartSend([]string{src}); err != nil {
		t.Fatalf("StartSend: %v", err)
	}

	// nobody receives; let the sender settle into "waiting"
	time.Sleep(1500 * time.Millisecond)
	sender.CancelTransfer()

	waitState(t, senderLog, "cancelled")
}

// Same as above, but with the shipped defaults (DisableLocal: false) so the
// sender also runs its local relay + multicast broadcast.
func TestCancelWhileWaitingDefaultSettings(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "cancel-me-local.txt")
	if err := os.WriteFile(src, []byte("cancel regression payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestAppDefault()
	if _, err := sender.StartSend([]string{src}); err != nil {
		t.Fatalf("StartSend: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)
	sender.CancelTransfer()

	waitState(t, senderLog, "cancelled")
}

// Cancelling mid-transfer on the receiver side must end both sides.
func TestCancelWhileTransferring(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "big.bin")
	// 64 MiB — big enough that localhost still takes a moment
	if err := os.WriteFile(src, make([]byte, 64*1024*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestApp()
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(true) }

	code, err := sender.StartSend([]string{src})
	if err != nil {
		t.Fatalf("StartSend: %v", err)
	}
	outDir := t.TempDir()
	if err := receiver.StartReceive(code, outDir); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}

	// wait for the transfer to actually start, then cancel the receiver
	waitState(t, receiverLog, "transferring")
	receiver.CancelTransfer()
	waitState(t, receiverLog, "cancelled")

	// the sender must also terminate (error or cancelled), not hang
	deadline := time.After(60 * time.Second)
	for {
		select {
		case <-senderLog.errs:
			return
		case <-senderLog.done:
			return
		case s := <-senderLog.states:
			if s == "cancelled" {
				return
			}
		case <-deadline:
			t.Fatal("sender did not terminate after receiver cancelled")
		}
	}
}
