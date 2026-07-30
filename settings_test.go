package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/schollz/croc/v10/src/comm"
	"github.com/schollz/croc/v10/src/models"
)

func TestSettingsRoundTrip(t *testing.T) {
	t.Setenv("CROC_CONFIG_DIR", t.TempDir())

	// missing file → defaults
	s := loadSettings()
	if s.RelayAddress != models.DEFAULT_RELAY || s.Curve != "p256" || s.HashAlgorithm != "xxhash" {
		t.Fatalf("unexpected defaults: %+v", s)
	}

	s.DownloadDir = filepath.Join(t.TempDir(), "dl")
	s.Socks5 = "user:pass@proxy.local:1080"
	s.HttpProxy = "http://proxy.local:8080"
	s.OnlyLocal = true
	if err := saveSettings(s); err != nil {
		t.Fatalf("saveSettings: %v", err)
	}

	got := loadSettings()
	if !reflect.DeepEqual(got, s) {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", got, s)
	}
}

// Upgrades from the pre-rename config file must keep working and migrate
// onto croc-desktop.json so later loads do not depend on the legacy path.
func TestSettingsLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CROC_CONFIG_DIR", dir)

	legacy := filepath.Join(dir, "croc-gui.json")
	if err := os.WriteFile(legacy, []byte(`{"curve":"p384","theme":"dark"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := loadSettings()
	if s.Curve != "p384" || s.Theme != "dark" {
		t.Fatalf("legacy settings not loaded: %+v", s)
	}
	// current name should now exist
	current := filepath.Join(dir, "croc-desktop.json")
	if _, err := os.Stat(current); err != nil {
		t.Fatalf("expected migration write to %s: %v", current, err)
	}
	// loading again should not require the legacy file
	_ = os.Remove(legacy)
	s2 := loadSettings()
	if s2.Curve != "p384" || s2.Theme != "dark" {
		t.Fatalf("post-migration load failed: %+v", s2)
	}
}

func TestBuildCrocOptions(t *testing.T) {
	// defaults are filled in when fields are empty
	opts := buildCrocOptions(Settings{}, true)
	if opts.RelayPassword != models.DEFAULT_PASSPHRASE {
		t.Errorf("relay password default: got %q", opts.RelayPassword)
	}
	if opts.Curve != "p256" || opts.HashAlgorithm != "xxhash" {
		t.Errorf("crypto defaults: got %q / %q", opts.Curve, opts.HashAlgorithm)
	}
	if !opts.IsSender || !opts.NoPrompt || !opts.DisableClipboard {
		t.Errorf("flags not wired: %+v", opts)
	}

	// a custom IPv4 relay disables the default IPv6 counterpart (CLI logic)
	opts = buildCrocOptions(Settings{RelayAddress: "relay.example.com:9009"}, false)
	if opts.RelayAddress6 != "" {
		t.Errorf("custom v4 relay should clear v6, got %q", opts.RelayAddress6)
	}
	if opts.IsSender {
		t.Error("IsSender should be false for receives")
	}

	// a custom IPv6 relay disables the IPv4 one
	opts = buildCrocOptions(Settings{
		RelayAddress:  models.DEFAULT_RELAY,
		RelayAddress6: "relay6.example.com:9009",
	}, false)
	if opts.RelayAddress != "" {
		t.Errorf("custom v6 relay should clear v4, got %q", opts.RelayAddress)
	}
}

func TestApplyProxySettings(t *testing.T) {
	oldSocks, oldHTTP := comm.Socks5Proxy, comm.HttpProxy
	defer func() { comm.Socks5Proxy, comm.HttpProxy = oldSocks, oldHTTP }()

	applyProxySettings(Settings{Socks5: "localhost:1080", HttpProxy: "http://localhost:8080"})
	if comm.Socks5Proxy != "localhost:1080" || comm.HttpProxy != "http://localhost:8080" {
		t.Fatalf("proxies not applied: %q / %q", comm.Socks5Proxy, comm.HttpProxy)
	}

	// clearing settings clears the globals too
	applyProxySettings(Settings{})
	if comm.Socks5Proxy != "" || comm.HttpProxy != "" {
		t.Fatalf("proxies not cleared: %q / %q", comm.Socks5Proxy, comm.HttpProxy)
	}
}

func TestBuildCrocOptionsTransferFlags(t *testing.T) {
	opts := buildCrocOptions(Settings{
		ZipFolder:      true,
		Exclude:        " *.log, node_modules ,,tmp",
		IP:             "10.0.0.1:9009",
		ThrottleUpload: "2m",
	}, true)
	if !opts.ZipFolder {
		t.Error("ZipFolder not mapped")
	}
	if len(opts.Exclude) != 3 || opts.Exclude[0] != "*.log" || opts.Exclude[2] != "tmp" {
		t.Errorf("Exclude not split/trimmed: %+v", opts.Exclude)
	}
	if opts.IP != "10.0.0.1:9009" {
		t.Errorf("IP not mapped: %q", opts.IP)
	}
	if opts.ThrottleUpload != "2m" {
		t.Errorf("ThrottleUpload not mapped: %q", opts.ThrottleUpload)
	}

	// invalid throttle values are dropped rather than panicking inside croc
	opts = buildCrocOptions(Settings{ThrottleUpload: "fast"}, true)
	if opts.ThrottleUpload != "" {
		t.Errorf("invalid throttle should be dropped, got %q", opts.ThrottleUpload)
	}
}

func TestValidThrottle(t *testing.T) {
	for _, ok := range []string{"", "500", "500k", "2M", "1g"} {
		if !validThrottle(ok) {
			t.Errorf("validThrottle(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"fast", "1.5m", "10 kb", "-5k", "k"} {
		if validThrottle(bad) {
			t.Errorf("validThrottle(%q) = true, want false", bad)
		}
	}
}
