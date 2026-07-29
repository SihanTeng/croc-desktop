# Browser E2E tests

End-to-end tests that drive the real app UI in Chromium (via the Wails dev
server's browser bridge) while throwaway croc peers (`peers/`) play the other
end of each transfer.

## Run

```sh
./run.sh           # from this directory
```

- With the app already running (`wails dev`), the tests use it; transfers go
  over the public croc relay (internet required).
- Otherwise `run.sh` boots a hermetic sandbox: a local relay (`peers relay`)
  plus the app with an isolated `CROC_CONFIG_DIR`, under `xvfb` when there is
  no display. This is the mode CI uses.

Tests are serialized (one transfer at a time) and clean up after themselves:
fixtures (`.e2e-tmp/`) are deleted and the app's history/logs are cleared in
the global teardown.

## Layout

- `peers/` — Go test peers: `send`, `recv` (auto-accept), `relay`
- `*.spec.js` — playwright suites (send text, receive file, decline, Esc
  cancel, history & logs)
- `helpers.mjs` — page reset, peer spawning, fixtures, unique code phrases
