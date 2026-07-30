//go:build android

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// modifyOptionsForPlatform adjusts application options for Android builds.
func modifyOptionsForPlatform(opts *application.Options) {
	opts.DisableDefaultSignalHandler = true
}
