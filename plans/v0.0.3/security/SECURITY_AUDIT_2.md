# Security audit 2 - ownframe

Scope: full repository (root library, `internal/`, `examples/`, `browser/`, `scripts/`, `frontend/`), excluding `temp/` and `.git`.
Method: 10 parallel domain audits (injection, cryptography/secrets, web/HTTP, authentication/authorization, concurrency, dependencies/supply chain, filesystem, memory safety, logging/errors, SSRF/IPC/clipboard/drag-drop) plus manual file reads. Read-only. No code changed. `go test -race`, `gosec`, and `govulncheck` were not re-run in this pass.
Findings are deduplicated across domains; overlapping IDs are merged and noted.

Severity scale (DREAD): Critical 8-10, High 6-7.9, Medium 4-5.9, Low 1-3.9. No Critical findings were identified.

## Summary

| Severity | Count |
| --- | --- |
| Critical | 0 |
| High | 3 |
| Medium | 16 |
| Low | 24 |

---

## Phase 1 - High (fix immediately)

- [x] **MEM-5** - High - `internal/masking/image_decode.go:19-35`, `internal/replay/image_cache.go:54`
  - Issue: `DecodeImage` runs `image.Decode` with no width, height, or pixel cap, then uploads via `ebiten.NewImageFromImage`. The image cache holds 64 entries with no total pixel budget. A large PNG or JPEG can OOM the process or panic past `MaxImageSize`. The emoji path resolves through the same decoder.
  - Proof: `src, _, e := image.Decode(bytes.NewReader(data))` followed only by `Bounds().Empty()` check; `imageCacheLimit` at `image_cache.go:14` with no pixel accounting, unlike raster, text, and chunk caches.
  - Solution: Reject `Dx` or `Dy` past a cap before decode or upload, mirroring `maxMaskPixels` in `internal/masking/limits.go:3`. Add pixel accounting to the image cache or reuse the existing budget pattern.

- [x] **RACE-3** - High (conditional, single-owner rule holds today) - `internal/page/page_frame.go:7`, `internal/frame/anim_step.go:58-70`, `internal/replay/draw_rect.go:21`
  - Issue: Tick patches retained ops in place (`op.X`, `op.Y`, `Alpha`, `SetXform`) while `DrawRect` reads the same ops. Safe today because tick and Draw run on one Ebiten goroutine and partial replay forces full replay when ticking. Any off-loop caller breaks it: audio callback, mobile inset callback, HTTP POST, background resolve.
  - Proof: `MarkRect` and `Anim.patch` mutate in place; `drawReplayPartial` full-replay guard at `internal/window/replay_partial.go:29` and `replay_partial_state.go:9` assumes same-goroutine use. `Anim.Bind` reads plain `uint64` generation at `internal/frame/anim_gen.go:22`.
  - Solution: Never share a Page between `Run` and `Serve`. Document single-owner rule. Keep all Tick, MarkRect, and Anim calls on the page loop. Add a debug assert or owner check if cheap.

- [x] **DEP-6** - High (no enforcement) - `.github/` missing, `AGENTS.md`, `Makefile:42-45`
  - Issue: CI directory does not exist. `AGENTS.md` claims GitHub CI runs `make test` and `make lint`, but `git ls-tree` shows no workflows and `ls` shows no `.github`. No gate enforces tests, lint, Dependabot, `govulncheck`, or `osv-scanner`.
  - Proof: `ls /home/chinmay/ChinmayPersonalProjects/ownframe/.github` returns no such file. `git log --all -- .github` is empty.
  - Solution: Recreate workflows that run `make test` and `make lint`, or fix `AGENTS.md`. Add scheduled `govulncheck` and dependency update checks.

---

## Phase 2 - Medium (next sprint)

- [x] **DEP-4** - Medium (still open from audit 1) - `examples/go.mod:6-7`
  - Issue: Examples pin stale `ownframe v0.0.2` and `blinkless v0.0.2` while root is at `v0.0.3-0.20261021155643-c1054ff46202`. Standalone `examples/` builds fail. Only `go.work` masks it.
  - Proof: `examples/go.mod:6-7` vs `go.mod:6`. Matches open item at `SECURITY_AUDIT.md:137-140`.
  - Solution: After publishing the matching release, bump both pins to the new tags. Drop workspace reliance for external builds.

