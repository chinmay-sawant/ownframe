# v0.0.3 - Phase 03: Public API Boundaries and Developer Ergonomics

> **Parent:** [README.md](README.md)
> **Status:** In progress (11/22 done, 1 partial `[~]`: `API-01`–`API-05`, `API-07`, `API-11`, `API-12`, `API-14`, `API-GATE-01`, `API-GATE-02`; `[~]` `API-09`)
> **Estimated effort:** 4–6 days

---

## Overview

Package `ownframe` currently spans **63 root `.go` files** and **149 exported methods** across `Page` and seven "door" wrapper types (`Form`, `Edit`, `Scroll`, `Window`, `Access`, `Probe`, `Testing`). While the `ownframe.Page` wrapper protects `*page.Page` from direct exposure, three ergonomic and architectural defects remain:

1. **Process-global window state leaks through an instance method:** In `page_window_door.go`, `w.CursorPosition()`, `w.Minimize()`, `w.ToggleMaximize()`, `w.Maximized()`, and `w.Close()` ignore `w.page` completely and call singleton `window.*` functions—blurring page-scoped state and process-global Ebiten host state.
2. **Micro-file sprawl at module root:** Dozens of 5–25 line files (`page_click_api.go`, `page_take_scroll.go`, `page_wrap.go`, `display_patch.go`, `drop.go`, `hover.go`, `page_draw.go`, `page_frame_api.go`, `page_geom_api.go`, `page_scroll_api.go`, `page_size_api.go`) force readers to jump across 63 root files just to understand the public surface.
3. **Four-path update mental model without high-level guardrails:** Real apps (`examples/code-view`, `examples/telegram`) have to manually juggle `SetData`, `PatchElement`/`PatchElements`, `SetScrollWindow` + counter-positioned pinned headers, and retained-op motion (`internal/frame` vs CSS motion) with silent footguns when mixed.

## Executive Summary

- Decouple process-global desktop window controls from `page.Window()` (or bind them cleanly to an active host handle) and document multi-page/headless behavior.
- Consolidate the 63 root `.go` files by domain and audit all 149 exported methods to remove/merge redundant surface while keeping gate tests green.
- Lift repeatable virtualization (`SetScrollWindow` + spacer/pinned-bar pattern) and update-mode guidance into reusable public helpers, and guard against dual-driving an element from CSS motion and manual `SetTick`.
- Drain the 16 remaining entries in `scripts/example-internal-allowlist.txt`.

## Current Source Baseline (`da4f12cc`)

| Surface | Baseline Measurement | Key Files / Evidence |
|---|---|---|
| Root Go files | **63 `.go` files** (non-test + test) | `api.go`, `page_*_api.go`, `page_*_door.go`, `page_testing_*.go`, `drop.go`, `hover.go`, etc. |
| Exported receiver methods in root `*.go` | **149 methods** | Spread across `Page`, `Form`, `Edit`, `Scroll`, `Window`, `Access`, `Probe`, `Testing` |
| Global calls on `Page.Window()` | **5 methods ignore `w.page`** | `page_window_door.go`: `CursorPosition` (L19), `Minimize` (L184), `ToggleMaximize` (L190), `Maximized` (L196), `Close` (L204) |
| Example `internal/` imports | **16 frozen pairs** | `scripts/example-internal-allowlist.txt` (15 in `examples/css-preview`, 1 in `examples/desktop-cat/cat/preview.go`) |

---

## Phase 3: Public API Boundaries and Developer Ergonomics

### 3.1 Separate process-global window controls from page-scoped `Window` door

