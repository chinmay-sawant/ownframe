# v0.0.3 - Chromium-style compositor optimizations

> **Parent:** (no v0.0.3 index yet; this optimizations folder is the first ledger)
> **Status:** planning only. No optimization behavior changes with these files. Rows are open unless a baseline row records checked-in code.
> **Estimated effort:** 10 optimization families across 2 detailed ledgers now, with the rest split into smaller ledgers before code starts.

---

## Overview

These ledgers line up ownframe's draw path with the moving parts named in
Chromium's compositor and paint pipeline, then map each one to the concrete
ownframe code that already does a piece of it. The result stays
`layout.DisplayList` (`layout/displaylist.go:129-138`): ownframe paints the
list through the tile grid, the replay primitive, and the GPU group compositor
in `internal/replay`. There is no page bitmap baseline and no PDF output.

Scope is ownframe only. The engine side (`github.com/chinmay-sawant/blinkless`)
is not edited by these files. Any engine change (property linkage on ops,
incremental layout, image payload metadata) lands in the blinkless checkout
first, is reviewed and pushed, and only then does this repo bump the pin in
`go.mod` and `examples/go.mod`. AGENTS.md reminds us that the blinkless
require is the only engine pin and that there is no replace directive.

A `[x]` row below is a baseline checked in current code on 2026-10-15.
A `[~]` row is a deliberate non-goal or deferred item, with a reason and a
next gate. A `[ ]` row is open work.

## Executive summary (index)

| # | Optimization | Chromium role | Current status | Owning phase file | Notes |
|---|---|---|---|---|---|
| 1 | Property trees | Transform, clip, and effect state that updates without repainting children | Not used | `phase-01-property-trees.md` | Baked `op.Transform()` already works (`internal/replay/xform_geom.go:17-29`); shader group compositing already works (`internal/replay/group_shader.go:24-29`) |
| 2 | Tiling and sparse raster | Cut painted content into tiles and re-paint only damaged tiles | Yes (core shipped) | `phase-02-tiling-and-sparse-rasterization.md` | Core tile grid shipped at 256 px (`internal/window/replay_tile.go:14`) with enhancement rows open |
| 3 | Double buffering (pending and active) | Separate the frame being built from the frame being shown | Not used | `phase-03-double-buffering-pending-active.md` (planned) | One live tile grid and one retained screen; no pending tree |
| 4 | Display list caching and partial invalidation | Reuse a built display list and invalidate by element | Partial | `plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md` | `diffDisplay` and `Invalidate` exist; per-element ownership is still a box scan |
| 5 | Damage tracking and partial swap | Keep a list of damaged screen rects and swap only them | Partial | `plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md` | Capped `dirtyRects` with coalescing shipped; platform swap-with-damage is not available through Ebiten |
| 6 | Occlusion culling | Skip paint for content hidden behind an opaque element | Partial | `phase-06-occlusion-culling.md` (planned) | Viewport visibility culling exists (`internal/replay/visible.go:11-56`); element-level occlusion does not |
| 7 | Thread-isolated scrolling | Scroll and composite on a thread that never touches page paint | Not used | `phase-07-thread-isolated-scrolling.md` (planned) | Ebiten's `Update` and `Draw` stay on one game goroutine |
| 8 | OOP-R (out of process raster) | Move raster and decode work into a second process | Not used | `plans/v0.0.2/multi-process-isolation.md` plus `phase-08-oop-r.md` (planned) | In-process today; multi-process plan is open |
| 9 | Directly composited images and solid color | Promote simple content to direct compositor layers | Partial | `phase-09-directly-composited-images-solid-color.md` (planned) | Replay draws fills and decoded images directly to tiles; no per-op layer promotion policy yet |
| 10 | Async image decoding | Decode image payloads off the draw path | Not used | `phase-10-async-image-decoding.md` (planned) | `decodedImage` decodes synchronously on first use (`internal/replay/image_cache.go:34-72`) |

## Phase-wise checklist (master status)

### Phase 01 - Property trees
Phase file: `phase-01-property-trees.md`
Status: Not used
Dependency: none (can ship either way)
- [ ] Define property tree nodes for transform, clip, and effect
- [ ] Resolve baked op.Transform through property nodes
- [ ] Resolve shader group compositing state through property nodes
- [ ] Add tests for state change without repaint
- [~] Baseline already works: baked `op.Transform()` (`internal/replay/xform_geom.go:17-29`); shader group compositing (`internal/replay/group_shader.go:24-29`)
Gate: `make golden` plus paced dump before and after

### Phase 02 - Tiling and sparse raster
Phase file: `phase-02-tiling-and-sparse-rasterization.md`
Status: Core shipped (256 px grid), enhancement rows open
Dependency: base for phases 3, 5, 6, 9
- [x] Tile grid shipped (`internal/window/replay_tile.go:14`)
- [ ] Re-paint only damaged tiles
- [ ] Add tile-level dirty tracking
- [ ] Add tile recycling and memory bounds
Gate: `make golden` plus paced dump before and after

### Phase 03 - Double buffering (pending and active)
Phase file: `phase-03-double-buffering-pending-active.md` (planned)
Status: Not used
Dependency: phase 02 and phase 05
- [ ] Introduce pending tile grid beside active tile grid
- [ ] Swap pending to active on frame commit
- [ ] Preserve TakeDirty behavior during the swap
- [ ] Add tests for in-flight frames
Gate: `make golden` plus paced dump before and after

