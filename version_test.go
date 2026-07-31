package main

import "testing"

func TestGetAppInfo(t *testing.T) {
	a := NewApp()
	info := a.GetAppInfo()
	if info.Version == "" {
		t.Fatal("expected non-empty Version")
	}
	if info.Version != AppVersion {
		t.Fatalf("Version = %q, want %q", info.Version, AppVersion)
	}
	if info.DefaultRelay == "" {
		t.Fatal("expected DefaultRelay from croc defaults")
	}
	if info.CrocVersion == "" || info.CrocVersion == "unknown" {
		// under go test build info should still list the dependency
		t.Logf("CrocVersion=%q (may be pseudo-version under replace)", info.CrocVersion)
	}
}
