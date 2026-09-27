# Pinned croc engine

This directory contains the `src` and `internal` library trees, module files,
license and third-party notices from **schollz/croc v11.5.4**, commit
`15a4577e56bf6bbad116576e448bd33120bd90c1`.
The exact module checksums and copied paths are in `UPSTREAM.json`.

Upstream does not expose the acceptance, overwrite and progress callbacks the
desktop requires. The Go module replacement uses this checked-in snapshot so a
clean checkout builds without an unpublished fork or a modified module cache.
The small diff from upstream is reviewable in `../../patches/croc-desktop.patch`.

The patch adds:

- `Hooks` installed before a transfer, for validated acceptance, overwrite/resume,
  send confirmation, state, progress and text delivery.
- A progress metadata snapshot under the existing progress lock, so callbacks
  do not read mutable file metadata from data-worker goroutines.
- Text delivery to a callback instead of stdout. Upstream still validates the
  offer, generates a safe temporary name, reads the payload and owns cleanup.
- `comm.WithProxy`, an immutable per-transfer context setting. Desktop transfers
  do not rewrite the CLI proxy globals while asynchronous dialers are active.
- `Options.DisableTailcat`, set by the desktop to preserve its existing
  LAN/relay/proxy network routing. Unmodified CLI behavior is the default when
  this option is false and no hooks are installed.

Receive-path, archive, hash, handshake and overwrite validation are retained.
Callbacks run after validation and do not bypass the overwrite eligibility gate.

## Verify and update

From the desktop repository root:

```sh
python3 scripts/check-croc-source.py
```

This downloads the exact upstream module through Go's checksum verification and
compares every included file with the reviewed patch. CI runs this check and the
upstream receive/security regression tests; nested-module tests are not included
automatically in the desktop's `go test ./...`.

For an engine update, update the root requirement and `UPSTREAM.json`, copy the
listed paths from the checksum-verified module, and port the existing patch.
Review all upstream changes around acceptance, file creation, progress and text
cleanup before regenerating the patch with `--write-patch`. Run desktop unit,
race, browser and official-CLI interoperability tests, plus the upstream security
tests. Keep the upstream license and notices. A future upstream callback API can
replace this snapshot with the official module directly.
