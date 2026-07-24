package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/schollz/croc/v10/src/models"
	"github.com/schollz/croc/v10/src/tcp"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const eventRelayState = "relay:state"

type relayState struct {
	Running bool     `json:"running"`
	Ports   []string `json:"ports,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// relayManager runs a croc relay (one control port plus transfer ports,
// mirroring `croc relay`) with cancellable contexts.
type relayManager struct {
	wailsCtx context.Context
	// emit publishes a frontend event; set by setWailsCtx, swappable in tests
	emit func(event string, data interface{})

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	ports   []string
}

func newRelayManager() *relayManager {
	return &relayManager{}
}

// setWailsCtx wires the manager to the Wails runtime.
func (r *relayManager) setWailsCtx(ctx context.Context) {
	r.wailsCtx = ctx
	r.emit = func(event string, data interface{}) {
		wailsruntime.EventsEmit(ctx, event, data)
	}
}

func (r *relayManager) emitEvent(event string, data interface{}) {
	if r.emit != nil {
		r.emit(event, data)
	}
}

func (r *relayManager) start(ports []string, password string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		return fmt.Errorf("relay is already running")
	}
	var clean []string
	for _, p := range ports {
		if p = strings.TrimSpace(p); p != "" {
			clean = append(clean, p)
		}
	}
	if len(clean) < 2 {
		return fmt.Errorf("relay requires at least two ports (control + transfer)")
	}
	if password == "" {
		password = models.DEFAULT_PASSPHRASE
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.ports = clean

	banner := strings.Join(clean[1:], ",")
	errCh := make(chan error, len(clean))
	for i, port := range clean {
		var banners []string
		if i == 0 {
			banners = append(banners, banner)
		}
		go func(portStr string, b ...string) {
			errCh <- tcp.Ignore(tcp.RunCtx(ctx, "error", "", portStr, password, b...))
		}(port, banners...)
	}
	r.running = true
	r.emitEvent(eventRelayState, relayState{Running: true, Ports: clean})

	// first server to exit shuts the whole relay down and reports why
	go func() {
		err := <-errCh
		r.mu.Lock()
		if r.cancel != nil {
			r.cancel()
		}
		r.running = false
		r.cancel = nil
		ports := r.ports
		r.mu.Unlock()
		state := relayState{Running: false, Ports: ports}
		if err != nil {
			state.Error = err.Error()
		}
		r.emitEvent(eventRelayState, state)
	}()
	return nil
}

func (r *relayManager) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.running = false
}

func (r *relayManager) state() relayState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return relayState{Running: r.running, Ports: r.ports}
}
