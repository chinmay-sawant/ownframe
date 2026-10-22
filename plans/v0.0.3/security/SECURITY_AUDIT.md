# Security audit — ownframe

Scope: full repository (root library, `internal/`, `examples/`, `cmd/`), excluding `temp/` and `.git`.  
Method: 10 parallel domain audits (injection, cryptography/secrets, web/HTTP, authentication/authorization, concurrency, filesystem, memory safety, logging/errors, cookies/IPC, dependencies) plus `govulncheck`.  
Findings are deduplicated across domains; overlapping IDs are merged and noted.

Severity scale (DREAD): Critical 8–10, High 6–7.9, Medium 4–5.9, Low 1–3.9. No Critical findings were identified.

## Summary

| Severity | Count |
| --- | --- |
| Critical | 0 |
| High | 8 |
| Medium | 24 |
| Low | 28 |

---

## Phase 1 — High (fix immediately)

- [x] **DEP-1** — High — `go.mod:3`, `examples/go.mod:3`
  - Issue: Both modules target Go 1.26.4; `govulncheck -scan=module` reports 24 stdlib/x/net vulnerabilities fixed in 1.26.10.
  - Proof: `govulncheck` output: "Your code may be affected by 24 vulnerabilities"; fixed versions `stdlib@go1.26.10`, `golang.org/x/net@v0.60.0`.
  - Solution: Install Go 1.26.10+, set `go 1.26.10` (or `toolchain go1.26.10`) in both `go.mod` files, rebuild.

- [x] **DEP-2** — High — `go.mod:37`
  - Issue: `golang.org/x/net` declared v0.55.0 but resolved v0.58.0; called CVEs fixed only in v0.60.0. Not tidied.
  - Proof: `govulncheck` traces through `internal/fetch/do.go` and `internal/web/serve_build.go`; fix version `golang.org/x/net@v0.60.0`.
  - Solution: Bump indirect requirement to `golang.org/x/net v0.60.0`, run `go mod tidy`.

- [x] **DEP-3** — High (supply chain) — `go.mod:46`
  - Issue: `replace github.com/chinmay-sawant/blinkless => ../blinkless` resolves the engine from an unversioned sibling checkout, bypassing checksum verification and breaking standalone builds (contradicts `AGENTS.md`).
  - Proof: `go.mod` line 46; `../blinkless` is a separate git repo, not a submodule; `go.sum` cannot pin it.
  - Solution: Remove the replace, require a published `blinkless` version matching the code used, or vendor it; update docs.

- [x] **INJ-1** — High — `examples/snipping-tool/snip/capture_win.go:26-36`
  - Issue: PowerShell script interpolates a caller-supplied path into a single-quoted string without escaping `'`.
  - Proof: `fmt.Sprintf(... -Path '%s' ..., path)` with no `''` escaping.
  - Solution: Escape `'` as `''` in `path`, or pass the path via env/stdin instead of shell string interpolation.

- [x] **INJ-4** — High — `internal/fetch/scheme.go:48, 60`
  - Issue: SSRF: `fetch.Do` accepts any http/https URL and follows redirects without blocking loopback/private/link-local/metadata addresses.
  - Proof: `fetch.Do(ctx, "GET", "http://127.0.0.1:8080/", nil, nil)` and `http://169.254.169.254/latest/meta-data/` pass `checkURL`/`redirect`; redirects to such addresses also accepted.
  - Solution: Validate resolved IPs in `checkURL` and `redirect`; reject loopback, private, link-local, multicast, metadata ranges; re-check after DNS resolution to prevent rebinding.

- [x] **CRYPTO-1** — High — `examples/lg-remote/ws/tcp.go:19`
  - Issue: TLS dial sets `InsecureSkipVerify: true`.
  - Proof: `tls.Config{InsecureSkipVerify: true, ServerName: ...}`.
  - Impact: MITM of the WebSocket connection to the LG TV.
  - Solution: Remove `InsecureSkipVerify`, keep `ServerName`, set `MinVersion: tls.VersionTLS12`.

