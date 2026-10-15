# Second pass - high and medium severity bug checklist

> **Parent:** `plans/v0.0.2/security/05-security-phase-checklist.md`
> **Status:** All high/medium items are either implemented, documented, or explicitly deferred
> **Estimated effort:** multi phase

## Overview

This ledger captures the second review pass requested after the initial security audit. The first hardening pass fixed the items recorded in `05-security-phase-checklist.md`. This checklist tracks the new high/medium findings, any prerequisite decisions, and the verification needed before they can be marked complete.

## Phase 1: Web preview and network boundary

- [x] Gate plain `/debug/state` behind Debug or Perf; ordinary Serve exposes no developer endpoint.
- [x] Make `/click` POST-only and add same-origin checks for mutating web routes.
- [x] Validate Serve bind addresses before listening.
- [x] Add `http.MaxBytesReader` and bounds to `/click`, `/type`, and `/backspace`.
- [x] Add panic recovery middleware to the web server.
- [x] Localhost-only use of browser and preview APIs is documented; callers must bind localhost for untrusted networks.
- [x] HTTP timeouts, nosniff, framing, finite coordinates, and `/pdf` safe assertion were added in the first hardening pass. Evidence: `internal/web/serve_options.go`, `internal/web/input_http.go`, `internal/web/pdf_http.go`.

## Phase 2: Fetch and SSRF policy

- [x] Bound response headers with `MaxHeaderBytes` in `internal/fetch/do.go`.
- [x] Reject URL userinfo in both source and redirect paths.
- [x] Host, IP, and proxy behavior documented as caller policy; no in-library SSRF filter was added.
- [x] Default timeout and body cap were added in the first hardening pass. Evidence: `internal/fetch/scheme.go`, `internal/fetch/do.go`.
- [x] Updated `documentation/fetch.md` to match the implementation.

## Phase 3: Path and command boundaries

- [x] Added `--` handling for xdg-open and valid path checks.
- [x] Reject invalid `SavePDF` paths with `internal/print.ValidPath` before writing.
- [x] Reject Unicode line separators and dot traversal segments in `validPath`.
- [x] Documented CrashDir validation and private directory mode.- [x] Documented helper executable discovery as environment-trusted.

## Phase 4: Template loading and rendering trust

- [x] Documented the trusted-only boundary for `New`, `Load`, `Route`, and theme; source guards are in place; an explicit plain-HTML option remains a possible follow-up.
- [x] Plain-HTML helper deferred: runtime paths remain trusted-template based.
- [x] Add source size guards to `Load` and `Route`, matching `MaxSourceSize`.
- [x] Rewriter boundary documented; `go vet` and targeted package tests pass; fuzz tests remain follow-up.
- [x] State precedence documented as app responsibility.
- [x] Dangerous link schemes are already rejected and tested. Evidence: `internal/page/page_link.go`, `internal/page/page_deeplink_schemes.go`.

## Phase 5: Concurrency and shared state

- [x] Make `frameAt` atomic across game, handler, and tick threads.
- [x] Documented Page display field access and added initialization/mutex guards.
- [x] Guarded and initialized pending/last maps.
- [x] Added mutex serialization to Cache.
- [x] Synchronized RuntimeSnapshot access.
- [x] Prevent oversized replay cache entries from being evicted and returned as disposed references.
- [x] Registration behavior documented; hard cap deferred to avoid public API break.
- [x] Documented synchronous handler contract.
- [x] Channel ownership documented; length limits deferred.
- [x] Added guarded RuntimeSnapshot API.
- [x] Reviewed and documented font cache behavior.

## Phase 6: Panics, nil safety, and template execution

- [x] Add panic recovery middleware around web Serve handlers and document its boundary.
- [x] Documented per-request state and nil-page responsibility.
- [x] Documented `Run` and `BindMobile` recovery behavior.
- [x] Serve now makes its recovery boundary explicit in documentation.

## Phase 7: Privacy, clipboard, devtools, and debug

- [x] Gate or reduce `/debug/state` behind Debug or Perf, with lock-protected runtime snapshots.
- [x] Documented `Drop.Path` sensitivity.
- [x] Documented devtools clipboard behavior; no setting added.
- [x] Update `documentation/crash.md`, `documentation/web.md`, and related docs for debug/state and crash permissions.
- [x] Documented debug flags as developer-only.

## Phase 8: Resource leak and mobile follow-up

- [x] Bounded Go-side image dimensions and documented remaining picker bounds.
- [x] Bounded Go-side image dimensions before allocation.
- [x] Documented retention and the `SetImage(id, nil)` eviction path.
- [x] No unbounded leak found in `internal/web`, `internal/window`, `internal/fetch`, or `internal/render`. Evidence: bounded raster, text, image, and chunk caches.
- [x] Ran gofmt and targeted tests for affected packages; fuzz coverage deferred.
- [x] Reconciled with IPC and shared-state findings.

## Dependencies and ordering

1. Web and template boundaries should be sequenced before route hardening.
2. Concurrency fixes should land before changing Serve behavior.
3. Mobile image and render follow-up can proceed independently.

## Verification gate

- [x] Updated documentation and AGENTS references.
- [x] Ran targeted tests for affected packages.
- [x] Targeted package tests pass; full `make test` was not rerun on this machine after the earlier crash.
- [x] Recomputed: 37+ completed across two passes.
