package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/croc/v10/src/croc"
	"github.com/schollz/croc/v10/src/utils"
	"github.com/skip2/go-qrcode"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound backend. All exported methods are callable from the
// frontend.
type App struct {
	tm *transferManager
	rm *relayManager

	settingsMu sync.Mutex
	settings   Settings
}

func NewApp() *App {
	return &App{
		tm:       newTransferManager(),
		rm:       newRelayManager(),
		settings: loadSettings(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.tm.setWailsCtx(ctx)
	a.rm.setWailsCtx(ctx)
	wailsruntime.OnFileDrop(ctx, a.onFileDrop)
}

func (a *App) shutdown(ctx context.Context) {
	a.tm.cancelTransfer()
	a.rm.stop()
}

// onFileDrop receives native file drops (paths, not web File objects) and
// forwards them to the frontend.
func (a *App) onFileDrop(x, y int, paths []string) {
	if a.tm.wailsCtx == nil {
		return
	}
	wailsruntime.EventsEmit(a.tm.wailsCtx, "files:dropped", paths)
}

// --- dialogs ---

func (a *App) PickFiles() ([]string, error) {
	return wailsruntime.OpenMultipleFilesDialog(a.tm.wailsCtx, wailsruntime.OpenDialogOptions{
		Title: "Choose files to send",
	})
}

func (a *App) PickDirectory() (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.tm.wailsCtx, wailsruntime.OpenDialogOptions{
		Title: "Choose a folder",
	})
}

func (a *App) GetDefaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dl := filepath.Join(home, "Downloads")
	if st, err := os.Stat(dl); err == nil && st.IsDir() {
		return dl, nil
	}
	return home, nil
}

// --- transfers ---

// StartSend begins sending the given files/folders and returns the code
// phrase the recipient should enter.
func (a *App) StartSend(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", fmt.Errorf("no files selected")
	}
	return a.startSend(paths, false, nil)
}

// StartSendText sends a text snippet; it is wrapped in a temp file, matching
// the CLI's --text behavior.
func (a *App) StartSendText(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no text to send")
	}
	f, err := os.CreateTemp("", "croc-text-")
	if err != nil {
		return "", err
	}
	if _, err = f.WriteString(text); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err = f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	cleanup := func() { os.Remove(f.Name()) }
	code, err := a.startSend([]string{f.Name()}, true, cleanup)
	if err != nil {
		cleanup()
	}
	return code, err
}

func (a *App) startSend(paths []string, sendingText bool, cleanup func()) (string, error) {
	ctx, err := a.tm.tryStart(true)
	if err != nil {
		return "", err
	}
	a.settingsMu.Lock()
	opts := buildCrocOptions(a.settings, true)
	a.settingsMu.Unlock()
	opts.SharedSecret = utils.GetRandomName()
	opts.SendingText = sendingText

	filesInfo, emptyFolders, totalFolders, err := croc.GetFilesInfoWithExactExclusions(paths, false, false, nil, nil)
	if err != nil {
		a.tm.reset()
		return "", err
	}
	if len(filesInfo) == 0 && len(emptyFolders) == 0 {
		a.tm.reset()
		return "", fmt.Errorf("nothing to send")
	}
	cr, err := croc.NewCtx(ctx, opts)
	if err != nil {
		a.tm.reset()
		return "", err
	}
	cr.SetHooks(a.tm.hooks())
	a.tm.setClient(cr)
	go func() {
		err := cr.Send(filesInfo, emptyFolders, totalFolders)
		if cleanup != nil {
			cleanup()
		}
		a.tm.finish(err)
	}()
	return opts.SharedSecret, nil
}

// StartReceive connects with the given code phrase and receives into outDir.
// Progress and prompts arrive as events.
func (a *App) StartReceive(code string, outDir string) error {
	code = strings.TrimSpace(code)
	if len(code) < 6 {
		return fmt.Errorf("code is too short (must be at least 6 characters)")
	}
	if outDir == "" {
		var err error
		outDir, err = a.GetDefaultDownloadDir()
		if err != nil {
			return err
		}
	}
	absDir, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return err
	}
	ctx, err := a.tm.tryStart(false)
	if err != nil {
		return err
	}
	a.settingsMu.Lock()
	opts := buildCrocOptions(a.settings, false)
	a.settingsMu.Unlock()
	opts.SharedSecret = code

	cr, err := croc.NewCtx(ctx, opts)
	if err != nil {
		a.tm.reset()
		return err
	}
	cr.SetHooks(a.tm.hooks())
	a.tm.setClient(cr)
	a.tm.setReceiveDir(absDir)
	// croc writes received files relative to the working directory (same
	// mechanism as the CLI's --out)
	oldWd, _ := os.Getwd()
	if err := os.Chdir(absDir); err != nil {
		a.tm.reset()
		return err
	}
	go func() {
		err := cr.Receive()
		os.Chdir(oldWd)
		a.tm.finish(err)
	}()
	return nil
}

func (a *App) CancelTransfer() {
	a.tm.cancelTransfer()
}

func (a *App) RespondAccept(accept bool) {
	a.tm.respondAccept(accept)
}

func (a *App) RespondOverwrite(overwrite bool) {
	a.tm.respondOverwrite(overwrite)
}

// GetQrPng returns a base64-encoded PNG QR code for the given text.
func (a *App) GetQrPng(text string) (string, error) {
	png, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(png), nil
}

// --- settings ---

func (a *App) GetSettings() Settings {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	return a.settings
}

func (a *App) SaveSettings(s Settings) error {
	if err := saveSettings(s); err != nil {
		return err
	}
	a.settingsMu.Lock()
	a.settings = s
	a.settingsMu.Unlock()
	return nil
}

// --- relay ---

func (a *App) StartRelay(ports []string, password string) error {
	return a.rm.start(ports, password)
}

func (a *App) StopRelay() {
	a.rm.stop()
}

func (a *App) GetRelayState() relayState {
	return a.rm.state()
}
