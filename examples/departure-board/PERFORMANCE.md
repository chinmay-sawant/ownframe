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
one or two rows. A ripple that touched all eight rows would cover the whole
board. On this board a rotation only changes status tiles (`statusAt`,
`board/data_seed.go`), one or two rows per feed, and the paced dump shows
bands of 188 by 50 CSS pixels with 10 operations meeting them.

The bar is metered so most frames report no damage and the window blits the
buffer it already has. At 10 Hz the frames in between report nothing.

## The damage list and the retarget

Three changes followed the frame-dirty fix.

The page reports a list (`TakeDirtyRects`). Rectangles within 8 px merge, the
list caps at 8 and collapses to its bounding box past that, and a take whose
rects cover more than a third of the content returns one bounding box. The
window repaints one region per rect and blits when the list is empty. The
board marks each changed flap with the box it cached at rebind
(`board/dirty.go`), so a bar update and a cascade band repaint as two regions
instead of one union.

A rotation no longer redraws. The tick calls `Rotate`, which updates the view,
stores it with `SetData` for the next real redraw, and retargets the flaps by
id; the page generation stays put. The tile grid is now fixed width: each
field pads to the widest value the seed can produce, with U+2800 BRAILLE
PATTERN BLANK as the filler, so every tile including the gaps has a text op
to retarget. The fixed grid moved the board from 631 and 640 operations to
844.

The devtools Ops tab caches its row list per `Page.Generation()` and culls
outline strokes to the page area the dock leaves visible
(`internal/window/devtools_sync.go`, `devtools_ops.go`), so a still board
rebuilds no rows.

## Measurements

The dump is headless (`examples/departure-board/dump`) and opens no window.

```
go run ./examples/departure-board -dump -out /tmp/opencode/board-dump-list
go run ./examples/departure-board -dump -pace 16ms -out /tmp/opencode/board-dump-list-paced
```

Without `-pace` the dump ticks as fast as it can, so a whole scenario finishes
in a few milliseconds and the two wall-clock driven updates, the clock text
and the bar fraction, mostly do not happen. Those numbers cannot show the bar
cadence. The second command sleeps 16 ms between ticks, matches the window's
60 ticks per second, and is the real-cadence table.

The `flap` scenario turns the feed before any tick has bound the flipper, so
its row measures the post-redraw steady state. The `cascade` row is the one
that flaps: it ticks three times first, then resyncs with `Advance`. The
`rotate` row does the same but calls `Rotate`, the tick's retarget path, and
the dump fails the run if the rotation moves the generation.

`rcts` is the mean damage-list length per frame and `aMean` the mean summed
list area on a frame that reported one. The first frame after a draw is
`full-buffer`: the initial `Redraw` marks the whole content, and the new
classification compares against that content rect, where the older dump
compared against the canvas and called the same frame `rect`. These runs were
taken at 20:26 local time, inside the board's night window, so the first tick
also pays the one-time night sheet redraw (7 to 11 ms) in `max`; every other
column is steady.

### Real cadence, `-pace 16ms`

| scenario | frames | ops | p50 ms | p95 ms | max ms | kinds | rMean px^2 | rMax px^2 | oMean | rcts | aMean px^2 |
| -------- | ------:| ---:| ------:| ------:| ------:| ----- | ----------:| ---------:| -----:| ----:| ----------:|
| warmup | 30 | 844 | 0.01 | 0.11 | 8.09 | blit 24, full-buffer 1, rect 5 | 126360 | 682000 | 45.3 | 1.0 | 126360 |
| flap | 120 | 844 | 0.01 | 0.05 | 8.10 | blit 97, full-buffer 1, rect 22 | 43440 | 682000 | 24.2 | 1.0 | 43440 |
| idle | 240 | 844 | 0.01 | 0.05 | 10.92 | blit 196, full-buffer 1, rect 43 | 32434 | 682000 | 20.7 | 1.0 | 29914 |
| loop | 240 | 844 | 0.02 | 0.05 | 8.18 | blit 196, full-buffer 1, rect 43 | 32434 | 682000 | 20.7 | 1.0 | 29914 |
| cascade | 150 | 844 | 0.02 | 0.04 | 7.89 | blit 111, full-buffer 2, rect 37 | 67362 | 682000 | 29.5 | 1.4 | 48991 |
| rotate | 150 | 844 | 0.02 | 0.05 | 8.15 | blit 112, full-buffer 1, rect 37 | 51187 | 682000 | 23.9 | 1.4 | 32333 |

`oMean` divides by every frame, so it carries the 844-operation `full-buffer`
frame 0 and, in `cascade`, the `Advance` frame at index 3. Between the frames
that reported damage, the worst steady `rect` frame replays 115 operations
(`cascade` and `rotate`), and the rest sit near 103.

### Flat out, no pace (not the cadence)

