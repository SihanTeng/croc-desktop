//go:build !ios && !android

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// modifyOptionsForPlatform is a no-op on desktop platforms.
func modifyOptionsForPlatform(_ *application.Options) {}
