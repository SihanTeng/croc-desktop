#!/usr/bin/env bash
# croc-desktop one-line installer (Linux / macOS)
#
#   curl -fsSL https://raw.githubusercontent.com/SihanTeng/croc-desktop/main/install.sh | bash
#
# Optional:
#   VERSION=v0.2.0  curl -fsSL ... | bash          # pin a release tag
#   INSTALL_DIR=~/bin curl -fsSL ... | bash       # Linux AppImage dest (default ~/.local/bin)
#   GITHUB_TOKEN=...  curl -fsSL ... | bash       # higher API rate limit
#
# Windows users:
#   irm https://raw.githubusercontent.com/SihanTeng/croc-desktop/main/install.ps1 | iex

set -euo pipefail

REPO="${REPO:-SihanTeng/croc-desktop}"
APP_NAME="croc-desktop"
GITHUB_API="https://api.github.com"
GITHUB_DL="https://github.com"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { printf '%b[INFO]%b %s\n'  "$GREEN"  "$NC" "$*"; }
warn()  { printf '%b[WARN]%b %s\n'  "$YELLOW" "$NC" "$*"; }
error() { printf '%b[ERROR]%b %s\n' "$RED"    "$NC" "$*" >&2; exit 1; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "Required command not found: $1"
}

curl_gh() {
  # Usage: curl_gh URL [extra curl args...]
  local url=$1
  shift
  local -a args=(-fsSL)
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    args+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi
  args+=(-H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28")
  curl "${args[@]}" "$@" "$url"
}

detect_platform() {
  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  case "$os" in
    linux)  OS=linux ;;
    darwin) OS=darwin ;;
    mingw*|msys*|cygwin*)
      error "Use PowerShell on Windows: irm https://raw.githubusercontent.com/${REPO}/main/install.ps1 | iex"
      ;;
    *) error "Unsupported OS: $os" ;;
  esac

  case "$arch" in
    x86_64|amd64) ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) error "Unsupported architecture: $arch" ;;
  esac

  # Map to release artifact names (see .github/workflows/release.yml).
  case "${OS}-${ARCH}" in
    linux-amd64)
      ASSET_SUFFIX="linux-amd64.AppImage"
      KIND=appimage
      ;;
    darwin-arm64)
      ASSET_SUFFIX="darwin-arm64.dmg"
      KIND=dmg
      ;;
    linux-arm64)
      error "No Linux arm64 release yet. Build from source or use an amd64 machine. See: https://github.com/${REPO}/releases"
      ;;
    darwin-amd64)
      error "No Intel macOS (amd64) release yet. Need Apple Silicon, or build from source. See: https://github.com/${REPO}/releases"
      ;;
    *)
      error "No prebuilt package for ${OS}-${ARCH}. See: https://github.com/${REPO}/releases"
      ;;
  esac

  info "Detected platform: ${OS}-${ARCH}"
}

get_latest_version() {
  if [ -n "${VERSION:-}" ]; then
    # Allow VERSION=0.2.0 or v0.2.0
    case "$VERSION" in
      v*) LATEST_VERSION="$VERSION" ;;
      *)  LATEST_VERSION="v${VERSION}" ;;
    esac
    info "Using requested version: ${LATEST_VERSION}"
    return
  fi

  info "Fetching latest release…"
  local json
  json=$(curl_gh "${GITHUB_API}/repos/${REPO}/releases/latest") || error "Failed to query GitHub releases for ${REPO}"

  LATEST_VERSION=$(printf '%s' "$json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)
  [ -n "$LATEST_VERSION" ] || error "Could not parse latest tag from GitHub API"
  info "Latest version: ${LATEST_VERSION}"
}

download_asset() {
  ASSET_NAME="${APP_NAME}_${LATEST_VERSION}_${ASSET_SUFFIX}"
  DOWNLOAD_URL="${GITHUB_DL}/${REPO}/releases/download/${LATEST_VERSION}/${ASSET_NAME}"

  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT
  DEST_FILE="${TMP_DIR}/${ASSET_NAME}"

  info "Downloading ${ASSET_NAME}…"
  info "  ${DOWNLOAD_URL}"
  if ! curl -fsSL -o "$DEST_FILE" "$DOWNLOAD_URL"; then
    error "Download failed. Check that ${LATEST_VERSION} has ${ASSET_NAME}: https://github.com/${REPO}/releases/tag/${LATEST_VERSION}"
  fi
  info "Download complete ($(du -h "$DEST_FILE" | awk '{print $1}'))"
}

