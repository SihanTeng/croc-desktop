package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
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
	hm *historyManager

	settingsMu sync.Mutex
	settings   Settings
}

func NewApp() *App {
	tm := newTransferManager()
	hm := newHistoryManager(historyPath())
	tm.history = hm
	return &App{
		tm:       tm,
		rm:       newRelayManager(),
		hm:       hm,
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
	return a.startSend(paths, "", nil)
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
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", err
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	cleanup := func() { _ = os.Remove(f.Name()) }
	code, err := a.startSend([]string{f.Name()}, text, cleanup)
	if err != nil {
		cleanup()
	}
	return code, err
}

func (a *App) startSend(paths []string, sendText string, cleanup func()) (string, error) {
	ctx, err := a.tm.tryStart(true)
	if err != nil {
		return "", err
	}
	a.settingsMu.Lock()
	opts := buildCrocOptions(a.settings, true)
	applyProxySettings(a.settings)
	a.settingsMu.Unlock()
	opts.SharedSecret = utils.GetRandomName()
	opts.SendingText = sendText != ""

	filesInfo, emptyFolders, totalFolders, err := croc.GetFilesInfoWithExactExclusions(paths, false, false, nil, nil)
	if err != nil {
		a.tm.reset()
		return "", err
	}
	if len(filesInfo) == 0 && len(emptyFolders) == 0 {
		a.tm.reset()
		return "", fmt.Errorf("nothing to send")
	}
	var sendFiles []historyFile
	var sendTotalSize int64
	for _, f := range filesInfo {
		sendTotalSize += f.Size
		sendFiles = append(sendFiles, historyFile{
			Name: filepath.Join(f.FolderRemote, f.Name),
			Size: f.Size,
		})
	}
	a.tm.setSendInfo(sendFiles, len(filesInfo), sendTotalSize, sendText)
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
	applyProxySettings(a.settings)
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
		_ = os.Chdir(oldWd)
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

// maxPreviewBytes caps how much data GetFileDataURL loads into memory and
// pushes across the bridge; larger files are listed without a preview.
const maxPreviewBytes = 64 << 20 // 64 MiB

// GetFileDataURL reads a local file and returns it as a data: URL so the
// frontend can preview received media without a local file server.
func (a *App) GetFileDataURL(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if st.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}
	if st.Size() > maxPreviewBytes {
		return "", fmt.Errorf("file is too large to preview")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// extension-based types cover markup/media the sniffer can't identify
	// (e.g. SVG); fall back to sniffing the content otherwise
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = http.DetectContentType(b)
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

// --- history ---

// GetHistory returns the recorded transfer history, newest first.
func (a *App) GetHistory() []historyItem {
	return a.hm.list()
}

// ClearHistory removes all recorded history entries.
func (a *App) ClearHistory() {
	a.hm.clear()
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
