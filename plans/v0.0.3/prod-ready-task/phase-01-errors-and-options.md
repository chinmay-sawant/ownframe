# v0.0.3 - Phase 1 field: Errors, options, and constructors

> **Parent:** [README.md](README.md)
> **Status:** closed 2026-10-22. Lands with or immediately after the Page wrap.
> **Estimated effort:** 2-4 days plus example call-site fixes for `PatchElement`.

---

**Closed:** 2026-10-22 on `feat/feat/rod-readiness`. `make test TEST_P=4` exit 0. `make size-check` clean. `bash scripts/check-example-internal.sh` clean (107 pairs). Login and clipboard examples have zero `ownframe/internal` imports.

Rows that named a 1.0 **non-goal** (Play Store, TalkBack, iOS, AT-SPI/UIA/NSAccessibility, widget kit, engine sticky, Serve-as-product, wasm clipboard) are checked because the README and `documentation/platforms.md` now state that limit. They are not shipped features.



## Overview

Go libraries return `error`. Callers use `errors.Is` / `errors.As`. Constructors take options that already exist (`NewWithOptions`, `WithPerf`). Panic is for bugs. `bool` is for predicates, not operations that fail for reasons.

## Executive summary

- `PatchElement(ctx, id, edits...) error` with sentinels.
- Keep `Config` + functional `Option`. No second builder.
- Wrap with `%w`. Root package holds sentinels apps switch on.
- `SetCrashDir` already validates; keep returning `error`.

## Current source baseline

```go
// internal/page/page_patch_element.go
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) bool {
    if p == nil || p.cache == nil || p.display == nil || id == "" {
        return false
    }
    if !p.retainState {
        return false
    }
    // refuse paths drop cache (rollbackPatch in page_patch_splice.go)
}
```

Existing sentinels at root include `ErrScheme`, `ErrPrivateAddress`, `ErrDowngrade`, `ErrNoHandler`, `ErrPanicked`, nav `ErrNoHistory`. Clipboard `Read() (string, error)` and `ErrUnavailable` already landed.

`NewWithOptions` and `WithPerf` exist (`options.go`). `Config.Validate` rejects bad source/size.

`SetCrashDir` rejects empty, control, dot-segment, traversal.

## Checklist

### 1.1 PatchElement

- [x] `ERR-01` change signature to `PatchElement(ctx context.Context, id string, edits ...ElementEdit) error`. Proof: `go doc` shows `error`; examples compile.
- [x] `ERR-02` `var ErrPatchRefused = errors.New("ownframe: element patch refused")` (wording free) in package `ownframe`. Proof: `errors.Is(err, ownframe.ErrPatchRefused)` in a test.
- [x] `ERR-03` typed reason optional in 1.0: `type PatchError struct { ID string; Reason string }` implementing `Is`. If skipped, document the sentinel only. Record the choice here. Preferred 1.0: sentinel plus `%w` with the engine reason string in `Error()`.
- [x] `ERR-04` nil page, empty id, missing cache, `RetainState` off: return `error`, not `false`. Distinct sentinels optional (`ErrNoRetainState`). Proof: table test.
- [x] `ERR-05` refuse still drops the cached tree (atomic). Document in `documentation/features.md` and `optimizations.md`. Proof: existing refuse tests updated to `error`; cache drop still asserted.
- [x] `ERR-06` css-preview and any `PatchElement` call sites. Proof: `git grep PatchElement` call sites compile.

### 1.2 Wrapping and sentinels

- [x] `ERR-07` public errors use `%w` when wrapping engine or OS errors. Proof: `errorlint` / existing golangci stays green; spot-check `PatchElement` and filepick.
- [x] `ERR-08` keep `ErrNoHistory`, `ErrScheme`, `ErrPrivateAddress`, `ErrDowngrade`, `ErrUnavailable`, `ErrPanicked`. Proof: `go doc` lists them.
- [x] `ERR-09` file picker: public function returns `error` including the fallback reason currently printed when `OWNFRAME_FILEPICK_DEBUG=1`. Proof: test with fake picker; debug env still optional.

### 1.3 Constructors and options

- [x] `ERR-10` keep `New(Config)` and `NewWithOptions(Config, ...Option)`. Do not add a builder type. Proof: `go doc`.
- [x] `ERR-11` keep `WithPerf`. Options that can fail must return `error` from `NewWithOptions` (current `Config.Validate` path). Proof: invalid Config still errors at `New`.
- [x] `ERR-12` no new `init()` in library packages. Audio `ensureAudio` in `run.go` stays a function called from `Run`/`BindMobile`. Proof: grep `^func init\(` in root `*.go` and `internal/**` recorded; new ones need this row updated.
- [x] `ERR-13` clipboard `UseMemory` stays explicit in tests, not `init()`. Proof: TEST-13.

### 1.4 Panic policy

- [x] `ERR-14` `Run` / `BindMobile` recover stays (`savePanic` in `run.go`). Recover converts to `error` plus crash file. Proof: `report_panic_test.go` / `run_report_test.go` still pass.
- [x] `ERR-15` library handlers that currently swallow panic in IPC (`ErrHandlerPanicked`) stay recovered. User click handlers: document whether they recover. Proof: `documentation/crash.md` sentence.
- [x] `ERR-16` no `log.Fatal` in library packages. Proof: grep `log.Fatal` in `internal/` and root `.go` excluding examples.

### 1.5 Context

- [x] `ERR-17` `Redraw`, `Run`, `Fetch`, `XHR`, `PatchElement` keep `ctx` first. Do not add `context.Background()` inside those if the caller passed ctx. Proof: no new `context.Background` in those funcs except documented top-level `Run` samples in README.
- [x] `ERR-18` Fetch already has 30s timeout and size cap. Keep. Do not hide a second timeout that ignores ctx. Proof: `internal/fetch` still respects ctx cancel.

## Dependencies

- Signature change is the wrap tag or the tag immediately after. Examples module replace must point at that tag ([phase-01-modules.md](phase-01-modules.md)).
- css-preview `ClickPaint` path uses patch; update in ERR-06.

## Risks

- `bool` callers `if page.PatchElement(...)` will not compile. That is intended. v0 break.
- Engine refuse reasons may be strings today. Do not invent a stable enum until blinkless has one.

## Closure gates

- [x] `make lint` exit 0
- [x] `make test` exit 0
- [x] `git grep 'PatchElement' -- '*.go'` call sites use `error`
- [x] `errors.Is` test for `ErrPatchRefused`
