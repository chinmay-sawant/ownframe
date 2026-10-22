# CSS preview click optimization

The CSS preview example has 43 Start buttons. Each click starts a CSS animation
or transition and updates the selected row and readout. This page records what
one click cost before any optimization, after each pass, and now.

All numbers are page rebuild time for one click, in milliseconds unless marked
as seconds. `p50`, `p95`, and `max` are per item over the 43 item walk.

## Results

The metric is the page rebuild one click caused.

| page rebuild | Original | First pass | Second pass | Third pass | Fourth pass | Now |
|---|---:|---:|---:|---:|---:|---:|
| average | 667 ms | 107 ms | 33.4 ms | 12.5 ms | 6.16 ms | 0 ms |
| p50 | 648 ms | 95 ms | 32.1 ms | 12.9 ms | 6.20 ms | 0 ms |
| p95 | 845 ms | 144 ms | 41.9 ms | 15.5 ms | 7.70 ms | 0 ms |
| max | 864 ms | 163 ms | 48.0 ms | 19.2 ms | 8.50 ms | 0 ms |
| total | 28.7 s | 4.6 s | 1.44 s | 0.54 s | 0.265 s | 0 s |

The zero in the current column is a measurement, not a missing number. A click
now patches the elements it changed and paints them, so the page's own redraw
counter does not move and there is no rebuild left to time. The click script
reports 0.0 ms for the rebuild of every one of the 43 clicks. The walk does
not become free: it still waits for each painted frame, and that wait is the
number in **Visible latency** further down. The total row is the sum of the
rebuild over the whole walk, so it is 0 s of page rebuild, not 0 s of walk.

The reductions below are about the rebuild only. Every cell is the drop from
the original, so the row reads against the same baseline as the table above.

| page rebuild | Original | First pass | Second pass | Third pass | Fourth pass | Now |
|---|---:|---:|---:|---:|---:|---:|
| average | baseline | 84.0% | 95.0% | 98.1% | 99.1% | 100% |
| p50 | baseline | 85.3% | 95.0% | 98.0% | 99.0% | 100% |
| p95 | baseline | 83.0% | 95.0% | 98.2% | 99.1% | 100% |
| max | baseline | 81.1% | 94.4% | 97.8% | 99.0% | 100% |
| total | baseline | 84.0% | 95.0% | 98.1% | 99.1% | 100% |

The last column is 100% because the click no longer rebuilds the page. It is a
drop in the rebuild, not in the time the walk takes.

The original total was 28.7 seconds of page rebuild, not 40 milliseconds. The
second-pass average was 33.4 ms, and the third-pass average was 12.5 ms. The
fourth-pass average was 6.16 ms. The current column is the real click script,
where the click no longer rebuilds the page at all. The
fourth-pass column is the mean of five `BenchmarkWindowedPointerCycle` runs,
and that benchmark runs at 960x860, while the current column comes from the
real script at 1908x999. The two columns are not the same measurement, and the
current column is the better number for the click path. The first run after
launch can be slower while the font and style caches warm. That benchmark
drives the page without the example's own handlers, so it still measures a full
rebuild, and it is the number to compare against when the page is used without
the element patch.

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
  `page.Testing().SelectAt` focused the button and redrew, then `page.Click` ran the handler and
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
  rows plus two rows of overscan. The template emits one spacer row per side,
  and that spacer spans both columns of the grid, so the grid keeps its full
  height, the scroll range does not move, and the metadata section does not
  move. Commit 8ec0125 gave each column its own spacer instead, so the outer
  grid keeps the one-pass layout. The window callback slices `view.Rows` from
  the full `view.Catalog`.
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
  The previous window added two rows of overscan above and below. The code
  change was the overscan constant dropping from 2 to 0, with a clamp at the
  end of the row list. The recorded averages of that pass went from 8.72 rows
  to 5.37 rows per click; nothing in this repo instruments those two numbers.
  Scroll range changes became about 15% more frequent, but the total scroll
  work stayed about the same because each rebuild is smaller.
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

Against the first pass, the fourth pass is 94.2% lower on average and 94.2%
lower on the 43-click total.

The page rebuild is now close to the one-row layout floor. The remaining work
is the display-list build for the rows on screen. Skipping that would mean
patching the retained display on Start instead of rebuilding the page, which
is a separate design with a different set of tradeoffs.

## Fifth pass, the click patches the page

