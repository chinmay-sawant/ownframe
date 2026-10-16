# Phase 02, tiling and sparse rasterization

> **Parent:** `plans/v0.0.3/optimizations/README.md` - tiling and sparse raster is optimization family 2.
> **Status:** shipped core is checked in; enhancement rows are open.
> **Estimated effort:** enhancement work is roughly two to three days after the
> measurements in Phase 02.4. Core rows below are evidence-backed and closed.

---

## Overview

Ownframe paints a `layout.DisplayList` through a row-major grid of 256 px
replay tiles. A dirty rect repaints only the tiles it covers, a frame blits only
the damaged tiles, and the retained screen keeps every untouched tile. The
result stays `layout.DisplayList`: there is no page bitmap baseline and no PDF
output.

Core lives in `internal/window/replay_tile*.go` and is wired from
`drawReplayPartial` in `internal/window/replay_partial.go`. The grid is allocated
by content size in `ensureTiles`, disposed with the content in `dropTiles`, and
reattached to `partialState.buf` as a zero-area handle so readers that ask
whether retained pixels exist keep working.

A `[x]` row is core code that is checked in on 2026-10-15. A `[ ]` row is open
enhancement work. `[~]` is not used here; enhancements are either open or moved
to a later phase ledger.

## Where it lives

Core files:

- `internal/window/replay_tile.go` - `tileSize`, `tileGrid`, `newTileGrid`,
  `gridTiles`, `rectAt`, `at`.
- `internal/window/replay_tile_range.go` - `tileRange`, `content`, `handle`,
  `dispose`.
- `internal/window/replay_tile_life.go` - `ensureTiles`, `dropTiles`,
  `dropRetainedTiles`.
- `internal/window/replay_tile_paint.go` - `paintTiles`, `paintTile`.
- `internal/window/replay_tile_copy.go` - `blitTiles`, `blitTile`.
- `internal/window/replay_tile_blit.go` - `blitFrame`, `showTilesWhole`,
  `withinTiles`.
- `internal/window/replay_tile_clip.go` - `dockClipX`, `clipDockX`, `dockStrip`.
- `internal/window/replay_plan_pick.go` - `planRepaints`, `maxRepaintRects`.
- `internal/window/frame_blit.go` - `blitPlan`, `maxBlitRects`.
- `internal/window/frame_blit_rects.go` - `mergeBlitRects`, `unionRects`,
  `chromeRects`.
- `internal/window/replay_partial.go` - `drawReplayPartial`.
- `internal/window/replay_partial_state.go` - `partialState`, `gridCurrent`.
- `internal/window/replay_buffer.go` - `applyRepaint`.

Related tests:

- `internal/window/replay_tile_grid_test.go`
- `internal/window/replay_tile_range_test.go`
- `internal/window/replay_tile_*_test.go`
- `internal/window/replay_damage_test.go`
- `internal/window/replay_plan_test.go`
- `internal/window/replay_plan_clip_test.go`
- `internal/window/replay_frame_test.go`

Engine side stays untouched. Any change to damage ownership or element linkage
lands in blinkless first and is pinned in `go.mod` after review.

## Phase 02.1, shipped tile grid and indexing

Goal: allocate one clipped tile image per covered 256 px cell, with exact
content-area accounting and a safe disposal path.

- [x] `tileSize` is 256 CSS pixels and is the only tile edge length
  (`internal/window/replay_tile.go:14`).
- [x] `newTileGrid` returns nil for non-positive size and otherwise allocates
  `cols * rows` tiles in row-major order.
- [x] `gridTiles` rounds up with `(n + tileSize - 1) / tileSize`.
- [x] `rectAt` clips every edge tile to content, so no tile lies outside the
  content and far-edge cells are remainders.
- [x] Tile images are bounds-checked against `rectAt`, so a loaded image has
  exactly `r.Dx()` by `r.Dy()` pixels.
- [x] `at` indexes row-major with `row * cols + col`.
- [x] Total tile area equals `width * height` for representative sizes, proven
  by `TestTileGridKeepsTheContentCost`.
- [x] Non-multiple sizes keep interior tiles full and clip only the far edges,
  proven by `TestEdgeTileIsClipped`.
