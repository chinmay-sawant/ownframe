# v0.0.4 - Phase 02: P1 Correctness & Security Hardening

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P1
> **Status:** Open (0 / 13 complete)

---

## Overview

This phase resolves verified functional correctness bugs (AT-SPI coalescing, emoji HTML rewriting, stale display caches, `NaN` propagation, `pprof` timeout, OpenType caret parity) and hardens network/HTTP boundaries in `internal/fetch` and `internal/web`.

---

## 2.1 Accessibility, Emoji & Text Correctness

- [ ] `CORR-ATSPI-01` Coalesce Linux AT-SPI snapshots by keeping the *newest* frame instead of dropping it (`internal/window/a11y_atspi.go:144-151`).
  - **Root cause:** When 1-slot channel `a.jobs` is full, `default:` drops the incoming newer `job` while leaving the older snapshot queued, contradicting the comment and risking stale AT-SPI state once desktop idle parking engages.
  - **Fix:** Under `a.mu`, drain any stale queued job (`select { case <-a.jobs: default: }`) and enqueue the newest `job`.
  - **Proof:** Unit test in `internal/window/` verifying rapid 3-snapshot burst delivers the latest generation to the AT-SPI worker.

- [ ] `CORR-ATSPI-02` Fix truncated D-Bus header loop exit in `internal/atspi/dbus_wire.go:239-249`.
  - **Root cause:** `break` inside the `switch` on truncated header field bounds exits only the `switch`, not the enclosing `for offset < len(fieldsBytes)` loop.
  - **Fix:** Use a labeled `break` or return a structured parse error on truncated header fields.
  - **Proof:** Unit/fuzz test in `internal/atspi/` feeding truncated header-field payloads and asserting clean error return.

- [ ] `CORR-UIA-01` Wire `ctxDone` auto-detach watcher for Windows UIA (`internal/window/a11y_uia_windows.go`, matching `internal/window/a11y_atspi.go:46-59`).
  - **Root cause:** Unlike Linux AT-SPI, `a11y_uia_windows.go` has no context-cancellation watcher, so `NewGame` hosts on Windows do not automatically detach UIA when their context cancels.
  - **Fix:** Add a `ctxDone` watcher invoking `s.closeA11y()` on context cancellation, guarded by `a.closeOnce`.
  - **Proof:** Headless lifecycle test on Windows verifying UIA detaches on context cancellation.

- [ ] `CORR-EMOJI-01` Fix unclosed HTML comment off-by-one and RCDATA (`<textarea>`/`<title>`) corruption in `internal/emoji/emoji_replace.go:33-47`.
  - **Root cause:** On unclosed `<!--`, `end` subtracts `len("<!--")` (4) while the slice adds `len("-->")` (3), leaving 1 trailing byte unconsumed (slicing mid-UTF-8 if the comment ends with a multi-byte rune). In addition, only `script` and `style` enter `rawText`, so `<textarea>hi 😂</textarea>` is rewritten to contain raw `<img ...>` markup.
  - **Fix:** On `end < 0`, write `html[i:]` and set `i = len(html)`. Extend the raw/RCDATA tag set to include `textarea` and `title`.
  - **Proof:** Unit + fuzz tests (`FuzzReplace` + explicit cases for unclosed comments ending in emoji and `<textarea>`/`<title>` with emoji).

- [ ] `CORR-EMOJI-02` Eliminate O(N²) `strings.ToLower(html[k:])` inside `rawText` (`internal/emoji/emoji_scan.go:64`).
  - **Root cause:** Every `<` inside `<script>` or `<style>` allocates and lowercases the entire remaining HTML tail.
  - **Fix:** Check `len(html)-k >= len(closeTag) && strings.EqualFold(html[k:k+len(closeTag)], closeTag)`.
  - **Proof:** Unit test + microbenchmark on a large `<script>` block with many `<` comparisons confirming linear time and zero tail copies.

- [ ] `CORR-TEXT-01` Include `op.FontFeatures()` in `internal/textrun/face.go:15-28` and avoid unbounded process-wide face retention.
  - **Root cause:** `textrun`'s `faceKey` omits `op.FontFeatures()` (present in `internal/replay/text_cache.go:24`), causing caret hit-testing widths to diverge from painted advances when OpenType features (`tnum`, `liga`, etc.) change glyph widths; `textFaces` also never trims on idle.
  - **Fix:** Include `features` in `textrun`'s `faceKey` and expose/wire a cache trim hook or bound.
  - **Proof:** Unit test comparing caret advance vs replay advance with `font-feature-settings: "tnum"` / `"liga" 0`.