### Phase 04 - Display list caching and partial invalidation
Phase file: `plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md`
Status: Partial
Dependency: phase 05 feeds invalidation decisions
- [x] diffDisplay exists
- [x] Invalidate exists
- [ ] Per-element ownership without a box scan
- [ ] Cache built display lists for reuse
Gate: `make golden` plus paced dump before and after

### Phase 05 - Damage tracking and partial swap
Phase file: `plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md`
Status: Partial
Dependency: base for phase 02 tile invalidation
- [x] Capped dirtyRects with coalescing shipped
- [ ] Swap only damaged rects on platforms that allow it
- [ ] Document Ebiten limitation
Gate: `make golden` plus paced dump before and after

### Phase 06 - Occlusion culling
Phase file: `phase-06-occlusion-culling.md` (planned)
Status: Partial
Dependency: phase 01 property tree clip and effect nodes
- [x] Viewport visibility culling exists (`internal/replay/visible.go:11-56`)
- [ ] Element-level occlusion behind opaque elements
- [ ] Parity proof against paint outline tooling
Gate: `make golden` plus paced dump before and after

### Phase 07 - Thread-isolated scrolling
Phase file: `phase-07-thread-isolated-scrolling.md` (planned)
Status: Not used
Dependency: phase 03 and cheap scroll offsets
- [ ] Move scroll handling off the page paint path
- [ ] Keep Ebiten Update and Draw contract intact
- [ ] Add latency measurement
Gate: `make golden` plus paced dump before and after

### Phase 08 - OOP-R (out of process raster)
Phase file: `plans/v0.0.2/multi-process-isolation.md` plus `phase-08-oop-r.md` (planned)
Status: Not used
Dependency: last, needs frame transport, damage, and partial repaint
- [ ] Move raster and decode into a second process
- [ ] Define frame transport contract
- [ ] Preserve partial repaint behavior
Gate: `make golden` plus paced dump before and after

### Phase 09 - Directly composited images and solid color
Phase file: `phase-09-directly-composited-images-solid-color.md` (planned)
Status: Partial
Dependency: phase 02 and group composite path
- [x] Replay draws fills and decoded images directly to tiles
- [ ] Add per-op layer promotion policy
- [ ] Add tests for promotion and demotion
Gate: `make golden` plus paced dump before and after

### Phase 10 - Async image decoding
Phase file: `phase-10-async-image-decoding.md` (planned)
Status: Not used
Dependency: independent, measure decode cost first
- [ ] Decode image payloads off the draw path
- [ ] Keep cache behavior correct under concurrency
- [ ] Add decode cost measurement
- [~] Current path decodes synchronously on first use (`internal/replay/image_cache.go:34-72`)
Gate: `make golden` plus paced dump before and after

## Dependency order

1. Tiling and sparse raster (phase 02) is the base. Partial repaint and damage tracking both build on the tile grid in `internal/window/replay_tile*.go`.
2. Damage tracking and the capped damage list (existing `plans/v0.0.2/...`) feed the tile invalidation decisions; complete before measuring tile work.
3. Display list caching and partial invalidation refine the damage path.
4. Property trees (phase 01) are independent of tiling and can ship either way; they only become useful once transform, clip, and effect state is expected to change without a full display-list rebuild.
5. Occlusion culling builds on visibility culling plus the property tree clip and effect nodes, and needs its own parity proof against the paint outline tooling.
6. Directly composited images and solid color builds on the group composite and tiling paths, so it lands after both.
7. Async image decoding is an independent utility that touches `internal/replay/image_cache.go`; it can land at any point but should be measured with decode cost first.
8. Double buffering (pending and active) depends on damage tracking and on property trees being usable as the pending state.
9. Thread-isolated scrolling depends on double buffering and on scroll offsets being cheap to update.
10. OOP-R is last: it needs frame transport, damage, and a stable partial repaint contract before raster moves out of process.

## Ground rules

- Scope is ownframe. blinkless is untouched. Engine work requires a separate blinkless PR, a pushed commit, and a pin bump here in `go.mod`.
- No new modules. The dependency allowlist test stays as it is.
- No optimization ships without a number. Use the dump harness (`go run ./examples/departure-board -dump -pace 16ms`) and record before and after in the ledger with the same command.
- Keep `TakeDirty()` working when adding list-based APIs; new methods arrive beside the old ones.
- Every Go file stays within the repo's size limit (2000 characters in ownframe), split inside the same package when needed. `gofmt` every edited file.
- Phase rows move to `[x]` only after the matching test, golden, or benchmark passes. Do not close a runtime item from plan prose.
- For documentation-only changes, do not run lint or test checks.
- No em dashes in any written output; use hyphens.

## Definition of done

- All 10 optimizations have a phase file or point to an existing ledger.
- Each phase records before and after numbers from the same dump command.
- No optimization ships from plan prose alone; rows stay `[ ]` until proof.
- `make golden` and `make test` pass before any runtime item closes.

## Status note

As of 2026-10-15, only phases 02 (core) and 04/05 (partial) and 09 (partial) have checked-in baselines. Phases 01, 03, 06, 07, 08, 10 are open. Nothing is committed; this file is planning only.
EOF