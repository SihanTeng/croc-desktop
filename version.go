package main

import (
	"runtime/debug"
	"strings"
)

// AppVersion is the croc-desktop product version. The git tag is the single
// source of truth: release builds inject it at link time via
//   -ldflags "-X main.AppVersion=1.2.3"
// (wired into build/*/Taskfile.yml as {{.VERSION}}). The fallback only shows
// for plain `go build` / `go test` runs outside the Taskfiles.
var AppVersion = "0.0.0-dev"

// AppInfo is shown on the Settings page (version + engine + defaults).
type AppInfo struct {
	Version       string `json:"version"`
	CrocVersion   string `json:"crocVersion"`
	DefaultRelay  string `json:"defaultRelay"`
	DefaultRelay6 string `json:"defaultRelay6"`
}

// GetAppInfo returns product/engine version and the stock public relays.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Version:       AppVersion,
		CrocVersion:   crocModuleVersion(),
		DefaultRelay:  defaultSettings().RelayAddress,
		DefaultRelay6: defaultSettings().RelayAddress6,
	}
}

func crocModuleVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/schollz/croc/v10" {
			v := d.Version
			if d.Replace != nil && d.Replace.Version != "" {
				// replaced modules often use pseudo-versions; prefer replace path tag when short
				v = d.Replace.Version
			}
			return strings.TrimSpace(v)
		}
	}
	// When running `go test` / some builds the main module path may list croc
	// only via replace; fall back to the go.mod requirement if present.
	for _, d := range bi.Deps {
		if strings.Contains(d.Path, "croc") && strings.Contains(d.Path, "v10") {
			return d.Version
		}
	}
	return "unknown"
}
