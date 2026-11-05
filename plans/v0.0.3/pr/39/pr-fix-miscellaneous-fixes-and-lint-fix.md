# PR: fix miscellaneous findings, lint gate, and security audit 3

Base: `master`. Head: `fix/miscellaneous-fixes-and-lint-fix`.
Commits on this branch range from replay and page fixes through `06653c4 security: close SECURITY_AUDIT_3 GEN-1 to GEN-50`.
Scale: 1304 files changed vs `origin/master`, about 19629 insertions and 4906 deletions. Go is 1229 of those files.

## Summary

This branch fixes the miscellaneous correctness and lint findings across the library and examples, wires the lint and CI gates that enforce them, and closes all 50 items in `plans/v0.0.3/security/SECURITY_AUDIT_3.md`. The pattern across the diff is hardening plus lint compliance, with a small number of real behavior fixes in animation fallback, chunk caching, image bounds, transition rebind, css-preview Play All timing, and error surfacing.

## Changes by area

Core page, state, render, frame. DOM caps of 256 depth and 20000 nodes, field maxlength clamped to 4096 with paste and programmatic writes capped at 16384 runes, `data:image/*` capped at 1 MiB with `file:` and `javascript:` rejected at resolve time, and open targets allowlisted to http, https, and mailto. `SetTick` is mutex guarded and `Page` is documented single owner. Failed redraws stay queued across batches. `SetText` follows textContent semantics. `ServeOptions{Perf,Debug}` are deprecated in favor of `InsecureAllowPprof`, `InsecureAllowDebugState`, and `InsecureAllowRemoteBind`, and non loopback bind now errors by default. `RunWithOptions` errors no longer embed the crash path, use `LastReportPath`. The old `state.setValueAtPath` shim is gone, use `setPath`.

Window, replay, bitmap. Additive composition with a missing keyframe now falls back to zero instead of the base paint, which fixes saturation. Chunk signatures fold group ID, mode, and isolate instead of the pointer address, so the cache hits across redraws. New bounded image cache with 64 entries and a 64 MiB pixel budget. Face cache keys on FNV hash plus length instead of an unsafe pointer. Masking prechecks size with `DecodeConfig` against a 16 MP cap. Transition rebind verifies ops are still owned by the box. Tile budget changed from an absolute 4 MP to content over 4x window area, so large pages keep zoom. Click through overlay ignores `Interactive` when `MousePassthrough` is set.

Web, fetch, IPC, OS integration. `requireLocalHost` rejects non loopback Host values. Mutating POST routes require Origin and fail closed without it. Per IP token bucket plus an 8 slot in flight semaphore, 64 KiB header cap, 256 KiB body cap. `/debug/state` and `/debug/pprof/*` exist only under the `ownframe_debug` build tag. Fetch uses a pinned transport with no proxy by default, TLS 1.2 minimum, 10s dial, TLS, and header timeouts, and a 1 MiB header cap, plus method and header allowlists, pre DNS and dial time SSRF checks, and redirect downgrade and private IP rejection with an `AllowPrivate(ctx)` escape hatch. New `secret.Secret` type redacts all format verbs and slog output. New `safego.Go` recovers into the crash report. Leak checks run in clipboard, fetch, ipc, and window, standing down under fuzz. Crash dir setup rejects dot segments and NUL, resolves symlinks, and pins 0700 and 0600. Clipboard race fixed with a new `Clear`. File picker WSL path rejects empty, NUL, and dot segments.

CI, supply chain, docs. New `ci.yml` with test, lint, sec, cross arch, secrets, lint actions, and reproducible jobs, with SHA pinned actions and read only permissions. Release gains SBOM, SLSA attestation, and keyless cosign signing. New CodeQL security extended, Semgrep custom rules, dependency review, and weekly govulncheck workflows. Golangci config grows to 19 linters including bodyclose, noctx, contextcheck, errorlint, forcetypeassert, gochecknoglobals, depguard, and forbidigo. `package.sh` links pie with fortify and relro. New scripts for Gradle verification metadata, reproducible double builds, secret scanning, and release signing. New `SECURITY.md`, `supply-chain.md`, `secrets.md`, `release-checklist.md`, and CONTRIBUTING guardrails for new module provenance.

Desktop cat, lg remote, telegram. Notify defaults to a 0600 Unix socket in the runtime dir, with TCP host port as explicit opt in that rejects wildcards and warns off loopback. Origin gate on notify and state, generic error strings, per IP token bucket and 4 in flight cap. Bidi controls stripped from message and source. LG remote websocket handshake has deadlines, scheme and host checks, bounded headers, 16 MiB frame and 4 MiB message caps, and wss now verifies with system roots at TLS 1.2 instead of skipping. TV pairing key is now `secret.Secret`. Telegram sets `FLAG_SECURE`, denies cleartext with a network security config, enables R8 and minify on release, caps photos at 32 MiB, and clamps insets.

