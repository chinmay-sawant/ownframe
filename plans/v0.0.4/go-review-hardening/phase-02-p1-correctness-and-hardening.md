# v0.0.4 - Phase 02: P1 Correctness & Security Hardening

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P1
> **Status:** In progress (12 / 13 `[x]`, 1 `[~]` waiting on a Windows host)

---

## Overview

This phase resolves verified functional correctness bugs (AT-SPI coalescing, emoji HTML rewriting, stale display caches, `NaN` propagation, `pprof` timeout, OpenType caret parity) and hardens network/HTTP boundaries in `internal/fetch` and `internal/web`.

---

## 2.1 Accessibility, Emoji & Text Correctness

- [x] `CORR-ATSPI-01` Coalesce Linux AT-SPI snapshots by keeping the *newest* frame instead of dropping it (`internal/window/a11y_atspi.go:144-151`).
  - **Root cause:** When 1-slot channel `a.jobs` is full, `default:` drops the incoming newer `job` while leaving the older snapshot queued, contradicting the comment and risking stale AT-SPI state once desktop idle parking engages.
  - **Fix:** Under `a.mu`, drain any stale queued job (`select { case <-a.jobs: default: }`) and enqueue the newest `job`.
  - **Proof:** Unit test in `internal/window/` verifying rapid 3-snapshot burst delivers the latest generation to the AT-SPI worker.
  - **Result (done, commit `17caf6c9`):** `publishA11y` drains the stale queued job (decrementing `queued`) under `a.mu` before queuing the newest. File: `internal/window/a11y_atspi.go`; test `TestLinuxA11yBurstKeepsNewestSnapshot` (3 generations, expects gen 3 and `queued == 1`) in `a11y_atspi_test.go`. Command: `go test -race -count=1 ./internal/window` exit 0.

- [x] `CORR-ATSPI-02` Fix truncated D-Bus header loop exit in `internal/atspi/dbus_wire.go:239-249`.
  - **Root cause:** `break` inside the `switch` on truncated header field bounds exits only the `switch`, not the enclosing `for offset < len(fieldsBytes)` loop.
  - **Fix:** Use a labeled `break` or return a structured parse error on truncated header fields.
  - **Proof:** Unit/fuzz test in `internal/atspi/` feeding truncated header-field payloads and asserting clean error return.
  - **Result (done, commit `17caf6c9`):** The truncated-field `break`s use a `fields:` label so the parse loop ends. File: `internal/atspi/dbus_wire.go`; test `TestReadMessageTruncatedFieldStopsParsing`. Red-then-green: old file gave `parsed past a truncated field: member "evil"`; fixed `go test -race -count=1 ./internal/atspi` exit 0. No dedicated fuzz target added.

- [~] `CORR-UIA-01` Wire `ctxDone` auto-detach watcher for Windows UIA (`internal/window/a11y_uia_windows.go`, matching `internal/window/a11y_atspi.go:46-59`).
  - **Root cause:** Unlike Linux AT-SPI, `a11y_uia_windows.go` has no context-cancellation watcher, so `NewGame` hosts on Windows do not automatically detach UIA when their context cancels.
  - **Fix:** Add a `ctxDone` watcher invoking `s.closeA11y()` on context cancellation, guarded by `a.closeOnce`.
  - **Proof:** Headless lifecycle test on Windows verifying UIA detaches on context cancellation.
  - **Result (blocked on host, commit `17caf6c9`):** Code done: `initA11y` starts `watchA11y(ctxDone)`; `closeA11y` marks `closed` under a new `a11yHost.mu`, `attachUIA` closes a host that attaches after close. Files: `internal/window/a11y_uia_windows.go`, test `TestWindowsA11yDetachesOnContextCancel` in `a11y_uia_windows_test.go`. Cross-compile evidence only: `GOOS=windows GOARCH=amd64|arm64 go vet ./internal/window` exit 0. **Reconsideration gate:** run `go test -race ./internal/window` on a Windows runner.

