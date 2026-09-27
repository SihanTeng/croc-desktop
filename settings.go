package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/schollz/croc/v11/src/comm"
	"github.com/schollz/croc/v11/src/croc"
	"github.com/schollz/croc/v11/src/models"
	"github.com/schollz/croc/v11/src/utils"
)

// SavedCode is a remembered receive code (favorite), managed from the
// Receive view.
type SavedCode struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// Settings holds the user-configurable options shown in the Settings view.
// It maps onto croc.Options; see buildCrocOptions.
type Settings struct {
	RelayAddress   string      `json:"relayAddress"`
	RelayAddress6  string      `json:"relayAddress6"`
	RelayPassword  string      `json:"relayPassword"`
	Curve          string      `json:"curve"`
	HashAlgorithm  string      `json:"hashAlgorithm"`
	OnlyLocal      bool        `json:"onlyLocal"`
	DisableLocal   bool        `json:"disableLocal"`
	NoCompress     bool        `json:"noCompress"`
	Overwrite      bool        `json:"overwrite"`
	DownloadDir    string      `json:"downloadDir"`
	Socks5         string      `json:"socks5"`
	HttpProxy      string      `json:"httpProxy"`
	ZipFolder      bool        `json:"zipFolder"`
	Exclude        string      `json:"exclude"`
	ThrottleUpload string      `json:"throttleUpload"`
	IP             string      `json:"ip"`
	Theme          string      `json:"theme"`    // system | light | dark
	Language       string      `json:"language"` // system | <locale>, e.g. en, zh-CN
	SavedCodes     []SavedCode `json:"savedCodes,omitempty"`
}

var settingsMu sync.Mutex

func defaultSettings() Settings {
	return Settings{
		RelayAddress:  models.DEFAULT_RELAY,
		RelayAddress6: models.DEFAULT_RELAY6,
		RelayPassword: models.DEFAULT_PASSPHRASE,
		Curve:         "p256",
		HashAlgorithm: "xxhash",
	}
}

func settingsFile() (string, error) {
	dir, err := utils.GetConfigDir(false)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "croc-desktop.json"), nil
}

func loadSettings() Settings {
	s := defaultSettings()
	f, err := settingsFile()
	if err != nil {
		return s
	}
	b, err := os.ReadFile(f)
	fromLegacy := false
	if err != nil {
		// fall back to the pre-rename settings file (croc-gui → croc-desktop)
		legacy := filepath.Join(filepath.Dir(f), "croc-gui.json")
		if b, err = os.ReadFile(legacy); err != nil {
			return s
		}
		fromLegacy = true
	}
	// keep defaults for anything missing from the file
	_ = json.Unmarshal(b, &s)
	// one-shot migrate: persist under the current name so upgrades keep
	// working without relying on the legacy path forever
	if fromLegacy {
		_ = saveSettings(s)
	}
	return s
}

func saveSettings(s Settings) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	// make sure the config directory exists
	if _, err := utils.GetConfigDir(true); err != nil {
		return err
	}
	f, err := settingsFile()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f, b, 0o644)
}

// withProxySettings freezes routing for this transfer, including lingering
// cancelled dialers. Never mutate croc's CLI-oriented process-wide globals.
func withProxySettings(ctx context.Context, s Settings) context.Context {
	return comm.WithProxy(ctx, s.Socks5, s.HttpProxy)
}

// throttleRe matches croc's upload-limit syntax: a number with an optional
// k/m/g (kilobytes..gigabytes per second) suffix. croc panics on anything
// else, so invalid values are dropped here instead.
var throttleRe = regexp.MustCompile(`^\d+[kKmMgG]?$`)

func validThrottle(s string) bool {
	return s == "" || throttleRe.MatchString(strings.TrimSpace(s))
}

// splitExclude turns the comma-separated exclude setting into patterns.
func splitExclude(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// buildCrocOptions maps Settings onto croc.Options, mirroring the CLI wiring
// in src/cli/cli.go. Relay ports follow the CLI default (base 9009, 4 transfer
// ports).
func buildCrocOptions(s Settings, isSender bool) croc.Options {
	opts := croc.Options{
		IsSender:         isSender,
		RelayAddress:     s.RelayAddress,
		RelayAddress6:    s.RelayAddress6,
		RelayPorts:       []string{"9009", "9010", "9011", "9012"},
		RelayPassword:    s.RelayPassword,
		Curve:            s.Curve,
		HashAlgorithm:    s.HashAlgorithm,
		OnlyLocal:        s.OnlyLocal,
		DisableLocal:     s.DisableLocal,
		NoCompress:       s.NoCompress,
		Overwrite:        s.Overwrite,
		ZipFolder:        s.ZipFolder,
		Exclude:          splitExclude(s.Exclude),
		IP:               s.IP,
		NoPrompt:         true,
		DisableClipboard: true,
		// Preserve desktop relay/proxy routing instead of starting Tailcat/DERP.
		DisableTailcat: true,
	}
	if validThrottle(s.ThrottleUpload) {
		opts.ThrottleUpload = s.ThrottleUpload
	}
	if opts.RelayPassword == "" {
		opts.RelayPassword = models.DEFAULT_PASSPHRASE
	}
	if opts.Curve == "" {
		opts.Curve = "p256"
	}
	if opts.HashAlgorithm == "" {
		opts.HashAlgorithm = "xxhash"
	}
	// same relay selection logic as the CLI: a custom relay disables the
	// counterpart address of the other IP family
	if opts.RelayAddress != models.DEFAULT_RELAY {
		opts.RelayAddress6 = ""
	} else if opts.RelayAddress6 != models.DEFAULT_RELAY6 {
		opts.RelayAddress = ""
	}
	return opts
}