- [x] **DEP-7** - Medium - `go.work.sum:8-9`, `go.mod:6`, `go.sum:31-32`
  - Issue: `go.work.sum` still pins prior blinkless pseudo `b26de136c260` after the engine bump to `c1054ff46202`. `go.sum` has the new hash. Workspace sum is stale.
  - Proof: `go.work.sum:8-9` vs `go.sum:31-32`.
  - Solution: Regenerate `go.work.sum` after the engine bump. Verify with `go work sync`.

- [x] **DEP-8** - Medium - `browser/go.mod:3`, `scripts/go.mod:3`, `frontend/go.mod:3`, `go.work:3-6`
  - Issue: Orphan modules still target `go 1.26.4` while root and examples are on `go 1.26.10`. They are not in `go.work`, so `make test`, `build`, and `lint` never vet them. No `toolchain` directive forces upgrade.
  - Proof: Module headers vs `go.mod:3` and `examples/go.mod:3`.
  - Solution: Align all modules to `go 1.26.10` or add them to the workspace. Add a `toolchain` line if the build needs a floor.

- [x] **DEP-9** - Medium - `Makefile:42-45`, `scripts/android*.sh`
  - Issue: No `govulncheck` gate exists. No Makefile target, no tool dep, no CI config. The only run is historic in `SECURITY_AUDIT.md:22-30`. Tool installs use unpinned `go install golangci-lint@v2.15.0` and `ebitenmobile@v2.10.4`, which mutate `GOPATH/bin` with no hash or provenance.
  - Proof: `grep govulncheck Makefile` empty. `grep ^tool go.mod` empty. `Makefile:12` version var with no hash.
  - Solution: Add `go get -tool golang.org/x/vuln/cmd/govulncheck` and a `make vuln` target. Run it in CI. Pin tool hashes or move installs to tool directives.

- [x] **DEP-10** - Medium - `go.mod:14-18,28,31-32`
  - Issue: Ancient indirects via Ebiten and blinkless with no in-repo upgrade path: freetype-go 2016, graphics-go 2016, xgb 2021, xgbutil 2019, poly2tri-go 2017, freetype 2017, rasterx 2022, scanx 2019. Pinned hashes limit tampering but code is unmaintained.
  - Proof: Version dates in `go.mod` indirect block.
  - Solution: Track upstream Ebiten and blinkless bumps. No direct fix in this repo.

- [x] **FETCH-1** - Medium - `internal/fetch/ssrf_check.go:48`, `internal/fetch/do.go:39`
  - Issue: DNS rebinding TOCTOU. Host is validated with `LookupIPAddr`, then `client.Do` re-resolves via the default transport. An attacker domain can answer the check with a public IP and the real fetch with loopback, metadata, or private IP. Each redirect hop re-checks but has the same race.
  - Proof: No custom `DialContext` or pinning in `internal/fetch`. Only hit for `Dial|Transport` is the `client.Do` line.
  - Solution: Resolve once, then dial the validated IP with a custom `DialContext` while verifying `req.Host`. Or capture validated addrs per request.

- [x] **RACE-4** - Medium - `internal/page/page_tick.go:9`, `internal/page/page_ticking.go:6`
  - Issue: `SetTick` and `Tick` read and write a plain func pointer with no lock or atomic. Game loop calls Tick in `internal/window/update.go:63` while web handlers, tests, or click paths can call SetTick from another goroutine.
  - Proof: `p.tick = fn` store vs `if p.tick != nil` load, no mutex.
  - Solution: Serialize SetTick on the page loop, or guard with mutex or atomic pointer.

- [x] **RACE-5** - Medium - `internal/page/page_image.go:23,17,76`, `internal/web/page_http.go:15,39`
  - Issue: `Display` and `Image` return raw pointers. `PNGWithContext` lazily mutates the `p.png` cache. Web `server.mu` serializes HTTP handlers but does not serialize against the Ebiten loop when one Page is shared between window and web.
  - Proof: Plain field access with no lock. Same root cause as RACE-3.
  - Solution: Same as RACE-3. Do not share a Page across transports. Copy or lock if sharing is ever required.

