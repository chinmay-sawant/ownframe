# Departure board performance

The board was reported at 13 to 14 FPS with a tick at 60 per second: about
23 ms in draw, about 60 ms per frame, and worse in fullscreen.

## Why it happened

A page with a registered tick keeps the window's full replay unless its tick
calls `Page.UseFrameDirty`. `internal/window/replay_frame.go` makes
`fullReplay` true when the page reports `Ticking` and `FrameDirty` is false.
`internal/window/replay_partial.go` then skips the persistent buffer and calls
`directReplay`, which walks the whole display list and draws every operation
whose bounds meet the window (`internal/window/replay_oversize.go`,
`internal/replay/visible.go`). The board's tick never called `UseFrameDirty`,
so all 631 to 640 operations were drawn again on every frame at 60 ticks per
second.

The dump run against the unfixed tree shows the shape exactly: every frame of
every scenario was `full-replay`, and the mean replayed operations per frame
equalled the full operation count.

## The fix

The tick leads with `a.page.UseFrameDirty()` (`board/tick.go`). It has to run
every tick because the window's `FrameDirty` call consumes the opt-in, and a
tick that changes nothing still keeps the window on the buffer path, where an
empty `TakeDirty` becomes a blit instead of a full replay.

The tick then names only what changed:

- The header clock, when its text changes (`setClockText`, `board/dirty.go`).
- The refresh bar, at most every six ticks and only when the whole-pixel
  width changed (`moveBar`, `board/dirty.go`; `barEvery = 6` is about 10 Hz
  at 60 ticks per second).
- Every flap tile whose glyph changed this frame (`flipper.step` returns the
  ids, `invalidateCells` marks them; `board/flip.go`, `board/dirty.go`).

The cascade is a row band (`rowDelay = 8`, `board/flip_rebind.go`). Every cell
in one screen row starts together, so one frame's changed cells stay inside
one or two rows. The window takes exactly one dirty rect per frame: the page
stores one rectangle and `TakeDirty` returns it
(`internal/page/page_dirty.go`), and `markRect` unions marks into a bounding
box (`internal/page/page_dirty_box.go`). A ripple that touched all eight rows
would union into the whole board, so the row band keeps the rect narrow. On
this board a rotation only changes status tiles (`statusAt`,
`board/data_seed.go`), one or two rows per feed, and the paced dump shows
bands of 188 by 50 CSS pixels with 10 operations meeting them.

The bar is metered for the same single-rect reason. If its fill moved every
frame, every frame would report a rect and the window would never blit; at
10 Hz the frames in between report nothing.

## Measurements

The dump is headless (`examples/departure-board/dump`) and opens no window.

```
go run ./examples/departure-board -dump -out /tmp/opencode/board-dump-verify
go run ./examples/departure-board -dump -pace 16ms -out /tmp/opencode/board-dump-verify-paced
```

Without `-pace` the dump ticks as fast as it can, so a whole scenario finishes
in a few milliseconds and the two wall-clock driven updates, the clock text
and the bar fraction, mostly do not happen. Those numbers cannot show the bar
cadence. The second command sleeps 16 ms between ticks, matches the window's
60 ticks per second, and is the real-cadence table.

The `flap` scenario turns the feed before any tick has bound the flipper, so
its row measures the post-redraw steady state. The `cascade` row is the one
that flaps: it ticks three times first, then rotates.

### Real cadence, `-pace 16ms`

| scenario | frames | ops | p50 ms | p95 ms | max ms | kinds | rMean px^2 | rMax px^2 | oMean |
| -------- | ------:| ---:| ------:| ------:| ------:| ----- | ----------:| ---------:| -----:|
| warmup | 30 | 640 | 0.01 | 0.13 | 0.28 | blit 24, rect 6 | 126360 | 682000 | 34.8 |
| flap | 120 | 631 | 0.01 | 0.07 | 0.40 | blit 97, rect 23 | 43440 | 682000 | 18.8 |
| idle | 240 | 640 | 0.02 | 0.09 | 0.28 | blit 196, rect 44 | 32434 | 682000 | 16.3 |
| loop | 240 | 631 | 0.02 | 0.09 | 1.02 | blit 195, rect 45 | 29249 | 682000 | 16.2 |
| cascade | 150 | 631 | 0.02 | 0.11 | 0.39 | blit 112, rect 38 | 64425 | 682000 | 29.7 |

### Flat out, no pace (not the cadence)