The fifth pass removed the rebuild from the click itself. The click state
lives in the parsed tree, the click patches the elements it changed, and the
page paints the patched region. Nothing is rebuilt.

- The click state lives in the parsed tree. The card's class and the state
  span's text are patched, not re-rendered from the template data, so the
  executed source does not change and the page keeps its display.
- `Page.PatchElement` re-lays out one element of the retained page and copies
  the patched page back into the display. The engine lays out the element's
  subtree alone and hands the whole patched page over, because a patch that
  changes the operation count moves every later operation index.
- `Handlers.ClickPaint` reports that the handler painted the frame, so the
  click skips the redraw and the redraw batch waits for no draw that will not
  come.
- `Page.SetAfterRedraw` puts the tree state back after the template replaced
  the tree, so a scroll or a reload does not lose the started cards.

Measured with the real click script, 43 clicks, 1908x999:

| Measure | Fourth pass | Now |
|---|---:|---:|
| Click redraw, p50 | 25.2 ms | 0.0 ms |
| Click redraw, average | 23.5 ms | 0.0 ms |
| press to started, p50 | 17.4 ms | 15.8 ms |
| press to painted, p50 | 50.9 ms | 51.0 ms |

The click does not rebuild the page, so there is no page rebuild to measure.
The last row is unchanged because it is dominated by the presentation frame of
the software compositor, not by page work.

The cost of the patch itself, measured by the harness on the whole 43 row
catalog:

| Path | Time | Allocated | Applied |
|---|---:|---:|---:|
| Full redraw, card class | 34.5 ms | 20.18 MB | |
| Full redraw, state class | 39.2 ms | 19.86 MB | |
| Element patch, card class | 6.08 ms | 6.31 MB | 30 of 30 |
| Element patch, state class | 4.72 ms | 5.99 MB | 30 of 30 |

The card row patch applies in every run, 30 of 30. The element the test
probes with a class that carries no rules is a different case, and the engine
still refuses it because the element changes size. The class the example uses
is spliced, 203 operations, and the running example applied 85 of 85 card
patches and 43 of 43 readout patches.

## The card patch and its guards

`Page.PatchElement` re-lays out one element of the retained page and copies its
operations into the retained display. It edits the cached tree first, so the
cascade sees the new class or the new text, then it asks Blinkless for that
element's operations, boxes, and old and new bounds. Blinkless refuses every
case it cannot represent, and the host falls back to a full rebuild.

The engine's guards:

| Guard | Why it exists | The CSS preview card |
|---|---|---|
| No context, or a cancelled one | The layout pass has to stop on demand | No |
| No styled document, no previous display, no element id | The patch needs a tree, a display, and an id | No |
| The id is not in the tree | The element is found by id | Yes. The state text has no id, and a row outside the catalog window is not in the tree |
| The element owns no box | The splice patches the element's box | No |
| The element is the document root | The root carries the canvas fill | No |
| The element is not a block box | A table box is placed by the table layout | No. A card is a grid item, and Blinkless builds grid items as block boxes |
| An ancestor carries a transform or an opacity | The subtree's operations carry that transform, and a subtree laid out on its own does not | No |
| The subtree holds a float, an out-of-flow box, a sticky box, an anchored element, or a named flow | Each is placed from state the kept result does not carry | Only for an out-of-flow box whose containing block is outside the subtree. The preview label and the preview target are absolute inside the relative stage, and the stage is inside the card, so the card is spliced |
| An anchored box sits outside the subtree | That box would go stale when the anchor moves | No |
| The document uses `@container` rules | A container query needs the used size of a container, and the kept state does not carry it | No today, but one such rule refuses every patch in the document |
| The element moved or changed size | Every box after it would be stale | Yes. A card is a grid item, so its width is the track width, while the splice laid the element out in the grid container's content width |
| The element owns no operations | There is no range of operations to replace | No. A card owns 204 operations in the fixture display |

The first refusal on the real fixture was the out-of-flow box, with the reason
the engine reported:

```
layout: relayout element: layout: element subtree cannot be spliced: the subtree holds an out-of-flow box
```

The guard now allows an out-of-flow box whose containing block is inside the
subtree, because the subtree lays that block out again and places the box
against it. The card row is spliced with that change.

