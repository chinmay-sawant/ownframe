# Phase 05, damage tracking and partial swap

Goal: take the shipped damage list (`Page.TakeDirtyRects`, `internal/page/page_dirty_take.go:8-22`) and the partial tile blits (`internal/window/replay_tile_blit.go:15-41`) one step further, and commit only the damaged screen regions at frame end. Today the page, the tile grid, and the buffer all speak in dirty rectangles, but the final present still covers the whole client area. This phase closes that gap where a platform backend allows it, and records the gaps where it does not.

## Where it lives

Baseline pieces that already ship, cited from the tree on 2026-10-16:

- `internal/page/page_dirty_take.go`: `TakeDirtyRects` and the bounding-box `TakeDirty`; `dirtyTracker.take` owns the damage state, capped at eight rects, with coalescing and total-coverage collapse.
- `internal/page/page_dirty_list.go` and `internal/page/page_dirty_merge.go`: coalesced insert, area-based collapse, ordering. The damage list is already the small capped list from phase 01 of the damage-and-invalidation plan.
- `internal/page/page_invalidate.go`: `Invalidate` feeds `markRect` and never blocks the render path. The element-id path is still a box scan and stays that way until the engine side lands.
- `internal/window/replay_taker.go`: `dirtyTaker`, `singleTaker`, `listTaker`, and `damageTaker()` adapt a page to the list method (`replay_taker.go:28-37`). `TakeDirty()` keeps working through the fallback wrapper.
- `internal/window/replay_partial.go`: `drawReplayPartial` reads `taker.TakeDirtyRects()` at line 48, builds per-rect plans at line 49-57, replays each rect into the same buffer at line 55, then blits only damaged tiles at 61.
- `internal/window/replay_plan.go`: `repaintPlan`, the four modes, and `maxRepaintRects = 8` (plan line 29). Tile side cap is `tileSize = 256` in `internal/window/replay_tile.go:14`.
- `internal/window/replay_tile_blit.go`: `blitFrame` copies only the damaged tiles on a normal frame (`blitFrame` lines 38-40) but always commits through Ebiten's normal `Update`/`Draw`, which presents the whole screen.
- `examples/departure-board/PERFORMANCE.md`: measured baseline, 115 ops on the worst cascade frame through 3 rects, down from 630 of 631 on the single union rect.

The engine side (blinkless) is untouched by this phase. Any ownership change stays additive and lands first.

## What Chrome does

Chromium's display compositor tracks damage as a list of screen rects and uses it for partial swap when the platform supports it. ownframe already has the damage list and the tile selection. What it lacks is the last hop: passing those rects to the windowing backend at present time. Ebiten's `screen.Draw` path presents the full client area, and there is no public swap-damage call in the version ownframe uses. The cost of that missing hop is one full-screen copy per frame on desktop, and the full client area on every touch scroll frame.

## Current status as of 2026-10-16

- [x] Page-side damage list shipped, capped at eight rects with coalescing and third-area collapse.
- [x] Window-side per-rect replay and per-tile blit shipped; untouched tiles come from the retained screen.
- [x] `TakeDirty()` remains the bounding-box wrapper, so examples and platform callers still compile.
- [ ] No damage rect reaches the present call. `blitFrame` never reports damaged screen rects upward.
- [ ] No partial-present probe exists, so we cannot tell from a run whether the backend accepted the damage.
- [ ] No platform backend in `blitFrame` or `frame commit` accepts a rect list.
- [ ] No swap-damage metric in the perf ring or devtools dock.
- [ ] Web and WASM builds still present the full canvas; grid blits are the only saving.

## Design

1. Keep the damage list authoritative. `TakeDirtyRects()` stays the single source. The bounding box `TakeDirty()` stays for callers that only want one rect, and new methods arrive beside it.
2. Introduce a screen-level damage carrier separate from page damage. Page damage is in CSS pixels and may span the content. Screen damage is in client pixels after scroll, clip, dock clip, chrome rects, and tile edges. The carrier lives in `internal/window`, next to `partial` state, so platform code never imports `internal/page`.
3. Compute screen damage in one place. `drawReplayPartial` already knows the final blit rects; extend it to accumulate those into `partial.swapDamage []image.Rectangle`, trimmed to the client area and coalesced with the same rules as the page list.
4. Add a present method with a list: `presentDamaged(dst *ebiten.Image, rects []image.Rectangle) bool`. It returns true when the backend accepted the partial present, false when it fell back to full. First implementation always returns false and presents in full, so behavior is unchanged and the call site exists for measurement.
5. Gate on a build-tag or runtime probe flag, `partialPresentProbe`, so a developer can log whether the current backend accepted damage without changing shipping code.
6. Keep the no-damage case cheap. An empty swap list must blit nothing and present whatever the backend considers the cached screen; do not fabricate a full present to satisfy a probe.
7. Do not change tile sizes or grid layout. Tile damage and swap damage are different units of work; tying them to the same grid size would make the swap path fragile.

## Checklist

### Phase A, baseline capture

- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` with a documented seed and store the numbers in `examples/departure-board/PERFORMANCE.md`: full-frame count, two-or-more-rect frames, blit tile count, and present size in pixels.
- [ ] Add a "presented area" column to the dump by having `blitFrame` report the total pixels it blits, clipped to the client area. This is a measurement, not a behavior change.
- [ ] Record before-and-after present area for the scroll path and the steady-state path separately. Scroll frames blit the whole content and must not be treated as a regression.
- [ ] Confirm `make test` passes on the tree before any window file changes.

### Phase B, screen damage contract

- [ ] Add `swapDamage []image.Rectangle` to the partial state struct and clear it at the start of each frame in `drawReplayPartial` (beside `partial.paints` at line 20).
- [ ] In `blitFrame`, when the damaged-tile branch runs (`blitFrame` lines 38-40), append each blit rect's screen-space region to `partial.swapDamage` instead of discarding it.
- [ ] Add `PageSwapDamage()` or a window-level accessor that returns the list without clearing, for tests and devtools. Document that it is valid only after a frame commit.
- [ ] Coalesce the accumulated rects with the same eight-rect cap and third-area collapse as the page damage. Reuse the helpers in `internal/page` through a small same-package adapter; do not duplicate the algorithm in `internal/window`.
- [ ] Make an empty damage list mean "no new pixels this frame", distinct from "full present needed". Fall back to full when the frame was stale, when chrome outgrew the grid, or when the oversized viewport path ran.

### Phase C, native partial-present probe

- [ ] Introduce `partialPresentProbe bool` on the shell, default false, and `SetPartialPresentProbe(bool)` for tests. No user-facing option.
- [ ] Add `presentDamaged(dst *ebiten.Image, rects []image.Rectangle) bool`. Default body: ignore rects, return false. `blitFrame` calls it when its existing branch would return, right before returning to `Draw`.
- [ ] Log one line per frame while the probe is on: `partialpresent: accepted=%t rects=%d area=%d/%d px`, with the damaged client area over the full client area.
- [ ] Add a test that sets the probe, runs a fixed interactive sequence (type into a field, scroll one step, hover), and asserts the logged counts look sane. The assertion checks the count of frames with `accepted=false`, which is every frame in this build.
- [ ] Keep the probe out of release builds by leaving it false by default and never wiring it to a hot key.

### Phase D, desktop backend

- [ ] Investigate Ebiten v2.10.x backends for any swap-damage entry point. Record the exact symbol search result, `go doc` output, and the backend source location in this file before writing code so the gap is explicit rather than implied.
- [ ] If a public symbol exists, implement `presentDamaged` through it and update the probe log to show `accepted=true` on a real run.
- [ ] If no symbol exists, document the required backend change as an ownframe-local fork or a dependency upgrade request, and leave `presentDamaged` at the always-false default. Do not vendor Ebiten without a written reason.
- [ ] For the native-dock path (`s.dockClipX` and chrome rects), keep chrome damage in the same list. A dock resize must damage both the dock region and the uncovered content tile.

### Phase E, platform backends

- [ ] WASM (`frontend` and `browser`): record whether `canvas.getContext('2d')` dirty rect present or `requestAnimationFrame` full redraw is the only path. Implement partial present via `putImageData` only when it is cheaper than the retained buffer blit.
- [ ] Android: `eglSwapBuffers` has no damage argument in the current bindings; record the finding and add a note to `plans/v0.0.2/android-runtime-upgrade-checklist.md`.
- [ ] Embedded or low-power targets: document the full-screen present as the expected path and mark it `[~]` with a reason.
- [ ] Add per-backend feature flags so `presentDamaged` can be compiled out where it is harmful, without build-tag forks in the window package.

### Phase F, metrics, tests, and parity

- [ ] Unit tests: damaged tile to client-pixel rect conversion, eight-rect cap on the window side, coalescing of overlapping tile rects, empty-list means no blit.
- [ ] Parity tests: each interactive test in `internal/page/page_picture_compare_test.go` (and its siblings) must still produce byte-identical output after partial present falls back to full. With `accepted=false` this is automatic, but the probe path must be under test.
- [ ] Add a benchmark, `BenchmarkPartialPresentProbe`, that runs the departure board at `-pace 16ms` with the probe on and reports damaged client area per frame.
- [ ] Add a devtools dock row, under the existing frame counters, showing `partialPresent rectangles` and `partialPresent accepted`. The dock must allocate nothing per frame on a still generation, matching phase 03.
- [ ] Run `go vet -p 1 ./... ./examples/...` and `make test` from the repo root. Fix build errors before closing any row.

### Phase G, docs and tracking

- [ ] Update `documentation/repaint.md` with the two-layer damage picture: page damage versus screen damage, the eight-rect cap at both layers, and the always-full-present default.
- [ ] Update `documentation/performance.md` with the measured present-area column and a statement that platform support is the open part of this phase.
- [ ] Update `docs/src/damage-and-swap.md` or create the page if it does not exist, without modifying other documentation pages.
- [ ] Update `PHASES.md` with commit hashes and the measurement numbers.
- [ ] Add a new row to `plans/v0.0.3/optimizations/README.md` executive summary already points here; only change the status from `Partial` to `In progress` when Phase B lands.

## Exit

- The present path has one named seam, `presentDamaged`, with a stable signature and a discoverable measurement. Every frame either presents the damaged client area or presents in full and records that fact.
- On a run with the probe disabled, numbers match the departure-board baseline on 2026-10-12: 115 ops worst cascade frame, 13 of 150 frames with two or more rects, no `full-buffer` frame after the initial draw.
- No parity test regresses, and no new allocation appears on a still generation with the devtools dock open.

## Risks

- Ebiten may never expose swap damage in its current API. Phase D captures that as a documented finding, not a silent no-op. The tile blit work still helps because it lowers the buffer copy cost even on a full present.
- Converting CSS damage to client damage needs scroll, clip, and dock state. Every conversion must be guarded so a conversion error falls back to full, never to a wrong rect.
- A stale frame and a partial present are a bad combination. The probe must clear its list on every stale frame and force a full present there.
- Measurement lines in the dump must stay behind flags. An unguarded per-frame printf on a production build will eat the saving the phase targets.

## Non-goals

- No change to `layout.DisplayList` or to blinkless. The engine stays a pure function returning ops.
- No second tile size or a dynamic grid. Tile geometry is a separate phase.
- No WASM-only shortcuts that diverge from the desktop code path; any backend difference is recorded in the feature flags, not in the replay logic.