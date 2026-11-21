# perf(page,motion,window,bitmap): optimize rendering/hover latency, fix translucent backdrop overblend, and bump blinkless to v0.1.3

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `v004/chore-performance-improve` (4 code commits, head `15255071`)
- Release cycle: v0.0.4

---

## Summary

- Upgrades both root `go.mod` and `examples/go.mod` to `github.com/chinmay-sawant/blinkless v0.1.3` (PR #42), bringing ~5× faster dashboard resize/layout and zero-fallback scoped hover patching (`patchBoxes` spare-capacity fix).
- Optimizes OwnFrame's software shape rasterizer, template execution (`draw.Execute`), PNG encoder buffer pooling, motion transition scanning, and batched hover patching (`render.RelayoutElements`).
- Fixes a visual regression where translucent `position: fixed` overlays (such as Command Palette `.app-overlay` and modal backdrops) repeatedly alpha-composited over retained screen pixels on partial/idle frames, fixes Loglens top-bar menu z-index layering above modal backdrops, and adds unit, GPU pixel, and benchmark regression tests registered in `scripts/check-graphics.sh`.

---

## Motivation / context

- Plans: `plans/v0.0.4/pr/pr-performance-improvements.md`
- Interactive profiling of `examples/builder/loglens -dev` and `examples/blinkless-all` exposed three bottlenecks and two visual regressions:
  1. Dashboard layout/resize took ~49.7 ms / 19.4 MB per frame, while hovering Edge Regions cells dropped FPS from ~85 FPS to <30 FPS due to full-layout fallbacks and per-hover motion/transition overhead.
  2. Opening the Command Palette (`⌘K`) or modal overlays over a retained surface (`ebiten.SetScreenClearedEveryFrame(false)`) caused progressive darkening and 256×256 tile checkerboarding on subsequent hover/idle frames, and top-bar menus (`z-index: 60`) painted un-dimmed above modal backdrops (`z-index: 50–55`).
- Issues: none. This work has no tracker issue.

---

## Changes

### 1. Software shape rasterizer & `blinkless-all` perf audit (`706b1c2a`)

- Optimized `internal/bitmap/shape.go` axis-aligned rounded fill and stroke rasterization with interior fast spans and named subpixel sample offsets (`subpixelLo`, `subpixelHi`), removing unused `shapeContains` from `internal/bitmap/shape_contains.go`.
- Added performance audit coverage for `examples/blinkless-all`.

### 2. Pooled PNG encoders & motion `StepLayout` spin fix (`8bd64568`)

- Pooled PNG encoder buffers on bitmap/draw paths to reduce steady-state heap churn.
- Fixed a `StepLayout` busy-spin condition on delayed CSS animations in `internal/motion`.

### 3. Zero-copy template execution & `blinkless` upgrade (`76d56744`)

- Switched `internal/page` template rendering to zero-copy `draw.Execute`, eliminating intermediate buffer copies on every template redraw.

### 4. Batched hover patching, motion fast paths, translucent backdrop fix, and `blinkless v0.1.3` (`15255071`)

- **Engine upgrade (`go.mod`, `examples/go.mod`, `go.sum`, `examples/go.sum`):** Pinned both modules identically to `github.com/chinmay-sawant/blinkless v0.1.3` (no `replace` for `blinkless`).
- **Batched hover patching (`internal/page/page.go`, `internal/page/patch.go`, `internal/page/render.go`):**
  - Defaulted `p.hoverPatch` to `nil` so `listHoverPatched` uses batched `render.RelayoutElements` (enabling `paintOnlyBatchDisplay` and single-pass multi-root relayout while preserving custom `hoverPatch` test hooks).
  - Removed redundant unconditional `updateMotionIndex` calls before `bindMotion` in `commitDisplay` and `PatchElements` (`bindMotion` already updates the motion index only when animations are declared or active).
- **Motion transition fast paths (`internal/motion/transition.go`, `internal/motion/transition_discrete.go`):**
  - Added an O(1) shared-`Boxes` + zero-alloc `hasDisplayTransition` early exit in `noteDisplayExits`.
  - Reused one `records` map (`fillTransitionRecords`) across all boxes in `CaptureTransitions`.
- **Translucent fixed backdrop over-blending fix (`internal/window/replay_tile.go`, `internal/window/replay_viewport.go`):**
  - In `blitFromTiles`, when `s.hasViewportOps()` is true, restore visible cached unfixed tiles via `BlendCopy` (or `showTilesWhole` when content is smaller than the viewport) before `paintPinnedChrome` runs. Clean tiles are never re-rasterized—only cached GPU tile textures are blitted—preventing translucent fixed backdrops/shadows from compositing repeatedly over their own retained pixels on partial or idle frames.
  - Set `options.Blend = ebiten.BlendCopy` in `drawViewport` so transparent windows on the viewport path also replace retained pixels cleanly before fixed/pinned draws.
- **Loglens modal layering (`examples/builder/loglens/menus.go`):**
  - Lowered `#top-nav` and `#top-menubar` from `ZIndex(60)` to `ZIndex(45)` so top-bar menus sit above the topbar (`40`) and dismiss layer (`39`) while dimming cleanly under modal backdrops (`50–55`).
- **Regression tests, benchmarks, and graphics gate (`internal/window/replay_test.go`, `internal/window/game_draw_test.go`, `examples/builder/loglens/backdrop_pixels_test.go`, `examples/builder/loglens/dismiss_test.go`, `examples/builder/loglens/layout_fixes_test.go`, `scripts/check-graphics.sh`, `scripts/test_graphics_gate.py`):**
  - `TestPinnedChromePartialFrameRestoresVisibleTiles` (headless unit test for partial damage, idle-clean pacing, and short-content fallback).
  - `TestTranslucentFixedBackdropNoOverblend` (GPU pixel test on retained surface for `rgba(0,0,0,0.5)` fixed overlay across partial/idle frames and dismiss).
  - `TestCommandBackdropPixelsStableAcrossFrames` (real `loglens` GPU pixel test opening `⌘K`, hovering items across partial/idle frames, and dismissing cleanly).
  - `TestModalBackdropPaintsAboveTopbarMenus` (verifies dialog, settings sheet, and drawer backdrops paint above top-bar menus).
  - `TestEdgeRegionHoverUsesScopedLayout`, `BenchmarkEdgeRegionHover`, and `BenchmarkCommandPaletteHover` (assert zero full-layout fallbacks/errors and track latency/allocs).
  - Registered both new GPU pixel tests in `scripts/check-graphics.sh` and `scripts/test_graphics_gate.py`.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | `examples/builder/loglens` full resize/layout: **49.7 ms -> 8.66–10.76 ms** (~5× faster). Edge-region hover: **19–50 ms full-layout fallback -> 4.12 ms/op scoped patch** (`BenchmarkEdgeRegionHover`, 0 fallbacks). Command palette hover: **4.43 ms/op** (`BenchmarkCommandPaletteHover`, 0 fallbacks). |
| **Memory** | Dashboard redraw allocations cut by >80% via `blinkless v0.1.3` + zero-copy template execution + pooled PNG encoders + reused transition maps (`1.78 MB/op, 3271 allocs/op` on edge-region hover). |
| **Behavior / correctness** | Eliminates translucent fixed backdrop progressive darkening and 256×256 tile checkerboard on retained surfaces; dims Loglens top-bar menus cleanly beneath modal backdrops; eliminates delayed-animation `StepLayout` spin. |
| **API / CLI** | No public Go API changes. |
| **Dependencies** | Upgraded `github.com/chinmay-sawant/blinkless` to `v0.1.3` in both root `go.mod` and `examples/go.mod`. |
| **Binary size / build time** | No new dependencies added. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing `Run`, `Serve`, and `Page` callers keep working without changes. |

---

## Test plan

- [x] `make lint` (all 10 checks: maintenance tooling, linter version regression, Go library and examples `golangci-lint v2.15.0`, docs-check, plans-check, repo-size-check, examples-check, internal-import-check, ui generated properties drift, website ESLint)
- [x] `make test` (all library and example packages)
- [x] `make graphics-check` (all 19 isolated Xvfb GPU pixel sessions including `TestTranslucentFixedBackdropNoOverblend` and `TestCommandBackdropPixelsStableAcrossFrames`)
- [x] `go test -bench "^Benchmark(EdgeRegionHover|CommandPaletteHover)$" -benchtime=10x -benchmem ./examples/builder/loglens`

### Commands

```sh
make lint
make test
make graphics-check
make docs-check plans-check examples-check tools-check repo-size-check
```

---

## Screenshots / sample output

```text
PASS maintenance tooling
PASS linter version regression
PASS Go library and examples (v2.15.0)
PASS documentation snippets
PASS plan states
PASS tracked file size guardrail
PASS example maintenance inventory
PASS example internal imports
PASS ui generated properties drift
PASS website ESLint

BenchmarkEdgeRegionHover-24         10   4118583 ns/op   1785098 B/op   3271 allocs/op
BenchmarkCommandPaletteHover-24     10   4430951 ns/op   1869841 B/op   3572 allocs/op

=== RUN   TestTranslucentFixedBackdropNoOverblend
--- PASS: TestTranslucentFixedBackdropNoOverblend (0.33s)
=== RUN   TestCommandBackdropPixelsStableAcrossFrames
--- PASS: TestCommandBackdropPixelsStableAcrossFrames (1.63s)
```

---

## Related issues

- Relates to `chinmay-sawant/blinkless#42` (`v0.1.3`).

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`)
- [x] Related issues filled (`chinmay-sawant/blinkless#42`)
- [x] Filled body committed under `plans/v0.0.4/pr/pr-performance-improvements.md`

---

## Follow-ups (out of scope)

- Further scoped hover allocation reductions in `examples/builder/loglens` and remaining v0.0.4 maintenance ledger items (Windows IME composition, native accessibility backends).

---

## Reviewer checklist

- [x] Behavior matches summary and test plan
- [x] No unrelated changes in diff
- [x] Public API / CLI changes documented
- [x] New rules have fixture coverage when applicable
- [x] PR has assignee and labels
- [x] Related issues use correct Closes/Relates keywords
- [x] No secrets or generated artifacts committed
- [x] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 23 | 1166 | 67 |
| `.md` | 1 | 177 | 0 |
| `.mod` | 2 | 2 | 2 |
| `.py` | 1 | 2 | 0 |
| `.sh` | 1 | 2 | 0 |
| `.sum` | 2 | 12 | 0 |
| **Total** | **30** | **1361** | **69** |