The host checks the page before it does anything: the page, the cache, and the
display have to exist, the id has to be non-empty, and the retention has to be
on. Then an edit has to change something, the re-cascade has to succeed, the
engine has to return a patch, and the patch has to carry a display. The host
takes the engine's whole patched display instead of copying the patch
operations, because a patch that changes the operation count moves every later
operation index, so the three checks Blinkless documents for a host that copies
the operations do not apply here.

The retained layout state is opt-in on both sides. `layout.Options.RetainState`
keeps the resolved styles and the box tree a patch needs, and
`Config.RetainState` asks the page for it. A page that never patches an
element pays nothing for it, and `Page.PatchElement` refuses before it edits
the tree when the page kept no state.

The retained state is not free. One laid out page keeps 201 KB alive with the
retention on and 34 KB with it off for 31 elements, 579 KB and 302 KB for 283
elements, and 1198 KB and 744 KB for 703. That is 167 KB, 277 KB, and 454 KB of
retained state. The cost is per layout, so a page that redraws often leaves
that much behind each time. The table is in `documentation/library-api.md` in
Blinkless, and the measurement is `BenchmarkDisplayRetainedBytes`. The style
store keeps the distinct styles of a page in backing arrays of eight records
that grow to sixty four, so a page with a few distinct styles keeps a small
array instead of a full one.

### The fallback

A refused patch leaves the cached tree edited and the retained display
untouched, because every engine refusal happens before the splice. The app
falls back to the path the click already used.

- Keep the state the click set: the started map, the selected name, the
  readout, and the start time. The click means the same thing either way.
- Redraw the whole page. The template prints the base class and the literal
  `Ready`, so the started class and the state text are not template data and
  the executed source does not change for them. The redraw lays the cached
  tree out again, and the tree still carries the edit the refused patch left
  in it, so the display shows the class the engine refused to splice. The
  readout is template data, so when it changes the executed source changes,
  the page drops the cache, and the tree is parsed again, which drops the
  stale tree edit with the old tree. The after-redraw hook puts the state back
  into the new tree. The redraw has to be one that lays the page out, not one
  that can be skipped as unchanged.
- Keep the sample's captured starting paint and its origin. The refused patch
  did not touch the display, so the capture is still the paint the transition
  starts from. A second capture would take the started paint as the starting
  value and the transition would not move.
- Let the tick re-bind. A full rebuild replaces the operations, so every
  sample's plan is stale. The rebuild bumps the page generation, and the tick
  re-binds the samples against the new display on its next frame.
- Fall back for the whole page when the click changes a second card. The
  selected class moves from the old row to the new one, and a one card patch
  cannot move a class between two elements. The readout in the metadata
  section is a second element, and its height moves with its text.

The user sees no difference. The card ends up started and selected, the readout
is the same, and the animation plays from the same phase. The difference is
the cost. The fallback pays the page rebuild of the fourth pass, 6.16 ms on
average with the catalog windowed, so the press to painted time stays at the
fourth pass number for the clicks that fall back.

The timings for the patch path itself are in the next section.

### The patch, measured

The harness is in `examples/css-preview`, and the numbers are in
[PATCH_COST.md](../examples/css-preview/PATCH_COST.md). It uses the whole
43 row catalog at 1908x999, the desktop window size, with the retained state
on.

| Path | Time | Allocated | Allocations | Patches applied |
|---|---:|---:|---:|---:|
| Full redraw, card class | 34.5 ms | 20.18 MB | 32188 | |
| Element patch, card class | 6.08 ms | 6.31 MB | 6875 | 30 of 30 |
| Full redraw, state class | 39.2 ms | 19.86 MB | 31656 | |
| Element patch, state class | 4.72 ms | 5.99 MB | 6347 | 30 of 30 |
| Full redraw, same source | 27.2 ms | 13.87 MB | 25304 | |

The redraw splits into three stages, measured by the page itself: the template
takes 0.58 ms, the cascade 0.09 ms, and the layout 27.7 ms. Laying out 43
cards, 9178 operations and every text run is the whole cost of a redraw. A
patch of one state span lays out that span and leaves the other 42 cards
alone, so it is about eight times faster than the full path.

The cascade is not saved, because the patch re-cascades the whole document, but
that stage costs 0.09 ms, so there is nothing to win there.

A patch still allocates 5.99 MB on the full catalog for the state span, and
6.31 MB for the card row. The host adopts the whole patched display the engine
hands back, so the engine rebuilds the operation list and the owners for the
whole page, and the page holds 9178 operations. The layout is skipped, the
copy is not. The windowed catalog holds about a fifth of the operations, so
the copy is about a fifth as big there.