- [x] **RACE-6** - Medium - `internal/replay/group_plan.go:48-71`
  - Issue: Plan cache miss drops the lock, walks `d.Ops` in `buildPlan`, then re-locks to insert. A concurrent Tick patch or Redraw replacement tears the read. Double miss also builds twice and evicts under lock.
  - Proof: Hit path holds `plans.mu`. Miss path builds outside the lock over mutable display.
  - Solution: Keep plan build on the draw path only. Clone ops under lock or skip cache on ticking frames.

- [x] **RACE-7** - Medium - `examples/music/engine.go:30-37`, `examples/music/engine_want.go:8-12`, `examples/music/engine_read.go:6-24`
  - Issue: `Engine` guards only `latest` with `latestMu`. `gen`, `applied`, `loading`, `clip`, `voice`, `wantPlay`, `volume`, and `err` are plain. `Want` from click races `Update` and `Position` from Tick.
  - Proof: `e.gen++` and `e.loading = true` on click goroutine vs `Update` on game loop. `voice` pointer swap with no lock.
  - Solution: Put Want, Update, and Position on the page loop, or add a mutex around Engine state.

- [x] **WEB-6** - Medium if bound to LAN, Low on loopback - `internal/web/serve_build.go:51-59`, `examples/desktop-cat/notify.go:26-30`, `examples/login/main.go:18`
  - Issue: No TLS anywhere. Plain HTTP serves the login demo, `frame.png`, click and type traffic, and cat notify. Loopback default limits sniffing, but LAN binds expose email, password, frames, and control traffic to passive capture. HSTS is correctly absent without TLS.
  - Proof: Zero hits for `ListenAndServeTLS` or `ServeTLS` repo-wide.
  - Solution: Keep loopback-only default. If LAN use is needed, terminate TLS in front or add TLS config plus HSTS. Document that LAN bind sends demo creds in clear.

- [x] **WEB-7 / AUTH-6** - Medium if bound off loopback, Low at default - `internal/web/http_security.go:36-55`, `examples/desktop-cat/notify_addr.go:22-45`, `examples/desktop-cat/cat/http_sec.go:19-23`
  - Issue: Non-loopback binds pass with only a log warning. Only Origin check guards mutating routes, and curl or non-browser peers forge Origin trivially. `checkOrigin` in desktop-cat fails open on missing Origin, unlike `requireOrigin` in `internal/web/http_origin.go:36`.
  - Proof: `validateBindAddr` rejects wildcard and empty but passes `192.168.x.x`. `checkNotifyAddr` warns with `needs auth` but still binds. `curl -H Origin:...` forges the check.
  - Solution: Fail closed on missing Origin for POST notify, matching `requireOrigin`. Reject non-loopback outright, or require a bearer token before binding off loopback.

- [x] **AUTH-7** - Medium when opted in, secure by default - `internal/web/serve_build.go:34-47`, `internal/web/debug_http.go:28-53`, `serve.go:9-34`
  - Issue: `GET /debug/state` and `GET /debug/pprof/*` have no per-request auth once enabled. State exposes width, height, generation, boxes with text and IDs, stats, op counts, and runtime snapshot. Plain `Serve` exposes nothing, which is correct.
  - Proof: Routes register when `Perf || Debug`. No token or `requireOrigin` on the GET.
  - Solution: Keep disabled by default. Bind debug mode to loopback only, or add a bearer token. Note in `ServeOptions` docs that debug exposes layout and field text.

- [x] **MEM-6** - Medium - `internal/replay/xform_scratch.go:25-27`, `internal/replay/xform_scratch_draw.go:18`
  - Issue: Transformed-shape scratch raster checks each side against `ebiten.MaxImageSize` only. A 16000 by 16000 box passes and allocates about 1 GiB. The replayable path has a 4M CSS px cap at `render/replayable_group.go:63-67`. The scratch fallback has no equivalent.
  - Proof: `oversizedImage` side check only, then `ebiten.NewImage`.
  - Solution: Add total pixel reject in `xformScratch`, mirroring `rasterBox` size check.

- [x] **MEM-7** - Medium - `examples/desktop-cat/nowplaying/string_windows.go:30-36`
  - Issue: Windows-only COM string copy trusts `length` from `WindowsGetStringRawBuffer` for `make([]uint16, length)` and raw copy. Source is the OS media session, trusted in practice, but a buggy provider gives OOM or over-read.
  - Proof: Unbounded `length` used for alloc and `RtlMoveMemory` byte count.
  - Solution: Cap length before alloc, for example reject past 4K chars. Media strings are small.

