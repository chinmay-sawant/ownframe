# Frames

`Page.SetTick` registers one function the window calls before it draws each
frame; `SetTick(nil)` removes it. The callback can move a bar, wave a level
meter, or call `Redraw`, so a page can animate without parsing the HTML again. The desktop window, the phone
build, and the WebAssembly canvas call it at Ebiten's default tick rate, about
sixty times a second. A desktop window that has drawn five clean frames in a
row and has seen no input for 250 ms stops presenting every frame: it waits in
the event loop and wakes on input, or at the next deadline the page declared.
`Page.WantFrameAt(t)` and `Page.WantFrameIn(d)` set that deadline, and
the window consumes and clears it through its inner screen. A page with no request falls back to
a keep-alive frame every 125 ms, so a tick that never asks still runs. A tick
that animates must leave damage. The inner screen's `Tick` diffs patched ops and sets the
frame-dirty flag itself. The window calls that tick on the inner screen. `page.Probe().MarkRect` and `page.Probe().Invalidate` still name extra regions.
A page that damages every frame is never paced.
`WindowOptions.NoIdlePace` turns the pacing off, and the phone and WebAssembly
builds keep the plain sixty tick loop. The Frame tab's PACING section shows
whether the window is idle, the next deadline in milliseconds, and how many
presents were skipped. Each frame runs
devtools sync, keys, resize, the passthrough update, pointer, drop, tick, a
reload poll, sync, the page scroll request, a devtools refresh, and wheel, in
that order. An error from the
callback stops the window, and `Run` returns it. `Serve` does not tick.

```go
page.SetTick(func(ctx context.Context) error {
    if dur <= 0 {
        return nil
    }
    page.Probe().PatchOp(opIndex, func(op *ownframe.DisplayOp) {
        op.W = barWidth * float64(pos) / float64(dur)
    })
    return nil
})
```

Here `opIndex` identifies the bar operation in the current display. Find it again after a redraw or resize. `PatchOp` and `Display` are on `page.Probe()`. `Boxes` stays on `page`. These helpers are for engine examples. Application animations can update data and call `Redraw` instead.

A page the replay accepts keeps its placement as `*ownframe.Display`.
`page.Probe().Display` returns that list and `Page.Boxes` returns the hit-test boxes
from the same placement. Operations carry points; boxes carry CSS pixels, so
multiply a box by `Display.PointsPerPixel` before comparing the two. The paint
fields of `ownframe.DisplayOp` (`X`, `Y`, `W`, `H`, `R`, `G`, `B`, `Alpha`, and
`Text`) can change between frames. Changing one changes the next drawn frame;
it does not parse, cascade, or lay out anything again. `Alpha` only takes
effect between 0 and 1; 0 means unset. To hide a text run, empty `Text`; to
hide a fill, collapse `W` and `H`.

A `Redraw` replaces the display list, so a callback that keeps an operation
pointer must find the operation again after any redraw. A window resize
relayouts the page and replaces the display list the same way. `page.Click` redraws
after its handler. `page.Edit().KeyUp` never draws, and `page.Edit().Copy` does not draw. `page.Edit().KeyDown`
draws only when Escape clears a focus or a caret key moves. The window calls those on the inner screen. The other input
handlers redraw after they run. The examples use `internal/frame`
to find operations: `frame.Fill` returns the first fill of a colour inside a
box, `frame.Fills` returns them left to right, `frame.Text` returns the first
text run, and `frame.BoxUnits` converts a hit-test box to display-list units. A
text operation carries its baseline in `Y`, so its box test differs from a
fill's centre test.

Anims built by `frame.TranslateTo`, `frame.FadeTo`, `frame.ScaleTo`, or
`frame.RotateTo` hold pointers into one display list. Call `Bind(page)` on the
anim so `Step` goes inactive once a `Redraw` replaces the list. An unbound anim
keeps the legacy behavior and must not outlive a redraw. No production code in
this tree constructs anims today, only tests, so every new call site starts
bound.