The harness prints what the engine says for each element it probes. The probe
puts a class that carries no rules on the element, so a card loses its padding
and its border and changes size, which the engine refuses:

| Element | Engine outcome |
|---|---|
| `catalog-state-01` | spliced, 1 operation |
| `start-property-webkit_animation` | spliced, 2 operations |
| `preview-metadata` | spliced, 1 operation |
| `catalog-row-01` | refused: the element changed size, 128.84 to 112.34 high |
| `catalog` | refused: the element changed size |
| `catalog-grid` | refused: the element changed size |
| `catalog-target-webkit_animation` | refused: the element changed size |
| `play-all` | refused: the element changed size, 24.75 to 30 high |
| `metadata` | refused: the element changed size, 400.05 to 408.93 high |

With the class the example uses, the card row is spliced, 203 operations, and
the state span inside it rides along with that one patch.

## Next: the two gaps against a browser

The first gap is closed. The second one is the list to work from next.

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

This gap is closed. A click in the CSS preview no longer re-lays out the
cards. The click state lives in the parsed tree, the click patches the
elements it changed, and the page paints the patched region.

- OwnFrame copies the changed card's operations into the retained display
  and marks its region dirty. `Page.PatchElement` does that for one element
  of the retained page, and it removes the rebuild for the selection, the
  state text, and the started class, and it helps every click, not only the
  catalog.
- Blinkless re-lays out one subtree and keeps the rest. `layout.RelayoutElement`
  re-resolves the styles of the subtree, places it, and splices its operations
  over the range the element owned, so the rest of the page keeps the
  operations and the boxes the host already had. The guards and the fallback
  are in the section above.

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
- Each row keeps its own start time, and the run does not stamp a shared one.
  The origins come from the drain clock, so a sample starts when the run
  starts rather than when the click that queued it was handled. A click that
  patches every card can take longer than the shortest animation, so a shared
  clock stamped before that work would leave every sample finished before the
  first tick reached it.
- A transition needs its starting paint captured while the row is still
  unstarted. A row that scrolls in for the first time renders once without the
  started class, captures that paint, then renders with the class and binds.
  Two renders for that card only, and only for the cards the user visits.
- The edge cases are a row that first appears after its animation would have
  finished, a row that is scrolled past, and a reset.

Measured at 1908x999, the press costs 6.7 ms with the catalog windowed, and
the display holds 1660 operations and 119 boxes. Before, the press rendered
all 43 rows, 9178 operations and 537 boxes, and cost 46 to 50 ms in total.


## Visible latency

The redraw time above is the page work. From the real click script, with
`-perf` on and the walk hook active:

| Segment | p50 | p95 | max | average | 43-click pass |
|---|---:|---:|---:|---:|---:|
| press to started | 15.8 ms | 30.9 ms | 32.2 ms | 16.5 ms | 0.68 s |
| press to painted | 51.0 ms | 51.6 ms | 52.3 ms | 51.0 ms | 2.19 s |

The 43-click pass column is the p50 times 43 clicks. It is the number that
replaces the old total rebuild row: a walk of the whole catalog now spends
about 2.2 s waiting for frames, against 28.7 s of page rebuild before any of
this work.

These are the numbers of the fifth pass, from the real click script with
`-perf` on and the walk hook active. The third-pass numbers were 9.9 ms and
50.6 ms on p50. `press to started` uses the handler timestamp in the app, so
polling latency does not count. `press to painted` waits for the next layout
counter, which the patch raises, and includes the window draw. It is
dominated by the presentation frame of the software compositor, so it did not
move when the page stopped rebuilding itself.

A separate run under a virtual X server, three passes of 43 clicks, gave
17.2 ms and 51.2 ms on p50, and 2.37 s per 43-click pass. The same run
reported a redraw of 0.0 ms on every click, which is the no-rebuild column
above measured directly.

## How the numbers were measured

Each column of the results table came from a different harness:

| Column | Harness | Size |
|---|---|---|
| Original | in-process pointer cycle, no OS window | 960x1000 page |
| First pass | in-process pointer cycle, no OS window | 960x1000 page |
| Second pass | real X11 click walk | 960x860 window |
| Third pass | real X11 click walk | 960x860 window |
| Fourth pass | `BenchmarkWindowedPointerCycle`, in process | 960x860 page |
| Now | real X11 click script | 1908x999 window |

