# v0.0.4 - Phase 01: P0 Safety, Security & Concurrency

> **Parent:** [README.md](README.md)
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Priority:** P0 (fix before performance or hygiene refactors)
> **Status:** Open (0 / 12 complete)

---

## Overview

This phase closes every confirmed memory-safety hazard, command/markup injection chain, credential leak, and data race found on `ownframe` `master` (`cf529aba`). Each item has verified file/line evidence and a concrete regression check.

---

## 1.1 Windows UIA COM Memory Safety & Subclass Concurrency

- [ ] `SEC-UIA-01` Fix COM object Use-After-Free and Go GC liveness race in `Host.objFor` and COM callbacks (`internal/uia/host_windows.go:131-154`, `internal/uia/com_windows.go:109-121, 483-485, 537-554`).
  - **Root cause:** `objFor(id)` returns `*comObj` without incrementing `o.ref` under `h.mu`. Between `objFor` and `t.give(...)`, `Publish()` can drop `ref` to `0` and delete `t` from `registry.m`, causing `give` to hand COM a raw slot pointer to an unregistered Go object eligible for GC.
  - **Fix:** Have `objFor` atomically acquire an owning reference while holding `h.mu` (refusing objects already released or when `h.closed` is true), make `give` transfer or retain that ref without a zero-ref gap, and ensure `release` only deletes from `registry.m` when `ref` transitions cleanly to `0` without resurrection.
  - **Proof:** Concurrent `Publish` + `navigate`/`fromPointAt`/`getFocus` stress test under `go test -race ./internal/uia` on Windows (`amd64`/`arm64`).

- [ ] `SEC-UIA-02` Pin `*comObj` lifetime across Win32 syscalls in `raise`, `raiseProperty`, and `wndProc` (`internal/uia/host_windows.go:198-249, 340-347`).
  - **Root cause:** `o.slotAddr(slotSimple)` returns a `uintptr`, after which Go's GC no longer considers `o` live across `procRaiseEvent.Call`, `procRaiseProperty.Call`, `procRaiseStructure.Call`, or `procReturnRaw.Call`; a concurrent `Publish()` or `Close()` can release `o` from `registry.m` mid-syscall.
  - **Fix:** Hold an owning `addRef()` + `defer o.release()` + `runtime.KeepAlive(o)` around every Win32 syscall taking `o.slotAddr(...)` (including `h.rootObj` in `wndProc` and `raise`).
  - **Proof:** Race/stress test exercising concurrent `Close()` / `Publish()` with `raise`/`raiseProperty`/`wndProc` on Windows + `go vet` / `checkptr=2`.

- [ ] `SEC-UIA-03` Fix data race on `Host.orig` during Win32 window subclassing (`internal/uia/host_windows.go:78-93, 349`).
  - **Root cause:** `AttachWindow` inserts `hosts.m[hwnd] = h` and calls `SetWindowLongPtrW` before storing `h.orig = old` without atomics or synchronization, so a message arriving between `SetWindowLongPtrW` and line 93 can call `CallWindowProcW(0, ...)`.
  - **Fix:** Store `orig` as `atomic.Uintptr`, pre-populate via `GetWindowLongPtrW` before exposing `h` in `hosts.m`, and update atomically with the return value of `SetWindowLongPtrW`.
  - **Proof:** Code inspection + `go test -race ./internal/uia` on Windows.

## 1.2 Command Injection, Theme XSS & Credential Leakage

- [ ] `SEC-WSL-01` Fix PowerShell command injection via Unicode smart quotes in the WSL file picker (`internal/filepick/pick_wsl_script.go:7-16`, `internal/filepick/pick_wsl.go:48-50`).
  - **Root cause:** `windowScript` only escapes ASCII `'` (`''`), but PowerShell parses Unicode smart single quotes (`‘` U+2018, `’` U+2019, `‚` U+201A, `‛` U+201B) as string delimiters, allowing breakout into arbitrary PowerShell commands when `title` contains them.
  - **Fix:** Pass `title` out-of-band via `cmd.Env` (e.g., `OWNFRAME_DIALOG_TITLE`) and make the PowerShell script 100% static (`$f.Title = $env:OWNFRAME_DIALOG_TITLE`). Also pass `--title=` (or `--` guard) for `kdialog` in `internal/filepick/pick_linux.go:24`.
  - **Proof:** Unit test in `internal/filepick/pick_wsl_script_test.go` asserting static script + smart-quote titles (`’ ; calc ; ‘`) never interpolate into `-Command`, plus `TestSec*` coverage.

