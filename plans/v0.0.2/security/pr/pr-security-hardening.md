## Summary

Adds a second security pass across the library. The web preview serves mutating routes over POST with same-origin checks, timeouts, body caps, bind validation, panic recovery, and gated developer endpoints. Fetch gets a default timeout and response caps. Print commands pass paths as arguments, crash reports are private by default, runtime template sources are capped at 8 MiB, and the page, replay, and IPC paths lose a few panics and lifecycle bugs.

## Motivation / context

- Audit and plans under `plans/v0.0.2/security/`: threat model (`00`), five finding files (`01`-`04`), and the two checklists (`05` first pass, `06` second pass).
- No linked issue covers this work.

## Changes

### Web preview and network boundary

- `internal/web/serve_build.go` runs an `http.Server` with ReadHeaderTimeout 5s, ReadTimeout 15s, WriteTimeout 15s, and IdleTimeout 60s, wrapped in panic recovery and security headers.
- Bind addresses with an empty or wildcard host are rejected before listening.
- `/click`, `/type`, and `/backspace` are POST-only, check the request origin, cap bodies at 256 KiB, and reject non-finite click coordinates.
- `/debug/state` is registered only with `ServeOptions{Debug: true}` or `{Perf: true}`; pprof only with `Perf`. `/pdf` uses a checked type assertion.
- The preview shell posts click coordinates and keeps the image map working without JavaScript falling back to a GET.

### Fetch and XHR

- 30 s default client timeout, 10 MiB response body cap, and 1 MiB response header cap.
- URLs with userinfo are rejected on the source request and on redirect targets.
- Host and SSRF policy stays with the caller and is documented.

### Paths, print, and crash reports

- `osascript` receives the PDF path as an argument; `lp`, `open`, and `xdg-open` get `--` before the path. PowerShell paths double single quotes.
- `print.ValidPath` rejects control characters, Unicode line separators, and `.`/`..` segments. `SavePDF` validates before any filesystem call and no longer creates a temp file on the error path.
- Crash directories are created 0700 and report files 0600. `SetCrashDir` validates and cleans its input and keeps the previous directory on a bad value.
- WSL file dialog results are validated before conversion to a Linux path.

### Template loading and rendering trust

- `New`, `Load`, `Route`, `SetTheme`, file-backed sources, and hot reload reject sources over `MaxSourceSize` (8 MiB) with `ErrSourceTooLarge`.
- Theme text is HTML-escaped when injected into generated PDF source.
- Regression tests cover the button rewriter against escaped quotes and nested `input` attributes.

### Runtime and concurrency

- The frame deadline uses atomic access, and the `pending`/`last` dirty maps are initialized and nil-guarded.
- The render `Cache` serializes its methods, and the replay raster, text, and chunk caches stop evicting and returning a disposed entry.
- `RuntimeSnapshot` is lock-protected, and web handlers recover panics.
- `ipc.Request` recovers a panicking handler and returns `ErrHandlerPanicked`.
- The mobile photo path decodes image dimensions first and rejects sides over 4096 px.

### Documentation

- New `documentation/security.md` collects the trust boundaries and residual risks.
- `web.md`, `fetch.md`, `crash.md`, `ipc.md`, `devtools.md`, `drag-drop.md`, `clipboard.md`, `performance.md`, and the feature docs get matching trust sections; `AGENTS.md` reflects the new behavior.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No library performance claim. The hardening adds a body cap and an origin check to mutating preview routes only. |
| **Memory** | Request and response caps bound transient allocations. A replay cache may keep one protected entry above its pixel budget instead of disposing it. |
| **Behavior / correctness** | Fixes nil-map panics in dirty tracking, returns of disposed replay images, and propagates panicking IPC handlers as an error. |
| **API / CLI** | Additive: `MaxSourceSize`, `ErrSourceTooLarge`, `ErrHandlerPanicked`, `ServeOptions.Debug`. `Page.Route` and `SetCrashDir` now return errors. |
| **Dependencies** | None. |
| **Binary size / build time** | Small additions only; no new dependency. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `Page.Route` returns `error` | Statement calls compile unchanged. Check the result when the HTML can exceed 8 MiB. |
| `SetCrashDir` returns `error` and `""` no longer restores the default | Keep the previous directory and pass a concrete path to change it. |
| `GET /click` is gone | Use `POST /click`. The bundled shell already posts. |
| Plain `Serve` no longer serves `/debug/state` | Pass `ServeOptions{Debug: true}` or `{Perf: true}`. |
| Wildcard bind addresses are rejected | Bind a concrete host. The loopback defaults keep working. |
| Sources over 8 MiB are rejected | Shrink or split the template and match `ErrSourceTooLarge`. |
| `SavePDF` can return `ErrBadPath` before `ErrNoPDF` | Pass a path without control characters or `.`/`..` segments. |
| Fetch rejects URL credentials and oversized responses | Remove userinfo from URLs and handle the returned errors. The new fetch sentinels are not re-exported at the module root yet. |
| Existing crash directories are chmod'ed to 0700 | Point `SetCrashDir` at a private directory. |