- [x] **FS-3 / AUTH-8** - Medium (demo boundary) - `examples/login/login/login_state.go:9-25`, `examples/input-lab/inputlab/status.go:58-62`, `examples/pet-care/app/auth.go:12`
  - Issue: Hardcoded demo creds with plaintext `==`, no hashing, no constant time compare, no session. Each site documents demo-only, but copy to production would be High. Passwords stay in Page memory plus undo stack.
  - Proof: `FormValue("email") == "secret"`. Zero hits for `bcrypt`, `argon2`, `subtle`, or session libs outside audit docs.
  - Solution: Keep demo-only. Add DO NOT COPY banner. Add logout that clears FormValue and undo stack. Production must verify server side over TLS with slow hash.

---

## Phase 3 - Low (opportunistic)

- [x] **INJ-5** - Low - `internal/page/page_new.go:8-12`, `internal/page/page_nav.go:16-32`, `internal/page/page_nav_route.go:8-20`, `internal/page/page_draw_stages.go:24`
  - Issue: `Page.Load`, `Route`, and `New` execute unsandboxed `html/template`. Documented as trusted author content. Risk only if a future caller forwards remote or user HTML into Load or Route. Attacker template could read app data or burn CPU.
  - Proof: `template.New("page").Parse(source)` then `p.tpl.Execute(&body, p.data)`. No scope caller forwards user input today. Telegram, spotify-player, and wispr flows use fixed IDs and allowlists.
  - Solution: Keep allowlisting before Load and Route. Keep `MaxSourceSize` and `ErrEmptyHTML` guards.

- [x] **INJ-6** - Low - `examples/desktop-cat/cat/http.go:23-28`, `examples/desktop-cat/cat/http_sec.go:17-23`, `examples/desktop-cat/cat/inbox.go:30-38`
  - Issue: POST notify accepts missing Origin. Impact is capped to bubble text spoofing: expression allowlist, 1 to 160 runes message, 32 runes source, strict JSON, `MaxBytesReader`, no script engine in ownframe, message rendered via escaped `html/template`.
  - Proof: `checkOrigin` true on empty Origin vs `requireOrigin` fail-closed in `internal/web`.
  - Solution: Same as WEB-7. Fail closed or gate curl bypass behind explicit flag. Add auth before exposing past loopback.

- [x] **CRYPTO-4** - Low - `examples/download-manager/scheduler/id.go:11`
  - Issue: Job IDs use 64-bit `rand` output (`job-<16 hex>`). Correct 128-bit pattern exists in tetris `game/id.go:12` and pet-care `pet_actions.go:56`, both with panic on failure. Local job identifier only, no network or capability use. Collision chance is negligible for demo but weak if ever used as unguessable token.
  - Proof: `var b [8]byte; rand.Read` vs `[16]byte` elsewhere.
  - Solution: Use 16 bytes for consistency.

- [x] **WEB-8** - Low - `examples/desktop-cat/notify.go:30`
  - Issue: Notify server lacks `IdleTimeout`. Preview server sets all four timeouts at `internal/web/serve_options.go:20-25`. Loopback default makes this local DoS at worst.
  - Proof: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `MaxHeaderBytes` set, no `IdleTimeout`.
  - Solution: Add `IdleTimeout`, for example 60s, mirroring serve options.

- [x] **WEB-9** - Low - `internal/web/http_security.go:12-18`, `examples/desktop-cat/cat/http_sec.go:9-15`, `browser/index.html:13-17`
  - Issue: Only `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY` are set. No CSP, Referrer-Policy, Permissions-Policy, or COOP and CORP. Shell HTML is escaped and redirects are fixed-target, so impact is limited today. Shell page uses inline onclick and script handlers, so future CSP needs nonces or hashes.
  - Proof: Header sets in both servers. No `Access-Control-Allow-*` anywhere, which is correct.
  - Solution: Add `Referrer-Policy: no-referrer` and Permissions-Policy. Plan nonce-based CSP before the shell page reflects richer content.

- [x] **AUTH-9** - Low - `examples/login/login/login_state.go:27-53`, `examples/login/login/login_history.go:5-29`
  - Issue: Plaintext password retained in Page store plus up to 64 undo snapshots. Display masking is paint-only. No zeroing, no scoping, no logout clearing.
  - Proof: `snapshot` copies both fields, `push` appends to capped undo, `restore` writes back.
  - Solution: Add logout that clears FormValue and undo and redo stacks. Never log or transmit FormValue.

