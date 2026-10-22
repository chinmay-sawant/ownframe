# v0.0.3 - Phase 1 field: Public API

> **Parent:** [README.md](README.md)
> **Status:** open. Depends on Phase 0 freeze so new aliases do not land during the wrap.
> **Estimated effort:** 1-2 weeks. This is the breaking v0 change. Do it before v1.

---

## Overview

Replace `type Page = page.Page` with a defined public type that holds `*page.Page`. Apps see about twenty methods. `internal/page.Page` keeps implementing `host.Screen` for `internal/window` and `internal/web`.

Go practice: defined types over aliases, small exported API, named packages for clipboard/crash/pagetest, no kitchen-sink growth on root.

## Executive summary

- Wrapper type in package `ownframe`.
- One name per idea. Delete synonyms in the same tag.
- Clipboard, crash, `pagetest` as public names.
- Fetch and IPC freeze or move. Do not add display-list mutation to root.

## Current source baseline

| Item | Path | Public today |
|---|---|---|
| Alias | `api.go` `type Page = page.Page` | every `internal/page` exported method |
| Config / Handlers / Box | `api.go` type aliases | keep as aliases **or** defined types copied; prefer defined structs if `page.Config` must change freely |
| Run / BindMobile | `run.go` | keep |
| Serve | `serve.go` | keep, document as debugger in Phase 2 |
| Fetch | `fetch.go` | freeze or `ownframe/fetch` |
| IPC | `ipc.go` | freeze or `ownframe/ipc` |
| Crash | `crash.go` | keep Report / SetCrashDir; typed errors already exist in `internal/crash` |
| Clipboard | no root file | examples import `internal/clipboard` |
| PatchElement | `internal/page/page_patch_element.go` | `bool` |
| Host methods on Page | `TakeDirty`, `PollReload`, `IMEReplace`, `AccessibilityPerform`, … | must leave the public type |

## Target public `Page` methods

Keep (names may match today):

- `SetData`, `Handle`, `AddHandlers`
- `SetTheme`, `SetTick`, `SetImage`
- `Redraw` (ctx first)
- `Load`, `Back`, `Forward`, `Route`, `HTML`
- `SetSize` / current size getters the examples need
- `PatchElement`
- `WantFrameAt` if apps use it; else keep on host only
- `SetScrollWindow`, `SetViewportPinZ` until Phase 3 replaces them
- Form value helpers the login/forms examples use (`FormValue`, focus helpers **one** name)

Do not keep on the public type:

- `TakeDirty`, `FallbackDisplay`, `SetDrawTime`, `IMEReplace`, `PollReload`
- `AccessibilityPerform` (tree snapshot may stay if documented; perform is host)
- `PatchOp`, `SetWindowing`, `StepScrollWindow` unless an example cannot move
- Synonym methods listed in 1.2

If an example needs a host method, that is a Phase 1 export decision recorded in `API-09`, not a silent keep.

## Checklist

### 1.1 Wrapper

- [ ] `API-01` `api.go` - change `type Page = page.Page` to `type Page struct { inner *page.Page }` (name of field free). `New` / `NewWithOptions` return `*Page`. Proof: `git grep 'type Page ='` empty. `go test ./...` on the library module compiles.
- [ ] `API-02` `internal/page` - keep `*page.Page` as `host.Screen`. Window and web import `internal/page` or `internal/host`, never the public wrapper. Proof: `internal/window` and `internal/web` have zero imports of the root package as a cycle. `var _ host.Screen = (*page.Page)(nil)` stays.
- [ ] `API-03` `run.go` `Run` / `RunWithOptions` / `BindMobile` - take `*ownframe.Page`, pass `inner` to `window.NewGame`. Proof: `go vet` on `./` and `./examples/login`.
- [ ] `API-04` `serve.go` - same unwrap. Proof: `go vet ./internal/web` plus login `-web` still builds (`go vet ./examples/login`).
- [ ] `API-05` unwrap helpers stay unexported (`func (p *Page) screen() host.Screen`). Proof: no exported `Inner()` unless a test package in `ownframe` needs it; prefer `pagetest`.
- [ ] `API-06` `Config`, `Handlers`, `Box` - either keep aliases with a comment "stable public shape" or copy the struct into `ownframe`. Do not leave callers unable to name `Handlers`. Proof: login `ownframe.Handlers{...}` still compiles.
- [ ] `API-07` `WindowOptions`, `ServeOptions`, `Option`, `WithPerf` - stay public. Proof: `go doc github.com/chinmay-sawant/ownframe` lists them.
- [ ] `API-08` compile-fire test `api_export_test.go` in package `ownframe_test` (external test) that only imports `ownframe` and constructs a page, SetData, Handle, Redraw. Proof: test passes and the file has zero `ownframe/internal` imports.
- [ ] `API-09` inventory of example call sites that fail after the wrap. Record the list in this row when the wrap PR is cut. Each extra method that must stay public gets a one-line reason. Proof: the PR description lists every extra export.

### 1.2 One name per idea (delete in the wrap tag)