- [x] **AUTH-2** — High — `internal/web/serve_build.go:21-51`, `internal/web/server.go`
  - Issue: `Serve` has no users/sessions/roles; any reachable peer can view the page and invoke `POST /click`, `/type`, `/backspace`, and (when enabled) `/debug/state` or `/debug/pprof/*`.
  - Proof: Mux registers mutating routes with only `sameOrigin` checks; `Serve` performs no identity check.
  - Solution: Keep documented localhost-only default; require reverse proxy/app-level auth before binding beyond loopback; keep `Debug`/`Perf` off in shipped binaries.

- [x] **RACE-1** — High — `internal/page/page_patch_state.go:38-45`, `internal/page/page_patch_set.go:15-30`
  - Issue: `BindState` listener appends to `Page.changedProps` from arbitrary goroutines while the page loop reads/drains the same slice without synchronization.
  - Proof: state store invokes listeners outside its lock (`internal/state/store_patch.go`, `patch_report.go`); `AddChangedProperty` appends to `p.changedProps` (`page_patch_set.go:19`); `ChangedProperties`/`TakeChangedProperties` read it (`page_patch_props.go:12,23`).
  - Solution: Protect `changedProps` with a mutex/atomic snapshot, or drop listener mutation and rely on `SyncState()`/store fallback under the page loop.

---

## Phase 2 — Medium (next sprint)

- [x] **CRYPTO-2** — Medium — `examples/download-manager/scheduler/id.go:11`, `examples/lg-remote/ws/accept.go:13`, `examples/lg-remote/ws/io.go:28`
  - Issue: `crypto/rand.Read` errors discarded; proceeds with zero-filled buffer.
  - Proof: `_, _ = rand.Read(...)`.
  - Solution: Check the error and fail/panic instead of continuing with zero bytes.

- [x] **WEB-1 / COOKIEIPC-5** — Medium — `internal/web/http_security.go:62-74`
  - Issue: CSRF check returns `true` when `Origin` is missing; mutating routes rely only on this.
  - Proof: `if origin == "" { return true }`.
  - Solution: Require matching `Origin` (or CSRF token) for state-changing routes; allow missing Origin only for safe GETs.

- [x] **WEB-2 / LOG-2** — Medium — `internal/web/serve_build.go:39-47`, `internal/web/debug_http.go:26-60`
  - Issue: Debug/profiling endpoints have no auth beyond opt-in flags; `/debug/state` exposes boxes/stats/ops/runtime and reload errors; `/debug/pprof/*` exposed when `Perf`.
  - Proof: `debug()` encodes `debugState` including `Boxes`, `Stats`, `Ops`, `Runtime`; pprof registered only when `opts.Perf`.
  - Solution: Keep disabled by default; bind to loopback or add token; sanitize `LastReloadError`.

- [x] **AUTH-1** — Medium — `examples/login/login/login_state.go:13`, `examples/input-lab/inputlab/status.go:60-61`
  - Issue: Hardcoded `secret`/`secret` checked in application code; no real auth boundary.
  - Proof: `if a.page.FormValue("email") == "secret" && a.page.FormValue("password") == "secret"`.
  - Solution: Treat as demos only; remove hardcoded creds or move verification to a trusted backend.

- [x] **AUTH-3 / COOKIEIPC-3** — Medium — `examples/desktop-cat/cat/http.go:10-42`, `examples/desktop-cat/notify.go:18-36`
  - Issue: Notification endpoints have no origin/CSRF/auth; `startNotifications` calls `net.Listen("tcp", addr)` without bind validation.
  - Proof: Handler never calls `sameOrigin`; no token or bind validation.
  - Solution: Add `sameOrigin` + `securityHeaders`, validate bind address (loopback default), document shared secret for remote use.

- [x] **RACE-2** — Medium — `internal/web/debug_runtime.go:15, 38-39`
  - Issue: Exported mutable `RuntimeSnapshot` func var bypasses its mutex on direct assignment, racing with `runtimeSnapshot()` reads.
  - Proof: `runtimeSnapshot()` reads `fn := RuntimeSnapshot` under `RLock`; direct assignment (as in tests) is unlocked.
  - Solution: Make variable unexported; expose only via `SetRuntimeSnapshot`/`GetRuntimeSnapshot`; update callers/tests.

- [x] **FS-1** — Medium — `internal/crash/validate.go:45, 54-58`
  - Issue: `validateDir` calls `filepath.Clean` before checking for `..`, so `/allowed/../secret` normalizes to `/secret` and bypasses traversal rejection.
  - Proof: `filepath.Clean("/allowed/../secret")` returns `/secret`.
  - Solution: Validate raw input segments before cleaning, or resolve final path and confirm it stays within trusted root.