- [x] `API-01` Audit every method in `page_window.go` and classify each as **page-scoped** (`PollReload`, `SetAnimations`, `Animations`, `SetTransitions`, `Transitions`, `SetPerf`, `Perf`, `SetDevTools`, `DevTools`, `SetTouchZoomAllowed`, `TouchZoomAllowed`, `WheelZoomAllowed`, `KeyZoomAllowed`, `SetTouchScrollOptions`, `SetHotReload`, `Watching`) vs **host-bound desktop controls** (`CursorPosition`, `Minimize`, `ToggleMaximize`, `Maximized`, `Close`). Evidence: all 21 methods classified; only those 5 called `internal/window`.
- [x] `API-02` Design and implement a clean host-control seam so page-scoped configuration and desktop-host window actions are not conflated: added unexported `windowHost` interface and `bindHost`/`hostControls` on `ownframe.Page` (`api.go`), attached via `desktopWindowHost` in `RunWithOptions` (detached with `defer page.bindHost(nil)`) and `NewGame` (`run.go`), with zero new methods on `internal/page.Page`. Detached pages, `Serve`, and headless tests safely no-op without mutating global Ebiten window state. Evidence: `TestPageWindowHostControlsDetachedAndBound` in `page_window_host_test.go` passed under `go test -race .`.
- [x] `API-03` Remove `page_window.go`'s direct calls to global `window.Minimize()`, `window.ToggleMaximize()`, `window.Maximized()`, `window.RequestClose()`, and `window.ContentCursorPosition()` in favor of `w.page.hostControls()`. Evidence: dropped `internal/window` import from `page_window.go`; `grep -n "window\." page_window.go` returns empty.
- [x] `API-04` Update `documentation/window.md` to document bound-host semantics while preserving all 25 existing example call sites without breaking changes. Evidence: updated `documentation/window.md`; `make build BUILD_P=4` and `make docs-check` exit 0.
- [x] `API-05` Add a host-control unit test covering nil page, detached page, and bound fake host (`Minimize`, `ToggleMaximize`, `Maximized`, `Close`, `CursorPosition`) without needing a live display server. Evidence: `TestPageWindowHostControlsDetachedAndBound` in `page_window_host_test.go` passes under `go test -race .`.

### 3.2 Consolidate the 63 root `.go` files and audit the 149 exported methods

- [ ] `API-06` Generate a complete method inventory of all 149 exported methods in root `*.go` (`go doc -all .`), marking each as: Core App API, Advanced/Escape Hatch, Testing-Only, or Redundant/Candidate-for-Removal. Proof: inventory attached or committed in review notes.
- [x] `API-07` Consolidate single-function micro-files at the module root into cohesive domain files using `scripts/merge-go.go` (48 non-test `.go` files -> **18** cohesive files: `api.go`, `clipboard.go`, `crash.go`, `elements.go`, `events.go`, `fetch.go`, `ipc.go`, `page.go`, `page_access.go`, `page_edit.go`, `page_form.go`, `page_nav.go`, `page_probe.go`, `page_scroll.go`, `page_testing.go`, `page_window.go`, `run.go`, `serve.go`). Evidence: `ls -1 *.go | grep -v _test.go | wc -l` = 18; updated `alias_allow_test.go` filenames (same 31 aliases); `go test -race .` (`TestRootAliasAllowlist`, `TestOwnframeTestNoInternal`), `make docs-check`, and `make build` exit 0.
- [ ] `API-08` Identify and remove or consolidate redundant exported methods and redundant `Probe()` / `Testing()` overlaps before locking `v0.1.0` / `v1.0` (remember `v0.0.x` may break; document any breaking renames in release notes). Keep `alias_gate_test.go` and `api_internal_gate_test.go` passing and tighten allowlists as aliases shrink. Proof: `go test -v -run 'TestAlias|TestAPI' .` exits 0 with reduced method count recorded.
- [~] `API-09` Verify `Page.Testing()` is strictly unused in `examples/` and non-test library files via an automated check. Evidence (partial): added AST-based `TestNoUnauthorizedPageTestingCalls` in `api_testing_test.go` enforcing zero non-test `.Testing()` calls across `internal/`, `components/`, `ui/`, and freezing the 5 pre-existing experimental example harnesses (`examples/blackhole/main.go`, `examples/blinkless-all/nav_scripted.go`, `examples/blinkless-all/snapshot.go`, `examples/blinkless-scroll/demo.go`, `examples/blinkless-scroll/main.go`) so no new non-test example can call `.Testing()`. Next gate: refactor those 5 experimental harnesses off `.Testing()` and empty `allowExamples`.
- [ ] `API-10` Ensure `go doc github.com/chinmay-sawant/ownframe` reads cleanly with package-level overview grouped by core workflow (`New` -> `SetData` -> `Handle` -> `Run`). Proof: manual review of `go doc .` output.

### 3.3 Simplify the 4-path update and virtualization mental model