- [ ] `SEC-WEB-01` Close the `-web` theme override `<style>` breakout XSS chain (`ui/values.go:486-546`, `ui/theme/overrides.go:28-45`, `internal/web/document.go:343-357`).
  - **Root cause:** `validateCSSValue` accepts arbitrary characters inside `"..."` / `'...'` (including `"</style><script>alert(1)</script>"`), `Overrides.Stylesheet()` emits it verbatim, and `previewHeadExtra` injects `theme` raw inside `<style id="of-theme">...</style>` with no CSP.
  - **Fix:** Reject case-insensitive `</` sequences inside `validateCSSValue` (and `Overrides.Stylesheet()`) AND escape `</` (e.g., `<\/`) when embedding `theme` inside `<style id="of-theme">` in `internal/web/document.go:346-348`.
  - **Proof:** Regression tests in `ui/`, `ui/theme/`, and `internal/web/` (`TestSecThemeStyleBreakout`) verifying `"</style><script>alert(1)</script>"` is rejected by `ui`/`theme` and neutralized by `previewHeadExtra`.

- [ ] `SEC-PWD-01` Prevent `Copy` and `Cut` from copying plaintext `type="password"` values to the OS clipboard (`internal/page/gesture.go:143-195`, `internal/page/form/control.go:40-46`).
  - **Root cause:** `focusedEditable(p)` accepts `type="password"`, and `Page.Copy`/`Page.Cut` return `c.Value` verbatim, leaking masked passwords onto the system clipboard (inconsistent with `internal/page/a11y/node.go:57` redaction).
  - **Fix:** When `strings.EqualFold(c.Type, "password")`, return `"", false, nil` from `Copy` and refuse copying plaintext on `Cut` (either no-op like browsers or clear without returning plaintext; match standard browser behavior: block both copy and cut on password fields).
  - **Proof:** Unit tests in `internal/page/` asserting `Copy` and `Cut` on a focused `type="password"` input return `("", false, nil)` and never expose the password value.

## 1.3 Web Server, Page & Clipboard Concurrency Fixes

- [ ] `CONC-WEB-01` Fix data race on `s.app.Boxes()` in `POST /click` (`internal/web/click_http.go:72-110`).
  - **Root cause:** `s.app.Boxes()` is iterated at lines 77 and 94 *before* `s.mu.Lock()` at line 109, racing directly with the 60 Hz `tickOnce` goroutine (`tick_loop.go:107`) and concurrent POST handlers mutating `p.boxes`.
  - **Fix:** Acquire `s.mu.Lock()` before resolving `id`/`action` against `s.app.Boxes()` (after cheap query/body parsing and `maxDOMIDLength` validation).
  - **Proof:** Concurrent `POST /click?id=...` + tick-loop race test passing under `go test -race ./internal/web`.

- [ ] `CONC-PAGE-01` Enforce `p.dataMu` lock discipline in `bindWrite` (`internal/page/source.go:298-326`).
  - **Root cause:** `page.go:58-84` specifies `dataMu` guards `p.data`, `p.lastSyncBody`, and `p.formDirty` so concurrent `SetData` never races `Redraw`, yet `bindWrite` reads `p.data` via `bindTarget`, mutates struct fields via reflection, and writes `p.formDirty = true` without holding `p.dataMu`.
  - **Fix:** Acquire `p.dataMu.Lock()` / `defer p.dataMu.Unlock()` in `bindWrite` (and audit `bindTarget` callers so lock acquisition is non-reentrant and consistent).
  - **Proof:** Race test exercising concurrent `SetData` + form edit (`Type`/`Click` on bound inputs) under `go test -race ./internal/page`.

