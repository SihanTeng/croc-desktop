package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// These tests drive the real GUI backend (App methods + croc.Hooks wiring)
// end-to-end through an in-process relay, without a Wails window.

var testRelayPorts = []string{"11009", "11010", "11011", "11012"}

func TestMain(m *testing.M) {
	rm := newRelayManager()
	if err := rm.start(testRelayPorts, "pass123"); err != nil {
		fmt.Println("starting test relay:", err)
		os.Exit(1)
	}
	for _, p := range testRelayPorts {
		if err := waitForPort(p, 10*time.Second); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func waitForPort(port string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort("127.0.0.1", port)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("port %s never came up", port)
}

// eventLog captures emitted frontend events in place of the Wails runtime.
type eventLog struct {
	mu       sync.Mutex
	events   []string
	done     chan donePayload
	errs     chan string
	states   chan string
	onAccept func()
}

func (l *eventLog) emit(event string, data interface{}) {
	l.mu.Lock()
	l.events = append(l.events, event)
	l.mu.Unlock()
	switch event {
	case eventDone:
		if p, ok := data.(donePayload); ok {
			l.done <- p
		}
	case eventError:
		l.errs <- fmt.Sprint(data)
	case eventState:
		l.states <- fmt.Sprint(data)
	case eventAccept:
		if l.onAccept != nil {
			go l.onAccept()
		}
	}
}

func (l *eventLog) saw(event string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.events {
		if e == event {
			return true
		}
	}
	return false
}

func newTestApp() (*App, *eventLog) {
	return newTestAppWithLocal(true)
}

// newTestAppDefault mirrors the shipped defaults (DisableLocal: false), so the
// local-relay + multicast paths run like in the real app.
func newTestAppDefault() (*App, *eventLog) {
	return newTestAppWithLocal(false)
}

func newTestAppWithLocal(disableLocal bool) (*App, *eventLog) {
	l := &eventLog{
		done:   make(chan donePayload, 1),
		errs:   make(chan string, 1),
		states: make(chan string, 16),
	}
	a := NewApp()
	a.tm.emit = l.emit
	a.settings = Settings{
		RelayAddress:  "127.0.0.1:" + testRelayPorts[0],
		RelayPassword: "pass123",
		DisableLocal:  disableLocal,
		Curve:         "p256",
		HashAlgorithm: "xxhash",
	}
	return a, l
}

func waitDone(t *testing.T, l *eventLog, what string) donePayload {
	t.Helper()
	select {
	case p := <-l.done:
		return p
	case e := <-l.errs:
		t.Fatalf("%s failed: %s", what, e)
	case <-time.After(60 * time.Second):
		t.Fatalf("timeout waiting for %s", what)
	}
	return donePayload{}
}

func waitState(t *testing.T, l *eventLog, want string) {
	t.Helper()
	deadline := time.After(60 * time.Second)
	for {
		select {
		case s := <-l.states:
			if s == want {
				return
			}
		case <-deadline:
			t.Fatalf("timeout waiting for state %q", want)
		}
	}
}

func TestFileTransfer(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "hello.txt")
	content := strings.Repeat("croc gui integration test payload. ", 40000) // ~1.4 MB
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestApp()
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(true) }

	code, err := sender.StartSend([]string{src})
	if err != nil {
		t.Fatalf("StartSend: %v", err)
	}
	if code == "" {
		t.Fatal("expected a code phrase")
	}

	outDir := t.TempDir()
	if err := receiver.StartReceive(code, outDir); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}

	waitDone(t, receiverLog, "receive")
	waitDone(t, senderLog, "send")

	got, err := os.ReadFile(filepath.Join(outDir, "hello.txt"))
	if err != nil {
		t.Fatalf("reading received file: %v", err)
	}
	if string(got) != content {
		t.Fatalf("content mismatch: got %d bytes, want %d", len(got), len(content))
	}
	if !receiverLog.saw(eventAccept) {
		t.Error("receiver never got the accept prompt event")
	}
	if !receiverLog.saw(eventProgress) {
		t.Error("receiver never got progress events")
	}
}

func TestTextTransfer(t *testing.T) {
	const message = "hello from the croc gui 🐊"

	sender, _ := newTestApp()
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(true) }

	code, err := sender.StartSendText(message)
	if err != nil {
		t.Fatalf("StartSendText: %v", err)
	}

	outDir := t.TempDir()
	if err := receiver.StartReceive(code, outDir); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}

	p := waitDone(t, receiverLog, "text receive")
	if !p.IsText {
		t.Fatal("expected done payload to be marked as text")
	}
	if p.Text != message {
		t.Fatalf("text mismatch: got %q, want %q", p.Text, message)
	}
}

func TestDeclineTransfer(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "nope.txt")
	if err := os.WriteFile(src, []byte("you shall not pass"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender, senderLog := newTestApp()
	receiver, receiverLog := newTestApp()
	receiverLog.onAccept = func() { receiver.RespondAccept(false) }

	code, err := sender.StartSend([]string{src})
	if err != nil {
		t.Fatalf("StartSend: %v", err)
	}
	outDir := t.TempDir()
	if err := receiver.StartReceive(code, outDir); err != nil {
		t.Fatalf("StartReceive: %v", err)
	}

	// the declining receiver finishes as "cancelled"; the sender gets a
	// friendly decline error
	waitState(t, receiverLog, "cancelled")
	select {
	case e := <-senderLog.errs:
		if e != "The recipient declined the transfer" {
			t.Fatalf("unexpected sender error: %q", e)
		}
	case <-senderLog.done:
		t.Fatal("declined transfer reported as done on sender")
	case <-time.After(60 * time.Second):
		t.Fatal("timeout waiting for sender error")
	}
}
