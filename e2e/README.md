# Browser E2E tests

End-to-end tests that drive the real app UI in Chromium while throwaway croc
peers (`peers/`) play the other end of each transfer. Specs call
`window.go.main.App` / `window.runtime`, which the frontend exposes as a thin
compatibility bridge over v3 bindings.

## Run

```sh
./run.sh           # from this directory
```

- If something already serves http://localhost:34115, the tests use it and
  transfers go over the public croc relay (internet required). This must be a
  server-mode instance to be useful: in `wails3 dev` the webview intercepts
  `/wails/*` natively, so an external browser's bound calls never reach the
  backend.
- Otherwise `run.sh` boots a hermetic sandbox: a local relay (`peers relay`),
  an isolated `CROC_CONFIG_DIR`, and the app built with `-tags server`
  (Wails v3 server mode — headless HTTP+WS, no webview/xvfb). This is the
  mode CI uses.

Tests are serialized (one transfer at a time) and clean up after themselves:
fixtures (`.e2e-tmp/`) are deleted and the app's history/logs are cleared in
the global teardown.

## Layout

- `peers/` — Go test peers: `send`, `recv` (auto-accept), `relay`
- `*.spec.js` — playwright suites (send text, receive file, decline, Esc
  cancel, history & logs)
- `helpers.mjs` — page reset, peer spawning, fixtures, unique code phrases
