<p align="center">
  <img src="build/appicon.png" alt="croc logo" width="96" height="96" />
</p>

<h1 align="center">croc-desktop</h1>

<p align="center">
  <strong>GUI for <a href="https://github.com/schollz/croc">croc</a></strong> — encrypted,
  peer-to-peer file and text transfer on <strong>Linux, macOS, Windows</strong>
  (desktop) and experimental <strong>iOS / Android</strong> via Wails v3.
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
- Responsive layout: side rail on desktop, bottom tabs on narrow / mobile screens

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

This repo tracks **Wails `v3.0.0-alpha2.119`** (see `go.mod`). Prefer the same
CLI version for reproducible builds:

```sh
# Linux: install CLI with gtk3 if you have webkit2gtk 4.1 but not gtk4
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.119

# Linux webview deps (project default = gtk3 / webkit2gtk 4.1), e.g.
#   webkit2gtk4.1-devel on Fedora, libwebkit2gtk-4.1-dev on Debian/Ubuntu
# Optional gtk4 path: install gtk4 + webkitgtk 6 and build with EXTRA_TAGS=

WEBKIT_DISABLE_DMABUF_RENDERER=1 wails3 dev -config ./build/config.yml -port 34115
# or: task dev
```

(`WEBKIT_DISABLE_DMABUF_RENDERER=1` works around a WebKitGTK crash on some
Wayland compositors.)

Dev mode loads the Vite server (port **34115**) and rebuilds Go on change.
Frontend bindings live under `frontend/bindings/` and are regenerated during
`wails3 build` / `wails3 generate bindings -ts`.

On Linux, unit tests and the app must use the **gtk3** build tag when gtk4
is not installed:

```sh
go test -tags gtk3 .
go build -tags gtk3 -o bin/croc-desktop .
```

## Build

```sh
wails3 build -tags gtk3   # Linux with webkit2gtk 4.1 (default for this repo)
# or: task build
# binary lands in bin/croc-desktop  (not build/bin/)
```

On macOS/Windows, omit the gtk3 tag (`wails3 build`). For Linux with gtk4 +
webkitgtk 6 installed: `EXTRA_TAGS= wails3 build`.

## Mobile (iOS / Android)

Wails v3 can build the same Go + React app for mobile. Support is still
**alpha** (needs full Xcode for iOS; Android SDK + NDK for Android). The UI is
responsive: desktop keeps the left side-rail; below ~720px it switches to a
bottom tab bar with safe-area padding.

Scaffold lives under `build/ios/` and `build/android/`. Platform option hooks
are `app_options_*.go` in the module root (`//go:build ios|android`).

```sh
# iOS Simulator (requires full Xcode on macOS)
wails3 task ios:run
# device: task ios:package IOS_PLATFORM=device CODESIGN_IDENTITY="Apple Development: …"

# Android emulator / device
wails3 task android:run
```

- iOS bundle ID / Android `applicationId`: **`com.schollz.croc-desktop`**
- Android Java package for the Wails bridge remains `com.wails.app` (upstream
  scaffold); only the Play Store id is product-branded.
- Resize the desktop window to ~360px wide to preview mobile chrome without a device.

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

## Data paths & upgrades

Settings and history live in the croc config directory (same as the CLI;
overridable with `CROC_CONFIG_DIR`):

| File | Purpose |
| --- | --- |
| `croc-desktop.json` | Settings (theme, relay, favorites, …) |
| `croc-desktop-history.json` | Transfer history |

**Backward compatibility:** if the new file is missing, the app still loads
the pre-rename names (`croc-gui.json`, `croc-gui-history.json`) and writes
the data under the current names on first use. Existing installs keep their
settings and history after upgrading.

## Contributing translations

UI strings live in plain JSON files under `frontend/src/i18n/locales/`
(`en.json` is the canonical key set). To add or improve a language, copy
`en.json` to `<locale>.json` (e.g. `pt-BR.json`) and translate the values —
Vite picks the file up automatically, no code changes needed. Keep the keys
and the `{placeholder}` variables intact.

## How it works

- Built with **Wails v3**: Go services bound to the frontend, events via
  `app.Event.Emit`, dialogs via `app.Dialog`, assets embedded from
  `frontend/dist`.
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

- **CI** (`ci.yml`, every push/PR): golangci-lint, gofmt, `go vet`, `go test`
  (all with `-tags gtk3` on Linux); frontend prettier, eslint, `tsc`, vitest,
  vite build; Playwright E2E.
- **Release** (`release.yml`, tags `v*` or manual dispatch): matrix build
  producing the AppImage / DMG / MSI above, with the tag in the file names,
  attached to the GitHub Release with a downloads table. Linux uses
  `wails3 build -tags gtk3`; binary path is `bin/`.
