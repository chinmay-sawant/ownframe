# Phase 03, invalidation cost

Goal: marking stays cheap when many elements change, hover costs nothing when
the pointer is still, and the devtools dock stops distorting the numbers the
user reads.

## Findings from the 2026-10-12 audit

Recorded in [audit-2026-10-12.md](audit-2026-10-12.md), with file references:

- `Page.Invalidate` does two map lookups plus `growBox`, which scans every box
  on each call (`internal/page/page_invalidate.go`,
  `internal/page/page_dirty_box.go`). `MarkRect` is one union with none of
  that. The board calls `Invalidate` once per changed cell.
- `markPending` allocates the pending map on the first invalidate of a frame,
  and `TakeDirty` nils it, so a busy frame allocates a map.
- `app.Hover` hit-tests every box every frame even when the pointer did not
  move (`internal/window/pointer.go`, `internal/page/pointer_state.go`).
- The Ops tab rebuilds one row per operation every frame, estimated at 0.5 to
  1.5 ms and 200 to 400 KB on this board, and the paint outline toggle strokes
  every operation with no visible-rect cull, estimated at 2 to 5 ms more
  (`internal/window/devtools_ops.go`, `devtools_ops_lines.go`).

## Checklist

- [x] Board: `dirty.go` marks the box rect of each changed cell with
      `MarkRect`, using the box the flipper already holds, instead of
      `Invalidate` per id.
- [x] Page: make the id path cheap for callers that must use ids, either a box
      id map or a batched `MarkIDs`. Keep the previous-rect union that makes a
      moved element repaint its old space.
- [x] Page: stop allocating the pending map per frame when nothing is pending.
- [x] Window: `pointer.go` and `hover_cursor.go` skip `Page.Hover` while the
      pointer point is unchanged; a press or a move still routes.
- [x] Devtools: cache the Ops row list per `Page.Generation()` and cull outline
      strokes to the visible list rect.
- [x] Benchmarks: `BenchmarkInvalidateMany` in `internal/page`; extend the
      devtools draw test with an allocation budget per frame.
- [x] Docs: one line in `documentation/devtools.md` saying readings taken with
      the Ops tab open include the dock's own work, and which tab to read for
      FPS.

## Exit

- The idle board with the dock open on the Ops tab holds 60 FPS.
- The dump's `AllocsFrame` stays under 20 KB per frame with animations on.
- The devtools draw test allocates nothing per frame when the generation did
  not move, and the outlines are culled.
- `make test` passes.

## Risks

- Culling outlines must not change the picking behavior of the Ops tab. Keep
  the pick path reading the full list.
- A cached row list is only valid for one generation; invalidate it on
  `SetTheme`, `Load`, and resize like the other caches.