## 2.2 Display Cache Invalidation & `NaN` Guards

- [ ] `CORR-CACHE-01` Fix stale `orderCache` and `planCache` hits on recycled `*layout.Display` pointers or changed `viewportPinZ` (`internal/window/replay.go:88-115`, `internal/replay/blend.go:76-95`, `internal/window/game.go:257`).
  - **Root cause:** `scanOrder()` keys only on `s.orderCache.display == s.display`, ignoring `s.viewportPinZ()` and `s.seq`. Likewise, `shell.setDisplay` skips `EvictDisplay` when `s.display == display`, so a recycled `*layout.Display` address with identical op/box counts hits stale `groupNode` (including `maskTex`/`backdropBox`).
  - **Fix:** Key `orderCache` on `(display, seq, viewportPinZ)` and call `EvictDisplay` (or bump a display revision) whenever a new display generation is installed, regardless of pointer equality.
  - **Proof:** Regression test toggling `SetViewportPinZ` and recycling a `*layout.Display` pointer with same op count but changed mask/color/backdrop.

- [ ] `CORR-NAN-01` Guard `math.IsNaN` / `math.IsInf` in audio playback and virtual windowing (`internal/page/audio.go:98-105`, `ui/virtual/virtual.go:23-67`).
  - **Root cause:** `< 0` and `> 1` clamps both evaluate `false` for `NaN`; storing `NaN` audio volume breaks `json.Marshal` in `/state` and `/events`, and `NaN` in `virtual.Window` produces `NaN` spacers and `math.MinInt` indices.
  - **Fix:** Clamp `math.IsNaN(volume)` (and `±Inf`) to `0` (or `1` for `+Inf`) in `audioSlot.setPlayback`, and reject/clamp non-finite floats in `ui/virtual.Window`.
  - **Proof:** Unit tests passing `math.NaN()` and `math.Inf(1)` into audio playback (verifying `/state` JSON succeeds) and `ui/virtual.Window`.

## 2.3 Fetch SSRF Hardening & Web Server Resilience

- [ ] `HARD-FETCH-01` Expand `internal/fetch` SSRF IP blocking with `net/netip`, isolate `AllowPrivate` connection pooling, export remaining sentinel errors, and remove brittle stdlib error-string matching (`internal/fetch/ssrf.go:24-43`, `scheme.go:15-31`, `dial_validated.go:10-45`, `do.go:53`, `fetch.go:13-20`).
  - **Root cause:** `net.IP.IsGlobalUnicast()` returns `true` for RFC 6598 CGNAT (`100.64.0.0/10`), NAT64 (`64:ff9b::/96`), benchmark (`198.18.0.0/15`), and `240.0.0.0/4`; `net.ParseIP` fails on zoned IPv6; one shared `http.Transport` pools keep-alive connections across normal and `AllowPrivate` calls; `AllowPrivate(nil)` panics; `checkURL` + `dialValidated` resolve DNS twice; root `fetch.go` omits sentinels (`ErrBodyTooLarge`, `ErrHeadersTooLarge`, `ErrTooManyRedirects`, `ErrCredentialsInURL`, `ErrMethod`, `ErrBlockedHeader`); `do.go:53` matches `"response headers exceeded"` by substring.
  - **Fix:** Use `net/netip.Addr` + `Unmap()` with explicit prefix blocks (CGNAT, NAT64, benchmark, reserved, zoned IPv6), give `AllowPrivate` its own transport (or disable keep-alive reuse across modes), guard `AllowPrivate(nil)`, eliminate redundant DNS resolution where safe, expose needed fetch sentinels cleanly, and replace brittle substring checks where typed checks apply.
  - **Proof:** Unit + fuzz tests (`FuzzBlockedIP`) covering `100.64.0.1`, `64:ff9b::169.254.169.254`, `198.18.0.1`, `fe80::1%eth0`, `AllowPrivate(nil)`, and `errors.Is` on exported fetch sentinels.

