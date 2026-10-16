# Phase 3: Double buffering (pending and active tree)

Recorded 2026-10-16 against the working tree on `master`, planning only. No
behavior change ships with this file. Rows stay open until code and a measurement
close them.

Parent ledger: [README.md](README.md). Depends on phase 01 property trees and on
the damage list in [plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md](../../v0.0.2/damage-and-invalidation/phase-01-damage-list.md)
for damage propagation across the swap.

## Why this is optional

Double buffering is not required for correctness today. Ebiten runs `Update`
and `Draw` on one game goroutine, so there is no reader observing a frame while
it is being built. The screen is presented once per `Draw`, and that draw reads
the same tile grid the previous frame tiled. Nothing tears.

The current buffer is already retained across frames. `partialState` holds one
`tileGrid` and one retained-content handle (`buf`) in
[internal/window/replay_partial_state.go:32-36](../../internal/window/replay_partial_state.go)
and [internal/window/replay_tile_life.go:9-33](../../internal/window/replay_tile_life.go).
Partial repaint paints directly into those retained tiles and the frame blits
only the damaged ones. Replacing that scheme with two full grids would double
the tile memory for content of every size without removing any work the current
path does.

A second copy also adds a swap step and two sets of tile bookkeeping. On the
departure board, tiles are 256 px (`internal/window/replay_tile.go:14`), a
content buffer already covers 800 by 36579 pixels in the worst baseline reported
in [documentation/performance-baseline.md:55](../../documentation/performance-baseline.md),
so a second grid is not free.

The reason to pay for it later is a thread boundary. Thread-isolated scrolling
(phase 07) and out-of-process raster (phase 10, and the multi-process plan in
[plans/v0.0.2/multi-process-isolation.md](../../v0.0.2/multi-process-isolation.md))
both need one side of the pipeline to read the on-screen frame while another side
builds the next one. With one tree the builder would have to block the reader, or
the reader would have to wait. Double buffering is the precondition for that,
not a feature on its own.

So the rule here: this phase stays optional until a measurement names an update
blocked behind a draw or a frame blocked behind a build. No row closes on
intent.

## Current single-buffer / tile state

### Page side

- One retained display list. `Page.display` is a single `*layout.Display` in
  [internal/page/page_types.go:29](../../internal/page/page_types.go). `Redraw`
  overwrites it in `commitDisplay`
  ([internal/page/page_draw_stages.go:64-80](../../internal/page/page_draw_stages.go)).
- One generation counter. `Page.generation` bumps on every committed display,
  and the window reads it through `app.Generation()` in
  [internal/window/replay_partial.go:56](../../internal/window/replay_partial.go).
- One dirty region. `dirtyTracker` in
  [internal/page/page_dirty_tracker.go](../../internal/page/page_dirty_tracker.go)
  holds one `rects` list, one `pending` id map, and one `last` id map. The
  display diff in [internal/page/page_dirty_diff.go](../../internal/page/page_dirty_diff.go)
  compares `prev` to `next` and dirties the changed bounds.
- `commitDisplay` records the diff, commits the list, commits accessibility, and
  runs `applyPending`, so a build both replaces the active list and updates the
  damage state in the same call.

### Window side

- One tile grid. `ensureTiles` allocates at most one `tileGrid` for the
  content size and reuses it through `gridCurrent`
  ([internal/window/replay_tile_life.go:9-33](../../internal/window/replay_tile_life.go)).
- One retained handle. `partialState.buf` stands for the grid; `dropTiles`
  clears both ([internal/window/replay_tile_life.go:35-52](../../internal/window/replay_tile_life.go)).
- Repaint paints into the same grid. `applyRepaint` repaints whole tiles for a
  buffer rebuild or the touched tiles for a rect plan
  ([internal/window/replay_buffer.go:11-22](../../internal/window/replay_buffer.go)).
- The frame blits from the same grid. `blitFrame` reads `s.partial.tiles`,
  `plans`, and `chrome`, and copies damaged tiles to the screen
  ([internal/window/replay_tile_blit.go:14-46](../../internal/window/replay_tile_blit.go)).
- A scroll only moves the blit; tiles are not repainted
  ([internal/window/replay_tile_scroll_test.go](../../internal/window/replay_tile_scroll_test.go)).

There is no pending frame anywhere. A `Redraw` commits immediately, the diff
dirties the current tracker, and the next `Draw` repaints into the same grid.

## Proposed pending / active model

Two named roles. `active` is what the screen currently shows and what the
window blits from. `pending` is what a build or a partial repaint is writing
into. The same two names apply on both sides of the boundary, but the data is
different, so the phases below split by owner.

### Page side: pending and active display

- Active display: `p.display`, the committed list the renderer last published.
  Stays the read source for `Page.Display()`, `Page.Boxes()`, and the frame
  helpers in [internal/frame](../../internal/frame).
- Pending display: `p.pendingDisplay`, a scratch `*layout.Display` that
  `Redraw` builds into. A second build does not touch `p.display`, so readers
  keep the stable list while it runs.
