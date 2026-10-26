# Repaint

A click that changes one box used to rebuild and replay the whole page. Now
the page reports the boxes it changed and the window repaints only those boxes.
This is the automatic path. The manual path, a frame callback that edits an
operation, is in [frames.md](frames.md).

## The page reports the regions

`page.Probe().TakeDirtyRects` returns the regions that changed since the previous
call, in CSS pixels, and clears them. A handler can declare the id it changed
with `page.Probe().Invalidate`; for handlers that forget, the page diffs the previous
display list against the new one and unions the bounds of every changed
operation. An operation removed from the list dirties its old bounds, so a
vanishing element leaves no ghost. A change the page cannot place, such as a
theme switch or a `Load`, returns the whole frame, and so does a diff whose
running bounding box covers more than a third of the frame. A generation with
many small changes inside one region stays partial: the changed boxes merge
within the eight-pixel gap, and when the list would pass the rect cap the
closest pair unions instead of the whole list collapsing.

A full-frame mark that is still pending when another display list lands is
diffed against the display the host last painted, not the display the page
committed last. When those changes stay inside the partial rules, the pending
full frame narrows to them, so a small hover patch that follows an unpainted
full commit keeps its own region. Without a painted display, or when the full
mark came from an in-place tick change the display diff cannot see, the frame
stays full.

The list holds one rect per separated change. A new rect merges into an
existing one when the two come within eight pixels of each other, so a shadow
that touches its neighbour repaints once. Past 24 rects the page collapses
the whole list to one bounding box. When the page hands the list over, each
rect is padded, clamped to the content, and coalesced again. If the surviving
rects cover more than a third of the content, the call returns one bounding
box instead. `page.Probe().TakeDirty` keeps its old signature and returns that bounding
box, so a caller that wants one rect still gets one.

```go
rects, ok := page.Probe().TakeDirtyRects()
```

`ok` false means nothing changed since the previous call. Each rect is
padded, so shadows and outlines repaint too.

## The window repaints the rects

The live window path is `drawReplayPartial` in `internal/window`. A
replayable page that can report dirty rects takes one of two retained buffers.
A page no more than three viewports tall keeps a 256 by 256 tile grid sized to
the content (`internal/window/replay_tile_paint.go` calls
`replay.DrawVisibleUnfixedZ`). Each dirty rect is filled with the page
background and replayed into every tile it touches, and only the damaged
regions are copied on a partial frame. The window retains the previous screen
between frames, so a clean frame needs no page copy. A stale screen or scroll
change copies the visible tiles again. The `bitmap fallback` badge draws on
the screen after the tiles, so a repaint never wipes it.

A taller page keeps a viewport image with 256 pixels of vertical overscan on
each side. Scrolling inside the overscan reuses that image. Crossing its edge
copies the overlapping pixels into a second image of the same size and replays
only the newly exposed strips. The two images alternate on later moves. A jump
with no overlap rebuilds the viewport. Dirty rects still repaint individually,
including changes inside the copied overlap. These render targets stay outside
Ebitengine's automatic sprite atlas, so switching between painting and copying
them does not migrate a large image through that atlas.

Warmup submits one band, at most 128 pixels tall and one viewport wide, per
frame into a reusable scratch image. A small page eventually warms in full;
a document beyond the 12-megapixel budget warms a neighbourhood around the
viewport and starts another sweep when scrolling leaves it. Warmup uses the
same background replay as the visible viewport. Pending bands keep the idle
pacer awake until the sweep finishes. This avoids submitting the whole page's
cold graphics work in its first frame.

Both buffers keep the page background only. A visible element with an active
2D transform or opacity animation can be promoted to a retained layer
(`internal/replay/layer_cache.go`), as long as no animation on that element
writes colour, layout, or SVG paint. Those writes change the whole subtree the
texture owns, so such an element stays direct and never re-bakes per frame. The
background replay skips the operations the layer owns, so their pixels are a
hole, and the window composites each layer after the background blit
(`replay.DrawVisibleUnfixedZBackground`, `replay.DrawLayers`). A layer that
only moves needs no background work. A promotion or demotion reports the pose
it leaves behind (`LayerTopologyDamage`). The viewport cache rebuilds its
background on a membership change because overscan can hold an older pose from
before the element was promoted. Scrolling also refreshes strips entering the
screen, even when the buffer's overscan already covers them: direct animations
keep sampling while offscreen without reporting repaint damage. A layer already held through a rest
pose keeps its texture, so an alternate animation does not demote and
re-promote at every turn. A layer whose content still changes every frame is
demoted after two frames and re-promoted after a cooldown. `LayerGeneration`
keys the cache to display content, not to a re-projection pass, so a 3D
animation does not re-bake every frame.

`internal/replay.DrawRect` is the same clip idea for tests and helpers. The
window does not call it. It walks `Display.Order`, skips any op whose painted
box does not meet the rect, and paints into a sub-image of the destination.

```go
replay.DrawRect(dst, display, rect, dx, dy)
```

