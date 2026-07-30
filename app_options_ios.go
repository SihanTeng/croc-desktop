//go:build ios

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// modifyOptionsForPlatform adjusts application options for iOS builds.
func modifyOptionsForPlatform(opts *application.Options) {
	// Signal handlers race with UIKit on iOS and can crash the process.
	opts.DisableDefaultSignalHandler = true
	opts.IOS.EnableInlineMediaPlayback = true
	opts.IOS.BackgroundColour = application.NewRGB(248, 250, 253)
}
