# CSS preview click optimization

The CSS preview example has 43 Start buttons. Each click starts a CSS animation
or transition and updates the selected row and readout. This page records what
one click cost before any optimization, after each pass, and now.

All numbers are page rebuild time for one click, in milliseconds unless marked
as seconds. `p50`, `p95`, and `max` are per item over the 43 item walk.

## Results

| Metric | Original | First pass | Second pass | Third pass | Current | Current vs original | Current vs third pass |
|---|---:|---:|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 12.5 ms | 6.16 ms | 108x | 2.03x |
| p50 | 648 ms | 95 ms | 32.1 ms | 12.9 ms | 6.20 ms | 105x | 2.08x |
| p95 | 845 ms | 144 ms | 41.9 ms | 15.5 ms | 7.70 ms | 110x | 2.01x |
| max | 864 ms | 163 ms | 48.0 ms | 19.2 ms | 8.50 ms | 102x | 2.26x |
| total | 28.7 s | 4.6 s | 1.44 s | 0.54 s | 0.265 s | 108x | 2.04x |

As reductions:

| Metric | Lower than original | Lower than first pass | Lower than second pass | Lower than third pass |
|---|---:|---:|---:|---:|
| average | 99.1% | 94.2% | 81.6% | 50.7% |
| p50 | 99.0% | 93.5% | 80.7% | 51.9% |
| p95 | 99.1% | 94.7% | 81.6% | 50.3% |
| max | 99.0% | 94.8% | 82.3% | 55.7% |
| total | 99.1% | 94.2% | 81.6% | 50.9% |

The original total was 28.7 seconds, not 40 milliseconds. The second-pass
average was 33.4 ms, and the third-pass average was 12.5 ms. The current
average is 6.16 ms. The current column is the mean of five
`BenchmarkWindowedPointerCycle` runs; the first run after launch can be slower
while the font and style caches warm.

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

## Fourth pass, 12.5 ms to 6.16 ms

The fourth pass stopped paying for layout that a click does not need.

- The catalog window now renders only the rows that intersect the viewport.
  The previous window added two rows of overscan above and below. That cut the
  average window from 8.72 rows to 5.37 rows per click. Scroll range changes
  became about 15% more frequent, but the total scroll work stayed about the
  same because each rebuild is smaller.
- The card header lost its flex formatting context. `.catalog-details h3` is a
  block, the property number is an inline-block badge, and the row and details
  containers are blocks. The playback row keeps its grid. The header flex was
  the largest single card cost at about 20% of the rebuild.
- Pages with only button controls skip the control-state source copy and the
  form style block. The catalog has 43 Start buttons and no input, textarea,
  or select.
- ASCII pages skip the emoji replacement scan copy. The catalog has no emoji.
- The box index map is cleared and refilled instead of allocated again. The
  immutable target list and its key and button maps are built once.
- The template builder grows from the previous source length instead of
  growing through repeated copies.

Controlled comparison, 43-click desktop pointer cycle:

| Comparison | Before | After |
|---|---:|---:|
| Average page rebuild | 11.7 to 13.0 ms | 6.0 to 6.4 ms |
| p50 page rebuild | 12.2 to 13.1 ms | 5.97 to 6.47 ms |
| p95 page rebuild | 14.3 to 17.7 ms | 7.3 to 8.0 ms |
| 43-click total | 0.50 to 0.56 s | 0.258 to 0.275 s |

The new `BenchmarkWindowedPointerCycle` drives the full windowed path:
960x860, perf on, the catalog window installed, scroll the next Start button
under a fixed line, then a batched press, focus, click, and release. It reports
`LastRedraw` for each of the 43 clicks. The numbers above are five runs of
`-benchtime=1x`.

The page rebuild is now close to the one-row layout floor. The remaining work
is the display-list build for the rows on screen. Skipping that would mean
patching the retained display on Start instead of rebuilding the page, which
is a separate design with a different set of tradeoffs.

## Next: the two gaps against a browser

This is the list to work from next. Nothing here is committed yet.

### What a browser does

Chrome keeps all 43 cards in the layout tree. Four things keep that cheap.

- Scrolling is a compositor operation. The offset moves on the compositor
  thread, so the main thread does no work while the page scrolls.
- Paint holding keeps the old tiles when a scroll outruns the main thread, so
  a scroll never waits for layout.
- Incremental layout re-lays out the changed card and its ancestors, not the
  page. A class change on one card is a fraction of a millisecond.
- Compositor animations move a layer for a transform or an opacity, so a
  running animation does not repaint the page.

`content-visibility: auto` with `contain-intrinsic-size` is the browser
answer to skipping off-screen layout. It needs a size placeholder to keep the
scrollbar stable, which is what the catalog spacer rows do.

### Gap one, a click re-lays out every visible card

A click in the CSS preview costs 7 to 10 ms with the catalog windowed and
about 21 ms when the whole 43 row catalog is rendered, which is what the Play
All button does today. The work is the display-list build, which is 93
percent of a click redraw.

- OwnFrame can copy the changed card's operations into the retained display
  and mark its region dirty, with `Page.PatchDisplay` and `Page.PatchOp`. That
  removes the rebuild for the selection, the state text, and the started
  class, and it helps every click, not only the catalog.