There are two ways to change the frame without a full `Redraw`. A tick
callback changes an operation in place, which suits an animation that runs
every frame. A click, key, or hover handler changes page content, and the
page reports the changed boxes with `page.Probe().TakeDirtyRects` so the window repaints
only those boxes; [repaint.md](repaint.md) has the call shapes. Reach for the
tick when the change repeats every frame, and for the content path when it is
one edit. A tick that patches operations is marked dirty by the library. The inner screen's `Tick` diffs
the display list, queues the changed boxes, and sets the frame-dirty flag, so
callers need not remember `page.Probe().UseFrameDirty`. The window then repaints
`page.Probe().TakeDirtyRects` and blits the buffer when that list is empty. A no-op tick
leaves the page clean, so idle park can sleep. `page.Probe().UseFrameDirty` remains a
working alias. `page.Probe().Invalidate` marks the elements that tick changed. `page.Probe().MarkRect`
marks a CSS-pixel line that has no element id. The window calls the inner screen.

A rotation can retarget operations in the tick instead of calling `Redraw`.
The [departure board](../examples/departure-board) does: when the feed is due,
the tick takes the next rows, updates each cell's target glyph by id, and
starts the row bands with the same math the rebind uses after a redraw.
Nothing is laid out. The same call stores the new rows with `SetData`, so the
next real `Redraw`, for a resize, theme, or night sheet, rebuilds the same
picture from the same data. Each flap keeps the box it paints in, computed
from its `ownframe.Box` at rebind, and the tick calls `page.Probe().MarkRect` on the boxes
whose glyph changed, so the window repaints one row band instead of the whole
board. `Advance`, the `SetData` plus `Redraw` pair, stays for tests and for
the dump.

## Cost

Changing an operation is cheap and can run every frame. A `Redraw` is not: the
engine lays the page out again, and parses and cascades when the executed
source changed. Redraw cost depends on the template and engine version. The
[Spotify measurement report](../examples/spotify-player/PERFORMANCE.md)
records a specific comparison after SVG artwork was replaced with PNGs.
Remeasure with the current benchmark before choosing an animation path.

## Limits

- The window replays the display list, so operation changes show in the
  desktop window, the phone build, and the WebAssembly canvas. `Serve` never
  ticks. `page.Probe().PNG()` paints the retained display when requested;
  operation patches and mutating ticks invalidate an already encoded picture.
  A no-op tick keeps the PNG cache. This software path
  does not capture the window framebuffer.
- A page that fell back to the bitmap path has no display list. `page.Probe().Display`
  returns nil and there is nothing to change.
- CSS `@keyframes` and `transition` use separate CSS-specific runtime packages.
  A caller binds Blinkless records to paint operations and supplies the clock,
  timeline values, trigger events, or before and after displays. The Page
  window does not start them automatically. See [CSS motion](css-motion.md).
  The `Page.SetTick` helpers above still suit Go-controlled motion.

The [Spotify player](../examples/spotify-player) example uses the tick for a
moving seek bar, a running clock, and a sine-driven equalizer. It plays
locally through [examples/music](../examples/music). The
[dino](../examples/dino) and [flappy-bird](../examples/flappy-bird) examples
step their games and paint the scene from the tick. The
[departure board](../examples/departure-board) example rotates its feed by
retargeting operations in the tick, opts into `page.Probe().UseFrameDirty` with one row band
per cascade frame and meters its refresh bar at 10 Hz.

## Hover without layout

`Handlers.Hover` receives the previous and next hit boxes. A callback that only changes retained paint operations can return `(true, nil)` to skip CSS relayout. The page invalidates both boxes. Return `(false, nil)` to use the ordinary CSS path; an error leaves the hover state unchanged. The callback must preserve geometry and handle both restoring the previous box and painting the next one. Rebind cached operation pointers after `page.Probe().Generation` changes. PNG paints the retained display through the software painter; see [screen.md](screen.md#png-and-the-web-page).

## Scroll windows

`Page.Scroll().SetWindowing(true)` enables `SetScrollWindow(func(offsetY, viewH int) bool)`. Return true after updating row data, or false when the existing overscan still covers the viewport. The host redraws only when true. `Redraw` also calls the callback before executing the template. The callback now returns a bool; callers using the earlier callback shape must add that return value. Preserve the full content height with spacers and keep stable row IDs. The fixed-height example is not a general variable-height or keyboard-focus virtualization implementation.
