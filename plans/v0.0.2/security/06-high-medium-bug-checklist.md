# Second pass - high and medium severity bug checklist

> **Parent:** `plans/v0.0.2/security/05-security-phase-checklist.md`
> **Status:** IPC/shared-state audit completed; rows below include those findings
> **Estimated effort:** multi phase

## Overview

This ledger captures the second review pass requested after the initial security audit. The first hardening pass fixed the items recorded in `05-security-phase-checklist.md`. This checklist tracks the new high/medium findings, any prerequisite decisions, and the verification needed before they can be marked complete.

## Phase 1: Web preview and network boundary

- [ ] Gate plain `/debug/state` or reduce its payload so box text and layout details are not exposed by default. Affects: `internal/web/serve_options.go`, `internal/web/debug_http.go`.
- [ ] Make `/click` POST-only or add origin and CSRF checks before mutating route state. Affects: `internal/web/serve_options.go`, `internal/web/input_http.go`, `internal/page/page_click.go`.
- [ ] Default or validate `Serve` bind addresses so empty or port-only values cannot bind all interfaces. Affects: `internal/web/serve_options.go`, `examples/desktop-cat/notify.go`.
- [ ] Add `http.MaxBytesReader` to `/backspace` and tighten request body handling. Affects: `internal/web/input_http.go`.
- [ ] Add panic recovery middleware around Serve handlers and document the boundary. Affects: `internal/web`, `serve.go`.
- [ ] Add an origin check for browser notification and preview APIs, or document localhost-only use. Affects: `examples/desktop-cat/cat`.
- [x] HTTP timeouts, nosniff, framing, finite coordinates, and `/pdf` safe assertion were added in the first hardening pass. Evidence: `internal/web/serve_options.go`, `internal/web/input_http.go`, `internal/web/pdf_http.go`.

## Phase 2: Fetch and SSRF policy

- [ ] Bound or reject oversized response headers, because only the body currently has a cap. Affects: `internal/fetch/do.go`.
- [ ] Reject or sanitize userinfo before redirects, especially HTTPS to HTTP transitions. Affects: `internal/fetch/scheme.go`.
- [ ] Decide and document whether host/IP controls belong in ownframe or in the caller. Affects: documentation and `internal/fetch`.
- [x] Default timeout and body cap were added in the first hardening pass. Evidence: `internal/fetch/scheme.go`, `internal/fetch/do.go`.
- [ ] Update `documentation/fetch.md` so it matches the implementation.

## Phase 3: Path and command boundaries

- [ ] Add `--` handling for `xdg-open`, matching `lp` and `open`. Affects: `internal/print/commands.go`.
- [ ] Either reject or validate arbitrary `SavePDF` paths before creating or truncating the destination. Affects: `internal/page/print.go`.
- [ ] Reject or clean Unicode line separators and `.` dot segments in `validPath` where helper output is reused. Affects: `internal/print/command_args.go`.
- [ ] Document `CrashDir` as requiring a caller-chosen private directory and preserve a clean preflight in examples. Affects: `internal/crash/validate.go`, examples.
- [ ] Keep helper executable discovery documented as environment-trusted behavior. Affects: docs.

## Phase 4: Template loading and rendering trust

- [ ] Document and enforce the trusted-only boundary for `New`, `Load`, `Route`, `Config.HTML`, `Config.File`, and themes. Affects: `internal/page/page_nav.go`, `internal/page/page_new.go`, `internal/page/theme.go`.
- [ ] Add a plain HTML parse option or helper that does not execute Go templates for runtime content.
- [ ] Add source size guards to `Load` and `Route`, matching `MaxSourceSize`. Affects: `internal/page/page_source_limit.go`.
- [ ] Fuzz the form HTML rewriters, button reconstruction, emoji replacement, and attribute unescaping. Affects: `internal/page/form*.go`, `internal/emoji/emoji_replace.go`.
- [ ] Define and implement template-versus-user state precedence so Redraw cannot become stale after binds.
- [~] Dangerous link schemes are already rejected and tested. Evidence: `internal/page/page_link.go`, `internal/page/page_deeplink_schemes.go`.

