# security - ownframe security findings and hardening plan

> **Parent:** `plans/v0.0.2/README.md` and `plans/v0.0.2/PHASES.md` lineage - security audit
> **Status:** findings documented, fixes partly applied
> **Estimated effort:** multiple phases

## Overview

This ledger records the October 2026 security audit of ownframe. Finding files are under `plans/v0.0.2/security/`. It tracks the audit findings, hardening work already applied, and pending work. Rows stay unchecked until source and tests support the result.

## Executive summary

ownframe has no browser runtime and does not execute template JavaScript. The main risks are untrusted template loading, an unauthenticated web preview, missing HTTP timeouts and body limits, deprecated server hardening, unbounded fetch behavior, local file access paths, and developer surfaces such as crash reports and debug endpoints. Several low risk command injection issues were already fixed. Remaining work is grouped below by risk.

## Phase 1: Template content and runtime loading

### 1.1 Template trust boundary
- [ ] Document in `docs/` that `Page.Load`, `Page.Route`, and template source fields are trusted author content.
- [ ] Add a no template execution parse option or helper.
- [x] Confirm dangerous outbound link schemes are rejected. Evidence: `internal/page/page_link.go:25` and tests there.
- [x] Confirm configured deep link schemes block JavaScript and other banned schemes. Evidence: `internal/page/page_deeplink_schemes.go:21`.

### 1.2 Post template HTML rewriters
- [ ] Fuzz `rewriteButtons`, control scan code, and emoji replacement with malformed and escaped attributes.
- [ ] Add regression tests for double escaped quotes and nested `input` attributes.

## Phase 2: Web server and network exposure

### 2.1 HTTP server
- [ ] Use `http.Server` with read, write, idle, and header timeouts.
- [ ] Add `MaxBytesReader` to `/type`.
- [ ] Reject non finite `/click` coordinates.
- [ ] Change `/click` from GET to POST or add origin and CSRF checks.
- [ ] Add `/debug/state` gating or payload reduction.
- [ ] Add defensive HTTP headers.
- [ ] Make `/pdf` type assertion safe.
- [ ] Recover or document panics in `Serve`.
- [x] Identify the web server surface. Evidence: `internal/web/serve_options.go` and related routes.

### 2.2 Fetch and XHR
- [ ] Add a default client timeout.
- [ ] Limit response sizes.
- [ ] Document host validation responsibilities.
- [ ] Add tests for timeouts, limits, and redirects.
- [ ] Decide whether SSRF controls belong in the library.

## Phase 3: File system, print, and commands

### 3.1 File paths
- [ ] Document path trust expectations.
- [ ] Add configuration side helper guidance and examples.
- [ ] Consider strict path option or symlink rejection.
- [x] Fix AppleScript path injection. Evidence: `internal/print/commands.go`.
- [x] Fix PowerShell interpolation in `examples/snipping-tool/snip/capture_win.go`.
- [x] Confirm PowerShell single quote escaping is used.
- [ ] Add `--` separators where helper commands accept paths.
- [ ] Add size cap to source reads.

### 3.2 PDF and crash reports
- [~] Fix theme injection into generated PDF source. Done 2026-10-15: `internal/page/print_source.go` now uses `html.EscapeString`. Reason: original gap closed. Next gate: add a regression test for `</style>`.
- [ ] Make crash directory mode `0700` and file mode `0600`.
- [ ] Validate crash directory input.

## Phase 4: Runtime, IPC, and local data handling

### 4.1 IPC
- [ ] Recover panics in `Request`.
- [ ] Document trusted in process boundary.
- [ ] Document non blocking handler expectation.
- [ ] Consider registration limits.
- [ ] Add timeout and cancellation guidance.
- [x] Confirm IPC is not reachable from web routes or mobile bindings.

### 4.2 Drag and drop, clipboard, focus
- [ ] Document `Drop.Path` as sensitive and desktop specific.
- [ ] Add guidance to copy `Drop.Read` bytes inside the handler.
- [ ] Confirm clipboard is not exposed through web endpoints. Coverage: `internal/web/input_http.go` has no clipboard route.
- [ ] Keep devtools off in shipped binaries or document exposure.

## Phase 5: Devtools and release closure

### 5.1 Devtools and debug surfaces
- [ ] Gate local devtools or document that it exposes rendered data.
- [ ] Keep debug environment variables off in shipped binaries.
- [ ] Document localhost binding and application protection expectations for `Serve`.

### 5.2 Verification
- [x] Run `go test -p 1 ./internal/print/... ./internal/filepick/...`. Outcome: pass on 2026-10-15.
- [ ] Run `make test` after all docs and hardening edits.
- [ ] Run targeted tests for web and fetch packages.
- [ ] Re review diffs and update statuses in this ledger.

## Dependencies

- Web server changes may require example updates if routes change behavior.
- Fetch warning guidance affects `examples/fetch`.
- Crash report permissions may affect tests that inspect temporary files.
- Template loading changes may break apps depending on remote template loads.<|end|>