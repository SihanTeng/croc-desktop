<p align="center">
  <img src="build/appicon.png" alt="croc logo" width="96" height="96" />
</p>

<h1 align="center">croc-desktop</h1>

<p align="center">
  <strong>Desktop GUI for <a href="https://github.com/schollz/croc">croc</a></strong> — encrypted,
  peer-to-peer file and text transfer on <strong>Linux, macOS, and Windows</strong>.
  No cloud upload. No terminal required.
</p>

<p align="center">
  <a href="https://github.com/SihanTeng/croc-desktop/actions/workflows/ci.yml"><img src="https://github.com/SihanTeng/croc-desktop/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="https://github.com/SihanTeng/croc-desktop/releases"><img src="https://img.shields.io/github/v/release/SihanTeng/croc-desktop?style=flat-square" alt="Release" /></a>
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-1F6FEB?style=flat-square" alt="Platforms" />
  <img src="https://img.shields.io/badge/built%20with-Wails%20v3-DF3A2C?style=flat-square" alt="Wails v3" />
</p>

<p align="center">
  <img src="docs/images/send.png" alt="Send view with transfer code and QR" width="49%" />
  <img src="docs/images/receive.png" alt="Receive view with inline file preview" width="49%" />
</p>

Unlike wrappers that spawn the croc CLI, croc-desktop links croc's Go packages
**in-process** (via a fork adding a hooks layer). That gives it things a
subprocess can't have: real accept/overwrite dialogs, structured progress
events, relay hosting, and accurate transfer state.

## Features

**Send & receive**

- Send files and folders by picker or drag & drop; send text snippets
- Transfer code phrase with QR code, copy buttons for the code and the CLI command
- Receive by pasting anything: a bare code, `croc <code>`, `CROC_SECRET=… croc`, or a share link
- Receive by pasting or uploading a screenshot of the sender's QR code
- Save favorite codes and re-receive with one click
- Accept/decline and overwrite/resume dialogs — never blind `--yes --overwrite`
- Live progress with file counts, "verifying" state, and stall hints
- Cancel any time (button or `Esc`), resume interrupted receives
- After a receive, previews of images, video, audio, and text inline
- "Send same files/text again" for repeat transfers
- Desktop notifications when a transfer completes or fails

**History & logs**

- Every transfer (completed, cancelled, failed) is recorded with files, sizes,
  and destination — searchable in the History tab, persisted across restarts
- Centralized leveled log (debug / info / warn / error) streamed live to the
  Logs tab with level filtering

**Power options**

- Run a croc relay from the app
- Custom relays, relay password, encryption curve, hash algorithm
- Local-only / disable-local modes, compression and overwrite defaults
- SOCKS5 and HTTP proxy support
- Zip folders before sending, exclude patterns, upload rate limit, manual
  sender address
- Dark theme (system-following or pinned)
- Multi-language UI (English, 简体中文, 繁體中文, Español, Français, Deutsch,
  日本語) — see *Contributing translations*

<p align="center">
  <img src="docs/images/history.png" alt="History tab" width="49%" />
  <img src="docs/images/logs.png" alt="Logs tab with level filter" width="49%" />
</p>

## Install