## Phase 5: Concurrency and shared state

- [ ] Make `frameAt` atomic or otherwise synchronized across game, handler, and tick threads. Affects: `internal/page/page_frame_deadline.go`.
- [ ] Serialize or document all concurrent reads and writes of `Page` display fields such as `png`, `img`, `display`, `width`, `height`, `generation`, and dirty rects. Affects: `internal/page/page_image.go`, `internal/page/page_draw.go`.
- [ ] Guard lazy initialization of `pending` and `last` maps. Affects: `internal/page/page_invalidate.go`.
- [ ] Define thread-safe contract for `render.Cache` reads versus `Relayout`. Affects: `internal/render/cache.go`.
- [ ] Synchronize assignment/read of the global `RuntimeSnapshot` function between window and HTTP handler paths. Affects: `internal/web/debug_http.go`.
- [ ] Split eviction results so an oversized replay cache entry cannot be evicted and returned as a disposed image reference. Affects: `internal/replay/chunk_cache.go`, `internal/replay/raster_cache.go`, `internal/replay/text_raster_ops.go`.
- [ ] Cap IPC registrations per channel or process, and document that listener lists grow until their cancel runs. Affects: `internal/ipc/ipc.go`, `internal/ipc/listen.go` if split from `ipc.go`.
- [ ] Add per-call handler timeouts or document the expected non-blocking handler contract. Affects: `internal/ipc/request.go`, `internal/ipc/send.go`.
- [ ] Validate and limit channel name length and ownership.
- [ ] Make the global `RuntimeSnapshot` hook safely replaceable or add a guarded registration API. Affects: `internal/web/debug_http.go`.
- [ ] Review long-running font cache eviction and pool retention decisions.

## Phase 6: Panics, nil safety, and template execution

- [ ] Add panic recovery or explicit contracts around template execution during Redraw and initial indexing. Affects: `internal/page/page_draw.go`, `internal/page/page_nav.go`.
- [ ] Document that web apps, not ownframe, are responsible for isolating per-request state and avoiding nil page objects.
- [ ] Keep `Run` and `BindMobile` recovery behavior documented for mobile callers.
- [ ] Add no-recovery boundary notes for `Serve`, unless a deliberate recovery path is implemented. Affects: `serve.go`.

## Phase 7: Privacy, clipboard, devtools, and debug

- [ ] Gate or reduce `/debug/state`, especially `Boxes` text and dimensions, unless debug is explicitly enabled.
- [ ] Document and discourage displaying `host.Drop.Path` in app UI.
- [ ] Add a clear user action requirement or setting before devtools writes the OS clipboard.
- [ ] Update `documentation/crash.md` so it matches the `0700`/`0600` implementation.
- [ ] Keep debug environment flags off in shipped binaries.

## Phase 8: Resource leak and mobile follow-up

- [ ] Bound image decoding and UI-thread work in the Telegram mobile camera/picker path.
- [ ] Bound Go-side image decode before allocation in `telegram/photo.go`.
- [ ] Track and evict cached captured photos if retention beyond one message is not needed.
- [x] No unbounded leak found in `internal/web`, `internal/window`, `internal/fetch`, or `internal/render`. Evidence: bounded raster, text, image, and chunk caches.
- [ ] Run targeted tests and add fuzz coverage for new boundaries.
- [ ] Reconcile this checklist with the IPC audit output and re-run `make test`.

## Dependencies and ordering

1. Web and template boundaries should be sequenced before route hardening.
2. Concurrency fixes should land before changing Serve behavior.
3. Mobile image and render follow-up can proceed independently.

## Verification gate

- [ ] Update documentation after each closed item.
- [ ] Run `go test -p 1` for affected packages after each phase.
- [ ] Run `make test` before marking a broad phase complete.
- [ ] Recompute finding counts and open rows before declaring the pass finished.