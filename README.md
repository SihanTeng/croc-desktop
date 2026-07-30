<p align="center">
  <img src="build/appicon.png" alt="croc logo" width="112" height="112" />
</p>

<h1 align="center">croc-desktop</h1>

<p align="center">
  <b>Send files and text between devices — simply, privately, and without the cloud.</b>
</p>

<p align="center">
  A friendly app for
  <a href="https://github.com/schollz/croc">croc</a>:
  end-to-end encrypted, peer-to-peer transfer.
  <br />
  <b>Desktop:</b> Linux · macOS · Windows &nbsp;·&nbsp;
  <b>Mobile:</b> iOS · Android
  <br />
  <sub>Built with <a href="https://v3.wails.io">Wails v3</a> (not v2)</sub>
</p>

<p align="center">
  <a href="https://github.com/SihanTeng/croc-desktop/releases/latest"><img src="https://img.shields.io/github/v/release/SihanTeng/croc-desktop?style=for-the-badge&label=Download&color=1F6FEB" alt="Download" /></a>
  &nbsp;
  <a href="https://github.com/SihanTeng/croc-desktop/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/SihanTeng/croc-desktop/ci.yml?branch=main&style=for-the-badge&label=CI" alt="CI" /></a>
  &nbsp;
  <img src="https://img.shields.io/badge/Desktop-Linux%20%7C%20macOS%20%7C%20Windows-1F6FEB?style=for-the-badge" alt="Desktop platforms" />
  &nbsp;
  <img src="https://img.shields.io/badge/Mobile-iOS%20%7C%20Android-0E8A16?style=for-the-badge" alt="Mobile platforms" />
  &nbsp;
  <a href="https://v3.wails.io"><img src="https://img.shields.io/badge/Wails-v3-DF3A2C?style=for-the-badge" alt="Wails v3" /></a>
</p>

<p align="center">
  <a href="#-install-desktop">Install desktop</a> ·
  <a href="#-mobile-ios--android">Mobile</a> ·
  <a href="#-how-to-use">How to use</a> ·
  <a href="#-features">Features</a> ·
  <a href="#-faq">FAQ</a> ·
  <a href="#-contributing">Contributing</a> ·
  <a href="#-development">Development</a>
</p>

---

<p align="center">
  <img src="docs/images/send.png" alt="Send screen with code and QR" width="48%" />
  &nbsp;
  <img src="docs/images/receive.png" alt="Receive screen with file preview" width="48%" />
</p>

<p align="center">
  <img src="docs/images/history.png" alt="History of past transfers" width="48%" />
  &nbsp;
  <img src="docs/images/logs.png" alt="Live logs" width="48%" />
</p>

## Why croc-desktop?

Moving a file to another computer should not mean uploading it to someone else’s
servers — or learning a command line.