- [x] **AUTH-10** - Low - `internal/page/page_nav.go:19-61`, `internal/page/page_nav_route.go:12-21`, `examples/spotify-player/player/controls.go:10-78`, `examples/telegram/telegram/actions.go:12-58`, `examples/wispr-flow-dashboard/app/actions.go:14-71`
  - Issue: Nav, Route, Load, player controls, telegram actions, and dashboard nav have no authz check. Correct for single-user local demos with no privileged operation. Telegram gift is double-gated by template `CanGift` plus server check in `gift.go:11`.
  - Proof: Action switches keyed on fixed IDs and allowlist maps. No JWT, session, or cookie libs in tree.
  - Solution: No change for demos. Require server-side permission checks if any screen gains accounts or tenants.

- [x] **RACE-8** - Low - `internal/render/cache.go:12-16`, `internal/render/cache_relayout.go:12`
  - Issue: `Styled` hands out unlocked `*css.Document`. Relayout methods hold `c.mu` plus `engineMu`, but the returned doc has no generation guard. Holder keeping it across pollReload or hover Relayout reads a replaced doc.
  - Proof: Unlock before return. Doc comment says caller must not retain across Relayout.
  - Solution: Keep the contract, or add generation guard or copy. Fragile but not a live race in current single-loop use.

- [x] **RACE-9** - Low - `internal/page/page_patch.go:34-47`, `internal/page/page_draw.go:24`
  - Issue: `sync.Pool` slice aliasing relies on serial Redraw. Overlapping Redraw from window sync and web reload on one Page could hand out the same backing array twice.
  - Proof: `rentOps` and `freeOps` with `buf[:0]` reset. `redrawBatch` and `afterRedraw` guards assume serial loop.
  - Solution: Same as RACE-3. One owner per Page. No concurrent Redraw.

- [x] **RACE-10** - Low - `internal/clipboard/clipboard.go:48-57`
  - Issue: `Read` split-lock TOCTOU. Unlocks between mode check and value read, so `UseMemory(false)` flipping in between returns stale memory in OS mode or misses fresh OS text.
  - Proof: `only := memory` under lock, unlock, then second lock for `mem`.
  - Solution: Hold the lock across both steps, or recheck mode under the second lock. `Write` already uses one section.

- [x] **RACE-11** - Low - `internal/clipboard/x11_read.go:33`, `internal/clipboard/x11_serve.go:22`
  - Issue: Each SelectionRequest spawns an unbounded goroutine. Core handoff is locked correctly, so this is availability only. A flooding requester causes goroutine burst contending `x.mu`.
  - Proof: `go x.onRequest(ev)` per 32-byte event with no semaphore.
  - Solution: Bound with semaphore or worker pool.

- [x] **FS-4** - Low - `internal/filepick/pick_linux.go:55,70`, `internal/filepick/pick_wsl_shell.go:19-31`, `internal/filepick/pick_wsl_path.go:41-47`
  - Issue: Dialog tool resolution via PATH and `/mnt` glob. Attacker with write to early PATH dir or mount plants fake `zenity` or `powershell.exe`. Requires already-compromised host. No shell involved, fixed argv vectors.
  - Proof: `exec.LookPath` per tool, `filepath.Glob("/mnt/*/Windows/...")`, fixed absolute `wslpath` fallback.
  - Solution: No change required for default use. Pin absolute tool paths in hardened deployments.

- [x] **FS-5** - Low - `internal/crash/validate.go:30-81`, `internal/crash/crash.go:28`, `internal/crash/create.go:23`
  - Issue: `SetDir` validation has no symlink canonicalization. Pre-created symlink at config or temp crash dir redirects reports there. File overwrite case is neutralized by timestamp names plus `O_EXCL`. Needs same-user FS write.
  - Proof: Rejects `..` pre- and post-Clean, controls, non-absolute, but no `EvalSymlinks` recheck. `MkdirAll` follows dir symlinks.
  - Solution: Add `EvalSymlinks` plus prefix recheck, or open with `O_NOFOLLOW`. Keep `0700` dir and `0600` file modes.