- Blinkless could re-lay out one subtree and keep the rest. The style
  resolution memo and the box cache are pass-local today, so nothing carries
  over between frames.

### Gap two, a running animation repaints the page

With 43 animations running, every frame replays the display list. A browser
moves a layer instead.

- The replay path already carries an op transform matrix, a group composite,
  and viewport-pinned ops. What is missing is a layer the window can blit
  without repainting the page under it.

### The Play All click cost

The Play All button keeps the catalog windowed, so a click stays at the
windowed cost while all 43 previews still play.

- The rows on screen start right away. The rows off screen wait. They keep
  their start time, so an animation is already at the right phase when its row
  scrolls in.
- A transition needs its starting paint captured while the row is still
  unstarted. A row that scrolls in for the first time renders once without the
  started class, captures that paint, then renders with the class and binds.
  Two renders for that card only, and only for the cards the user visits.
- The edge cases are a row that first appears after its animation would have
  finished, a row that is scrolled past, and a reset.

Measured at 1908x999, the press costs 6.7 ms with the catalog windowed, and
the display holds 1660 operations and 119 boxes. Before, the press rendered
all 43 rows, 8800 operations and 537 boxes, and cost 46 to 50 ms in total.


## Visible latency

The redraw time above is the page work. From the real click script, with
`-perf` on and the walk hook active:

| Segment | p50 | p95 | max | average |
|---|---:|---:|---:|---:|
| press to started | 9.9 ms | 18.2 ms | 30.4 ms | 10.4 ms |
| press to painted | 50.6 ms | 51.0 ms | 51.9 ms | 50.7 ms |

These are third-pass numbers. The fourth pass cuts the page rebuild further,
but the fixed window draw and presentation frame still dominate `press to
painted`. `press to started` uses the handler timestamp in the app, so polling
latency does not count. `press to painted` waits for the next layout counter
and includes the redraw plus the window draw.

## How the numbers were measured

The current column comes from `BenchmarkWindowedPointerCycle`, which follows
the same 43-click desktop path as the X11 script but reads `LastRedraw` in
process. That removes the software X server and presentation frame from the
measurement. The benchmark runs one warm pass, then reports the average, p50,
p95, and max over one or more measured passes. The numbers above are the mean
of five `-benchtime=1x` runs. The third-pass and earlier columns came from the
real X11 walk and the earlier pointer cycle measurement.

The original and first-pass columns come from the earlier pointer cycle
measurement in this optimization work. That measurement called the page
pointer methods directly. The third-pass column is the same page rebuild
metric from a real window. The fourth-pass column uses the in-process
windowed benchmark.

Two same-benchmark comparisons give the cleanest picture:

- `BenchmarkClickRedraw` with the old grid path: 87.6 to 90.9 ms. With the
  one-column and multi-column one-pass grid: 26.7 to 27.8 ms. The benchmark
  uses the full fixture, so it does not include the catalog window.
- `BenchmarkPointerCycleAll43` with two page builds: 56.5 ms per click. With
  the redraw batch: 30.0 ms per click. This is also the full fixture.
- The third pass's real window with the catalog window: 12.2 to 12.9 ms p50
  rebuild.
- The fourth pass's in-process windowed benchmark: 6.0 to 6.4 ms average
  rebuild over five runs.

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

Run the windowed benchmark that matches the current column:

```sh
go test ./examples/css-preview -run '^$' \
  -bench '^BenchmarkWindowedPointerCycle$' \
  -benchtime=1x -count=5
```

Use `-benchtime=1x` for one 43-click pass per run, or a higher value for a
longer run. The full-fixture benchmarks need at least `-benchtime=3x` before
their numbers settle:

```sh
go test ./examples/css-preview -run '^$' \
  -bench 'BenchmarkClickRedraw|BenchmarkPointerCycleAll43' \
  -benchtime=3x -count=3
```

`BenchmarkWindowedPointerCycle` runs the windowed, batched click sequence and
reports `LastRedraw` per click. `BenchmarkPointerCycleAll43` runs the same
pointer sequence on the full fixture. `BenchmarkPointerCycleAll43NoBatch` runs
it without the batch. `BenchmarkClickRedraw` measures one `SetData` plus
`Redraw` on the full fixture.

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
- `examples/css-preview/window_bench_test.go`,
  `window_bench_setup_test.go`, `window_bench_click_test.go`, and
  `window_bench_pass_test.go` run the windowed 43-click benchmark.
- `examples/css-preview/catalog_target_cache.go` and `model.go` cache the
  immutable target list and its lookup maps.
- `examples/css-preview/testdata/css-preview/styles.css` holds the flat card
  header and row styles.
- `internal/emoji/emoji_ascii.go` skips the replacement copy on ASCII pages.
- `internal/page/form_spans.go` skips the control rewrite when only buttons
  are present.
- `internal/page/page_boxindex.go` reuses the box index map.
- `internal/page/page_draw_stages.go` sizes the template builder from the
  previous source.
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

Ownframe pins Blinkless
`v0.0.3-0.20261020132936-e0c3c9ca4f30`, which includes both one-pass grid
paths. The 12.5 ms and 6.16 ms columns use that pin.