- [x] **MEM-1** — Medium — `internal/bitmap/font.go:18`
  - Issue: Cache key derived from `uintptr(unsafe.Pointer(&data[0]))` without `len(data) > 0` check and without retaining `data`.
  - Proof: `key := faceKey{ptr: uintptr(unsafe.Pointer(&data[0])), size: size}`.
  - Solution: Guard `len(data) == 0`; key cache on stable identity (content hash or retain slice) rather than raw pointer.

- [x] **MEM-2** — Medium — `internal/clipboard/clipboard_darwin.go:49`
  - Issue: `C.ownframePut(p, C.int(len(text)))` truncates length to `C.int`; huge strings cause negative/wrapped length and OOB read in CGO.
  - Proof: `p := unsafe.Pointer(unsafe.StringData(text)); C.ownframePut(p, C.int(len(text)))`.
  - Solution: Validate length against `math.MaxInt32` (or use `size_t`/`NSUInteger`) before calling C.

- [x] **MEM-3** — Medium — `internal/clipboard/clipboard_utf.go:8`
  - Issue: `(len(u)+1)*2` can overflow on 32-bit, allocating too-small buffer before writing `len(u)` UTF-16 code units.
  - Proof: `b := make([]byte, (len(u)+1)*2)` then `b[i*2] = byte(v)`.
  - Solution: Check overflow before allocating; return error/empty result when size exceeds `math.MaxInt/2`.

- [x] **LOG-1** — Medium — `internal/web/page_http.go:27`, `internal/web/click_http.go:65`, `internal/web/input_http.go:26,53`
  - Issue: Handlers return raw `err.Error()` to clients.
  - Proof: `http.Error(w, err.Error(), http.StatusInternalServerError)`.
  - Solution: Return generic `500 internal server error`; log full error server-side (sanitized).

- [x] **COOKIEIPC-1** — Medium — `internal/ipc/ipc.go:33`
  - Issue: IPC registry is process-global and unauthenticated; `Listen` appends any callback, `Handle` replaces handler without owner identity.
  - Proof: global `ears`/`asks` maps keyed only by channel name; `Handle` writes `asks[channel] = ask{...}`.
  - Solution: Treat IPC as in-process convenience only; use owner-prefixed channels; keep secrets/capabilities out of payloads; add owner tokens if isolation needed.

- [x] **COOKIEIPC-2** — Medium — `internal/ipc/request.go:36`
  - Issue: `Request` runs registered handler synchronously in caller's goroutine with no per-call timeout.
  - Proof: `item.fn(ctx, payload)` called directly inside `wrap()` with no deadline/worker.
  - Solution: Document non-blocking handler contract; require contexts with deadlines; consider timeout/worker pool if untrusted code can register.

- [x] **INJ-2** — Medium — `examples/teams/store/seed.go:54-58`
  - Issue: Seed SQL file lines executed directly via `tx.Exec(line)`.
  - Proof: loop over `strings.Split` lines; `tx.Exec(line)`.
  - Solution: Parse seed data in Go and insert via parameterized statements; or validate seed integrity (checksum) and disable multi-statement execution; document trusted build input.

- [ ] **DEP-4** — Medium — `examples/go.mod:6-7`
  - Issue: Examples pin stale `ownframe v0.0.2` and `blinkless v0.0.2`, which predate current root; standalone builds fail.
  - Proof: `go mod tidy` in `examples/` fails resolving `internal/...` against v0.0.2; builds only via `go.work`.
  - Solution: After publishing matching release, bump both pins to new tags; drop workspace reliance for external builds.

---

## Phase 3 — Low (opportunistic)

- [x] **INJ-3** — Low — `internal/page/page_nav.go:29`, `page_nav_route.go:27`, `page_reload_apply.go:35`, `internal/web/serve_build.go:31`
  - Issue: `html/template` parses author-controlled source; template actions can call methods on data. File watch reloads template automatically.
  - Proof: `template.New("page").Parse(source)` where source may come from `Config.File`/`ThemeFile`.
  - Solution: Treat template files as trusted; ensure file permissions; disable reload in production (`-reload=false`); document trust boundary.

