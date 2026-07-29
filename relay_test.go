package main

import (
	"testing"
	"time"
)

func TestRelayStartStop(t *testing.T) {
	rm := newRelayManager()

	// too few ports is an error
	if err := rm.start([]string{"29209"}, "pass123"); err == nil {
		t.Fatal("expected an error for a single port")
	}
	if rm.state().Running {
		t.Fatal("relay should not be running after a failed start")
	}

	ports := []string{"29209", "29210", "29211", "29212", "29213"}
	if err := rm.start(ports, "pass123"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := waitForPort("29209", 10*time.Second); err != nil {
		t.Fatalf("relay port never came up: %v", err)
	}
	st := rm.state()
	if !st.Running || len(st.Ports) != len(ports) {
		t.Fatalf("unexpected state after start: %+v", st)
	}

	// a second start while running is rejected
	if err := rm.start(ports, "pass123"); err == nil {
		t.Fatal("expected an error for a duplicate start")
	}

	// blank entries are dropped before validation
	rm.stop()
	if rm.state().Running {
		t.Fatal("relay still running after stop")
	}
	if err := rm.start([]string{" 29209 ", "", "29210"}, "pass123"); err != nil {
		t.Fatalf("restart with padded/blank ports: %v", err)
	}
	if got := rm.state().Ports; len(got) != 2 || got[0] != "29209" {
		t.Fatalf("ports not cleaned: %+v", got)
	}
	rm.stop()
}

func TestRelayPasswordDefault(t *testing.T) {
	rm := newRelayManager()
	defer rm.stop()
	// empty password must not error (falls back to the croc default)
	if err := rm.start([]string{"29219", "29220"}, ""); err != nil {
		t.Fatalf("start with empty password: %v", err)
	}
	if err := waitForPort("29219", 10*time.Second); err != nil {
		t.Fatal(err)
	}
}