**croc-desktop** is a full GUI for [croc](https://github.com/schollz/croc):

| | |
| --- | --- |
| **Private by default** | End-to-end encrypted, peer-to-peer. Nothing is stored in the cloud. |
| **One code (or QR)** | Share a short phrase or scan a QR — done. |
| **You stay in control** | Accept or decline each transfer. Overwrite and resume when you choose. |
| **Works with the CLI** | Codes are the same as `croc` on the command line. Mix and match devices. |
| **No terminal required** | Point, click, drop, paste. |

Unlike thin wrappers that only launch the `croc` CLI in a window, this app runs
croc **inside the process**, so you get real dialogs, live progress, history,
and relay hosting.

> **Stack note:** this project uses **[Wails v3](https://v3.wails.io)**
> (`github.com/wailsapp/wails/v3`). Older docs or the `main` branch may still
> mention Wails v2 — that is obsolete here.

## Install (desktop)

Download the latest **desktop** build for your system:

**[Releases →](https://github.com/SihanTeng/croc-desktop/releases/latest)**

| Platform | File | Tips |
| --- | --- | --- |
| **Linux** | `croc-desktop_*_linux-amd64.AppImage` | `chmod +x` the file, then double-click or run it |
| **macOS** (Apple Silicon) | `croc-desktop_*_darwin-arm64.dmg` | Open the DMG and drag the app to Applications. First open may need **System Settings → Privacy & Security** (ad-hoc signed) |
| **Windows** | `croc-desktop_*_windows-amd64.msi` | Run the installer |

> Both sides of a transfer need either this app or the [croc CLI](https://github.com/schollz/croc).

## Mobile (iOS · Android)

Same app, same UI (bottom tabs on a phone-sized screen). Powered by **Wails v3**
native mobile targets.

| Client | Status | Where it lives | How to run |
| --- | --- | --- | --- |
| **iOS** | Experimental — build from source | Scaffold in [`build/ios/`](build/ios/) | Simulator / device (needs full **Xcode** on macOS) |
| **Android** | Experimental — build from source | Scaffold in [`build/android/`](build/android/) | Emulator / device (needs **Android SDK + NDK**) |

**There are no App Store / Play Store packages yet** — mobile is for people who
build from this repo (or a CI you set up). Desktop remains the supported
download path on [Releases](https://github.com/SihanTeng/croc-desktop/releases).

### Build & run mobile clients

Prerequisites: same as [Development](#-development) (Go, Node, Wails v3 CLI),
plus the platform SDK.

```sh
# Install Wails v3 CLI (pin matches go.mod)
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.119

# Frontend assets (embedded into the app)
npm --prefix frontend ci && npm --prefix frontend run build

# ── iOS (macOS + Xcode only) ─────────────────────────────────
wails3 task ios:run                          # Simulator
# Device:
# wails3 task ios:package IOS_PLATFORM=device \
#   CODESIGN_IDENTITY="Apple Development: You (TEAMID)"

# ── Android ──────────────────────────────────────────────────
wails3 task android:run                      # emulator or adb device
```

| | Value |
| --- | --- |
| iOS bundle ID | `com.schollz.croc-desktop` |
| Android applicationId | `com.schollz.crocdesktop` (no hyphens) |
| Config | [`build/config.yml`](build/config.yml) (`ios:` block) |
| Platform hooks | [`app_options_ios.go`](app_options_ios.go), [`app_options_android.go`](app_options_android.go) |

**UI tip:** resize the desktop window to ~360px wide to preview the mobile
bottom-tab layout without a phone.

## How to use

### Send something

1. Open **Send**.
2. Drop files or a folder, browse to pick them, or switch to **Text** for a snippet.
3. Press **Send**. You’ll get a short **code** and a **QR**.
4. Share the code (or the QR image) with the person receiving.
5. Wait for them to accept — progress shows live. Cancel anytime with the button or `Esc`.

### Receive something

1. Open **Receive**.
2. Paste the code (or a whole `croc …` command / share link) — or load a **QR screenshot**.
3. Choose a download folder if you like, then **Receive**.
4. **Accept** (or decline) when the prompt appears.
5. When it’s done, open previews for images, video, audio, or text right in the app.

### Everyday extras

- **History** — past sends and receives, searchable.
- **Favorites** — save codes you reuse often.
- **Send again** — re-send the same files or text without re-picking.
- **Settings** — theme (light / dark / system), language, custom relays, proxies, and more.
- **Relay** — optionally host a relay from the app for locked-down networks.

## Features

### Transfer

- Files & folders (drag and drop or picker)
- Text snippets
- QR code + copy code / copy CLI command
- Paste codes in any common format (`code`, `croc code`, `CROC_SECRET=…`, share links)
- Decode codes from QR screenshots
- Accept / decline and overwrite / resume prompts
- Live progress, stall hints, cancel anytime
- Resume interrupted receives
- Inline previews after receive
- Desktop notifications on finish or failure

### App

- Transfer **history** that survives restarts
- Live **logs** with level filters
- Optional in-app **relay**
- Custom relay address & password, crypto curve, hash algorithm
- Local-only / disable-local modes, compression, proxies (SOCKS5 / HTTP)
- Zip folders, exclude patterns, upload rate limit
- Light / dark / system theme
- Languages: English, 简体中文, 繁體中文, Español, Français, Deutsch, 日本語
- Responsive layout: side rail on desktop; **bottom tabs on phones / small windows**

## FAQ

<details>
<summary><b>Is my data uploaded to a server?</b></summary>

<br>

No. Transfers are end-to-end encrypted and peer-to-peer. A public **relay** may
help peers find each other (same idea as the croc CLI), but file contents are
not stored there. You can run your own relay from the app or point at a private
one in Settings.

</details>

<details>
<summary><b>Do both people need this app?</b></summary>

<br>

No. One side can use **croc-desktop** and the other the official
[`croc` CLI](https://github.com/schollz/croc) (or vice versa). Codes are
compatible.

</details>

<details>
<summary><b>Does it work offline / on a LAN only?</b></summary>

<br>

Yes. Enable **local-only** (or use local discovery) in Settings when both
devices are on the same network. You can also run a relay yourself under the
**Relay** tab.

</details>

<details>
<summary><b>Where are settings and history stored?</b></summary>

<br>

In the same config area as the croc CLI (override with the `CROC_CONFIG_DIR`
environment variable):

- `croc-desktop.json` — preferences
- `croc-desktop-history.json` — transfer history

Older installs that used `croc-gui.json` are upgraded automatically on first run.

</details>

<details>
<summary><b>macOS says the app is from an unidentified developer</b></summary>

<br>

Release builds are currently **ad-hoc signed**, not notarized. Open
**System Settings → Privacy & Security** and allow the app after the first
blocked launch, or right-click → **Open**.

</details>

<details>
<summary><b>Is this still Wails v2?</b></summary>

<br>

**No.** This branch uses **Wails v3** (`github.com/wailsapp/wails/v3`, see
`go.mod`). Docs: [v3.wails.io](https://v3.wails.io). If you still see “Wails v2”
on GitHub, you are looking at an older branch (e.g. `main` before merge).

</details>

<details>
<summary><b>Where do I get the mobile apps?</b></summary>

<br>

See **[Mobile (iOS · Android)](#-mobile-ios--android)** above. They are built
from this repo with `wails3 task ios:run` / `android:run` — not published as
store installers yet.

</details>

## Contributing

Contributions are welcome — bug reports, translations, docs, and pull requests.

- Be kind; we follow the [Code of Conduct](CODE_OF_CONDUCT.md).
- Prefer small, focused pull requests.
- For UI strings, edit JSON under `frontend/src/i18n/locales/` (`en.json` is the source of truth). Copy it to a new locale file (e.g. `pt-BR.json`) and translate values; keep keys and `{placeholders}` intact.

## Development

<details>
<summary><b>Build from source (developers)</b></summary>

<br>

**Stack:** Go (see `go.mod`), Node.js, **[Wails v3](https://v3.wails.io)** —
module `github.com/wailsapp/wails/v3` @ **v3.0.0-alpha2.119** (not the v2 CLI).

```sh
# Wails v3 CLI (use -tags gtk3 on Linux without gtk4)
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.119

# Linux: webkit2gtk 4.1 + GTK3 (project default)
#   Fedora: webkit2gtk4.1-devel
#   Debian/Ubuntu: libgtk-3-dev libwebkit2gtk-4.1-dev

# Dev with hot reload (desktop)
WEBKIT_DISABLE_DMABUF_RENDERER=1 wails3 dev -config ./build/config.yml -port 34115

# Production desktop binary → bin/croc-desktop
wails3 build -tags gtk3   # omit -tags gtk3 on macOS/Windows

# Mobile — see “Mobile (iOS · Android)” section
wails3 task ios:run
wails3 task android:run

# Tests
npm --prefix frontend ci && npm --prefix frontend run build
go test -tags gtk3 .
npm --prefix frontend test
./e2e/run.sh
```

More detail: `e2e/README.md`, `design.md`, and Taskfiles under `build/`.

**How it works (short):** the app depends on a croc fork with a hooks layer so
progress and prompts are events in the UI, not a spawned CLI. See `go.mod`
`replace` for the pin.

</details>

## Acknowledgments

- [croc](https://github.com/schollz/croc) by [@schollz](https://github.com/schollz) — the protocol and transfer engine
- [Wails v3](https://v3.wails.io) — Go + web UI for desktop and mobile

---

<p align="center">
  <sub>Made for people who just want to send a file.</sub>
</p>