Spotify player and music. Engine is now mutex guarded across want, update, readers, play, and pause, decoding unlocked and discarding stale generations. `Sound.Close` pauses and stops reading instead of closing the player. Main returns error once, pprof moves behind `InsecureAllowPprof`, and the per frame seek bar patch is gone so only time labels animate. Playback constants are named.

Spreadsheet, teams, download manager. Spreadsheet is const extraction and errcheck with identical values. Teams moves exec and query calls to context variants, uses errors.Is, scrubs CR and LF from log lines, rejects stacked semicolons per line, and drops the dead clone helper. Download manager grows job IDs from 8 to 16 bytes of hex, panics on CSPRNG failure instead of ignoring it, converts main to run returning error, and tightens fresh dir and lock permissions to 0750 and 0600.

Tetris, system monitor, live log viewer, snipping tool. Tetris surfaces score apply errors that were swallowed. Live log viewer logs instead of `log.Fatal` so defers run. Error wrapping moves from `%v` to `%w` where matching matters. The rest is const extraction and checked closes.

Misc examples. Css preview Play All is the one real fix here. Samples take origin at bind, offscreen rows queue through patch grid, and readout goes through a setter, so short animations no longer expire during patch. Login gains `Logout` that clears fields, undo, redo, view, and the OS clipboard, stores the password as `secret.Secret`, and warns that web demo creds cross loopback in plaintext. Fetch tests use `AllowPrivate`. Departure board follows the pprof rename. Pet care passes ctx into click, joins close errors, and panics on `rand.Read` failure instead of falling back to time. Everything else is lint only.

Plans and audits. Audit 1 with 60 findings and audit 2 with 43 findings are recorded closed. Audit 3 replaced the draft with the 50 item GEN list and all 50 are now checked. The css preview optimization plan collapses to fixed after merge plus 5 still open items. AGENTS.md and README record the size gate and lint gate.

## Breaking changes and migration

- `ServeOptions{Perf,Debug}` still compile as deprecated. Move to `InsecureAllowPprof` and `InsecureAllowDebugState`.
- Non loopback Serve bind now errors. Set `InsecureAllowRemoteBind` to keep it.
- Desktop cat notify defaults to a Unix socket. Pass `-notify-addr host:port` to keep TCP.
- Custom hostname previews need the configured bind name to pass the Host allowlist.
- Clients that POST without Origin now get 403 on mutating routes.
- Intranet and metadata URL fetches need `AllowPrivate` or the proxy opt in.
- Debug endpoints 404 without the `ownframe_debug` tag.
- Download manager job IDs are longer and CSPRNG failure now panics.
- Fresh data dirs use 0750 and lock files 0600.
- Live log viewer returns 0 instead of exiting 1 on Run failure.
- Crash path no longer comes from the Run error, read `LastReportPath`.

## Test plan

- `go vet -p 1 ./... ./examples/...` passes.
- `go test -p 1 -run Sec ./... ./examples/...` passes across all packages.
- `go test -count=1` passes for `internal/page`, `internal/web`, `internal/fetch`, `internal/secret`, `internal/safego`, and `internal/replay`.
- `bash scripts/check-file-size.sh` passes with zero over limit files.
- Full `go test -p 1 ./... ./examples/...` was started but exceeded the 120s tool timeout, so it is not claimed green here.
- `gitleaks detect` over 451 commits reports no leaks. Actionlint and zizmor report no findings on the touched workflows.

## Risks and follow ups

- Previously valid large templates, long inputs, big images, and odd links now error by design.
- Unix socket default and strict Origin break old TCP clients and bare curl without an Origin header.
- TV TLS verification may fail against self signed TV chains.
- `singleStatement` rejects semicolons inside string literals and fails seed at startup.
- Longer job IDs break persisted `job-<16hex>` assumptions. Group shared data dirs may need permission work.
- `PHASES.md` and `RENDERING-FINDINGS.md` are stale relative to the replay and motion fixes.
- Audit file and line proofs will rot as code moves. The examples module still cannot be scanned alone by govulncheck.
- The css preview plan still lists 5 open items for release.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
|-----------|-------|------------|-----------|
| `.go` | 1229 | 13158 | 4444 |
| `.md` | 32 | 1484 | 176 |
| `.yml` | 8 | 674 | 0 |
| `.json` | 2 | 3712 | 147 |
| `.sh` | 6 | 228 | 34 |
| `.mod` | 5 | 42 | 14 |
| `.sum` | 3 | 49 | 32 |
| `.jsx` | 3 | 66 | 18 |
| `.js}` | 2 | 50 | 34 |
| `.js` | 1 | 41 | 0 |
| `.xml` | 2 | 7 | 0 |
| `.css` | 1 | 5 | 2 |
| `.gradle` | 1 | 8 | 0 |
| `.java` | 1 | 5 | 0 |
| `.html` | 1 | 2 | 2 |
| Total files 1304, insertions 19629, deletions 4906 |

Related: `plans/v0.0.3/security/SECURITY_AUDIT_3.md`, branch `fix/miscellaneous-fixes-and-lint-fix`, closure commit `06653c4`.