- [x] `CORR-EMOJI-01` Fix unclosed HTML comment off-by-one and RCDATA (`<textarea>`/`<title>`) corruption in `internal/emoji/emoji_replace.go:33-47`.
  - **Root cause:** On unclosed `<!--`, `end` subtracts `len("<!--")` (4) while the slice adds `len("-->")` (3), leaving 1 trailing byte unconsumed (slicing mid-UTF-8 if the comment ends with a multi-byte rune). In addition, only `script` and `style` enter `rawText`, so `<textarea>hi 😂</textarea>` is rewritten to contain raw `<img ...>` markup.
  - **Fix:** On `end < 0`, write `html[i:]` and set `i = len(html)`. Extend the raw/RCDATA tag set to include `textarea` and `title`.
  - **Proof:** Unit + fuzz tests (`FuzzReplace` + explicit cases for unclosed comments ending in emoji and `<textarea>`/`<title>` with emoji).
  - **Result (done, commit `17caf6c9`):** Unclosed `<!--` now copies the tail and stops; `rawTextTag` adds `textarea` and `title` to `script`/`style`. Files: `internal/emoji/emoji_replace.go`, `emoji_scan.go`; tests `TestReplaceUnclosedCommentEndingInEmoji`, `TestReplaceRCDATAUntouched`, and `FuzzReplace` seeds plus a valid-UTF-8-out invariant. Commands: `go test -count=1 ./internal/emoji` exit 0; `go test -run xxx -fuzz FuzzReplace -fuzztime 15s ./internal/emoji` exit 0 (987k execs, no crash).

- [x] `CORR-EMOJI-02` Eliminate O(N²) `strings.ToLower(html[k:])` inside `rawText` (`internal/emoji/emoji_scan.go:64`).
  - **Root cause:** Every `<` inside `<script>` or `<style>` allocates and lowercases the entire remaining HTML tail.
  - **Fix:** Check `len(html)-k >= len(closeTag) && strings.EqualFold(html[k:k+len(closeTag)], closeTag)`.
  - **Proof:** Unit test + microbenchmark on a large `<script>` block with many `<` comparisons confirming linear time and zero tail copies.
  - **Result (done, commit `17caf6c9`):** `rawText` compares `strings.EqualFold(html[k:k+len(close)], close)` with a length guard: no tail copy. Test `TestRawTextLinear` (400 KB script with 200k `<`, asserts <= 2 allocs) passes; `BenchmarkRawTextManyAngles` is written but not run (benchmarks run in the primary session). Command: `go test -count=1 ./internal/emoji` exit 0.

- [x] `CORR-TEXT-01` Include `op.FontFeatures()` in `internal/textrun/face.go:15-28` and avoid unbounded process-wide face retention.
  - **Root cause:** `textrun`'s `faceKey` omits `op.FontFeatures()` (present in `internal/replay/text_cache.go:24`), causing caret hit-testing widths to diverge from painted advances when OpenType features (`tnum`, `liga`, etc.) change glyph widths; `textFaces` also never trims on idle.
  - **Fix:** Include `features` in `textrun`'s `faceKey` and expose/wire a cache trim hook or bound.
  - **Proof:** Unit test comparing caret advance vs replay advance with `font-feature-settings: "tnum"` / `"liga" 0`.
  - **Result (done, commit `17caf6c9`):** `textrun.faceKey` carries `op.FontFeatures()` and the face applies the features like `internal/replay`; new `textrun.Trim()` drops both maps and `paceGame.frame` calls it with `renderer.SetIdle()` (`internal/window/idle.go`). The existing `maxSources`/`maxFaces` bound stays. Files: `internal/textrun/face.go`, `internal/window/idle.go`; tests `TestFaceKeyIncludesFontFeatures` (renders `font-feature-settings: "tnum" 1`; faces must differ) and `TestTrimDropsAndRebuildsFaces` in `face_features_test.go`/`export_test.go`. Command: `go test -race -count=1 ./internal/textrun` exit 0. Not done: a caret-vs-replay pixel-advance comparison, because the bundled font has no `tnum` glyph change to measure; the cache-key test is the regression guard.

## 2.2 Display Cache Invalidation & `NaN` Guards