- [x] `content` returns `image.Rect(0, 0, width, height)`.
- [x] `tileRange` clamps to content, returns an empty span for a miss, and
  uses `(Max-1)/tileSize + 1` so partial tail cells count.
- [x] `handle` returns a zero-area sub-image of the first tile, preserving the
  retained-content contract without copying pixels.
- [x] `dispose` calls `Dispose` on every non-nil tile.
- [x] `ensureTiles` keeps the grid when size is unchanged, drops and rebuilds
  it when the content size changes, and leaves no grid for zero size.
- [x] `dropTiles` clears both `tiles` and `buf`.
- [x] `dropRetainedTiles` drops the grid when `buf` was lost, keeping the
  handle and tiles consistent.
- [x] `gridCurrent` reports true only for a non-nil grid with matching
  content size.

## Phase 02.2, shipped paint, blit, and plan paths

Goal: repaint only damaged tiles and restore only the screen rects that need
refresh.

- [x] `drawReplayPartial` runs tile allocation, painting, and blitting on the
  same frame and returns a bool rather than paging through a second path.
- [x] `drawReplayPartial` clears `paints`, `paintPx`, `blits`, and `blitPx`
  at frame start.
- [x] Scroll changes are detected before plans are made and recorded in
  `partialState.scroll`.
- [x] A stale frame or non-live retained state forces the tile path to repair
  the screen instead of assuming current pixels.
- [x] Oversized content falls back to `drawViewport` and leaves the frame
  clock untouched instead of corrupting tile timing.
- [x] `dockClipX` converts the dock panel into page pixels with scroll
  accounted for and clamps it to content.
- [x] `clipDockX` trims a page rect at the dock clip and returns empty when
  nothing remains.
- [x] `dockStrip` returns the strip the previous dock hid and only when the
  visible clip grew.
- [x] `planRepaints` rebuilds when the grid is missing or the content size
  changed.
- [x] `planRepaints` turns no-damage frames into `repaintBlit` when the grid
  holds the current generation, or `repaintRect` for a newly exposed dock
  strip.
- [x] `planRepaints` skips rects clipped fully out by the dock.
- [x] `planRepaints` rebuilds on a full-content rect, a rect outside content,
  or more than `maxRepaintRects` plans.
- [x] `planRepaints` caps to `maxRepaintRects` through the rebuild branch, not
  by silently dropping rects.
- [x] `applyRepaint` rebuilds the grid and repaints all of it for
  `repaintBuffer`, and repaints only covering tiles for `repaintRect`.
- [x] `paintTiles` trims rect to dock clip, computes the tile span, and
  visits every tile in that span.
- [x] `paintTile` intersects the dirty rect with the tile, returns early when
  empty, and uses a `SubImage` at the exact part.
- [x] `paintTile` fills the part with `pageBackground` before replay, so
  stale tile pixels do not bleed through.
- [x] `paintTile` calls `replay.DrawVisibleUnfixedZ` with `-tile.Min` offsets so
  page coordinates land on the tile.
- [x] Fixed and viewport-pinned operations stay outside the tile buffer, as in
  the old single-buffer path.
- [x] `blitFrame` uses `showTilesWhole` or `paintFullReplay` when stale, the
  grid missing, or chrome reaches past the grid.
- [x] `blitFrame` performs one whole-content copy when `whole` is true,
  matching the scroll case.
- [x] `blitFrame` copies only merged damage rects otherwise.
- [x] `showTilesWhole` fills the screen with page background and copies every
  tile, replacing a second full replay.
- [x] `withinTiles` rejects chrome that extends past content, forcing a full
  rebuild instead of a clipped restore.
- [x] `blitTiles` clips to dock, computes the tile span, and bounds each tile
  copy.
- [x] `blitTile` intersects source and destination, uses `BlendCopy`, and
  translates draw coordinates into screen space.
- [x] `blitTile` counts blits and blit pixels in `partialState`.
- [x] `paintTile` counts tile paints and painted pixels in `partialState`.
- [x] `blitPlan` merges overlapping rects, adds chrome rects, and collapses to
  one union past `maxBlitRects`.
