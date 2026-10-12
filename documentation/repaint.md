# Repaint

A click that changes one box used to rebuild and replay the whole page. Now
the page reports the boxes it changed and the window repaints only those boxes.
This is the automatic path. The manual path, a frame callback that edits an
operation, is in [frames.md](frames.md).

## The page reports the regions

`Page.TakeDirtyRects` returns the regions that changed since the previous
call, in CSS pixels, and clears them. A handler can declare the id it changed
with `Invalidate`; for handlers that forget, the page diffs the previous
display list against the new one and unions the bounds of every changed
operation. An operation removed from the list dirties its old bounds, so a
vanishing element leaves no ghost. A change the page cannot place, such as a
theme switch or a `Load`, returns the whole frame, and so does a diff that
changes more than eight operations or covers more than a third of the frame.

The list holds one rect per separated change. A new rect merges into an
existing one when the two come within eight pixels of each other, so a shadow
that touches its neighbour repaints once. Past eight rects the page collapses
the whole list to one bounding box. When the page hands the list over, each
rect is padded, clamped to the content, and coalesced again. If the surviving
rects cover more than a third of the content, the call returns one bounding
box instead. `TakeDirty` keeps its old signature and returns that bounding
box, so a caller that wants one rect still gets one.

```go
rects, ok := page.TakeDirtyRects()
```

`ok` false means nothing changed since the previous call. Each rect is
padded, so shadows and outlines repaint too.

## The window repaints the rects

`internal/replay.DrawRect` is `Draw` for one region. It walks `Display.Order`
the same way, skips any op whose painted box does not meet the rect, and
paints the ops it keeps into a sub-image of the destination. The sub-image
matters: the display list carries a page background fill that covers the
whole canvas and therefore meets every rect. Without the clip, that one op
would repaint the background over everything outside the dirty rect, and the
rest of the page would vanish until its own box is dirtied again.

```go
replay.DrawRect(dst, display, rect, dx, dy)
```

Text needs care. An op carries its baseline in `Y`, so `DrawRect` tests the
line box from the font ascent above the baseline to `InkDescent` below it,
not `Y` alone. Strokes and lines grow by half a stroke, and an op whose ink
cannot be bounded, such as a rotated run, is always drawn.

For content within the replay cache limits, the window keeps one
content-sized image for the replay path. The content is the canvas grown to
cover every box that overflows it, the same measure the scroll clamp uses, so
a page with a fixed root stays painted when it scrolls.
Each dirty rect is filled with the page background and replayed into that
buffer, and the visible page is blitted from the buffer to the screen each
frame. The screen is cleared between frames, so the blit covers the viewport
and only the buffer repaint is partial. The `bitmap fallback` badge draws on
the screen after the blit, so a repaint never wipes it.

Each rect in the list gets one partial repaint against that buffer. A list
longer than the eight-rect cap, or a rect equal to the whole content, falls
back to one full buffer rebuild. When the list is empty and the generation
has not changed, the window replays nothing and blits the buffer it already
has. A screen that only implements `TakeDirty` is read as a one-element
list, so its single rect still takes the partial path.

The buffer is rebuilt in full when the generation changes without a usable
rect, when the content changes size, and when a rect is the whole content. A
scroll offset change only moves the blit, so scrolling does not replay the
list.

## Oversized pages

When content exceeds the cache limits, a static page with `TakeDirty` keeps one viewport image with at most 256 pixels of vertical overscan on each side. Scrolling inside that area reuses the image. Crossing its edge moves the area and repaints visible operations. Dirty changes follow the list: a single rect repaints that rect, several repaint their bounding box, and the work covers only the intersection with the cached area. A generation change without a dirty region rebuilds it. Resizing disposes the old image. The viewport path disposes the content buffers, so they are not retained together.

`DrawVisible` filters operations before replay while preserving paint order and retaining operations whose ink cannot be bounded. Ticking pages replay visible operations each frame. This still scans the operation list; row windowing reduces the list itself.

## What still repaints in full

- A fallback page with no display list. `imageout.RenderLayout` paints the
  whole canvas, and a partial bitmap path needs a new engine call; the
  decision is recorded in
  [the plan](../plans/v0.0.2/incremental-repaint.md). The badge stays
  visible.
- A page with a frame callback that did not call `Page.UseFrameDirty`.
  A tick changes operations in place and leaves no dirty rect, so
  `Page.Ticking` keeps the full replay. `UseFrameDirty` opts that tick
  into the partial buffer. `Invalidate` and `MarkRect` name the regions.
- A stretched or zoomed page. `internal/window/replay_scale.go` replays the
  whole list into a canvas-sized buffer each frame and scales it to the
  window.
- `Page.PNG` paints from the template source on demand, so `GET /frame.png`
  always costs a full paint.

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
retarget path: `Page.Generation()` does not move, `RedrawNS` stays the initial
draw, and no frame after the first is full-content. The full tables and the
command lines are in `examples/departure-board/PERFORMANCE.md`.

## Proving it with counters

The DIRTY section of the Frame tab reports dirty rects taken, full frame fallbacks, operations repainted, and operations skipped, and `GET /debug/state` serves the same data in web mode. The full checklist, budgets, and soak and baseline scripts live in [performance.md](performance.md).
