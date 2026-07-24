# croc GUI

A cross-platform desktop GUI for [croc](https://github.com/schollz/croc), built
with [Wails v2](https://wails.io) (Go backend) and React + TypeScript
(frontend). It uses croc's Go packages in-process — no CLI subprocess.

Features:

- Send files/folders (picker or drag & drop) with code phrase + QR code
- Send text snippets (shown inline on the receiving end)
- Receive files with accept/decline and overwrite/resume dialogs, live progress
- Receive by pasting or uploading a screenshot of the sender's QR code
  (decoded in the background — the image is never displayed)
- Run a croc relay from the app
- Settings: custom relays, relay password, curve/hash, local-only, etc.

## Prerequisites

- Go (matching the repo's `go.mod`)
- Node.js + npm
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Platform webview deps — check with `wails doctor`. On Linux you need GTK3 +
  webkit2gtk dev packages (e.g. `webkit2gtk4.1-devel` on Fedora,
  `libwebkit2gtk-4.1-dev` on Debian/Ubuntu).

## Develop

```sh
cd croc-gui
wails dev        # on distros with webkit2gtk 4.1: wails dev -tags webkit2_41
```

## Build

```sh
cd croc-gui
wails build                 # add -tags webkit2_41 where applicable
# binary lands in gui/build/bin/
```

Cross-platform builds run on the target OS (or CI): `wails build -platform
windows/amd64,darwin/universal,linux/amd64` from a machine/CI with the
respective toolchains.

## Test

The backend has headless integration tests that drive the real transfer path
(`App` methods → `src/croc` → in-process relay) without a window:

```sh
cd croc-gui
npm --prefix frontend install && npm --prefix frontend run build  # provides frontend/dist for go:embed
go test .
```

## How it works

- `src/croc/hooks.go` (in the croc module) defines `croc.Hooks` —
  progress/state/prompt callbacks installed via `Client.SetHooks`. When hooks
  are set, the terminal progress bar is silenced and stdin prompts
  (accept/overwrite/ask) are routed through the hooks. CLI behavior is
  unchanged when hooks are nil.
- `transfer.go` bridges hooks to Wails events consumed by the React UI.
- Transfers run with `croc.NewCtx`, so Cancel is a context cancellation.
- `relay.go` wraps `tcp.RunCtx` for in-app relays.
- Settings persist to `<croc config dir>/croc-gui.json`.

This directory is a separate Go module (`replace github.com/schollz/croc/v10
=> ../`) so the CLI's dependency set stays untouched.