- [x] `mergeBlitRects` repeatedly unions overlapping rects until no overlap
  remains.
- [x] `chromeRects` moves previous screen chrome into buffer space using the
  current scroll.
- [x] Two separated changes produce two `repaintRect` plans instead of one
  union, proven by `TestTwoRectsRepaintSeparately`.
- [x] A page that only implements `TakeDirty` is wrapped as a one-element list
  and still paints, proven by `TestSingleRectTakerStillRepaints`.

## Phase 02.3, enhancement work

Goal: reduce scan cost, reduce allocation churn, and make tiling behavior
testable without changing the output.

- [ ] Add per-tile operation slices built from the display list, so
  `paintTile` replays only ops that intersect that tile instead of scanning the
  global list every time.
- [ ] Define tile operation invalidation rules for element ids, layout box
  changes, and scroll offsets.
- [ ] Add benchmarks that compare global scan cost to tile-local replay cost.
- [ ] Reuse tile images across content resizes when only width or height grows,
  rather than disposing the whole grid.
- [ ] Add a tile image pool for `ensureTiles` and measure allocation reduction.
- [ ] Add a content-size threshold above which the grid is divided into fewer,
  larger tiles, and record the reason.
- [ ] Store a dirty bitmask in `partialState` and skip tiles that are already
  clean even when a broader rect is planned.
- [ ] Coalesce dirty rects to tile boundaries before `planRepaints`, then
  measure rect count and worst-frame ops.
- [ ] Add parallel tile painting with deterministic output and verify no race
  with counters and dock clipping.
- [ ] Expose `paints`, `paintPx`, `blits`, and `blitPx` read-only through a
  devtools surface without allocating per frame.
- [ ] Add regression tests for `blitFrame` fallback when chrome is outside the
  grid and when the retained handle is dropped.
- [ ] Add tests for dock strip catch-up after a dock hide or show transition.
- [ ] Add tests for tile range with one-row, one-column, and 1 px content.
- [ ] Add tests for `blitTile` when the destination intersection is empty.
- [ ] Add error or fallback handling for `ebiten.NewImage` failure, if Ebiten
  can return a nil image.
- [ ] Add lifecycle documentation for ensure/drop/rebuild transitions and
  retained handles.

## Phase 02.4, performance and closure gate

Goal: close the tiling family with measured proof and docs.

- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` before the
  enhancement pass and record worst-frame and mean numbers.
- [ ] Run the same command after the enhancement pass and record the same
  metrics.
- [ ] Confirm improvement is ratio-based against baseline, not a fixed op count.
- [ ] Run targeted tests for `internal/window` after every edit.
- [ ] Run `make golden` after tiling, paint, or blit changes.
- [ ] Run `make test` once at session end.
- [ ] Update `documentation/repaint.md` with the tile grid, plan, and blit
  contract.
- [ ] Update `documentation/performance.md` or its baseline section with tile
  counters and the measured gate.
- [ ] Leave all core rows `[x]` and every unproven enhancement row `[ ]`.

## Exit criteria

- The tile grid covers exactly the content area at 256 px edges, with
  remainders only on far edges.
- A separated change repaints fewer tile pixels than a full-content rebuild,
  and a full-content change still falls back safely.
- Scroll, dock, and chrome cases restore the screen without painting stale
  tile pixels.
- Two separated changes produce two planes through `repaintRect`.
- `make golden` and `make test` pass after any tiling behavior change.

## Risks

- Tiling adds per-tile clear and sub-image calls; the enhancement gate must
  prove the change beats the baseline by ratio, not by assumption.
- Content that outgrows the canvas can rebuild or retain the wrong grid; the
  oversized viewport path keeps its separate cached image and must remain
  authoritative for that case.
- Resize reuse may retain stale tile images; any reuse must preserve generation
  and dirty state checks.
- Parallel painting can introduce races in counters, dock clip, or plan state;
  keep counters atomic or serialize them and measure contention.
- Engine-side damage semantics are not owned here; do not imply blinkless support
  that does not exist yet.

EOF
wc -l /home/chinmay/ChinmayPersonalProjects/ownframe/plans/v0.0.3/optimizations/phase-02-tiling-and-sparse-rasterization.md