package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/croc/v10/src/croc"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// events emitted to the frontend
const (
	eventState     = "transfer:state"     // string: connecting | waiting | transferring | cancelled
	eventProgress  = "transfer:progress"  // croc.ProgressEvent
	eventAccept    = "transfer:accept"    // acceptPayload
	eventOverwrite = "transfer:overwrite" // overwritePayload
	eventDone      = "transfer:done"      // donePayload
	eventError     = "transfer:error"     // string
)

type fileEntry struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
	Size   int64  `json:"size"`
}

type acceptPayload struct {
	Files        []fileEntry `json:"files"`
	TotalSize    int64       `json:"totalSize"`
	TotalFolders int         `json:"totalFolders"`
	SenderID     string      `json:"senderId"`
	IsText       bool        `json:"isText"`
}

type overwritePayload struct {
	Path      string  `json:"path"`
	ResumePct float64 `json:"resumePct"`
}

// receivedFile is a file written to disk by a completed receive; Path is
// absolute, Name is relative to the receive directory.
type receivedFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type donePayload struct {
	IsText bool           `json:"isText"`
	Text   string         `json:"text,omitempty"`
	Files  []receivedFile `json:"files,omitempty"`
}

// transferManager runs at most one croc transfer at a time and bridges the
// croc.Hooks callbacks to Wails frontend events.
type transferManager struct {
	wailsCtx context.Context
	// emit publishes a frontend event; set by setWailsCtx, swappable in tests
	emit func(event string, data interface{})

	mu              sync.Mutex
	running         bool
	cancelRequested bool
	// declined is set when the local user declines an incoming transfer, so
	// finish() reports a cancellation instead of a scary "refused" error
	declined bool
	isSender bool

	ctx    context.Context
	cancel context.CancelFunc
	// client is the in-flight croc client; its connections are force-closed
	// on cancel so blocked network reads unwind immediately
	client *croc.Client

	acceptChan    chan bool
	overwriteChan chan bool

	lastAccept *acceptPayload
	receiveDir string

	// history records every transfer outcome; send-side metadata below is
	// captured at start since the accept payload only exists on the receiver
	history        *historyManager
	logger         *logManager
	sendFiles      []historyFile
	sendTotalFiles int
	sendTotalSize  int64
	sendText       string
}

func newTransferManager() *transferManager {
	return &transferManager{}
}

// setWailsCtx wires the manager to the Wails runtime.
func (t *transferManager) setWailsCtx(ctx context.Context) {
	t.wailsCtx = ctx
	t.emit = func(event string, data interface{}) {
		wailsruntime.EventsEmit(ctx, event, data)
	}
}

func (t *transferManager) emitEvent(event string, data interface{}) {
	if t.emit != nil {
		t.emit(event, data)
	}
}

// tryStart marks a transfer as running and returns its cancellable context.
func (t *transferManager) tryStart(isSender bool) (context.Context, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		return nil, fmt.Errorf("a transfer is already in progress")
	}
	t.running = true
	t.cancelRequested = false
	t.declined = false
	t.isSender = isSender
	t.lastAccept = nil
	t.client = nil
	t.sendFiles = nil
	t.sendTotalFiles = 0
	t.sendTotalSize = 0
	t.sendText = ""
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.acceptChan = make(chan bool, 1)
	t.overwriteChan = make(chan bool, 1)
	if t.logger != nil {
		dir := "receive"
		if isSender {
			dir = "send"
		}
		t.logger.log(levelDebug, "transfer", "transfer started (%s)", dir)
	}
	return t.ctx, nil
}

// reset clears the running flag without emitting anything; used when a
// transfer fails before it starts.
func (t *transferManager) reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = false
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