- [x] `CORR-CACHE-01` Fix stale `orderCache` and `planCache` hits on recycled `*layout.Display` pointers or changed `viewportPinZ` (`internal/window/replay.go:88-115`, `internal/replay/blend.go:76-95`, `internal/window/game.go:257`).
  - **Root cause:** `scanOrder()` keys only on `s.orderCache.display == s.display`, ignoring `s.viewportPinZ()` and `s.seq`. Likewise, `shell.setDisplay` skips `EvictDisplay` when `s.display == display`, so a recycled `*layout.Display` address with identical op/box counts hits stale `groupNode` (including `maskTex`/`backdropBox`).
  - **Fix:** Key `orderCache` on `(display, seq, viewportPinZ)` and call `EvictDisplay` (or bump a display revision) whenever a new display generation is installed, regardless of pointer equality.
  - **Proof:** Regression test toggling `SetViewportPinZ` and recycling a `*layout.Display` pointer with same op count but changed mask/color/backdrop.
  - **Result (done, commit `17caf6c9`):** `orderCache` keys on `(display, seq, viewportPinZ)`; `setDisplay` calls `EvictDisplay` even when the pointer repeats. New `Renderer.CachePlan`/`PlanCached` (`internal/replay/plan_probe.go`) let the window test observe eviction. Files: `internal/window/replay.go`, `game.go`; tests `TestScanOrderRescansWhenPinZChanges`, `TestScanOrderRescansWhenGenerationChanges`, `TestSetDisplayEvictsRecycledPointer` in `internal/window/order_cache_test.go`. Command: `go test -race -count=1 ./internal/window ./internal/replay` exit 0. The tests drive the cache logic, not a full recycled-pointer paint, so `make graphics-check` still has to confirm no paint change.

- [x] `CORR-NAN-01` Guard `math.IsNaN` / `math.IsInf` in audio playback and virtual windowing (`internal/page/audio.go:98-105`, `ui/virtual/virtual.go:23-67`).
  - **Root cause:** `< 0` and `> 1` clamps both evaluate `false` for `NaN`; storing `NaN` audio volume breaks `json.Marshal` in `/state` and `/events`, and `NaN` in `virtual.Window` produces `NaN` spacers and `math.MinInt` indices.
  - **Fix:** Clamp `math.IsNaN(volume)` (and `±Inf`) to `0` (or `1` for `+Inf`) in `audioSlot.setPlayback`, and reject/clamp non-finite floats in `ui/virtual.Window`.
  - **Proof:** Unit tests passing `math.NaN()` and `math.Inf(1)` into audio playback (verifying `/state` JSON succeeds) and `ui/virtual.Window`.
  - **Result (done, commit `17caf6c9`):** `setPlayback` maps NaN to 0 (+Inf already clamps to 1); `virtual.Window` rejects non-finite `rowH`, zeroes NaN `scrollY`/`viewportH`, clamps `viewportH` to the content height, and clamps the row quotients as floats before `int`. Files: `internal/page/audio.go`, `ui/virtual/virtual.go`; tests `TestAudioPlaybackNonFinite`, `TestWindowNonFiniteInputs`. Command: `go test -race -count=1 ./internal/page ./ui/virtual` exit 0.

## 2.3 Fetch SSRF Hardening & Web Server Resilience

- [x] `HARD-FETCH-01` Expand `internal/fetch` SSRF IP blocking with `net/netip`, isolate `AllowPrivate` connection pooling, export remaining sentinel errors, and remove brittle stdlib error-string matching (`internal/fetch/ssrf.go:24-43`, `scheme.go:15-31`, `dial_validated.go:10-45`, `do.go:53`, `fetch.go:13-20`).
  - **Root cause:** `net.IP.IsGlobalUnicast()` returns `true` for RFC 6598 CGNAT (`100.64.0.0/10`), NAT64 (`64:ff9b::/96`), benchmark (`198.18.0.0/15`), and `240.0.0.0/4`; `net.ParseIP` fails on zoned IPv6; one shared `http.Transport` pools keep-alive connections across normal and `AllowPrivate` calls; `AllowPrivate(nil)` panics; `checkURL` + `dialValidated` resolve DNS twice; root `fetch.go` omits sentinels (`ErrBodyTooLarge`, `ErrHeadersTooLarge`, `ErrTooManyRedirects`, `ErrCredentialsInURL`, `ErrMethod`, `ErrBlockedHeader`); `do.go:53` matches `"response headers exceeded"` by substring.
  - **Fix:** Use `net/netip.Addr` + `Unmap()` with explicit prefix blocks (CGNAT, NAT64, benchmark, reserved, zoned IPv6), give `AllowPrivate` its own transport (or disable keep-alive reuse across modes), guard `AllowPrivate(nil)`, eliminate redundant DNS resolution where safe, expose needed fetch sentinels cleanly, and replace brittle substring checks where typed checks apply.
  - **Proof:** Unit + fuzz tests (`FuzzBlockedIP`) covering `100.64.0.1`, `64:ff9b::169.254.169.254`, `198.18.0.1`, `fe80::1%eth0`, `AllowPrivate(nil)`, and `errors.Is` on exported fetch sentinels.
  - **Result (done, commit `17caf6c9`):** `blockedAddr(netip.Addr)` unmaps, drops zones, and blocks CGNAT, NAT64, benchmark, TEST-NETs, `240/4`, Teredo, 6to4, documentation ranges; literals parse with `netip.ParseAddr` (zoned IPv6 works); `AllowPrivate` uses its own `privateClient`/transport; `AllowPrivate(nil)` returns nil (Do then returns `ErrNilContext`); `checkHost` skips the second DNS lookup unless a system proxy is on (the pinned dialer still resolves once and rejects blocked answers, and `Do` maps its error back to `ErrPrivateAddress`/`ErrDowngrade`); the six sentinels are re-exported from the module root (`fetch.go`); the header-size substring check is isolated in `headersTooLarge` with a comment, since net/http offers no typed error. Files: `internal/fetch/*.go`, `fetch.go`; tests in `ssrf_ranges_test.go`, `fuzz_blocked_test.go`, root `fetch_sentinels_test.go`. Commands: `go test -race -count=1 ./internal/fetch` exit 0; `go test -count=1 -run FetchSentinels .` exit 0; `go test -run xxx -fuzz FuzzBlockedIP -fuzztime 15s ./internal/fetch` exit 0 (127k execs).