- Generation semantics: an active generation and a pending generation. Readers
  continue to see the old generation until the swap. `dirtyFromDisplay` compares
  the pending previous active with the pending candidate, so damage is computed
  before the swap and applied as one batch.
- Atomic swap: a single method installs pending as active, bumps the active
  generation, moves the dirty region into the active tracker, and clears the
  scratch. The swap is the only point where the display pointer changes.
- Keep `TakeDirty()` and `TakeDirtyRects()` unchanged. They keep reporting the
  active region. A pending dirty region is promoted at the same time as the
  list.

### Window side: pending and active tile tree

- Active grid: `s.partial.activeTiles` plus `s.partial.activeHandle`, the grid
  the screen currently blits from. This replaces the current `tiles`/`buf`
  naming once the second grid exists.
- Pending grid: `s.partial.pendingTiles` plus its handle. Repaint plans paint
  into the pending grid. The active grid keeps showing the previous frame until
  the paint finishes.
- Swap at the present boundary. The frame's `Draw` chooses the grid to blit
  explicitly: after all rects are painted, the pending grid becomes active by
  pointer swap, the old active becomes the new scratch, and the next build
  reuses it. No tile is freed on a steady frame.
- Damage across the swap. The damage list order does not change: the generator
  still calls `TakeDirtyRects`; only the destination image change. A rect that
  only touches the active grid stays on the active path; a rebuild allocates or
  reuses the pending grid and then swaps.
- Keep the current grid sizes. The pending grid can be allocated to the active
  handle's pixels when the content size is unchanged, so no second full
  allocation occurs on a steady size. A size change frees both, as today.
- One extra frame of latency is acceptable only when the presenter (Ebiten
  `Draw`) is scheduled between build and present; today both run in `Draw`, so
  the latency stays zero on desktop. Async rasterization later is the consumer.

### Shared constraints

- No public API change. `Page.Display()`, `Page.FallbackDisplay()`,
  `Page.Image()`, and the dirty calls keep their signatures. Window options do
  not gain a flag unless measurement asks for one.
- Keep the opt-out. `fullReplay` and the ticking-page override still force the
  full replay path, and ` oversized` still drops tiles (see
  [internal/window/replay_partial.go:45-60](../../internal/window/replay_partial.go)).
- Keep `TakeDirty()` working beside any list API, per the damage-ledger ground
  rules.

## Phases

| Phase | Owner | Job |
|---|---|---|
| 0 | ownframe | Record a number that shows the current path is blocked, otherwise stop |
| 1 | ownframe | Name active and pending on the page side without changing pointer timing |
| 2 | ownframe | Build into pending and swap atomically |
| 3 | ownframe | Record dirty regions against pending, promote at swap |
| 4 | ownframe | Add the second tile grid and two handles |
| 5 | ownframe | Paint into pending tiles and swap at the present boundary |
| 6 | ownframe | Keep fallback, header, and dock clipping correct across the swap |
| 7 | ownframe | Measure memory, frame time, and skip counts before and after |
| 8 | docs | Repaint docs, PHASES.md, perf ledger |

Phases 1 to 3 need no window change. Phases 4 to 6 need no page change except
the generation and damage contract. Phase 7 is the gate; phase 8 only runs when
it passes.

## Checklist

### Phase 0 - prove the gap

- [ ] Name a concrete caller that cannot build a frame while another frame is
      on screen: a page that updates while a slow paint is in progress, or a
      scroll that blocks a repaint.
- [ ] Record the current `make test` and departure-board dump baseline:
      `go run ./examples/departure-board -dump -pace 16ms`.
- [ ] Record tile count, grid bytes, and draw time from the Frame tab and the
      Ops tab of the devtools dock, with the dock open and closed.
- [ ] Stop if the measured frame time is under budget or no caller is blocked.
      Document the reason and close the phase as deferred.

### Phase 1 - name active and pending on the page side

- [ ] Add `pendingDisplay *layout.Display` beside `display` in
      [internal/page/page_types.go:29](../../internal/page/page_types.go). Keep
      `display` as the active field.
- [ ] Add `activeGeneration` and `pendingGeneration` while keeping `generation`
      as the public active value.
- [ ] Add read-only accessors `ActiveDisplay()`, `PendingDisplay()` that never
      return the build buffer while a build is in progress.
- [ ] Keep `Page.Display()` returning the active display. No caller changes.
- [ ] Split `commitDisplay` so it reads as: build pending, diff, apply damage,
      mark pending dirty, then swap. The first three steps write only to
      `pending*` fields.
- [ ] Ensure `applyPending` runs during the pending build, not after the swap.
- [ ] Benchmark `Page.Redraw` for the tempo example to keep the build time flat.
- [ ] Tests: `Page.Display()` is unchanged while `Redraw` is halfway through.

### Phase 2 - build into pending and swap atomically

- [ ] Implement `commitPending()` that installs `pendingDisplay` as `display`,
      bumps `generation`, promotes `pendingGeneration`, and clears
      `pendingDisplay`.
