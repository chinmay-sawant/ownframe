# CSS preview click optimization

The CSS preview example has 43 Start buttons. Each click starts a CSS animation
or transition and updates the selected row and readout. This page records what
one click cost before any optimization, after the first pass, and now.

All numbers are page rebuild time for one click, in milliseconds unless marked
as seconds. `p50`, `p95`, and `max` are per item over the 43 item walk.

## Results

| Metric | Original | First pass | Current | Current vs original | Current vs first pass |
|---|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 20.0x | 3.20x |
| p50 | 648 ms | 95 ms | 32.1 ms | 20.2x | 2.96x |
| p95 | 845 ms | 144 ms | 41.9 ms | 20.2x | 3.44x |
| max | 864 ms | 163 ms | 48.0 ms | 18.0x | 3.40x |
| total | 28.7 s | 4.6 s | 1.44 s | 19.9x | 3.19x |

As reductions:

| Metric | Lower than original | Lower than first pass |
|---|---:|---:|
| average | 95.0% | 68.8% |
| p50 | 95.0% | 66.2% |
| p95 | 95.0% | 70.9% |
| max | 94.4% | 70.6% |
| total | 95.0% | 68.7% |

The original total was 28.7 seconds, not 40 milliseconds. The current max is
48.0 ms and the current average is 33.4 ms.

## First pass, 667 ms to 107 ms

The first pass removed work that painted the same pixels and reduced the
layout cost of generated content.

- Skip a hover, press, or release redraw when the styled sheets cannot match
  that state. The CSS preview only uses `:focus-visible`, so hover and press
  were rebuilding the page for nothing.
- Use `RedrawIfChanged` for the generic click path.
- Recover a press and release that land between two frames as one click.
- Repaint only the tick's damaged boxes while up to eight samples animate,
  then fall back to a full replay.
- Index Blinkless selectors by the rightmost id, class, or tag, and skip the
  generated-content scan when no rule declares `content`.

That pass was 6.2x faster than the original on average (667 ms to 107 ms),
6.8x on p50, 5.9x on p95, 5.3x on max, and 6.2x on total.

## Second pass, 107 ms to 33.4 ms

The second pass removed a second full page build and made the grid build each
item once.

- A desktop press ran two page builds before any frame was drawn:
  `SelectAt` focused the button and redrew, then `Click` ran the handler and
  redrew. The focus build was never presented.
- Add a redraw batch around press, focus, click, and release. The page records
  that a draw is pending and draws the final focus, active, source, and
  selection state once. A plain `Redraw` in the batch forces the final draw;
  an `IfChanged`-only batch still honors the unchanged check. The window,
  touch tap, devtools picker, and web click route all use it.
- Blinkless now lays out a one-column grid in one pass. A row with one item
  sizes to that item, so the item can be built at its final position instead
  of being measured with `noEmit` and built again. The CSS preview uses a
  one-column grid for every card and every details block.
- Return early from soft-hyphen scanning when the text has no soft hyphen.
- Cache the last font face key so an inline run reusing one resolved style
  does not hash the face map for every rune.

Controlled comparisons on the same machine:

| Comparison | Before | After |
|---|---:|---:|
| One click rebuild, old vs new grid | 87.6 to 90.9 ms | 33.3 to 36.8 ms |
| 43 item pointer cycle, two builds vs batched | 56.5 ms/click | 30.0 ms/click |

The final column in the results table includes both changes. The grid change
halves the rebuild. The redraw batch removes the extra focus build that the
real window paid on every first click.

## Visible latency

The redraw time above is the page work. From the real click script, with
`-perf` on and the walk hook active:

| Segment | p50 | p95 | max | average |
|---|---:|---:|---:|---:|
| press to started | 7.6 ms | 18.5 ms | 21.4 ms | 8.9 ms |
| press to painted | 50.5 ms | 53.3 ms | 54.4 ms | 51.0 ms |

`press to started` uses the handler timestamp in the app, so polling latency
does not count. `press to painted` waits for the next layout counter and
includes the redraw plus the window draw.

## How the numbers were measured

The current column comes from 43 real X11 clicks in the desktop window. The
script starts at the press, not at the start of its 80 ms hover settle or 50 ms
hold. It reads `LastRedraw` from `/state` for the rebuild time and uses the
app's started timestamp for the handler time.

The original and first-pass columns come from the earlier pointer cycle
measurement in this optimization work. That measurement called the page
pointer methods directly. The current column is the same page rebuild metric
from a real window, after the test harness was corrected.

Two same-benchmark comparisons give the cleanest picture:

- `BenchmarkClickRedraw` with the old grid path: 87.6 to 90.9 ms. With the new
  grid path: 33.3 to 36.8 ms.
- `BenchmarkPointerCycleAll43` with two page builds: 56.5 ms per click. With
  the redraw batch: 30.0 ms per click.

The machine was shared during some earlier runs, so treat the absolute
original numbers as recorded history. The multipliers are the comparison this
work reports.

## Reproduce

Run the real click walk from the repository root:

```sh
CSS_PREVIEW_WALK=1 go run ./examples/css-preview -perf
python3 scripts/css-preview-click-test.py
```

The script resets the started rows at the start and stops the previous sample
before each click, so it measures one click at a time instead of a soak of 43
animations. `/state` reports the started timestamp, the last redraw, and the
pipeline counters. `/reset` clears the rows, and `/goto` scrolls the next Start
button under a fixed row.

Run the benchmarks:

```sh
go test ./examples/css-preview -run '^$' \
  -bench 'BenchmarkClickRedraw|BenchmarkPointerCycleAll43' \
  -benchtime=1x -count=1
```

`BenchmarkPointerCycleAll43` runs the batched window sequence.
`BenchmarkPointerCycleAll43NoBatch` runs the same sequence without the batch.
`BenchmarkClickRedraw` measures one `SetData` plus `Redraw`.

## Files

Ownframe:

- `internal/host/batch.go` defines `RedrawBatcher` and `Batch`.
- `internal/page/page_redraw_batch.go` defers and flushes the draw.
- `internal/page/page_draw.go` records the pending draw.
- `internal/page/page_render_key.go` stores the batch fields with the render
  state.
- `internal/window/drag_select.go`, `tap_at.go`,
  `devtools_pick_content.go`, and `internal/web/click_http.go` wrap the
  pointer sequence.
- `examples/css-preview/walk_trace.go`, `walk_reset.go`, and
  `scripts/css-preview-click-test.py` drive and measure the walk.
- `examples/css-preview/batch_click_test.go` covers the batched press through
  a real animation start.

Blinkless:

- `internal/layout/grid_single.go` is the one-pass single-column grid.
- `internal/layout/grid.go` selects it after grid placement.
- `internal/layout/inline_hyphenation.go` skips scanning text with no soft
  hyphen.
- `internal/layout/layout.go` holds the one-entry face cache.

Ownframe pins Blinkless `v0.0.3-0.20261020122842-3fa0a49a895b`.
