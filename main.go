package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "croc-desktop",
		Width:     1024,
		Height:    720,
		MinWidth:  820,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 253, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Linux: &linux.Options{
			Icon: appIcon,
			// the Wayland app_id is derived from the prgname; pinning it makes
			// the window match croc-desktop.desktop (and its icon) regardless of
			// the binary name, e.g. the wails-dev binary
			ProgramName: "croc-desktop",
		},
		Bind: []interface{}{
			app,
		},
		// route native file drops to the runtime callback; elements styled
		// with --wails-drop-target: drop opt in as drop targets
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