- [ ] Add `swapDisplay()` as the single pointer write site.
- [ ] Move `stats.layouts`, `stats.repaints`, and `stats.lastDraw` updates to
      the swap, so perf counts a published frame, not a build.
- [ ] Move `setBoxes`, `commitAccessibility`, and `p.img = nil` to the swap.
- [ ] On a full fallback path (`paintFallback`), build and commit the pending
      bitmap through the same swap, including `fallbackDisplay`.
- [ ] Keep cache behavior: `render.Replayable` and the cache unchanged.
- [ ] Tests: a reader called between build start and swap sees the old display
      and the old generation.

### Phase 3 - damage across the pending / active boundary

- [ ] Compute `diffDisplay(p.display, pending)` during build and store the
      result as `pendingDiff`, not into the active tracker.
- [ ] Promote the rect list, `pending` id map, and full flag at the swap.
- [ ] Promote `changedProps` and `lastPatch` at the swap so devtools reads them
      after publication.
- [ ] Keep `TakeDirty()` returning the active region, with the promoted list.
- [ ] Tests: a change made during a pending build is not visible through
      `TakeDirty()` until the swap; after the swap it is.

### Phase 4 - add the second tile grid

- [ ] Rename `partialState.tiles` to `activeTiles` and `partialState.buf` to
      `activeHandle`. Update all read sites, including `blitFrame`,
      `showTilesWhole`, `withinTiles`, and `drop*` methods
      ([internal/window/replay_tile_blit.go](../../internal/window/replay_tile_blit.go),
      [internal/window/replay_tile_life.go](../../internal/window/replay_tile_life.go)).
- [ ] Add `pendingTiles *tileGrid` and `pendingHandle *ebiten.Image`.
- [ ] Allocate the pending grid from the same content size in `ensureTiles`,
      reusing the active allocation when the size matches.
- [ ] Drop both grids when the size changes, and keep the zero-size guard.
- [ ] Add `handleFor(tileGrid) *ebiten.Image` returning the retained handle, so
      the two grids do not share a backing image.
- [ ] Tests: content size change frees both; same size reuses pending.

### Phase 5 - paint into pending, swap at present

- [ ] Add `paintTilesPending(display, rect)` writing into `pendingTiles`.
- [ ] Route `applyRepaint` to the pending grid and leave the active grid
      untouched during the frame.
- [ ] Swap active and pending at the end of `drawReplayPartial`, after every
      plan runs but before `blitFrame`.
- [ ] Keep `fullReplay` and `oversized` paths working: both allocate into the
      pending grid and still swap, so a full rebuild publishes atomically.
- [ ] A blit-only frame skips both grids and leaves the handles as they are.
- [ ] Tests: a frame that repaints one rect touches only the pending tiles for
      that rect and publishes the old active tiles unchanged.

### Phase 6 - fallback, chrome, and devtools across the swap

- [ ] Keep `paintFullReplay` writing into the pending grid when tiles exist, so
      the active grid stays stable even on a full fallback.
- [ ] Preserve dock clip behavior:
      `clipX`, `withinTiles`, `chromeRects`, and the strip repaint on dock close
      ([internal/window/replay_partial.go:56-92](../../internal/window/replay_partial.go)).
- [ ] Keep the `bitmap fallback` badge draw after the tiles, as today.
- [ ] Keep `partial.gen = app.Generation()` after a successful paint.
- [ ] Devtools counters (`paints`, `paintPx`, `blits`, `blitPx`) measure the
      frame that was just published, not the one just built.
- [ ] Tests: dock close after the swap still repaints the hidden strip once.

### Phase 7 - measurement and gate

- [ ] Run `make test`.
- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` before and after;
      save both outputs.
- [ ] Compare departure-board frame time, heap bytes, allocs per frame, repaint
      counts, and tile counts.
- [ ] Compare a large scrolled page (`examples/tall` if available, else a page
      over 800 by 10000 pixels) for memory.
- [ ] Gate: publish only if one of: tile allocation drops, frame p99 drops by
      more than 10 percent, or a documented blocked caller disappears. Record
      the reason if the gate fails and defer the phase.

### Phase 8 - docs

- [ ] Update [documentation/repaint.md](../../documentation/repaint.md) with the
      active / pending terms and the swap boundary.
- [ ] Update [documentation/performance.md](../../documentation/performance.md)
      and [documentation/performance-baseline.md](../../documentation/performance-baseline.md).
- [ ] Update [PHASES.md](../../PHASES.md) with commit hashes.
- [ ] Add citations back to this file for any line that moved.

## Definition of done

- A build in progress does not change `Page.Display()` or `Page.Generation()`
  until the swap.
- A partial repaint paints into the pending grid and the active grid stays
  readable until publication.
- `TakeDirty()` reports the same region before and after, only at the swap.
- `make test` passes, and the departure-board dump shows no regression in
  frame time or memory when the gate passes.
- The fallback path, dock clip, and scroll-only frames behave exactly as
  before.

## Status

Open. Optional. No code changes attached to this file. The phase only closes on
a measurement that names a blocked build or a blocked draw; otherwise it stays
deferred with the reason recorded here.