- [x] `HARD-DOM-01` Enforce `CheckDOM` limits on expanded template output and treat `<style>`/`<script>` as raw text (`internal/page/source/dom.go:13-49`, `internal/page/render.go:143-155`).
  - **Root cause:** `CheckDOM` only inspects raw template text at `New`/`Load`/`Route`, so `{{range}}` expansion can exceed node/depth limits unchecked; `CheckDOM` also counts `<...>` inside `<style>`/`<script>` blocks as tags.
  - **Fix:** Skip `<script>`/`<style>` raw-text interiors in `CheckDOM` and enforce a bounded node/depth check on post-template executed output (or wire through the cached parse pass).
  - **Proof:** Unit + fuzz tests (`FuzzDOMLimits`) covering large `{{range}}` expansions and `<style>`/`<script>` containing `<`.
  - **Result (done, commit `17caf6c9`):** `CheckDOM` skips `<script>`/`<style>` bodies (`rawTextEnd`, no allocation) and `execTemplate` runs `checkDOMLimits` on the executed output. Files: `internal/page/source/dom.go`, `internal/page/render.go`; tests in `internal/page/dom_expand_test.go` (a `{{range}}` of 20010 items fails Redraw with `ErrTooManyNodes`; 20010 `<b>` inside `<style>` passes) and updated `FuzzDOMLimits` oracle/seeds. Commands: `go test -race -count=1 ./internal/page/...` exit 0; `go test -run xxx -fuzz FuzzDOMLimits -fuzztime 15s ./internal/page` exit 0. The per-redraw cost of the extra linear scan is not measured; benchmark in the primary session.

- [x] `HARD-WEB-01` Stop holding `s.mu.Lock()` across HTTP socket writes (`internal/web/page_http.go:21-110`, `dom_http.go:94-148`, `state_http.go:43-54`, `debug_http.go:27-54`).
  - **Root cause:** All five handlers hold `s.mu` while writing directly to `http.ResponseWriter` (with `WriteTimeout: 0`), allowing a single slow client to stall the 60 Hz tick loop and all other routes.
  - **Fix:** Snapshot or encode response bytes into a buffer under `s.mu`, release `s.mu.Unlock()`, then write headers and body to `w`.
  - **Proof:** Test with a blocking/slow `http.ResponseWriter` verifying concurrent routes and ticks remain unblocked under `go test -race ./internal/web`.
  - **Result (done, commit `17caf6c9`):** `page`, `frame`, `dom`, `state`, `debug` build a `reply` (buffered body) under `s.mu`, unlock, then write. Files: `internal/web/page_http.go`, `dom_http.go`, `state_http.go`, `debug_http.go`; test `TestHandlersReleaseLockBeforeSocketWrite` (blocking `ResponseWriter`, five routes) in `slow_writer_test.go`. Red-then-green: old handlers fail all five with `server mutex held across the socket write`; fixed `go test -race -count=1 ./internal/web` exit 0.