Download from [Releases](https://github.com/SihanTeng/croc-desktop/releases):

| Platform | File | Notes |
| --- | --- | --- |
| Linux | `croc-desktop_*_linux-amd64.AppImage` | `chmod +x`, run |
| macOS (Apple Silicon) | `croc-desktop_*_darwin-arm64.dmg` | ad-hoc signed — Gatekeeper warns on first launch until the app is notarized |
| Windows | `croc-desktop_*_windows-amd64.msi` | WiX installer |

## Develop

Prerequisites: Go (see `go.mod`), Node.js + npm, the Wails v3 CLI, and
platform webview deps (`wails3 doctor`).

```sh
# Install the CLI (use -tags gtk3 on Linux if you have webkit2gtk 4.1 but not gtk4)
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@latest

# Linux: GTK3 + webkit2gtk 4.1 (this project’s default), e.g.
#   webkit2gtk4.1-devel on Fedora, libwebkit2gtk-4.1-dev on Debian/Ubuntu
# Optional native gtk4 path: install gtk4 + webkitgtk 6 and build with EXTRA_TAGS=

WEBKIT_DISABLE_DMABUF_RENDERER=1 wails3 dev   # or: task dev
```

(`WEBKIT_DISABLE_DMABUF_RENDERER=1` works around a WebKitGTK crash on some
Wayland compositors.)

Dev mode loads the Vite server (default port **34115** for this project) and
rebuilds Go on change. Frontend bindings live under `frontend/bindings/` and
are regenerated during `wails3 build` / `wails3 generate bindings`.

## Build

```sh
wails3 build -tags gtk3   # Linux with webkit2gtk 4.1 (default for this repo)
# or: task build
# binary lands in bin/croc-desktop
```

On macOS/Windows, omit the gtk3 tag (`wails3 build`). For Linux with gtk4 +
webkitgtk 6 installed: `EXTRA_TAGS= wails3 build`.

## Mobile (iOS / Android)

Wails v3 builds the same Go + React app for mobile. The UI is responsive:
desktop keeps the left side-rail; below ~720px (phones and the mobile
webview) it switches to a bottom tab bar with safe-area padding.

Scaffold lives under `build/ios/` and `build/android/`. You need the platform
SDKs (`wails3 doctor` reports what it can see).

```sh
# iOS Simulator (requires full Xcode on macOS)
wails3 task ios:run
# or: task ios:package IOS_PLATFORM=device CODESIGN_IDENTITY="Apple Development: …"

# Android emulator / device (requires Android SDK + NDK)
wails3 task android:run
```

Bundle / application IDs default to `com.schollz.croc-desktop` (see
`build/config.yml` and the platform Taskfiles). Resize the desktop window
down to ~360px wide to preview the mobile chrome without a device.

## Test

Three layers:

- **Backend** (`go test -tags gtk3 .`): unit tests for settings/history/logger/code
  parsing plus headless integration tests that drive the real transfer path
  (`App` methods → `src/croc` → in-process relay) without a window:

  ```sh
  npm --prefix frontend install && npm --prefix frontend run build  # frontend/dist for go:embed
  go test -tags gtk3 .
  ```

- **Frontend** (`npm --prefix frontend run test`): vitest unit tests for the
  pure helpers (byte formatting, preview-kind mapping, data-URL decoding).

- **Browser E2E** (`./e2e/run.sh`): playwright drives the real app UI in a
  browser against throwaway croc peers — send text, receive with preview,
  decline, Esc cancel, history & logs. Uses the running `wails3 dev` instance
  if there is one, otherwise boots a hermetic sandbox (local relay + isolated
  config). See `e2e/README.md`.

Frontend checks: `npm --prefix frontend run typecheck && npm --prefix frontend run lint`.

## Contributing translations

UI strings live in plain JSON files under `frontend/src/i18n/locales/`
(`en.json` is the canonical key set). To add or improve a language, copy
`en.json` to `<locale>.json` (e.g. `pt-BR.json`) and translate the values —
Vite picks the file up automatically, no code changes needed. Keep the keys
and the `{placeholder}` variables intact.

## How it works

- `src/croc/hooks.go` (in the croc module) defines `croc.Hooks` —
  progress/state/prompt callbacks installed via `Client.SetHooks`. With hooks
  set, the terminal progress bar is silenced and stdin prompts
  (accept/overwrite/ask) route through the hooks; CLI behavior is unchanged
  when hooks are nil.
- `transfer.go` bridges hooks to Wails events consumed by the React UI.
  Transfers run with `croc.NewCtx`, so Cancel is a context cancellation plus
  connection teardown.
- `history.go` persists transfer history and `logger.go` the centralized
  leveled log (both in-process; history on disk next to settings).
- `relay.go` wraps `tcp.RunCtx` for in-app relays.
- Settings persist to `<croc config dir>/croc-desktop.json`.

This is a separate Go module so the CLI's dependency set stays untouched.
croc is consumed as a plain module dependency — no sibling checkout needed:

```
replace github.com/schollz/croc/v10 => github.com/SihanTeng/croc/v10 v10.0.0-...-51660d6d7730
```

The replace points at the `gui-hooks` branch of the `SihanTeng/croc` fork,
which carries the hooks layer (`src/croc/hooks.go`) this app needs. Once the
hooks land upstream in `schollz/croc`, the `replace` line can be deleted.

> [!NOTE]
> When the hooks change, push the croc repo's `main` to the fork branch and
> refresh the pinned pseudo-version here:
>
> ```sh
> cd croc && git push origin main:gui-hooks
> cd ../croc-desktop && go get github.com/SihanTeng/croc/v10@gui-hooks && go mod tidy
> ```

## CI & releases

- **CI** (`ci.yml`, every push/PR): golangci-lint, gofmt, `go vet`, `go test`;
  frontend prettier, eslint, `tsc --noEmit`, vite build.
- **Release** (`release.yml`, tags `v*` or manual dispatch): matrix build
  producing the AppImage / DMG / MSI above, with the tag in the file names,
  attached to the GitHub Release with a downloads table.