Keep the first name. Delete the rest. Update tests in the same PR.

- [ ] `API-10` `internal/page/form_focus_last.go` - keep `LastFocusID`. Remove `PreviousFocusID`, `BlurredFocusID`, `LastFocused`. Proof: `git grep PreviousFocusID` empty in `.go` files.
- [ ] `API-11` `internal/page/page_patch_state.go` - keep `State`, `SetState`, `BindState`, `SyncState`. Remove `StateStore`, `Store`, `SetStateStore`, `SetStore`, `AttachState`, `BindStore`, `AttachStore`, `SyncChangedProperties`, `ChangedPropertiesFromStore`. Proof: grep empty for those names as methods.
- [ ] `API-12` `internal/page/page_patch_props.go` - keep `ChangedProperties`, `TakeChangedProperties`. Remove `ChangedProps`, `LastChangedProperties`, `LastChangedProps`, `TakeChangedProps`, `PropertiesChanged`, `Changed`. Proof: grep empty.
- [ ] `API-13` `internal/page/page_element.go` - keep `Element`. Remove `ElementHandle`, `RetainedElement` as a public alias (implementation may stay unexported), `RetainedNode`, `HandleElement`. Proof: public `go doc` shows one element getter.
- [ ] `API-14` sweep `Deprecated: use` comments in `internal/page`. After deletion, those comments go with the funcs. Proof: no deprecated synonym left.

### 1.3 Named packages instead of Page junk

- [ ] `API-15` public clipboard. Preferred: `package ownframe` wrappers `Read` / `Write` / `UseMemory` **or** `package clipboard` at `github.com/chinmay-sawant/ownframe/clipboard` if that path can be public (it cannot if it lives under `internal/`). Move or wrap `internal/clipboard` so `examples/login` imports `ownframe` or a non-internal path. Proof: `examples/login/login/login_logout.go` has zero `internal/clipboard`.
- [ ] `API-16` `UseMemory` stays a test seam, documented, not `init()`. Proof: clipboard tests still call `UseMemory`; no OS clipboard in `go test`.
- [ ] `API-17` crash stays `Report`, `SetCrashDir`, `LastReportPath` on the root or `ownframe/crash`. Keep path validation in `internal/crash`. Proof: `go doc` shows them; tests reject empty, control, dot-segment, traversal.
- [ ] `API-18` `ownframe/pagetest` (new public package) with find-box, Click, Type helpers copied from example test patterns. Library tests may keep using `internal/page` directly. Example tests migrate when they touch the wrapped type. Proof: login tests compile against `pagetest` or public `Page` methods only.
- [ ] `API-19` secret helper used by `login_state.go` (`internal/secret`) gets a public name or login stops importing it (inline). Proof: `git grep internal/secret examples/login` empty.

### 1.4 Kitchen sink

- [ ] `API-20` `fetch.go` - freeze the current funcs. New fetch API goes in `ownframe/fetch` or stays in `internal/fetch` with the existing root wrappers only. Do not add cookies, cache, or extra methods on root. Proof: `git log` after freeze has no new exported names in `fetch.go` unless this row is updated.
- [ ] `API-21` `ipc.go` - same freeze. In-process bus is not a 1.0 feature. Proof: no new exported IPC names.
- [ ] `API-22` display-list mutation (`PatchOp`, `Display`, `DisplayOp`) - leave on `internal/page` or a clearly experimental comment. Do not add more root aliases. Proof: `display.go` / `display_patch.go` either stay documented as experimental or move off root in this PR.

### 1.5 Types callers can name

- [ ] `API-23` `Stats()` - re-export `type Stats` in `ownframe` **or** unexport the method on the public type. Do not return `host.Stats` from an exported method. Proof: `go doc` of public `Page.Stats` shows `ownframe.Stats` or the method is gone.
- [ ] `API-24` `State()` - re-export a store type **or** hide it. Callers of `*state.Store` in examples must compile. Proof: `go vet ./examples/...` after the wrap.

## Dependencies

- Phase 0 gates (`BLEED-04` allowlist) shrink as synonyms die. Update the allowlist in the wrap PR.
- File-size cap: wrapper methods that only delegate can live in small files. Do not create synonym files.
- Examples module will not compile against published v0.0.2 after this. Coordinate [phase-01-modules.md](phase-01-modules.md) `MOD-01` in the same release train.

## Risks

- `type Page =` means every exported `page.Page` method is already user API. The wrap is a break. Tag notes must say so.
- Window code that type-asserts `*page.Page` from `host.Screen` keeps working. Window code that imported `ownframe` (it should not) would cycle.
- `AddHandlers` already exists (`page_add_handlers.go`). Keep it. Do not invent a second merge API.

## Closure gates

- [ ] `make lint` exit 0
- [ ] `make test` exit 0
- [ ] `go vet -p 1 ./... ./examples/...` exit 0
- [ ] External test `API-08` green
- [ ] `git grep 'type Page ='` empty
- [ ] `git grep 'ownframe/internal' examples/login` empty