- [x] `HARD-WEB-02` Cap SSE `/events` subscribers, fix `closed` header race, validate pre-bound `opts.Listener`, avoid emitting plaintext password values under remote bind, and add `sameOrigin(r)` to `/debug/state` (`internal/web/events_http.go:25-103`, `serve_build.go:32-41, 85`, `debug_http.go:27`, `document.go:250-251`).
  - **Root cause:** `/events` allows unlimited concurrent subscribers and writes `"retry: 3000\n\n"` before checking `s.subscribe(1)`; `checkBindAddr` skips loopback check when `opts.Listener != nil`; `debug` omits `sameOrigin(r)` (defense-in-depth); `fmt.Println("http://"+addr+"/")` prints `:0` on ephemeral ports; `document.go:250-251` emits plaintext `value="..."` on `<input type="password">` in `GET /` and `GET /dom`.
  - **Fix:** Enforce a reasonable subscriber cap (return `503`), subscribe before writing headers, validate `opts.Listener.Addr()` loopback when `!opts.InsecureAllowRemoteBind`, add `sameOrigin(r)` in `debug`, omit/redact password values from serialized HTML/DOM JSON (or gate strictly), and log `ln.Addr().String()`.
  - **Proof:** Unit tests in `internal/web/` covering subscriber cap, closed-server `/events`, remote pre-bound listener rejection, password redaction, and cross-origin `/debug/state`.
  - **Result (done, commit `17caf6c9`):** `/events` subscribes first (cap `maxSubscribers = 64`, 503 when full or closed, no `retry:` line), `checkBindAddr` judges a pre-bound listener by `Listener.Addr()` (unix sockets exempt), `/debug/state` requires `sameOrigin`, the banner prints `ln.Addr()`, and `type=password` values are blanked in `/` and `/dom` when `InsecureAllowRemoteBind` is set (`redactPasswords`; loopback binds keep the live value so typing still round-trips). Files: `internal/web/events_http.go`, `serve_insecure.go`, `serve_build.go`, `server.go`, `page_http.go`, `dom_http.go`, `debug_http.go`; tests in `sec_web02_test.go`. Command: `go test -race -count=1 ./internal/web` exit 0.

- [x] `HARD-DBG-01` Fix debug server `WriteTimeout` killing default 30s `pprof` captures and add shutdown timeout (`internal/window/debug_server_tag.go:21, 51, 74-79`).
  - **Root cause:** `debugWriteTimeout = 15 * time.Second` aborts `GET /debug/pprof/profile` (30s default), and `hs.Shutdown(context.Background())` can hang indefinitely on an open connection.
  - **Fix:** Set `WriteTimeout: 0` (or `>= 60s` / per-route `http.ResponseController`) for pprof compatibility and bound `hs.Shutdown` with a 2–5s `context.WithTimeout`.
  - **Proof:** Test with `-tags ownframe_debug` verifying clean shutdown with an idle open connection and profile endpoint deadline > 30s.
  - **Result (done, commit `17caf6c9`):** `WriteTimeout` is 0; `withDebugWriteDeadline` sets a per-route deadline (`seconds`+30 s for pprof profile/trace, capped at 10 min; 15 s elsewhere); shutdown uses a 3 s context then `hs.Close()`. File: `internal/window/debug_server_tag.go`; tests `TestDebugWriteDeadlineCoversProfileCapture`, `TestDebugServerShutdownBoundedByOpenRequest` (profile `seconds=20` open, `stop()` returns in about 3.3 s). Command: `go test -tags ownframe_debug -race -count=1 -run "DebugServer|DebugWriteDeadline" ./internal/window` exit 0. A real 30 s `go tool pprof` capture was not run end to end.

---

## Closure Gates

- [ ] `P1A-GATE-01` `make test` and `make test-race` pass cleanly across all modified packages (`emoji`, `textrun`, `fetch`, `web`, `window`, `replay`, `page`, `atspi`, `ui/virtual`). **Status:** partial. Per-package `go test -race -count=1` passed for `emoji`, `textrun`, `fetch`, `web`, `window`, `replay`, `page/...`, `atspi`, `ui/...`; whole-repo `make test` and `make test-race` pending primary-session run.
- [x] `P1A-GATE-02` `make lint` and `make graphics-check` pass with exit `0`. **Status:** done. Primary session (commit 17caf6c9): `make lint` exit `0`; `make graphics-check` exit `0`.
- [ ] `P1A-GATE-03` All 7 fuzz targets (`FuzzReplace`, `FuzzBlockedIP`, etc.) pass a 15s smoke run with zero crashes. **Status:** partial. 15 s smoke passed for the targets these rows touch (`FuzzReplace`, `FuzzBlockedIP`, `FuzzDOMLimits`); the other four targets pending primary-session run.
