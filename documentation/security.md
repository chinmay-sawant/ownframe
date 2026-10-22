# Security

ownframe treats templates, themes, and source files passed to `New`, `Load`,
`Route`, `Config`, and `SetData` as trusted author content. End user input,
remote URLs, and filesystem paths passed through app code need validation in the
app. This page collects the trust boundaries and links to the full audit.

## Audit

The complete audit lives under
[`plans/v0.0.2/security/`](../plans/v0.0.2/security/). Start with the overview,
then read the focused notes:

- [00-overview-threat-model-and-priority.md](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md)
- [01-template-content-xss-and-injection.md](../plans/v0.0.2/security/01-template-content-xss-and-injection.md)
- [02-web-server-network-and-fetch.md](../plans/v0.0.2/security/02-web-server-network-and-fetch.md)
- [03-file-paths-print-commands-and-local-data.md](../plans/v0.0.2/security/03-file-paths-print-commands-and-local-data.md)
- [04-runtime-devtools-crash-and-residual-risks.md](../plans/v0.0.2/security/04-runtime-devtools-crash-and-residual-risks.md)
- [05-security-phase-checklist.md](../plans/v0.0.2/security/05-security-phase-checklist.md)

## Assumptions

- Templates, themes, and files named in `Config` and passed to `Page` are
  trusted application content.
- Path inputs such as `Config.File`, `ThemeFile`, and `SetCrashDir`
  are trusted configuration.
- `Serve`, devtools, and debug flags only make sense on a machine the caller
  controls.
- `Fetch` and `XHR` load remote URLs chosen and validated by the app.

## Release refresh

Before tagging a release, re-check the assumptions above and the data
flow below against what actually shipped. New screens, inputs, fetch
callers, or mobile surfaces since the last tag can move a trust
boundary without touching this file.

```
author template/theme ──► New/Load/Route ──► layout ──► window / Serve
untrusted text/URLs ──► app validation ──► FormValue/fetch ──► display
OS/user files ──► dialog or Drop.Read ──► app ──► SetImage ──► display
```

The gate is documentation/release-checklist.md: the threat-model item
must be signed off before the tag.

## Drop path handling

`Drop.Path` is desktop only and is an absolute local path. It is sensitive
and must not be surfaced in UI, logs, crash reports, web previews, or IPC.
Copy bytes from `Drop.Read` during the `Drop` call; the read is only valid
for that frame. See [drag-drop.md](drag-drop.md).

## Mobile captured images

Examples that accept camera or gallery images (such as `examples/telegram`)
decode them inside `addPhoto` before allocation. Callers should bound image
dimensions before decoding and reject oversized or malformed payloads rather
than allocating from an untrusted source. Photos are queued on a bounded
channel and applied on the game loop.

## Debug and developer surfaces

- `Serve` exposes no `/debug/state` endpoint. `ServeWithOptions` registers
  `GET /debug/state` only with `ServeOptions` debug opt-ins. Treat it as privileged: it exposes layout
  boxes with field text, stats, and runtime snapshot on loopback only.
  Reduce or gate it before exposing a preview. `ServeOptions` pprof opt-ins
  are the only ones that add `GET /debug/pprof/*`.
  Both registrations additionally need `-tags ownframe_debug`: in a
  default build the routes answer 404 and `net/http/pprof` is not
  linked. See `supply-chain.md` ("Compiling out debug surfaces").
- DevTools summarizes rendered data on screen. Keep it off in shipped binaries
  unless debugging is intended. The examples enable `Config.DevTools` for
  debugging; a shipped binary leaves it false.
- Crash reports can contain stack traces, panic text, page titles, and source
  paths. Directory mode is `0700` and file mode `0600`, but callers should
  still choose a private directory on shared machines.
- `SetCrashDir` and other path APIs are trusted; the crash directory is not
  sandboxed.
- Environment debug flags (`OWNFRAME_FILEPICK_DEBUG`, `OWNFRAME_PACE_DEBUG`)
  are dev-only and off unless explicitly set. They write to stderr and may
  include paths or command failure details, so keep them off in shipped binaries.

## Boundary by API

- Web mode: [web.md](web.md)
- Fetch and XHR: [fetch.md](fetch.md)
- Crash reports: [crash.md](crash.md)
- IPC: [ipc.md](ipc.md)
- Drag and drop: [drag-drop.md](drag-drop.md)

## Local notification endpoints

`examples/desktop-cat` serves `POST /notify` on `127.0.0.1:6969` by default.
`POST /notify` and `GET /state` require a matching `Origin` against the
request scheme and host; a missing `Origin` returns 403, and curl callers
send `Origin` explicitly. The handler sets `nosniff`, `X-Frame-Options: DENY`,
`Referrer-Policy: no-referrer`, and a restrictive `Permissions-Policy`,
caps the body at 4 KiB, sets an idle timeout, and
returns generic `invalid notification` errors. Empty and wildcard binds are
rejected; other non-loopback binds log a warning. Remote use needs an
authenticated reverse proxy or a shared secret checked before `Submit`, plus
firewall rules for the chosen address.