- [ ] `HARD-DOM-01` Enforce `CheckDOM` limits on expanded template output and treat `<style>`/`<script>` as raw text (`internal/page/source/dom.go:13-49`, `internal/page/render.go:143-155`).
  - **Root cause:** `CheckDOM` only inspects raw template text at `New`/`Load`/`Route`, so `{{range}}` expansion can exceed node/depth limits unchecked; `CheckDOM` also counts `<...>` inside `<style>`/`<script>` blocks as tags.
  - **Fix:** Skip `<script>`/`<style>` raw-text interiors in `CheckDOM` and enforce a bounded node/depth check on post-template executed output (or wire through the cached parse pass).
  - **Proof:** Unit + fuzz tests (`FuzzDOMLimits`) covering large `{{range}}` expansions and `<style>`/`<script>` containing `<`.

- [ ] `HARD-WEB-01` Stop holding `s.mu.Lock()` across HTTP socket writes (`internal/web/page_http.go:21-110`, `dom_http.go:94-148`, `state_http.go:43-54`, `debug_http.go:27-54`).
  - **Root cause:** All five handlers hold `s.mu` while writing directly to `http.ResponseWriter` (with `WriteTimeout: 0`), allowing a single slow client to stall the 60 Hz tick loop and all other routes.
  - **Fix:** Snapshot or encode response bytes into a buffer under `s.mu`, release `s.mu.Unlock()`, then write headers and body to `w`.
  - **Proof:** Test with a blocking/slow `http.ResponseWriter` verifying concurrent routes and ticks remain unblocked under `go test -race ./internal/web`.

- [ ] `HARD-WEB-02` Cap SSE `/events` subscribers, fix `closed` header race, validate pre-bound `opts.Listener`, avoid emitting plaintext password values under remote bind, and add `sameOrigin(r)` to `/debug/state` (`internal/web/events_http.go:25-103`, `serve_build.go:32-41, 85`, `debug_http.go:27`, `document.go:250-251`).
  - **Root cause:** `/events` allows unlimited concurrent subscribers and writes `"retry: 3000\n\n"` before checking `s.subscribe(1)`; `checkBindAddr` skips loopback check when `opts.Listener != nil`; `debug` omits `sameOrigin(r)` (defense-in-depth); `fmt.Println("http://"+addr+"/")` prints `:0` on ephemeral ports; `document.go:250-251` emits plaintext `value="..."` on `<input type="password">` in `GET /` and `GET /dom`.
  - **Fix:** Enforce a reasonable subscriber cap (return `503`), subscribe before writing headers, validate `opts.Listener.Addr()` loopback when `!opts.InsecureAllowRemoteBind`, add `sameOrigin(r)` in `debug`, omit/redact password values from serialized HTML/DOM JSON (or gate strictly), and log `ln.Addr().String()`.
  - **Proof:** Unit tests in `internal/web/` covering subscriber cap, closed-server `/events`, remote pre-bound listener rejection, password redaction, and cross-origin `/debug/state`.

- [ ] `HARD-DBG-01` Fix debug server `WriteTimeout` killing default 30s `pprof` captures and add shutdown timeout (`internal/window/debug_server_tag.go:21, 51, 74-79`).
  - **Root cause:** `debugWriteTimeout = 15 * time.Second` aborts `GET /debug/pprof/profile` (30s default), and `hs.Shutdown(context.Background())` can hang indefinitely on an open connection.
  - **Fix:** Set `WriteTimeout: 0` (or `>= 60s` / per-route `http.ResponseController`) for pprof compatibility and bound `hs.Shutdown` with a 2–5s `context.WithTimeout`.
  - **Proof:** Test with `-tags ownframe_debug` verifying clean shutdown with an idle open connection and profile endpoint deadline > 30s.

---

## Closure Gates

- [ ] `P1A-GATE-01` `make test` and `make test-race` pass cleanly across all modified packages (`emoji`, `textrun`, `fetch`, `web`, `window`, `replay`, `page`, `atspi`, `ui/virtual`).
- [ ] `P1A-GATE-02` `make lint` and `make graphics-check` pass with exit `0`.
- [ ] `P1A-GATE-03` All 7 fuzz targets (`FuzzReplace`, `FuzzBlockedIP`, etc.) pass a 15s smoke run with zero crashes.