The window does not call `DrawRect`. `drawReplayPartial` in
`internal/window/replay_partial.go` drives the tile grid; each damaged tile
calls `replay.DrawVisibleUnfixedZ`. `DrawRect` is the test and helper path.

Text needs care. An op carries its baseline in `Y`, so a rect test uses the
line box from the font ascent above the baseline to `InkDescent` below it,
not `Y` alone. Strokes and lines grow by half a stroke, and an op whose ink
cannot be bounded, such as a rotated run, is always drawn.

The content is the canvas grown to cover every box that overflows it, the
same measure the scroll clamp uses, so a page with a fixed root stays painted
when it scrolls.

Pinned chrome stays on the tile path. `Page.Scroll().SetViewportPinZ` and
`Page.SetScrollWindow` keep bars at the viewport ([screen.md](screen.md)).
Tiles hold only unfixed operations (`replay.DrawVisibleUnfixedZ`). After the
dirty-tile blit, `drawReplayPartial` paints fixed and pin-z operations in
screen space (`replay.DrawFixed`, `replay.DrawPinnedZ`), clipped to the
strip the devtools dock leaves visible. Telegram uses that overlay. A
scroll copies message tiles and then paints the header and composer, so
`hasViewportOps` does not force a full replay. The measured table below is
the forms family; a pinned page uses the same 256 px grid.

Each rect in the list gets one partial repaint against that buffer. A list
longer than the 24-rect cap, or a rect equal to the whole content, falls
back to one full buffer rebuild. When the list is empty and the generation
has not changed, the window replays nothing and blits the buffer it already
has. A screen that only implements `TakeDirty` is read as a one-element
list, so its single rect still takes the partial path.

The buffer is rebuilt in full when the generation changes without a usable
rect, when the content changes size, and when a rect is the whole content. A
scroll offset change only moves the blit, so scrolling does not replay the
list.

While the devtools dock is open, the window clips the tile repaint and the
tile blit to the strip the panel leaves visible, so hidden tiles are neither
painted nor copied. When the dock closes, the strip that was hidden is
repainted once, because tiles under the panel may be stale.

## Repeated shapes come from the raster cache

Ebiten tessellates an antialiased vector path into several stencil passes, so
a page that paints many copies of one shape pays for every copy on every full
replay. The departure board paints 273 rounded fills and 270 rounded borders,
and one full replay grew the graphics command queue's vertex buffer to about
87 MB, which the process then held for its lifetime.

`internal/replay` bakes a vector shape once into an image and blits the baked
pixels for every repeat. The raster is a white mask, and the blit applies the
operation's colour through `ColorScale`, so one image serves every colour of
the same shape and a colour animation does not re-rasterize its path per
frame. A `rasterKey` names the shape exactly: the variant (fill or stroke),
the size, the corner radii, the stroke width and mask, and the sub-pixel
offset of the op's origin. The integer part of the position is not in the
key: the raster bakes with that part moved to the image's origin and blits
back at it, so one image serves every copy of a shape in every integer
position. Source-over compositing is associative, so compositing the baked
raster is pixel-identical to drawing the path there. A gated GPU test
compares the two for fills, borders, transparent colors, and two positions,
and a second compares a `DrawRect` sub-image clip against the direct draw.

The cache holds at most 256 rasters and 4 megapixels; past either bound the
oldest entry is disposed. A shape larger than 512 px a side, or an unknown
stroke mask bit, falls through to the direct path. The cache is
process-wide and keyed per shape, like the decoded-image cache.

Text runs take the same route: a run bakes once into a white mask and blits
with the op color, so one raster serves every color of the same run while its
face, content, size, fake-bold flag, letter spacing, and sub-pixel origin
match, and a damaged rect over static text stops reshaping it. The run cache
is bounded at 512 entries and 4 megapixels with least-recently-used eviction,
and an oversized run draws directly.

## Element paint chunks

A dirty rect used to replay every operation it met, changed or not. The
painter now groups consecutive operations that sit inside the same element
box from `Display.Boxes` into a chunk, rasterizes the chunk once into a
bounded image, and blits it on later damage. A content signature covers every
field that affects pixels, including an operation's baked matrix, so a tick
that patches an operation re-bakes only its chunk and a transform that moves
each frame re-bakes only the chunk it belongs to. Chunks paint in order, which
keeps overlaps correct, and stay direct for fixed or z-indexed operations,
translucent runs, ink that cannot be bounded, boxes over half the content, and
everything inside a non-normal or masked group. The cache holds a bounded number
of chunks and pixels, and an evicted chunk re-bakes on the next draw.

## Transforms and blend groups

An op with a finite 2D transform replays through its matrix: fills and strokes
blit a cached local raster through the op matrix, lines rasterize locally, and
text draws with the matrix on the text options. A blend or isolation group
buffers its members into a layer and composites the layer once with the CSS
blend formula in a Kage shader, so a partially transparent backdrop stays
exact. Nested groups buffer inside the parent, up to four deep.