- [x] **FS-6** - Info - `internal/page/page_source.go:48-72`, `internal/page/page_reload.go:57-95`, `internal/page/page_watch_api.go:26-31`
  - Issue: `Config.File` and `ThemeFile` plus poll reload follow symlinks and use mtime plus size for change check. A same-size and same-mtime swap is missed until stat differs. Trusted author boundary, documented at `internal/page/page_new.go:8-11`. TOCTOU growth is caught by post-read length check against 8 MiB.
  - Proof: `Stat`, 8 MiB check, `ReadFile`, re-length check, `sameStat` early return.
  - Solution: No change while input stays author-trusted. Use `os.Root` if untrusted paths are ever accepted. Go 1.26.10 has `os.Root` available. Nothing in `internal/` uses it today.

- [x] **MEM-8** - Low - `internal/frame/anim.go:59`, `internal/frame/anim_step.go:13,48-54`, `internal/frame/anim_kinds.go:15-33`
  - Issue: `Anim.Active` is nil-safe but `Step` and `Tick` deref nil. Op helpers deref every op with no nil check. Safe today because `Fills` and `Text` never return nil, but a caller-built slice with nil panics. NaN coords would give junk damage rect only, no OOB.
  - Proof: `stale()` deref vs nil-safe `Active`. `TranslateTo`, `FadeTo`, `SetAlpha`, `setScale`, `localBox` with no nil skip.
  - Solution: Nil-guard Step and Tick to match Active. Skip nil ops defensively.

- [x] **MEM-9** - Low - `examples/desktop-cat/nowplaying/com_windows.go:12,21-25`
  - Issue: COM vtable dispatch has no index bound. Table is `[40]uintptr`, `call` indexes directly then jumps. All current sites use constants 0, 2, 6, 7, 8, and 9, so no live OOB. Future caller with index past 40 is arbitrary call.
  - Proof: `o.table[index]` with no range check. `release` nil-checks but `call`, `get`, `query`, and `text` do not.
  - Solution: Check `0 <= index < len(table)` and nil receiver before dispatch.

- [x] **FETCH-2** - Low - `internal/fetch/do.go:17-37`
  - Issue: Caller-controlled headers and method. No denylist means in-process caller can set `Cookie`, `Authorization`, or `Host`. Jar stays nil so no automatic storage, but explicit exfil to a public host is possible. Empty method becomes GET, otherwise passed verbatim including TRACE and CONNECT. Only response has caps.
  - Proof: Header loop with `req.Header.Set`. No method allowlist.
  - Solution: If XHR is ever exposed past trusted callers, allowlist methods and block sensitive headers.

- [x] **IPC-1** - Low - `internal/ipc/send.go:10-22`, `internal/ipc/handle.go:11-29`, `internal/ipc/request.go:40-71`, `ipc.go:15-42`
  - Issue: IPC is global namespace with no auth, origin check, or payload validation. Any linked code can Listen or Handle on victim channel and spoof, eavesdrop, or cancel. Last writer wins on Handle. Acceptable only because in-process. Panic isolation, empty channel, and flood cap at 256 listeners are correct.
  - Proof: No identity or size checks. Docs state no socket and no isolation boundary.
  - Solution: Keep secrets out of payloads. Prefix names per owner. Do not bridge to Serve or notify port without token and allowlist.

- [x] **CLIP-1** - Low - `internal/clipboard/clipboard.go:29-43`, `internal/clipboard/clipboard_windows.go:33-63`, `internal/window/chord.go:7-12`
  - Issue: Sensitive clipboard data lingers. Write persists indefinitely. Windows buffer never zeroed. In-memory string cannot be wiped. Password copy via Ctrl-C stays after app exit. Serve exposes no clipboard route, which is correct. Tests use `UseMemory`, which is correct.
  - Proof: No `Clear` API. `grep Clear|Zero|wipe` hits only constants and tests.
  - Solution: Add `clipboard.Clear`, auto-clear password fields after timeout, consider byte slice plus zeroing for mem store.

- [x] **DROP-1** - Low - `internal/host/drop.go:5-14`, `internal/window/drop.go:55-79`, `internal/page/page_drop.go:14-26`, `examples/drop/drop/drop_handler.go:12-17`
  - Issue: Absolute host path stored verbatim and rendered, with no file count or byte cap. Bulk drop causes memory and CPU load before Redraw. Late `Read` after frame fails closed, which is correct. No dir recursion, which is correct. Callers must copy `Read` bytes during the call.
  - Proof: `onDrop` stores `file.Path` into view. No limit on files, size, or `Read` bytes in window or page drop paths.
  - Solution: Cap files and per-file bytes, stream instead of `ReadFile` for large drops, display base name only, redact Path from logs and debug state.