// finish reports the end of a transfer to the frontend and records it in the
// transfer history.
func (t *transferManager) finish(err error) {
	t.mu.Lock()
	accept := t.lastAccept
	dir := t.receiveDir
	cancelRequested := t.cancelRequested
	declined := t.declined
	isSender := t.isSender
	sendFiles := t.sendFiles
	sendTotalFiles := t.sendTotalFiles
	sendTotalSize := t.sendTotalSize
	sendText := t.sendText
	history := t.history
	t.lastAccept = nil
	t.sendFiles = nil
	t.sendTotalFiles = 0
	t.sendTotalSize = 0
	t.sendText = ""
	t.running = false
	t.cancelRequested = false
	t.mu.Unlock()

	status := "completed"
	errMsg := ""
	var payload donePayload
	switch {
	case cancelRequested || declined || errors.Is(err, context.Canceled):
		status = "cancelled"
	case err != nil:
		status = "error"
		errMsg = err.Error()
		if isSender && strings.Contains(errMsg, "refusing files") {
			errMsg = "The recipient declined the transfer"
		} else {
			errMsg = friendlyTransferError(errMsg)
		}
	default:
		// for received text, read it back so the UI can show it inline, then
		// remove the wrapper file (matching the CLI, which doesn't keep it)
		if accept != nil && accept.IsText && len(accept.Files) > 0 {
			textFile := filepath.Join(dir, accept.Files[0].Folder, accept.Files[0].Name)
			b, rerr := os.ReadFile(textFile)
			if rerr == nil {
				payload.IsText = true
				payload.Text = string(b)
				_ = os.Remove(textFile)
			}
		}
		// include the received files so the UI can display them; text transfers
		// show the message inline instead
		if !payload.IsText && accept != nil {
			for _, f := range accept.Files {
				payload.Files = append(payload.Files, receivedFile{
					Name: filepath.Join(f.Folder, f.Name),
					Path: filepath.Join(dir, f.Folder, f.Name),
					Size: f.Size,
				})
			}
		}
	}

	if history != nil {
		history.add(buildHistoryItem(isSender, status, errMsg, accept, payload, dir,
			sendFiles, sendTotalFiles, sendTotalSize, sendText))
	}

	switch status {
	case "cancelled":
		if t.logger != nil {
			t.logger.log(levelWarn, "transfer", "transfer cancelled")
		}
		t.emitEvent(eventState, "cancelled")
	case "error":
		if t.logger != nil {
			t.logger.log(levelError, "transfer", "transfer failed: %s", errMsg)
		}
		notify("croc-desktop", errMsg)
		t.emitEvent(eventError, errMsg)
	default:
		if t.logger != nil {
			t.logger.log(levelInfo, "transfer", "transfer complete")
		}
		notify("croc-desktop", "Transfer complete")
		t.emitEvent(eventDone, payload)
	}
}

// friendlyTransferError rewrites known croc error strings into actionable
// messages; unknown errors pass through unchanged.
func friendlyTransferError(msg string) string {
	switch {
	case strings.Contains(msg, "password mismatch"):
		return "The relay rejected the connection (password mismatch) — check the relay password in Settings"
	case strings.Contains(msg, "could not secure channel"):
		return "Couldn't establish a secure channel — double-check the code and try again"
	case strings.Contains(msg, "found no addresses to connect"),
		strings.Contains(msg, "could not reconnect to any relay"):
		return "Couldn't reach the relay — check the relay address in Settings and your network connection"
	case strings.Contains(msg, "room full"):
		return "That code is already in use on the relay — wait a moment or pick a different code"
	case strings.Contains(msg, "not enough disk space"):
		return "Not enough disk space to receive the files"
	default:
		return msg
	}
}

// buildHistoryItem assembles the history entry for a finished transfer.
func buildHistoryItem(isSender bool, status, errMsg string, accept *acceptPayload,
	payload donePayload, dir string, sendFiles []historyFile, sendTotalFiles int,
	sendTotalSize int64, sendText string) historyItem {
	item := historyItem{
		Time:   time.Now(),
		Status: status,
		Error:  errMsg,
	}
	if isSender {
		item.Direction = "send"
		if sendText != "" {
			item.IsText = true
			item.Text = truncateText(sendText, maxHistoryText)
		} else {
			item.TotalFiles = sendTotalFiles
			item.TotalSize = sendTotalSize
			for i, f := range sendFiles {
				if i >= maxHistoryFiles {
					break
				}
				item.Files = append(item.Files, f)
			}
		}
		return item
	}
	item.Direction = "receive"
	item.Dir = dir
	if payload.IsText {
		item.IsText = true
		item.Text = truncateText(payload.Text, maxHistoryText)
		return item
	}
	if accept != nil {
		item.IsText = accept.IsText
		item.TotalFiles = len(accept.Files)
		item.TotalSize = accept.TotalSize
		for i, f := range accept.Files {
			if i >= maxHistoryFiles {
				break
			}
			item.Files = append(item.Files, historyFile{
				Name: filepath.Join(f.Folder, f.Name),
				Size: f.Size,
			})
		}
	}
	return item
}