Residual risks that ownframe cannot remove on its own include unsandboxed Go
templates, missing host allowlists in `Fetch`, trusted path APIs, and an
unauthenticated `Serve`. The current implementation does add transport
timeouts on `Serve`, `nosniff` and framing headers, MaxBytes parsing for
`/type`, finite coordinate checks, a 30 second fetch timeout, a 10 MiB fetch
body cap, crash files at mode `0600`, and a source size limit of
`MaxSourceSize` (8 MiB). The audit tracks the remaining items.


## Helper executables

Native helper tools such as `zenity`, `powershell.exe`, and `pwsh.exe` are discovered through `PATH` and their normal OS locations, including WSL `/mnt` paths. Treat the executable search path and installed helpers as trusted local environment components.

In hardened deployments, pin absolute dialog tool paths instead of relying on `PATH` lookup (for example, resolve `zenity` once at startup and reject unexpected locations). No behavior change by default: the file picker keeps `PATH` discovery unless the app pins paths itself.

## Devtools clipboard

The devtools dock has no setting for OS clipboard writes: clicking an Ops row copies its displayed text to the clipboard. Keep devtools closed or disabled when rendered data or coordinates are sensitive.


## Captured photo retention

The Telegram example registers each captured photo under a `photo-N` key and keeps it in `Page` image memory until the app calls `SetImage(id, nil)`. There is no automatic eviction. An app that handles sensitive images should remove the entry when the message or chat no longer needs it.

## Supply chain and CI (audit 3, GEN-1 to GEN-10)

Go modules: CI sets `GOFLAGS=-mod=readonly` and runs `go mod verify`
for both modules, so builds fail when the cache disagrees with `go.sum`.
Never set `GONOSUMDB`, `GOINSECURE`, or `GOFLAGS=-insecure`; `GOPRIVATE`
stays unset so the proxy and the checksum database verify every module.
`GOTOOLCHAIN=local` is set in CI and both `go.mod` files carry a
`toolchain go1.26.10` line, so the compiler never silently upgrades
itself; bump that line on purpose.

Tags are immutable: proxy.golang.org caches them forever, so a bad or
vulnerable release is fixed with a `retract` directive in `go.mod` plus a
new tag, never by re-tagging. `v*` tags and `master` are protected and
force-push is off.

Transitive parser surface: blinkless pulls goldmark, fpdf, star-tex,
minify, brotli, and canvas. `go mod why -m` shows they arrive only via
`blinkless/internal/...`; nothing in ownframe imports them directly, and
a `depguard` rule in `.golangci.yml` fails the gate on any direct
import, so untrusted input reaches them only through the reviewed
engine path. Re-run `go mod why -m` and `go mod graph` on every engine
bump and review what changed before committing.

Frontend: CI installs with `npm ci --ignore-scripts` (lockfile-only, no
install-time code execution) and runs `npm audit signatures`.

Android: Maven artifacts are trusted on TLS only. Run
`sh scripts/gen-gradle-verification.sh` from a provisioned machine and
commit `examples/telegram/android/gradle/verification-metadata.xml`;
unsigned or mismatched artifacts then fail the build. Re-run after any
AGP or dependency bump. See `examples/telegram/android/README.md`.

Dependencies on PRs: `dependency-review.yml` fails on high-severity
CVEs and non-allowlisted licenses, and the `go-licenses check` step
covers both modules. Every new module needs manual provenance review
(owner, pkg.go.dev, stars) before it merges; see CONTRIBUTING.md.

Workflows: actions are pinned by full commit SHA (the comment names the
source tag), top-level `permissions` is `contents: read`, and there is
no `pull_request_target` with a checkout. The `lint-actions` job runs
actionlint (pinned tarball, SHA checked) and zizmor (pinned version).

Secrets: `.gitleaks.toml` is the shared config for the pre-commit hook
(`.pre-commit-config.yaml`) and the CI `secrets` job, which scans full
history. Also turn on GitHub push protection under repository Settings
> Code security so secrets never reach history.

## Reporting vulnerabilities

Do not open a public issue for a suspected vulnerability. The process
is in `SECURITY.md`: private reporting through the repository Security
tab, response targets, and public GHSA disclosure that feeds the Go
vulnerability database. Maintainers enable the reporting entry point
under Settings > Code security.

## Release hardening checklist

Branch protection, signed commits and tags, SBOM and SLSA provenance,
reproducible builds, the `ownframe_debug` build tag, binary hardening,
and OS signing live in `supply-chain.md`. The packaging side (archive
layouts, signing steps, sandboxing, updater rules) is in
`packaging.md`.
