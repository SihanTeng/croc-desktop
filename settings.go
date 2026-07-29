package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/croc/v10/src/comm"
	"github.com/schollz/croc/v10/src/croc"
	"github.com/schollz/croc/v10/src/models"
	"github.com/schollz/croc/v10/src/utils"
)

// Settings holds the user-configurable options shown in the Settings view.
// It maps onto croc.Options; see buildCrocOptions.
type Settings struct {
	RelayAddress  string `json:"relayAddress"`
	RelayAddress6 string `json:"relayAddress6"`
	RelayPassword string `json:"relayPassword"`
	Curve         string `json:"curve"`
	HashAlgorithm string `json:"hashAlgorithm"`
	OnlyLocal     bool   `json:"onlyLocal"`
	DisableLocal  bool   `json:"disableLocal"`
	NoCompress    bool   `json:"noCompress"`
	Overwrite     bool   `json:"overwrite"`
	DownloadDir   string `json:"downloadDir"`
	Socks5        string `json:"socks5"`
	HttpProxy     string `json:"httpProxy"`
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
	return filepath.Join(dir, "croc-gui.json"), nil
}

func loadSettings() Settings {
	s := defaultSettings()
	f, err := settingsFile()
	if err != nil {
		return s
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return s
	}
	// keep defaults for anything missing from the file
	_ = json.Unmarshal(b, &s)
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

// applyProxySettings wires the proxy settings into croc's comm package, which
// (like the CLI) reads them from package-level variables.
func applyProxySettings(s Settings) {
	comm.Socks5Proxy = s.Socks5
	comm.HttpProxy = s.HttpProxy
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
		NoPrompt:         true,
		DisableClipboard: true,
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
