# Frames

`Page.SetTick` registers one function the window calls before it draws each
frame; `SetTick(nil)` removes it. The callback can move a bar, wave a level
meter, or call `Redraw`, so a page can animate without parsing the HTML again. The desktop window, the phone
build, and the WebAssembly canvas call it at Ebiten's default tick rate, about
sixty times a second, and this library has no setting for it. Each frame runs
keys, pointer, resize, tick, sync, and wheel, in that order. An error from the
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

A page the replay accepts keeps its placement as `*gpui.Display`.
`Page.Display` returns that list and `Page.Boxes` returns the hit-test boxes
from the same placement. Operations carry points; boxes carry CSS pixels, so
multiply a box by `Display.PixelPerPoint` before comparing the two. The paint
fields of `gpui.DisplayOp` (`X`, `Y`, `W`, `H`, `R`, `G`, `B`, `Alpha`, and
`Text`) can change between frames. Changing one changes the next drawn frame;
it does not parse, cascade, or lay out anything again. `Alpha` only takes
effect between 0 and 1; 0 means unset. To hide a text run, empty `Text`; to
hide a fill, collapse `W` and `H`.

A `Redraw` replaces the display list, so a callback that keeps an operation
pointer must find the operation again after any redraw. `Click` redraws after
its handler. `KeyDown` and `KeyUp` never draw, and `Copy` does not draw. The
other input handlers redraw after they run. The examples use `internal/frame`
to find operations: `frame.Fill` returns the first fill of a colour inside a
box, `frame.Fills` returns them left to right, `frame.Text` returns the first
text run, and `frame.BoxUnits` converts a hit-test box to display-list units. A
text operation carries its baseline in `Y`, so its box test differs from a
fill's centre test.

## Cost

Changing an operation is cheap and can run every frame. A `Redraw` is not: the
engine parses the HTML, applies the CSS, and lays the page out again. On the
audio player example that is about 130 ms, and on the Spotify example about
400 ms, mostly from re-rasterizing the SVG artwork. An animation should change
operations per frame and reserve `Redraw` for real content changes, such as
moving the active row when a track ends.

## Limits

- The window replays the display list, so operation changes show in the
  desktop window, the phone build, and the WebAssembly canvas. `Serve` never
  ticks, and `Page.PNG` paints from the template source, so a `-web` page
  shows the last `Redraw` and not the frame changes.
- A page that fell back to the bitmap path has no display list; `Display`
  returns nil and there is nothing to change.
- CSS `@keyframes` and `transition` are permanent non-goals of the engine
  (see [theming.md](theming.md)), which is why frames move from Go.

The [audio player](../examples/audio-player) and
[Spotify player](../examples/spotify-player) examples use the tick for a
moving seek bar, a running clock, and a sine-driven equalizer. They play
locally through [examples/music](../examples/music). The
[dino](../examples/dino) example steps the game and paints the scene from the
tick.