The fourth-pass column comes from `BenchmarkWindowedPointerCycle`, which follows
the same 43-click desktop path as the X11 script but reads `LastRedraw` in
process. That removes the software X server and presentation frame from the
measurement. The benchmark runs one warm pass, then reports the average, p50,
p95, and max over one or more measured passes. The numbers above are the mean
of five `-benchtime=1x` runs. The third-pass and earlier columns came from the
real X11 walk and the earlier pointer cycle measurement.

The in-process benchmark barely moves with the window size: it reads 5.17 ms
p50 at 960x860 and about 6.1 ms p50 at 1908x999. The 25.2 ms p50 that the
real X11 walk reports for the same fourth pass code at 1908x999 is a different
measurement, because the walk drives the real window and reads the page's own
redraw. The walk hook closes the devtools dock, so the viewport is the full
window width.

The fifth-pass column comes from the real X11 walk, because the benchmark
drives the page without the example's own handlers and therefore still
measures a full rebuild. The walk script reports the last redraw the page
made, and it reports zero when the click painted the frame without a redraw,
so the number is the page rebuild the click caused.

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

The patch cost numbers come from the Blinkless commit `c1054ff`, which is the
pinned engine, so the wall clock is not comparable with the earlier recorded
run and the allocation counts are the stable part.

## Reproduce

Run the real click walk from the repository root:

```sh
CSS_PREVIEW_WALK=1 go run ./examples/css-preview -perf -width 1908 -height 999
python3 scripts/css-preview-click-test.py --runs 3
```

The example defaults to 960x860, so a run without the flags measures the
smaller window.

The script resets the started rows at the start and stops the previous sample
before each click, so it measures one click at a time instead of a soak of 43
animations. `/state` reports the started timestamp, the last redraw, and the
pipeline counters. `/reset` clears the rows, and `/goto` scrolls the next Start
button under a fixed row. It closes the app with `/quit` after the last pass;
use `--no-close` to keep the window open.

Run the windowed benchmark that matches the fourth-pass column:

```sh
go test ./examples/css-preview -run '^$' \
  -bench '^BenchmarkWindowedPointerCycle$' \
  -benchtime=1x -count=5
```

Run the element patch harness and its parity gate:

```sh
cd examples
go test -p 2 -run 'TestPatch' -v ./css-preview/
go test -p 2 -run XXX -bench BenchmarkPatchCost -benchtime 30x -count 5 ./css-preview/
```

The report of that harness is
[PATCH_COST.md](../examples/css-preview/PATCH_COST.md). The run and profile
notes for the example are in
[PROFILING.md](../examples/css-preview/PROFILING.md).

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
- `internal/page/page_after_redraw.go` defers and flushes the draw.
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
- `internal/page/page_draw.go` sizes the template builder from the
  previous source.
- `examples/css-preview/batch_click_test.go` covers the batched press through
  a real animation start.
- `internal/page/page_patch_element.go` patches one element of the retained
  display.
- `internal/render/cache_edit.go` edits the cached tree and re-cascades it
  without replacing the cached styled document.
- `internal/page/page_after_redraw.go` runs a callback after a committed
  redraw, so an app that keeps its state in the tree can reapply it.
- `internal/page/page_click_paint.go` runs the paint-only click handler and
  clears the redraw batch when the handler painted the frame.
- `examples/css-preview/catalog_state.go`, `catalog_click_paint.go`, and
  `catalog_sync.go` keep the click state in the parsed tree and put it back
  after a rebuild.
- `examples/css-preview/patch_cost_*.go` and
  [PATCH_COST.md](../examples/css-preview/PATCH_COST.md) measure an element
  patch against a full redraw and gate the patch against a full redraw
  operation by operation.

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
`v0.0.3-0.20261021155643-c1054ff46202`, which carries both one-pass grid
paths and the element patch work. The 12.5 ms and 6.16 ms columns use the
earlier pin, `e0c3c9c`, and the patch cost numbers use `c1054ff`. The main
module has no replace, so the build and the module zip carry the same engine.
The examples module still pins Blinkless `v0.0.2`, and the patch cost harness
calls `layout.RelayoutElement`, which only the newer pin carries, so the
examples module builds through the committed `go.work` until the release
process bumps its pin.