- [x] **MOBILE-1** - Low - `examples/telegram/mobile/api.go:30-41`, `examples/telegram/telegram/photo.go:16-92`, `examples/telegram/telegram/insets.go:5-31`
  - Issue: `SetPhoto` has no byte length cap before enqueue. Dimension cap at 4096 via `DecodeConfig` checks pixels, not compressed length. `SetInsets` applies raw int64 with no clamp, so negative or huge values corrupt pinned bar layout. Source is system bars via MainActivity, trusted, but exported binding has no defense in depth. Queue drop on full and attach-ID early return are correct.
  - Proof: `photos=make(chan []byte,2)` with arbitrary `len(data)`. Insets stored then applied without `clamp`.
  - Solution: Reject `len(data)` past 32 MiB or stream with limit. Clamp insets to 0 to 2048. Move attach-ID check before decode.

- [x] **DEP-11** - Low - `examples/telegram/android/app/build.gradle:5,10`, `build.gradle:9`, `gradle/wrapper/gradle-wrapper.properties:3`
  - Issue: Android behind current SDK and AGP. Compile and target 34, AGP 8.5.1, Gradle 8.7, minSdk 23. No wrapper checksum stanza. Supply chain trust is TLS plus wrapper only, with Maven Central and Google repos and no lockfile.
  - Proof: Gradle files listed above. Only jar plus properties in wrapper dir.
  - Solution: Plan bump to SDK 35 or 36 and AGP 8.7 or later. Add wrapper validation.

- [x] **DEP-12** - Low - `frontend/package.json:14-36`, `frontend/eslint.config.js:26`
  - Issue: Overrides entry forces `eslint-plugin-react` peer to installed ESLint 10, with version pin workaround in config. Lockfile has no top-level overrides key, so the override is invisible in lock diff. Frontend deps are otherwise sane: React 19.3.0 locked, Vite 7.3.7, highlight.js pinned.
  - Proof: `overrides` in package.json vs missing key in `package-lock.json`. Locked plugin still declares old peer range.
  - Solution: Reverify after any npm install or ESLint 10 bump. Remove workaround when plugin supports ESLint 10.

- [x] **DEP-13** - Low - `examples/go.mod:11,49-51`
  - Issue: Large examples-only TCB: `modernc.org/sqlite v1.56.0` plus libc, memory, mathutil, uuid, go-mp3, pprof. Pure Go but large. Not shipped in the library.
  - Proof: Examples module requirements.
  - Solution: No library change. Track updates on the examples schedule.

---

## Notes

- Overlapping findings merged: WEB-7=AUTH-6; WEB-2 and LOG-2 from audit 1 stay fixed; INJ-4 from audit 1 stays fixed for direct SSRF, FETCH-1 is the residual rebinding leg; AUTH-3 from audit 1 stays partly open as AUTH-6 here because missing-Origin bypass remains.
- Stale claims in audit 1 corrected: `http.Error(w, err.Error(), 500)` no longer present. Current handlers return generic 500 with sanitized server log. `sameOrigin` now compares scheme plus host. Click query has count and length caps plus `NaN` and `Inf` reject. `MaxHeaderBytes` is set. Crash perms and `SetDir` traversal checks verified.
- Pass items with no finding: SQL injection, command injection in scope, `text/template` use, `template.HTML` bypass except two internally computed ints, stored and reflected XSS sinks escaped with no JS engine, `GET /frame.png` and safe GETs, method enforcement, CORS defaults, clipboard route absence, crash report contents and perms, `OWNFRAME_FILEPICK_DEBUG` gating, IPC panic isolation, render and replay cache locks, atomic handoffs, tile grid ownership, `math/rand` non-security uses annotated, SHA1 only for RFC 6455, TLS `InsecureSkipVerify` gone, `replace` directives gone, browser and android scripts with no network fetch.
- Recommended verification after fixes: `go test -race ./...`, `gosec ./...`, `govulncheck ./...` in both modules, plus fuzzing of image, emoji, drop, and fetch redirect paths.
