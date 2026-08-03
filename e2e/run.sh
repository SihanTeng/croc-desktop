#!/usr/bin/env bash
# Run the browser E2E suite.
#
# - If something already serves http://localhost:34115, the tests run against
#   it (transfers go over the public croc relay). Only a server-mode instance
#   is useful there: `wails3 dev` gives external browsers no /wails bridge.
# - Otherwise a hermetic sandbox is booted: a local relay, an isolated
#   CROC_CONFIG_DIR, and the app in Wails v3 *server mode* (headless HTTP+WS,
#   no webview/X). This is the mode CI uses.
set -euo pipefail
cd "$(dirname "$0")"

URL="${E2E_APP_URL:-http://localhost:34115}"
RELAY_BASE=29309

APP_PID=""
RELAY_PID=""
SANDBOX=""
cleanup() {
  [ -n "$APP_PID" ] && kill "$APP_PID" 2>/dev/null || true
  [ -n "$RELAY_PID" ] && kill "$RELAY_PID" 2>/dev/null || true
  [ -n "$SANDBOX" ] && rm -rf "$SANDBOX" || true
}
trap cleanup EXIT

if curl -sf "$URL" -o /dev/null 2>&1; then
  echo ">> using the running app at $URL"
else
  echo ">> no app at $URL; booting a sandboxed instance"
  SANDBOX="$(mktemp -d)"
  export E2E_RELAY="127.0.0.1:$RELAY_BASE"
  mkdir -p "$SANDBOX/config"
  # settings file name matches settings.go (croc-desktop.json); the load
  # path still accepts legacy croc-gui.json for upgraded installs
  cat > "$SANDBOX/config/croc-desktop.json" <<JSON
{"relayAddress":"127.0.0.1:$RELAY_BASE","relayPassword":"pass123","disableLocal":true,"curve":"p256","hashAlgorithm":"xxhash"}
JSON
  (cd .. && go build -o "$SANDBOX/peers" ./e2e/peers)
  "$SANDBOX/peers" relay "$RELAY_BASE" &
  RELAY_PID=$!
  # the app binary boots in ~1s; give the relay a moment to bind before any
  # test transfer tries to dial it
  echo ">> waiting for the sandbox relay"
  relay_up=0
  for _ in $(seq 1 60); do
    if (echo > "/dev/tcp/127.0.0.1/$RELAY_BASE") 2>/dev/null; then relay_up=1; break; fi
    sleep 1
  done
  [ "$relay_up" = 1 ] || { echo "!! relay never came up"; exit 1; }

  # Wails v3 dev mode has no bridge for external browsers (the webview
  # intercepts /wails natively), so the sandbox runs the app built with
  # -tags server: it serves the embedded frontend plus the /wails HTTP
  # transport (calls) and WebSocket (events) itself. CGO off — no
  # gtk/webkit/wails3 CLI/xvfb needed.
  [ -f ../frontend/dist/index.html ] || {
    echo "!! build the frontend first: npm --prefix frontend ci && npm --prefix frontend run build"
    exit 1
  }
  (cd .. && CGO_ENABLED=0 go build -tags server -o "$SANDBOX/croc-desktop-server" .)
  CROC_CONFIG_DIR="$SANDBOX/config" WAILS_SERVER_PORT=34115 \
    "$SANDBOX/croc-desktop-server" &
  APP_PID=$!

  echo ">> waiting for the app (first build can take a few minutes)"
  for _ in $(seq 1 240); do
    curl -sf "$URL" -o /dev/null 2>&1 && break
    sleep 1
  done
  curl -sf "$URL" -o /dev/null 2>&1 || { echo "!! app never came up"; exit 1; }
fi

npm ci --no-audit --no-fund
# the bundled browser is only needed when no system Chrome is configured
# (CI); local runs use the system Chrome (see playwright.config.js)
if [ "${CI:-}" = "true" ]; then
  npx playwright install chromium
fi
npx playwright test "$@"
