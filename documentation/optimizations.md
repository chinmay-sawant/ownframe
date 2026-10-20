# CSS preview click optimization

The CSS preview example has 43 Start buttons. Each click starts a CSS animation
or transition and updates the selected row and readout. This page records what
one click cost before any optimization, after the first pass, and now.

All numbers are page rebuild time for one click, in milliseconds unless marked
as seconds. `p50`, `p95`, and `max` are per item over the 43 item walk.

## Results

| Metric | Original | First pass | Second pass | Current | Current vs original | Current vs first pass | Current vs second pass |
|---|---:|---:|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 12.5 ms | 53.4x | 8.56x | 2.67x |
| p50 | 648 ms | 95 ms | 32.1 ms | 12.9 ms | 50.2x | 7.36x | 2.49x |
| p95 | 845 ms | 144 ms | 41.9 ms | 15.5 ms | 54.5x | 9.29x | 2.70x |
| max | 864 ms | 163 ms | 48.0 ms | 19.2 ms | 45.0x | 8.49x | 2.50x |
| total | 28.7 s | 4.6 s | 1.44 s | 0.54 s | 53.1x | 8.52x | 2.67x |

As reductions:

| Metric | Lower than original | Lower than first pass | Lower than second pass |
|---|---:|---:|---:|
| average | 98.1% | 88.3% | 62.6% |
| p50 | 98.0% | 86.4% | 59.8% |
| p95 | 98.2% | 89.2% | 63.0% |
| max | 97.8% | 88.2% | 60.0% |
| total | 98.1% | 88.3% | 62.5% |

The original total was 28.7 seconds, not 40 milliseconds. The second-pass
average was 33.4 ms. The current average is 12.5 ms. The current column is a
steady-state real-window run; the first run after launch can be slower while
the font and style caches warm.

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

## Third pass, 33.4 ms to 12.5 ms

The third pass stopped laying out cards that are not on screen and removed the
last duplicate grid build.

- The catalog now uses `Page.SetScrollWindow` with a window of the visible card
  rows plus two rows of overscan. Top and bottom spacer rows keep the grid at
  its full height, so the scroll range and the metadata section do not move.
  The window callback slices `view.Rows` from the full `view.Catalog`.
- The geometry is measured once after the full setup draw:
  `catalog_metrics.go` records each row top and height, the grid gap, the grid
  bottom, and the full content height. A window change uses those numbers
  without reading boxes from the windowed frame.
- Blinkless now lays out the two-column card grid and the two-column playback
  grid in one pass, just like the one-column grids from the second pass. A row
  builds its items once, takes the tallest height, stretches the shorter item
  by extending its border box and regenerating its deferred chrome, and shifts
  the rest. Percentage-height descendants, bottom-anchored absolutes, flex and
  multicol items, transforms, sticky or fixed items, and group-creating styles
  fall back to the two-pass path.
- `spliceInlineParts` grows inline item slices geometrically instead of
  allocating the exact new length on every split.
- Internal page hit tests scan the owned box slice instead of copying all
  boxes through `Page.Boxes()`.
- The app's initial readout is two short lines instead of a 43-line display
  list dump, so the metadata section stays a stable height across a click.

Controlled comparison:

| Comparison | Before | After |
|---|---:|---:|
| Full-page click rebuild, old vs one-pass grid | 33.3 to 36.8 ms | 26.7 to 27.8 ms |
| Real window click rebuild, full page vs windowed | 32.1 to 33.4 ms | 12.2 to 12.9 ms |
| Real window p95 | 41.9 ms | 15.5 to 17.9 ms |
| Real window 43-click total | 1.44 s | 0.54 to 0.55 s |

The visible `press to painted` time stays around 50 ms because the window
draw and the presentation frame are a fixed cost. The page rebuild is the part
this pass cut.

## Visible latency

The redraw time above is the page work. From the real click script, with
`-perf` on and the walk hook active:

| Segment | p50 | p95 | max | average |
|---|---:|---:|---:|---:|
| press to started | 9.9 ms | 18.2 ms | 30.4 ms | 10.4 ms |
| press to painted | 50.6 ms | 51.0 ms | 51.9 ms | 50.7 ms |

`press to started` uses the handler timestamp in the app, so polling latency
does not count. `press to painted` waits for the next layout counter and
includes the redraw plus the window draw.

## How the numbers were measured

The current column comes from 43 real X11 clicks in the desktop window with
the catalog window active. The script starts at the press, not at the start of
its 80 ms hover settle or 50 ms hold. It reads `LastRedraw` from `/state` for
the rebuild time and uses the app's started timestamp for the handler time.
The reported run is the steady state after two warm passes. The first run after
launch can be slower while the shared font and face caches warm.

The original and first-pass columns come from the earlier pointer cycle
measurement in this optimization work. That measurement called the page
pointer methods directly. The current column is the same page rebuild metric
from a real window, after the test harness was corrected.

Two same-benchmark comparisons give the cleanest picture:

- `BenchmarkClickRedraw` with the old grid path: 87.6 to 90.9 ms. With the
  one-column and multi-column one-pass grid: 26.7 to 27.8 ms. The benchmark
  uses the full fixture, so it does not include the catalog window.
- `BenchmarkPointerCycleAll43` with two page builds: 56.5 ms per click. With
  the redraw batch: 30.0 ms per click. This is also the full fixture.
- The real window with the catalog window: 12.2 to 12.9 ms p50 rebuild.

The machine was shared during some earlier runs, so treat the absolute
original numbers as recorded history. The multipliers are the comparison this
work reports.

## Reproduce

Run the real click walk from the repository root:

```sh
CSS_PREVIEW_WALK=1 go run ./examples/css-preview -perf
python3 scripts/css-preview-click-test.py --runs 3
```

The script resets the started rows at the start and stops the previous sample
before each click, so it measures one click at a time instead of a soak of 43
animations. `/state` reports the started timestamp, the last redraw, and the
pipeline counters. `/reset` clears the rows, and `/goto` scrolls the next Start
button under a fixed row. It closes the app with `/quit` after the last pass;
use `--no-close` to keep the window open.

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
- `examples/css-preview/walk_trace.go`, `walk_step.go`, `walk_reset.go`, and
  `scripts/css-preview-click-test.py` drive and measure the walk.
- `examples/css-preview/catalog_metrics.go` and
  `catalog_metrics_range.go` measure the full-page row geometry and pick the
  visible row range.
- `examples/css-preview/catalog_window.go` installs the scroll window and
  updates the spacer rows.
- `examples/css-preview/catalog_window_test.go` checks that the window keeps
  the full grid and metadata geometry.
- `examples/css-preview/batch_click_test.go` covers the batched press through
  a real animation start.

Blinkless:

- `internal/layout/grid_single.go` is the one-pass single-column grid.
- `internal/layout/grid_auto_row.go` is the one-pass multi-column grid used by
  the catalog and playback rows.
- `internal/layout/grid.go` selects the one-pass paths after grid placement.
- `internal/layout/inline.go` grows inline item slices geometrically.
- `internal/layout/inline_hyphenation.go` skips scanning text with no soft
  hyphen.
- `internal/layout/layout.go` holds the one-entry face cache.

Ownframe currently pins Blinkless
`v0.0.3-0.20261020122842-3fa0a49a895b`. That pin includes the one-pass
single-column grid, not the multi-column `grid_auto_row.go`. The 12.5 ms
number also uses the local multi-column grid, so shipping it needs a Blinkless
commit and a pin bump.
