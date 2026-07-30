package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	svc := NewApp()

	app := application.New(application.Options{
		Name:        "croc-desktop",
		Description: "Desktop GUI for croc — encrypted peer-to-peer file transfer",
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		// the Wayland app_id is derived from the prgname; pinning it makes
		// the window match croc-desktop.desktop (and its icon) regardless of
		// the binary name
		Linux: application.LinuxOptions{
			ProgramName: "croc-desktop",
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "croc-desktop",
		Width:            1024,
		Height:           720,
		MinWidth:         820,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(248, 250, 253),
		// native file drops: elements with data-file-drop-target opt in as
		// drop targets (see frontend DropZone).
		EnableFileDrop: true,
		URL:            "/",
	})

	// Bridge native file drops to the frontend event the Send view listens for.
	win.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		paths := event.Context().DroppedFiles()
		if len(paths) == 0 {
			return
		}
		app.Event.Emit("files:dropped", paths)
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
