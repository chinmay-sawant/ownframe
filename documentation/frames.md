# Frames

`Page.SetTick` registers one function the window calls before it draws each
frame; `SetTick(nil)` removes it. The callback can move a bar, wave a level
meter, or call `Redraw`, so a page can animate without parsing the HTML again. The desktop window, the phone
build, and the WebAssembly canvas call it at Ebiten's default tick rate, about
sixty times a second, and this library has no setting for it. Each frame runs
devtools sync, keys, resize, the passthrough update, pointer, drop, tick, a
reload poll, sync, the page scroll request, a devtools refresh, and wheel, in
that order. An error from the
callback stops the window, and `Run` returns it. `Serve` does not tick.

```go
page.SetTick(func(ctx context.Context) error {
    d := page.Display()
    if d == nil {
        return nil // the page fell back to the bitmap path
    }

    if fill := frame.Fill(d, seekBox, accent); fill != nil && dur > 0 {
        fill.W = barWidth * float64(pos) / float64(dur)
    }

    return nil
})
```

A page the replay accepts keeps its placement as `*ownframe.Display`.
`Page.Display` returns that list and `Page.Boxes` returns the hit-test boxes
from the same placement. Operations carry points; boxes carry CSS pixels, so
multiply a box by `Display.PointsPerPixel` before comparing the two. The paint
fields of `ownframe.DisplayOp` (`X`, `Y`, `W`, `H`, `R`, `G`, `B`, `Alpha`, and
`Text`) can change between frames. Changing one changes the next drawn frame;
it does not parse, cascade, or lay out anything again. `Alpha` only takes
effect between 0 and 1; 0 means unset. To hide a text run, empty `Text`; to
hide a fill, collapse `W` and `H`.

A `Redraw` replaces the display list, so a callback that keeps an operation
pointer must find the operation again after any redraw. A window resize
relayouts the page and replaces the display list the same way. `Click` redraws
after its handler. `KeyUp` never draws, and `Copy` does not draw; `KeyDown`
draws only when Escape clears a focus or a caret key moves. The other input
handlers redraw after they run. The examples use `internal/frame`
to find operations: `frame.Fill` returns the first fill of a colour inside a
box, `frame.Fills` returns them left to right, `frame.Text` returns the first
text run, and `frame.BoxUnits` converts a hit-test box to display-list units. A
text operation carries its baseline in `Y`, so its box test differs from a
fill's centre test.

There are two ways to change the frame without a full `Redraw`. A tick
callback changes an operation in place, which suits an animation that runs
every frame. A click, key, or hover handler changes page content, and the
page reports the changed box with `TakeDirty` so the window repaints only
that box; [repaint.md](repaint.md) has the call shapes. Reach for the tick
when the change repeats every frame, and for the content path when it is one
edit. A page with a tick registered keeps the full replay, because the window
cannot tell which operation the callback changed. A tick that calls
`Page.UseFrameDirty` is the exception: the window repaints the `TakeDirty`
rect, and blits the cached buffer when that rect is empty. `Invalidate`
marks the elements that tick changed. `MarkRect` marks a CSS-pixel line
that has no element id. A tick that never calls
`UseFrameDirty` still replays the whole list.

## Cost

Changing an operation is cheap and can run every frame. A `Redraw` is not: the
engine lays the page out again, and parses and cascades when the executed
source changed. On the Spotify example that is about 400 ms, mostly from
re-rasterizing the SVG artwork. An animation should change operations per
frame and reserve `Redraw` for real content changes, such as moving the active
row when a track ends.

## Limits

- The window replays the display list, so operation changes show in the
  desktop window, the phone build, and the WebAssembly canvas. `Serve` never
  ticks, and `Page.PNG` paints from the template source, so a `-web` page
  shows the last `Redraw` and not the frame changes.
- A page that fell back to the bitmap path has no display list; `Display`
  returns nil and there is nothing to change.
- CSS `@keyframes` and `transition` are permanent non-goals of the engine
  (see [theming.md](theming.md)), which is why frames move from Go.

The [Spotify player](../examples/spotify-player) example uses the tick for a
moving seek bar, a running clock, and a sine-driven equalizer. It plays
locally through [examples/music](../examples/music). The
[dino](../examples/dino) and [flappy-bird](../examples/flappy-bird) examples
step their games and paint the scene from the tick.

## Hover without layout

`Handlers.Hover` receives the previous and next hit boxes. A callback that only changes retained paint operations can return `(true, nil)` to skip CSS relayout. The page invalidates both boxes. Return `(false, nil)` to use the ordinary CSS path; an error leaves the hover state unchanged. The callback must preserve geometry and handle both restoring the previous box and painting the next one. Rebind cached operation pointers after `Generation` changes. This affects desktop replay; PNG and PDF render from source.

## Scroll windows

`SetWindowing(true)` enables `SetScrollWindow(func(offsetY, viewH int) bool)`. Return true after updating row data, or false when the existing overscan still covers the viewport. The host redraws only when true. `Redraw` also calls the callback before executing the template. The callback now returns a bool; callers using the earlier callback shape must add that return value. Preserve the full content height with spacers and keep stable row IDs. The fixed-height example is not a general variable-height or keyboard-focus virtualization implementation.