install_linux_appimage() {
  need_cmd chmod
  local install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
  mkdir -p "$install_dir"
  local target="${install_dir}/${APP_NAME}"

  install -m 0755 "$DEST_FILE" "$target"
  info "Installed AppImage → ${target}"

  # Desktop entry so the app shows up in menus (best-effort).
  local apps_dir="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
  mkdir -p "$apps_dir"
  cat > "${apps_dir}/${APP_NAME}.desktop" <<EOF
[Desktop Entry]
Name=croc-desktop
Comment=Easily and securely transfer files between computers
Exec=${target}
Icon=application-x-executable
Type=Application
Categories=Network;Utility;
Terminal=false
StartupWMClass=croc-desktop
EOF
  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$apps_dir" 2>/dev/null || true
  fi
  info "Desktop entry → ${apps_dir}/${APP_NAME}.desktop"

  case ":$PATH:" in
    *":${install_dir}:"*) ;;
    *)
      warn "${install_dir} is not on your PATH."
      warn "Add this to your shell rc (~/.bashrc / ~/.zshrc):"
      printf '\n    export PATH="$PATH:%s"\n\n' "$install_dir"
      ;;
  esac

  echo
  info "Run:  ${APP_NAME}"
  info "  or: ${target}"
}

install_macos_dmg() {
  need_cmd hdiutil
  need_cmd ditto

  local mount_point
  mount_point=$(mktemp -d "${TMPDIR:-/tmp}/croc-dmg.XXXXXX")

  info "Mounting DMG…"
  # -nobrowse keeps Finder quiet; capture the device for clean detach.
  local attach_out device
  attach_out=$(hdiutil attach -nobrowse -readonly -mountpoint "$mount_point" "$DEST_FILE")
  device=$(printf '%s\n' "$attach_out" | awk '/^\/dev\// { print $1; exit }')

  cleanup_dmg() {
    if [ -n "${device:-}" ]; then
      hdiutil detach "$device" -quiet 2>/dev/null || hdiutil detach "$mount_point" -quiet 2>/dev/null || true
    else
      hdiutil detach "$mount_point" -quiet 2>/dev/null || true
    fi
    rmdir "$mount_point" 2>/dev/null || true
  }
  trap 'cleanup_dmg; rm -rf "$TMP_DIR"' EXIT

  local app_src
  app_src=$(find "$mount_point" -maxdepth 2 -name '*.app' -type d | head -1)
  if [ -z "$app_src" ]; then
    # Fallback: bare binary in the DMG
    local bin_src
    bin_src=$(find "$mount_point" -maxdepth 2 -type f -name "${APP_NAME}" | head -1)
    if [ -n "$bin_src" ]; then
      local install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
      mkdir -p "$install_dir"
      install -m 0755 "$bin_src" "${install_dir}/${APP_NAME}"
      info "Installed binary → ${install_dir}/${APP_NAME}"
      cleanup_dmg
      trap 'rm -rf "$TMP_DIR"' EXIT
      return
    fi
    error "No .app (or ${APP_NAME} binary) found inside the DMG"
  fi

  local app_name
  app_name=$(basename "$app_src")
  local dest="/Applications/${app_name}"

  info "Installing ${app_name} → /Applications…"
  if [ -d "$dest" ]; then
    if [ -w "/Applications" ] || [ -w "$dest" ]; then
      rm -rf "$dest"
    else
      sudo rm -rf "$dest"
    fi
  fi

  if [ -w "/Applications" ]; then
    ditto "$app_src" "$dest"
  else
    warn "Need admin rights to write /Applications"
    sudo ditto "$app_src" "$dest"
  fi

  # Remove quarantine so first launch is less painful (ad-hoc signed builds).
  if command -v xattr >/dev/null 2>&1; then
    xattr -dr com.apple.quarantine "$dest" 2>/dev/null || true
  fi

  cleanup_dmg
  trap 'rm -rf "$TMP_DIR"' EXIT

  info "Installed → ${dest}"
  warn "Release builds are ad-hoc signed. If Gatekeeper blocks open:"
  warn "  System Settings → Privacy & Security → Open Anyway"
  echo
  info "Open from Launchpad / Applications, or:"
  info "  open ${dest}"
}

main() {
  printf '\n'
  printf '%bcroc-desktop%b — install\n' "$CYAN" "$NC"
  printf 'Repo: https://github.com/%s\n\n' "$REPO"

  need_cmd curl
  need_cmd uname
  need_cmd mktemp

  detect_platform
  get_latest_version
  download_asset

  case "$KIND" in
    appimage) install_linux_appimage ;;
    dmg)      install_macos_dmg ;;
    *)        error "Internal error: unknown package kind $KIND" ;;
  esac

  printf '\n'
  info "Done. Happy transferring."
  printf '\n'
}

main "$@"
