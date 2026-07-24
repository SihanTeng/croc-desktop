package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

type donePayload struct {
	IsText bool   `json:"isText"`
	Text   string `json:"text,omitempty"`
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
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.acceptChan = make(chan bool, 1)
	t.overwriteChan = make(chan bool, 1)
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

// finish reports the end of a transfer to the frontend.
func (t *transferManager) finish(err error) {
	t.mu.Lock()
	accept := t.lastAccept
	dir := t.receiveDir
	cancelRequested := t.cancelRequested
	declined := t.declined
	isSender := t.isSender
	t.lastAccept = nil
	t.running = false
	t.cancelRequested = false
	t.mu.Unlock()

	switch {
	case cancelRequested || declined || errors.Is(err, context.Canceled):
		t.emitEvent(eventState, "cancelled")
		return
	case err != nil:
		msg := err.Error()
		if isSender && strings.Contains(msg, "refusing files") {
			msg = "The recipient declined the transfer"
		}
		t.emitEvent(eventError, msg)
		return
	}
	payload := donePayload{}
	// for received text, read it back so the UI can show it inline, then
	// remove the wrapper file (matching the CLI, which doesn't keep it)
	if accept != nil && accept.IsText && len(accept.Files) > 0 {
		textFile := filepath.Join(dir, accept.Files[0].Folder, accept.Files[0].Name)
		b, rerr := os.ReadFile(textFile)
		if rerr == nil {
			payload.IsText = true
			payload.Text = string(b)
			os.Remove(textFile)
		}
	}
	t.emitEvent(eventDone, payload)
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

func (t *transferManager) respondAccept(ok bool) {
	t.mu.Lock()
	if !ok {
		t.declined = true
	}
	ch := t.acceptChan
	t.mu.Unlock()
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