- [x] `API-11` Write a concise decision-tree guide (`documentation/updates.md`, linked from `documentation/README.md`) answering "Which update API do I call?" covering `SetData`, `PatchElement`/`PatchElements`, `ui/virtual.Window` + `SetScrollWindow`, and CSS motion vs `SetTick`. Evidence: created `documentation/updates.md` and linked in `documentation/README.md`; `make docs-check` exits 0.
- [x] `API-12` Extract fixed-row virtualized-list windowing math into a reusable, well-tested helper package `ui/virtual` (`Range`, `Window(scrollY, viewportH, rowH, totalRows, overscan)` with top/bottom spacer calculations). Evidence: created `ui/virtual/virtual.go` and `ui/virtual/virtual_test.go`; `go test -race -v ./ui/virtual/...` passed (empty/invalid, top/mid/bottom clamping, spacer invariant, small list, resize).
- [ ] `API-13` Add a pinned-chrome helper or documented pattern in `ui/` so apps don't have to hand-roll fragile `SetScrollWindow` counter-offsets for sticky headers/composers (and link to `blinkless` sticky support once adopted per `plans/v0.0.4/04-css-conformance-and-engine-adoption.md`). Proof: helper + test or verified sticky replacement documented.
- [x] `API-14` Add a defensive runtime diagnostic counter when `Page.SetTick` mutates retained ops on an element that also has an active automatic CSS animation or transition (`MotionStats.ManualOverlap` and `ManualOverlapID`, checked via `motion.Runtime.CheckManualOverlap` during `noteTickDamage`). Evidence: added `ManualOverlap` / `ManualOverlapID` in `internal/motion/stats.go`, `CheckManualOverlap` in `internal/motion/runtime.go`, wired in `internal/page/page_tick.go`, and verified with `TestTickManualOverlapDiagnostic` in `internal/page/motion_overlap_test.go` (`go test -race` passed).
- [ ] `API-15` Document or eliminate the one-tick delay quirk when 3D transforms auto-enable `Config.RetainState` (ensure callers either opt in explicitly via `Config` or see clear documentation + probe visibility). Proof: test covering 3D transform first-tick behavior and updated `documentation/css-motion.md`.

### 3.4 Drain the example `internal/` import allowlist

- [ ] `API-16` Audit the 15 `examples/css-preview/*` entries in `scripts/example-internal-allowlist.txt` (`internal/animation`, `cssanimation`, `cssmotion`, `csstransition`, `csstrigger`, `viewtransition`): relocate engine-internal unit/conformance tests into `internal/motion/` (or `internal/viewtransition/`) tests so `examples/css-preview` exercises only public `ownframe` APIs. Proof: `examples/css-preview` entries removed from `scripts/example-internal-allowlist.txt` and moved tests pass in the library module.
- [ ] `API-17` Re-evaluate the final remaining entry (`examples/desktop-cat/cat/preview.go` -> `internal/web`): either expose a minimal public `ServeOptions` / custom-preview hook if broadly useful, or document why it remains the sole explicit allowlist entry. Proof: allowlist reduced from 16 entries to <=1 entry.
- [ ] `API-18` Run `make internal-import-check` and `make examples-check` to lock in the reduced allowlist. Proof: both commands exit 0.

---

## Dependencies

- `API-02` / `API-03` should complete before `Phase 05` refactors `internal/window` so window host bindings don't get rewritten twice.
- `API-16` coordinates with `Phase 04` (`internal/motion` consolidation) — relocating `examples/css-preview` internal tests either precedes or lands alongside `Phase 04` package merges so import paths only change once.

## Risks

- **Breaking `examples/` or downstream snippets:** Any method renamed or moved off `Page.Window()` must be updated across `examples/`, `documentation/`, and `scripts/check-doc-snippets.py` in the same PR.
- **Over-exposing internals to drain allowlist:** Do NOT add ugly public forwarding APIs solely to empty `scripts/example-internal-allowlist.txt` (respecting the `INTERNAL-IMPORTS` principle in `plans/README.md`). Move internal-testing code into library tests instead.

## Closure Gates

- [x] `API-GATE-01` Zero process-global `window.*` singleton calls left on `Page.Window()` without explicit page-to-host binding. Evidence: `page_window.go` no longer imports `internal/window`; `TestPageWindowHostControlsDetachedAndBound` passes under `-race`.
- [x] `API-GATE-02` Root non-test `.go` files consolidated to <=18; `TestRootAliasAllowlist` and `TestOwnframeTestNoInternal` gates pass. Evidence: 48 -> 18 root non-test `.go` files; `go test -race .` exited 0.
- [ ] `API-GATE-03` `scripts/example-internal-allowlist.txt` reduced from 16 pairs to <=1 pair.
- [ ] `API-GATE-04` `make test`, `make test-race`, `make lint`, `make docs-check`, and `make internal-import-check` exit 0.