Automatic CSS motion makes the matrix path run per frame. A 2D transform sample
writes the operation's matrix, and the dirty-bounds helper
(`internal/page/dirty`, `OpBounds`) maps the operation's local paint box through
that matrix, so a damage rect covers where the ink lands after the move rather
than the laid-out box. A 3D transform or perspective sample does not write a
matrix: the page hands the values to the engine's re-projection pass and marks
the patch's old and new painted boxes (`internal/page/motion_wire.go`,
`adoptRestamp`). A matrix the replay cannot bound still falls to the bitmap, as
the next paragraph lists.

The gate still sends several cases to the bitmap: non-separable blend modes
(hue, saturation, color, luminosity), transformed grids, rotated or oblique
glyphs, groups whose members would split across the window's fixed or pinned
layers, and matrices that cannot be bounded.

## Oversized pages

When content exceeds the cache limits, a static page with `TakeDirty` keeps
viewport buffers with at most 256 pixels of vertical overscan on each side.
Scrolling inside that area reuses the current image. Crossing its edge copies
the overlap and paints the newly exposed strips; a jump with no overlap
rebuilds the image. Dirty rects repaint only their intersections with the cached
area. A generation change without dirty regions rebuilds it. Resizing disposes
the old buffers. The viewport path drops the content-sized tile grid.

A stationary frame can park on the desktop idle clock after warmup finishes.
A scroll, generation change, dirty region, focus or size change, or frame
callback keeps the frame clock running.

`DrawVisible` filters operations before replay while preserving paint order and retaining operations whose ink cannot be bounded. Ticking pages replay visible operations each frame. This still scans the operation list; row windowing reduces the list itself.

## What still repaints in full

- A fallback page with no display list. `imageout.RenderLayout` paints the
  whole canvas, and a partial bitmap path needs a new engine call; the
  decision is recorded in
  [the plan](../plans/v0.0.2/incremental-repaint.md). The badge stays
  visible.
- A ticker that does not implement `FrameDirty` (not a `Page`). A page
  tick that patches ops is marked dirty by the inner screen's `Tick` and stays on the
  tile path. A no-op tick blits the buffer. `page.Probe().UseFrameDirty` is an alias.
  The window calls the inner screen.
- A stretched or zoomed page. `internal/window/replay_scale.go` replays the
  whole list into a canvas-sized buffer each frame and scales it to the
  window.
- `page.Probe().PNG()` paints the retained display on the first request after a redraw and caches the encoded picture. It does not use window tiles. Paint damage, including tick-only edits, invalidates the encoded picture; see [screen.md](screen.md#png-and-the-web-page).

## Measured numbers

Against the examples in this tree, at 480x640, at this commit. The rect is
the element's box plus 8 px, and the count is the operations that meet it:

| Page | Ops | Repainted | Skipped |
|------|-----|-----------|---------|
| platform, `#count` | 11 | 5 | 6 |
| login, `#message` | 18 | 3 | 15 |
| states, `#hover-btn` | 16 | 4 | 12 |
| forms, `#remember` | 45 | 5 | 40 |

A full-canvas background meets every rect, so the repainted count is not the
whole story. The automatic diff rect is the changed operations' own bounds,
which is often tighter than the element box, and the ops it skips are the
small text and border runs a partial repaint exists to avoid. The op filter
runs at about 12 ns per op (`go test -bench BenchmarkDrawRectScan -v
./internal/replay`, which prints the platform counts). One `#inc` click
through the handler, the template execute, the relayout, and the dirty rect
measures about 0.5 ms on this machine (`BenchmarkClickCount` in the platform
example).

### The damage list on the departure board

The departure-board dump measures the real cadence (`-dump -pace 16ms`,
2026-10-12). The fixed tile grid puts the board at 844 operations, up from
631 and 640 on the old grid.

| scenario | ops | worst steady frame: replayed ops | rects on that frame | frames with 2+ rects |
|----------|----:|--------------------------------:|--------------------:|---------------------:|
| cascade | 844 | 115 | 3 | 13 of 150 |
| rotate | 844 | 115 | 3 | 13 of 150 |
| idle | 844 | 105 | 2 | 1 of 240 |

The worst steady frame skips the first frame after a draw, which repaints the
whole content once. On the same board the single union rect replayed 630 of
631 operations on the worst cascade frame, so the list removes about 82
percent of the worst-frame operation count. The `rotate` row takes the tick's
retarget path: `page.Probe().Generation()` does not move, `RedrawNS` stays the initial
draw, and no frame after the first is full-content. The full tables and the
command lines are in `examples/departure-board/PERFORMANCE.md`.

## Proving it with counters

The PIPELINE DETAIL group of the Frame tab reports dirty ops, dirty regions, and changed ops with Perf on, and `GET /debug/state` serves the same data in web mode. The full checklist, budgets, and soak and baseline scripts live in [performance.md](performance.md).

## Validate repaint behavior

The software-picture comparison tests do not capture window tiles. Run `make graphics-check` for retained-window pixel checks, damage blits, scrolling, dock restoration, and replay parity. See [testing.md](testing.md) for prerequisites and platform limits.
