#!/usr/bin/env bash
# Run the browser E2E suite.
#
# - If an app is already serving http://localhost:34115 (e.g. `wails3 dev` /
#   `task dev`), the tests run against it (transfers go over the public croc
#   relay).
# - Otherwise a hermetic sandbox is booted: a local relay plus an isolated
#   CROC_CONFIG_DIR, with the app started under xvfb when there is no display.
set -euo pipefail
cd "$(dirname "$0")"

URL="${E2E_APP_URL:-http://localhost:34115}"
WAILS3="$(command -v wails3 || echo "$(go env GOPATH)/bin/wails3")"
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
  cat > "$SANDBOX/config/croc-gui.json" <<JSON
{"relayAddress":"127.0.0.1:$RELAY_BASE","relayPassword":"pass123","disableLocal":true,"curve":"p256","hashAlgorithm":"xxhash"}
JSON
  (cd .. && go run ./e2e/peers relay "$RELAY_BASE") &
  RELAY_PID=$!

  LAUNCHER=""
  if [ -z "${DISPLAY:-}" ]; then LAUNCHER="xvfb-run -a"; fi
  # gtk3 tag: use webkit2gtk 4.1 (same stack as the previous Wails v2 build).
  # VITE_PORT keeps the E2E default of 34115.
  (cd .. && CROC_CONFIG_DIR="$SANDBOX/config" WEBKIT_DISABLE_DMABUF_RENDERER=1 \
    WAILS_VITE_PORT=34115 \
    $LAUNCHER "$WAILS3" dev -config ./build/config.yml -port 34115) &
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