| scenario | frames | ops | p50 ms | p95 ms | max ms | kinds | rMean px^2 | rMax px^2 | oMean |
| -------- | ------:| ---:| ------:| ------:| ------:| ----- | ----------:| ---------:| -----:|
| warmup | 30 | 640 | 0.00 | 0.00 | 0.26 | blit 28, rect 2 | 348616 | 682000 | 24.0 |
| flap | 120 | 631 | 0.00 | 0.00 | 0.23 | blit 118, rect 2 | 348616 | 682000 | 5.9 |
| idle | 240 | 640 | 0.00 | 0.00 | 0.21 | blit 238, rect 2 | 348616 | 682000 | 3.0 |
| loop | 240 | 631 | 0.00 | 0.00 | 0.23 | blit 238, rect 2 | 348616 | 682000 | 3.0 |
| cascade | 150 | 631 | 0.00 | 0.01 | 0.28 | blit 135, rect 15 | 99122 | 682000 | 9.7 |

### Before and after

The before column is the same dump run against the unfixed tree. That run
predates the `cascade` scenario, and every scenario it measured was 100
percent `full-replay` with the full operation count as the mean replayed ops.

| scenario | before: kinds | before oMean | after (paced): kinds | after (paced) oMean |
| -------- | ------------- | -----------: | -------------------- | ------------------: |
| warmup | full-replay 30 | 640.0 | blit 24, rect 6 | 34.8 |
| flap | full-replay 120 | 631.0 | blit 97, rect 23 | 18.8 |
| idle | full-replay 240 | 640.0 | blit 196, rect 44 | 16.3 |
| loop | full-replay 240 | 631.0 | blit 195, rect 45 | 16.2 |
| cascade | not measured | | blit 112, rect 38 | 29.7 |

`oMean` is the mean replayed operations per frame. It is an approximation:
the dump counts operations whose raw `X`, `Y`, `W`, `H` box meets the rect
(`dump/measure_frame.go`), while the real replay filter converts point-space
bounds to CSS pixels and pads them for strokes and antialiasing
(`internal/replay/op_bounds.go`, `internal/replay/draw_rect.go`). Treat it as
the shape of the difference, not an exact draw count. The dump also measures
the tick, not the GPU draw, so it implies the draw drop from the operation
counts rather than timing the draw itself.

## Verifying by hand

Run `go run ./examples/departure-board` and press F12 (or Ctrl+Shift+I) for
the devtools dock. Read the Frame tab or keep the dock closed. The Ops tab
rebuilds one row per painted operation, about 640 for this board, every frame,
and its Show paint outlines toggle strokes every operation without culling
(`internal/window/devtools_ops_lines.go`, `internal/window/devtools_ops.go`).
The frame audit estimated that work at 0.5 to 3 ms per frame with the tab open
and several ms more with outlines on, so it lands in the numbers you are
trying to read. The board does not open with `WindowOptions.Perf`, so the
dock's average and p95 rows read `-`; the Frame and Draw values are still
live.

Expect the FPS to hold at 60, a mostly flat Frame row while the board idles,
and a short rise during a cascade. Fullscreen costs pixels in the blit; the
operation list does not grow with the window.

## Follow-ups

- One dirty rect per frame is the limit this fix lives within. A multi-rect
  `TakeDirty` would let the bar and a cascade band repaint separately instead
  of unioning into one bounding box. The paced cascade shows the union: when
  a bar update lands on a band frame the rect grows to 952 by 274 with 604
  operations meeting it by the dump's measure. Files:
  `internal/page/page_dirty.go`, `internal/page/page_dirty_box.go`,
  `internal/window/replay_partial.go`, `internal/window/replay_plan.go`.
- `Page.Invalidate` scans every box for each call (`markElement` calls
  `growBox`, `internal/page/page_invalidate.go`,
  `internal/page/page_dirty_box.go`) and allocates the pending map on the
  first invalidate of a frame. `Page.MarkRect` is one union with none of that
  (`internal/page/page_frame.go`, `internal/page/page_dirty_box.go`). The
  flipper already holds each cell's operation and box, so marking rects
  directly in `board/dirty.go`, or batching ids in the library, removes the
  scans and the per-frame map. The bar would also shrink from a full-track
  mark to the changed sliver.
- `app.Hover` hit-tests every box every frame even when the mouse is still.
  `internal/window/pointer.go` calls `hoverCursor`, which calls `Page.Hover`,
  and `internal/page/pointer_state.go` runs `boxIDAt`, a linear scan of every
  box, before `setHover` sees that the id did not change. Skipping the call
  when the point is unchanged removes the scan. Files:
  `internal/window/pointer.go`, `internal/window/hover_cursor.go`,
  `internal/page/pointer_state.go`.
