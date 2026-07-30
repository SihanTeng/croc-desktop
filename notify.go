package main

import (
	"github.com/gen2brain/beeep"
)

// notify posts a desktop notification; failures are ignored (headless
// sessions, missing notification daemons) — notifications are best-effort.
func notify(title, message string) {
	_ = beeep.Notify(title, message, "")
}