| scenario | frames | ops | p50 ms | p95 ms | max ms | kinds | rMean px^2 | rMax px^2 | oMean | rcts | aMean px^2 |
| -------- | ------:| ---:| ------:| ------:| ------:| ----- | ----------:| ---------:| -----:| ----:| ----------:|
| warmup | 30 | 844 | 0.00 | 0.00 | 7.78 | blit 28, full-buffer 1, rect 1 | 348616 | 682000 | 31.6 | 1.0 | 348616 |
| flap | 120 | 844 | 0.00 | 0.00 | 8.46 | blit 118, full-buffer 1, rect 1 | 348616 | 682000 | 7.9 | 1.0 | 348616 |
| idle | 240 | 844 | 0.00 | 0.00 | 7.16 | blit 238, full-buffer 1, rect 1 | 348616 | 682000 | 3.9 | 1.0 | 348616 |
| loop | 240 | 844 | 0.00 | 0.00 | 7.58 | blit 237, full-buffer 1, rect 2 | 237488 | 682000 | 4.4 | 1.0 | 237488 |
| cascade | 150 | 844 | 0.00 | 0.00 | 8.84 | blit 134, full-buffer 2, rect 14 | 96602 | 682000 | 13.0 | 1.8 | 95789 |
| rotate | 150 | 844 | 0.00 | 0.00 | 8.51 | blit 135, full-buffer 1, rect 14 | 57575 | 682000 | 7.4 | 1.9 | 56708 |

### Phase by phase

The before column is `board-dump-verify-paced`, the frame-dirty fix with the
single union rect; the after column is `board-dump-list-paced`. The before
baseline ran on the old tile grid at 631 and 640 operations, and the fixed
grid moved the after board to 844. Compare the fractions, not the raw counts.

| Phase | Before | After |
| ----- | ------ | ----- |
| Damage list | the paced `cascade` worst steady frame replayed 630 of 631 operations through the one union rect | the paced `cascade` worst steady `rect` frame replays 115 of 844, three rects summing 27,532 px squared; 13 of 150 frames report two or more rects |
| Rotation retarget | a rotation called `Advance` (`SetData` plus `Redraw`): on the same grid the paced `cascade` row spends 15.46 ms of `RedrawNS` for the initial draw plus the rotation, and its rotation frame rebuilds the whole buffer | the paced `rotate` row leaves the generation alone and adds no `RedrawNS` (its 6.95 ms is the initial draw); no frame after the first is `full-buffer`, and the worst band is 115 operations |
| Devtools dock | the Ops tab rebuilt one row per painted operation every frame and stroked every outline with no cull; the frame audit estimated 0.5 to 1.5 ms and 200 to 400 KB per frame, plus several ms with the outlines on | the Ops rows are cached per generation (844 rows on this board) and allocate nothing per frame on a still generation (`TestDevOpsRowsStillGenerationNoAlloc`); the outline strokes cull to the visible page rect (`TestDevOpsOutlineCull`) |

`oMean` is the mean replayed operations per frame. It is an approximation:
the dump counts operations whose raw `X`, `Y`, `W`, `H` box meets a rect
(`dump/measure_frame.go`), while the real replay filter converts point-space
bounds to CSS pixels and pads them for strokes and antialiasing
(`internal/replay/op_bounds.go`, `internal/replay/draw_rect.go`). Treat it as
the shape of the difference, not an exact draw count. The dump also measures
the tick, not the GPU draw, so it implies the draw drop from the operation
counts rather than timing the draw itself.

## Verifying by hand

Run `go run ./examples/departure-board` and press F12 (or Ctrl+Shift+I) for
the devtools dock. Read the Frame tab or keep the dock closed. The Ops tab
caches its rows per generation and culls outline strokes to the page area the
dock leaves visible (`internal/window/devtools_sync.go`,
`internal/window/devtools_ops.go`), so a still board pays only for the rows on
screen. Read fps and frame time on the Frame tab. The board does not open with
`WindowOptions.Perf`, so the dock's average and p95 rows read `-`; the Frame
and Draw values are still live.

Expect the FPS to hold at 60, a mostly flat Frame row while the board idles,
and a short rise during a cascade. Fullscreen costs pixels in the blit; the
operation list does not grow with the window.

## Open items

- The plan's phase 1 exit asked for a worst cascade frame under 60 approximate
  operations. The paced cascade measures 115 (three rects, 27,532 px squared)
  on the 844-operation board, down from 630 through the one union rect. The
  remaining cost is several row bands active at once.
- `Page.Invalidate` still walks every box per call through `growBox`
  (`internal/page/page_dirty_box.go`), and `BenchmarkInvalidateMany` marks
  3000 ids on a 3000-box page in about 31 ms. A caller with many ids should
  mark rects, as the board does.
- The dump reads `max` on the tick, so it includes the one-time night sheet
  redraw when the run lands between 20:00 and 06:00 local time, and frame 0
  always carries the initial draw's damage.