func (t *transferManager) cancelTransfer() {
	t.mu.Lock()
	t.cancelRequested = true
	cancel := t.cancel
	client := t.client
	t.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	// force-close connections so goroutines parked in network reads (e.g.
	// waiting on a relay that doesn't ping) unwind immediately
	if client != nil {
		client.CloseConnections()
	}
}

// setClient registers the in-flight croc client for cancel teardown.
func (t *transferManager) setClient(cr *croc.Client) {
	t.mu.Lock()
	t.client = cr
	t.mu.Unlock()
}

// setSendInfo captures send-side metadata for the history entry; the accept
// payload that describes the files only exists on the receiver.
func (t *transferManager) setSendInfo(files []historyFile, totalFiles int, totalSize int64, text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sendFiles = files
	t.sendTotalFiles = totalFiles
	t.sendTotalSize = totalSize
	t.sendText = text
}

func (t *transferManager) respondAccept(ok bool) {
	t.mu.Lock()
	if !ok {
		t.declined = true
	}
	ch := t.acceptChan
	t.mu.Unlock()
	if t.logger != nil {
		if ok {
			t.logger.log(levelInfo, "transfer", "transfer accepted")
		} else {
			t.logger.log(levelWarn, "transfer", "transfer declined")
		}
	}
	if ch == nil {
		return
	}
	select {
	case ch <- ok:
	default:
	}
}

func (t *transferManager) respondOverwrite(ok bool) {
	t.mu.Lock()
	ch := t.overwriteChan
	t.mu.Unlock()
	if t.logger != nil {
		if ok {
			t.logger.log(levelInfo, "transfer", "overwrite accepted")
		} else {
			t.logger.log(levelInfo, "transfer", "overwrite declined (skip)")
		}
	}
	if ch == nil {
		return
	}
	select {
	case ch <- ok:
	default:
	}
}

func (t *transferManager) setReceiveDir(dir string) {
	t.mu.Lock()
	t.receiveDir = dir
	t.mu.Unlock()
}

// hooks builds the croc.Hooks that divert progress and prompts to the
// frontend.
func (t *transferManager) hooks() *croc.Hooks {
	return &croc.Hooks{
		OnProgress: func(p croc.ProgressEvent) {
			t.emitEvent(eventProgress, p)
		},
		OnStateChange: func(state string) {
			if t.logger != nil {
				t.logger.log(levelDebug, "transfer", "state → %s", state)
			}
			t.emitEvent(eventState, state)
		},
		OnAcceptRequest: func(req croc.AcceptRequest) bool {
			payload := acceptPayload{
				TotalSize:    req.TotalSize,
				TotalFolders: req.TotalFolders,
				SenderID:     req.SenderID,
				IsText:       req.IsText,
			}
			for _, f := range req.Files {
				payload.Files = append(payload.Files, fileEntry{
					Name:   f.Name,
					Folder: f.FolderRemote,
					Size:   f.Size,
				})
			}
			t.mu.Lock()
			t.lastAccept = &payload
			ch := t.acceptChan
			ctx := t.ctx
			t.mu.Unlock()
			if t.logger != nil {
				what := fmt.Sprintf("%d file(s), %s", len(payload.Files), humanBytes(payload.TotalSize))
				if payload.IsText {
					what = "a text message"
				}
				t.logger.log(levelInfo, "transfer", "incoming transfer: %s", what)
			}
			t.emitEvent(eventAccept, payload)
			select {
			case ok := <-ch:
				return ok
			case <-ctx.Done():
				return false
			}
		},
		OnOverwriteRequest: func(path string, resumePct float64) bool {
			t.mu.Lock()
			ch := t.overwriteChan
			ctx := t.ctx
			t.mu.Unlock()
			if t.logger != nil {
				t.logger.log(levelInfo, "transfer", "overwrite prompt for %s (%.0f%% already present)", path, resumePct)
			}
			t.emitEvent(eventOverwrite, overwritePayload{
				Path:      path,
				ResumePct: resumePct,
			})
			select {
			case ok := <-ch:
				return ok
			case <-ctx.Done():
				return false
			}
		},
		// the GUI user explicitly initiated the send, so no extra confirmation
		OnConfirmSendToPeer: func(machineID string) bool { return true },
	}
}

// humanBytes formats a byte count compactly for log messages.
func humanBytes(n int64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