- [ ] `CONC-PAGE-02` Enforce `p.dataMu` on `p.formDirty` and `p.lastSyncBody` in `scroll.go` and `nav.go` (`internal/page/scroll.go:60-79`, `internal/page/nav.go:76-105`).
  - **Root cause:** `SetScrollOffset` (`scroll.go:65`) and `applyScrollWindow` (`scroll.go:77`) write `p.formDirty = true` without `p.dataMu`; `ReplaceHTML` (`nav.go:86, 99`) reads and restores `p.lastSyncBody` without `p.dataMu`.
  - **Fix:** Call `markFormDirty(p)` in `SetScrollOffset` and `applyScrollWindow`, and guard `p.lastSyncBody` snapshot/rollback in `ReplaceHTML` under `p.dataMu`.
  - **Proof:** Race test exercising concurrent `SetData` + `SetScrollOffset` / `ReplaceHTML` rollback under `go test -race ./internal/page`.

- [ ] `CONC-CLIP-01` Lock OS thread in Windows clipboard backend (`internal/clipboard/clipboard_windows.go:56-100`, `internal/clipboard/clipboard_windows_read.go:10-52`).
  - **Root cause:** Win32 `OpenClipboard(0)` binds clipboard ownership to the calling OS thread; Go goroutine migration before `CloseClipboard` causes `CloseClipboard` to fail and can leave the desktop clipboard locked system-wide.
  - **Fix:** Call `runtime.LockOSThread()` / `defer runtime.UnlockOSThread()` in `writeOS`, `readOS`, and `clearOS`, add a brief bounded retry loop (e.g., 3 retries with short backoff) for transient `OpenClipboard` contention, and cap `cmd.Output()` size in `internal/clipboard/clipboard_wayland_run.go:42`.
  - **Proof:** Unit/build verification on Windows + bounded output test for Wayland runner (`clipboard.UseMemory(true)` for non-native unit tests).

## 1.4 WSLg Native X11 Process-Abort Guards

- [ ] `SEC-X11-01` Install a non-fatal Xlib error handler around `ownframe_find_window` / `ownframe_client` (`internal/window/window_drag_native_linux.go:12-30`).
  - **Root cause:** On WSLg, `ownframe_find_window` walks all top-level windows via `XQueryTree` and `XGetWindowProperty` without `XSetErrorHandler`; if any unrelated X window closes mid-walk, Xlib invokes its default fatal error handler and calls `exit(1)`.
  - **Fix:** Install a thread-safe/scoped Xlib error handler (capturing `BadWindow`/`BadMatch` without exiting) around the query walk and restore the previous handler before returning.
  - **Proof:** Review of C glue + Linux cgo build (`go vet ./internal/window`) + WSLg manual verification record if host is available.

- [ ] `SEC-X11-02` Guard `XResizeWindow` against non-positive dimensions (`internal/window/window_drag_native_linux.go:120-143`).
  - **Root cause:** `setNativeWindowRect` passes `C.uint(w), C.uint(h)` directly to `C.XResizeWindow` without checking `w > 0 && h > 0`; zero/negative dimensions trigger fatal `BadValue` from the X server.
  - **Fix:** Return `false` early when `w <= 0 || h <= 0`.
  - **Proof:** Unit test covering non-positive inputs + `go test ./internal/window`.

---

## Closure Gates

- [ ] `P0-GATE-01` All new security and concurrency regression tests pass locally under `make test` and `make test-race` (with `go test -gcflags=all=-d=checkptr=2 ./...`).
- [ ] `P0-GATE-02` `make lint` and `make graphics-check` exit `0` with zero new `//nolint` directives lacking a written reason.
- [ ] `P0-GATE-03` Windows `amd64` / `arm64` CI runs `go test -race ./internal/uia ./internal/clipboard ./internal/window` cleanly (or records explicit host/CI artifact links).