- [x] **CRYPTO-3** — Low — `examples/tetris/game/id.go:22`, `examples/pet-care/app/pet_actions.go:60`
  - Issue: IDs fall back to `time.Now().UnixNano()` when `crypto/rand` fails.
  - Proof: `fmt.Sprintf("g%016x%08x", time.Now().UnixNano(), idSeq.Add(1))` and `fmt.Sprintf("%x", time.Now().UnixNano())`.
  - Solution: Drop time-based fallback; return error or panic when system random source fails.

- [x] **WEB-3** — Low — `internal/web/click_http.go:65`, `internal/web/input_http.go:26,53`, `internal/web/page_http.go:27`
  - Issue: Handler errors returned verbatim to clients (overlap LOG-1).
  - Proof: `http.Error(w, err.Error(), http.StatusInternalServerError)`.
  - Solution: Log full error server-side; return generic 500.

- [x] **WEB-4** — Low — `internal/web/serve_build.go:57`
  - Issue: No `MaxHeaderBytes`; click parsing accepts unbounded query parameters.
  - Proof: `http.Server` sets only timeout fields; `click` calls `r.URL.Query()` without limits.
  - Solution: Set `MaxHeaderBytes` (e.g. 64 KiB); bound/validate click query parameters.

- [x] **WEB-5** — Low — `internal/web/http_security.go:73`
  - Issue: `sameOrigin` compares only host, ignores scheme.
  - Proof: `strings.EqualFold(u.Host, r.Host)` without scheme comparison.
  - Solution: Compare both scheme and host, allowing only the scheme actually used by the preview.

- [x] **AUTH-4** — Low — `internal/window/devtools_own.go:12-18`; examples using `Config{DevTools: true}`
  - Issue: DevTools inspector shipped enabled in examples; exposes element JSON, boxes, coordinates, metrics, ops.
  - Proof: `available := s.dev.available || s.dev.on || insp.DevTools()`.
  - Solution: Ship examples/production with `Config.DevTools` false unless debugging; document exposure.

- [x] **AUTH-5** — Low — `documentation/features.md`, `internal/web`
  - Issue: No session/cookie primitives; `Serve` cannot establish/validate authenticated session on its own.
  - Proof: No `http.Cookie`, `SetCookie`, or session-token handling found.
  - Solution: Document that authentication is application-level; require apps to bind locally or use authenticated proxy.

- [x] **FS-2** — Low — `internal/filepick/pick_wsl_path.go:19-33`
  - Issue: UNC branch validates dot segments, but `wslpath -u` branch returns converted path without `hasDotSegment` or absoluteness check.
  - Proof: `fromWindowsPath` returns `strings.TrimRight(string(out), "\r\n")` directly.
  - Solution: Apply `hasDotSegment` to converted Linux path; reject non-absolute or `.`/`..` results.

- [x] **MEM-4** — Low — `internal/state/path_read.go:30`, `internal/state/path_setseg.go:36`
  - Issue: Reflection helpers call `f.Interface()`/`f.Set(...)` without `CanInterface()`/`CanSet()` checks; unexported matching field panics.
  - Proof: `f := rv.FieldByName(seg); ... cur = f.Interface()` and `f := elem.FieldByName(seg); ... f.Set(v)`.
  - Solution: Skip fields where `!f.CanInterface()`/`!f.CanSet()`; return unchanged/false.

- [x] **LOG-3** — Low — `internal/web/http_security.go:28`, `internal/web/debug_runtime.go:47`, `internal/web/page_http.go:36,57`, `internal/web/debug_http.go:60`, `internal/clipboard/x11_recover.go:9`, `internal/window/reload.go:50`, `internal/web/reload.go:38`, `internal/filepick/pick_darwin.go:41`, `internal/filepick/pick_wsl_shell.go:42`
  - Issue: Logs use error/panic values directly, allowing newline-bearing values to forge log lines.
  - Proof: `log.Printf("web handler panic: %v", rec)` and similar.
  - Solution: Strip CR/LF from logged values or use structured escaped logging.

- [x] **LOG-4** — Low — `internal/page/page_draw_commit.go:17`, `internal/page/page_patch_element.go:65`, `internal/page/page_patch_splice.go:16`
  - Issue: Unconditional `println("ZZDEBUG ...")` statements print internal box/op counts.
  - Proof: `println("ZZDEBUG commit boxes", len(display.Boxes), len(display.Ops))`.
  - Solution: Remove debug prints.

