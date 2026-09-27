# croc v11 desktop migration

The desktop now uses **croc v11.5.4**, the latest published release checked on
2026-09-27, building on PR #2's v11 import migration. This includes the v11.5.4
fixes for skipped archive extraction and accidental stdout-file deletion.

## User-visible changes

- **macOS 13 Ventura or newer is required.** Croc v11.5.4 requires Go 1.27,
  whose runtime no longer supports macOS 12. The native build target, app bundle
  metadata and Homebrew template now agree with this minimum.
  See the [Go 1.27 release notes](https://go.dev/doc/go1.27#darwin).

- **Both peers must use croc v11.** The v11 handshake binds authentication to the
  participants and session. v10 clients (including older desktop releases) are
  incompatible; the desktop explains that both devices need updating when it
  receives the incompatible-protocol error. There is no insecure v10 fallback.
- **Generated codes have three words**, replacing the old numeric prefix and
  three words. Copy, paste, QR, history and favorites continue to accept the code
  as text. Legacy/custom code syntax is still parsed by v11, but that does not
  make a v10 peer compatible.
- **Copy command matches the local OS.** On macOS/Linux it now produces
  `CROC_SECRET='code' croc`, because the v11 CLI rejects positional receive codes
  on Unix. Windows retains `croc code`. Recipients on a different OS can copy
  the bare code or scan the QR and enter it in their desktop/CLI prompt.
- **Text is limited to 1 MiB of UTF-8 bytes**, matching upstream validation.
  Oversized text is rejected before a transfer starts, with instructions to send
  it as a file. The text view shows the limit in every supported language.
- **Receive paths are more restrictive.** Unsafe traversal, symlinks escaping
  the receive root, sensitive destinations and invalid/duplicate metadata are
  rejected before the acceptance dialog. Existing-file protection also covers
  empty files and symlinks; declining an archive does not extract or remove it.
- **Hosted relays now rate-limit joins** (upstream defaults: 30 per source IP
  and per room in a sliding minute). A throttled transfer asks you to wait and
  retry. The test-only loopback relays use larger limits for rapid test runs.
- **LAN, custom relays and proxies retain their routing.** The desktop disables
  the new optional Tailcat/DERP transport. Normal v11 CLI peers fall back to the
  croc relay/LAN path; a CLI sender forcing `--transport derp` must instead use
  `--transport auto` or `--transport relay` with this desktop.

Accept/decline, overwrite/resume, live progress, cancellation, inline text display
and history remain desktop features. Text arrives through a callback before
upstream removes its temporary receive artifact; it is not printed to stdout.
Existing settings and history paths do not change. The app's product version is
still supplied by its release tag; Settings reports engine `v11.5.4`.

## Implementation and verification

The reviewed callback patch and its update/check instructions are in
[`third_party/croc/UPSTREAM.md`](../third_party/croc/UPSTREAM.md). Upstream source
is pinned and checked in because the official library does not yet expose the
GUI hooks. The desktop build does not depend on an unpublished fork commit.

```sh
npm --prefix frontend ci
npm --prefix frontend run build
python3 scripts/check-croc-source.py
CGO_ENABLED=0 go test -tags server -timeout 300s ./...
go test -race -tags server -timeout 300s .
./e2e/run.sh
```

For interoperability tests, install the unmodified upstream CLI into a temporary
directory, set `CROC_TEST_CLI` to its absolute executable path, and run
`go test -tags server -run TestOfficialCLIInterop -v .`. The tests send files and
text in both directions through a local relay and verify contents and text-file
cleanup. CI builds the official CLI independently of the local Go replacement.

References: [v11.0.0](https://github.com/schollz/croc/releases/tag/v11.0.0),
[v11.5.4](https://github.com/schollz/croc/releases/tag/v11.5.4).

## Verification results (2026-09-27)

- Full desktop Go suite, including official v11.5.4 CLI file/text transfers in
  both directions: passed. Race detector and `go vet`: passed.
- Upstream receive/path/handshake regression selection and SOCKS5 transfer test:
  passed. Pinned-source provenance check: passed (seven patched/added files).
- Frontend build, lint, formatting and 26 unit tests: passed.
- All seven browser transfer tests: passed, including the exact engine version
  and oversized-text rejection.
- Native macOS: built and transferred text in both directions with the official
  CLI, verified acceptance and inline display, cancellation, version display and
  the copied Unix command by pasting it into the app.
- Windows amd64 desktop: cross-compiled successfully. Windows/Linux native UI
  runtime and public-relay/WAN behavior were not exercised in this migration.

Native macOS validation used isolated settings and a loopback relay. The final
production build and Wails release build passed with the corrected macOS 13
deployment target; the binary reports `minos 13.0`. This is not validation on a
macOS 13 machine or a signed distribution release. No release was published.

## Final review

Two additional merge blockers were fixed during final review:

- The existing golangci-lint v1.64.8 fails to decode Go 1.27 export data. CI now
  pins v2.14.0, with the configuration migrated and the existing checks retained.
  Installed npm dependencies and the checksum-pinned upstream snapshot are
  excluded from project formatting; upstream provenance and security tests run
  separately. CI now also runs the race detector.
- macOS 12 was incorrectly advertised by the app bundle and Homebrew template.
  Build flags, bundle metadata, install documentation and the cask template now
  require macOS 13, matching the Go runtime.

The final linter reports zero issues. GUI-specific tests additionally confirm
resume percentage reporting and that declining an archive preserves it without
extraction across reconnects. The complete backend suite, including the official
CLI, passes under the race detector. The actual Wails release build (including
binding generation) passes and its binary declares macOS 13 as its minimum.

PR #2's reviewed base is `8e522a3599f1cb6cd7427d5be7e2626c97056ec7` and its head
is `dba8ffff9a40f4103ae1818db7b067765c42a920`. The follow-up is based directly on
that head. At review time the PR is still a draft, and GitHub's fork workflow run
requires maintainer approval. Local verification does not replace a green run
on the updated PR head before merging. Windows/Linux native runtime validation
and testing on macOS 13 hardware remain outside the verified evidence.