## Test plan

- [x] `go vet -p 1 ./...` (exit 0)
- [x] `go vet -p 1 ./examples/...` (exit 0; every example compiles against the new API)
- [x] `go test -p 1 ./internal/web ./internal/fetch ./internal/print ./internal/filepick ./internal/crash ./internal/ipc -count=1` (all pass)
- [x] `go test -p 1 ./internal/page` including the new regression tests (pass)
- [x] `gofmt -l` clean on the changed Go files
- [ ] Full `make test` was not run on this branch
- [ ] Rewriter fuzzing and redirect-scheme tests stay follow-ups

### Commands

```sh
go vet -p 1 ./... ./examples/...
go test -p 1 ./internal/web ./internal/fetch ./internal/print ./internal/filepick ./internal/crash ./internal/ipc -count=1
go test -p 1 ./internal/page -run 'Regression|Oversized|StyleBreakout' -count=1
```

## Screenshots / sample output

No UI change. Recorded test output:

```
ok  github.com/chinmay-sawant/ownframe/internal/web      0.012s
ok  github.com/chinmay-sawant/ownframe/internal/fetch    0.210s
ok  github.com/chinmay-sawant/ownframe/internal/print    0.002s
ok  github.com/chinmay-sawant/ownframe/internal/filepick 0.234s
ok  github.com/chinmay-sawant/ownframe/internal/crash    0.003s
ok  github.com/chinmay-sawant/ownframe/internal/ipc      0.002s
```

## Related issues

None. No issue exists for this audit.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] Labels applied (`enhancement`, `documentation`).
- [x] Related issues checked; none exist for this work.
- [x] This PR record is committed under `plans/v0.0.2/security/pr/pr-security-hardening.md`.

## Follow-ups (out of scope)

- No template-execution sandbox or plain HTML parse option; runtime template sources stay trusted author content.
- Missing tests for the new web middleware (`sameOrigin`, `panicRecovery`, `validateBindAddr`, route gating), `ErrHandlerPanicked`, the protected replay eviction, and the photo bounds.
- `Config.File` and `ThemeFile` still follow symlinks; no strict path option.
- IPC has no registration cap and no per-request timeout.
- The new fetch errors and limits are not re-exported at the module root.
- `SetRuntimeSnapshot` has no caller; the window sampler writes `RuntimeSnapshot` directly.
- `SetCrashDir("")` cleanup helpers in `examples/crash` and the root tests still expect the old reset behavior.
- These files went over the repo's 2000-character Go limit on this branch: `api.go` (2044), `internal/page/page_nav.go` (2225), `internal/page/page_new.go` (2372), `internal/render/cache.go` (2538). `internal/page/page_types.go` was already over (2165 before, 2542 now).
- The first-pass ledger `05-security-phase-checklist.md` still shows unchecked boxes for work this branch implements, carries a stray `<|end|>` token at line 98, and should be reconciled with `06`.
- Full `make test` and a rebuilt docs site before release.

## Reviewer checklist

- [ ] Behavior matches the summary and test plan.
- [ ] No unrelated changes in the diff.
- [ ] Public API changes are documented.
- [ ] PR has an assignee and labels.
- [ ] Related issues use the correct closing keywords.
- [ ] No secrets or generated artifacts were committed.
- [ ] Diff statistics were generated from the merge base.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 55 | 1245 | 264 |
| `.md` | 19 | 930 | 19 |
| **Total** | **74** | **2175** | **283** |