- [x] **LOG-5** — Low — `internal/filepick/pick_darwin.go:37-42`, `internal/filepick/pick_wsl_shell.go:38-43`, `internal/window/idle_debug.go:16-21`
  - Issue: Debug env vars enable stderr diagnostics that may include paths/command failure details.
  - Proof: `debugf` and `paceDebugWriter`.
  - Solution: Keep flags off in shipped binaries; document as dev-only.

- [x] **LOG-6** — Low — `internal/page/page_reload_note.go:9,20`, `internal/page/page_reload_state.go:7-11`, `internal/window/reload.go:32-50`, `internal/web/reload.go:20-38`
  - Issue: Hot reload errors (including file paths) stored in page stats, printed to stderr, exposed via debug state.
  - Proof: `p.stats.reloadErr = err.Error()` and `fmt.Fprintln(os.Stderr, line)`.
  - Solution: Avoid exposing reload error text outside debug mode; sanitize stderr output.

- [x] **LOG-7** — Low — `examples/desktop-cat/cat/http.go:28,37`
  - Issue: Notification API returns raw JSON decode/submit error details to clients.
  - Proof: `reply(w, 400, map[string]string{"error": err.Error()})`.
  - Solution: Return generic validation errors.

- [x] **LOG-8** — Low — `report.go:23`, `run.go:57`
  - Issue: `reportError` returns an error containing crash report file path.
  - Proof: `fmt.Errorf("ownframe: %w: %s", err, path)`.
  - Solution: Return original error without path; expose path only via explicit API.

- [x] **LOG-9** — Low — `examples/download-manager/ui/wire/wire_dir.go:30,43`, `examples/teams/app/db.go:17`, `examples/teams/app/state.go:46`
  - Issue: Example apps log directory paths and DB errors to stdout/stderr.
  - Proof: `log.Printf("download-manager: open %s: %v", dir, err)`.
  - Solution: Remove verbose path logging or sanitize output in shipped examples.

- [x] **COOKIEIPC-4** — Low — `examples/desktop-cat/notify.go:22`
  - Issue: Notification listener binds to caller-supplied address with no validation.
  - Proof: `listener, err := net.Listen("tcp", addr)` with no `validateBindAddr` equivalent.
  - Solution: Reuse `web.validateBindAddr`-style checks or require opt-in for non-loopback binds; document localhost-only.

- [x] **COOKIEIPC-6** — Low — `internal/fetch/scheme.go:64`
  - Issue: Fetch allows `https`→`http` redirect and performs no host/IP validation (SSRF redirect leg; see INJ-4).
  - Proof: `redirect` only checks `web(req.URL)` (http or https); no scheme downgrade or private-range rejection.
  - Solution: Document caller-side URL validation (already noted as residual risk); optionally reject downgrades and private IP ranges in transport.

- [x] **COOKIEIPC-7** — Low — `internal/window/devtools_press.go:48`
  - Issue: DevTools writes rendered operation text to OS clipboard on click without setting/consent.
  - Proof: `clipboard.Write(text)` in `devActOp` branch.
  - Solution: Keep DevTools disabled in shipped binaries (documented) or add opt-in for clipboard writes.

- [x] **DEP-5** — Low — `go.sum`
  - Issue: `go mod tidy -diff` removes 8 stale `blinkless` checksum lines from prior pseudo-versions.
  - Proof: `go mod tidy -diff` shows only those removals.
  - Solution: Run `go mod tidy` on root after resolving DEP-3.

---

## Notes

- Overlapping findings merged: WEB-1=COOKIEIPC-5; WEB-2=LOG-2; WEB-3=LOG-1; AUTH-3=COOKIEIPC-3; INJ-4=COOKIEIPC-6 (redirect leg).
- `examples/teams` SQL uses parameterized queries; the only SQL string building is over hardcoded table-name slices (not user input).
- Hardcoded sign-in values in demo apps are documented test credentials, not production secrets.
- No cookie/session handling exists in the library (documented as absent); authentication is application-level.
- Recommended verification after fixes: `go test -race ./...`, `gosec ./...`, `govulncheck ./...`, and fuzzing of parsers/input handlers.



