# DevTools

The window can draw an inspector over the page. It outlines the element under
the cursor, pins one to read its properties as JSON, outlines the display
operations, and prints the frame numbers. The overlay is window chrome, like
the scrollbar thumbs and the fallback badge. It never enters `Page.PNG`, the
display list, or the box list.

## Turning it on

Press F12 or Ctrl+Shift+I. The window owns both keys while a page has an
inspector, so `Handlers.KeyDown` and `Handlers.KeyUp` never see them, and
the matching release is swallowed too. Pressing the key again closes the
overlay and sends one hover for the current cursor, so a stale `:hover` does
not stick.

`Config.DevTools` starts the overlay on, and `Page.SetDevTools` and
`Page.DevTools` read and change the flag at runtime. Use those in a browser
and on a phone: a browser may keep F12 for itself, and a phone has no
keyboard.

```go
page, err := ownframe.New(ownframe.Config{
    HTML:     html,
    Width:    720,
    Height:   560,
    DevTools: true,
})
```

## The dock

The panel is a full-height dock on the right edge, 340 pixels wide by
default, drawn in the Chrome DevTools dark theme. Drag the strip on its left
edge to resize it; the width clamps to 220 pixels and to the window. The
wheel scrolls the dock content while the pointer is over it, and a thumb on
the right edge shows how far the content runs. Opening, closing, or resizing
the dock shrinks or restores the page viewport like a docked browser
inspector, so the page relayouts to the strip left of the dock. The fallback
badge shifts left when the dock is on, so it stays visible.

| Tab | Content |
|---|---|
| Elements | The picked element as pretty JSON. Click a node opener to fold it to `{...}`. |
| Frame | Section headers and right-aligned counters for the window, the frame, the pipeline, and reload. |
| Ops | The outline toggle, the count per kind with a colour chip, and every operation in paint order. Click a row to outline that operation and copy its line to the clipboard. |

The title bar holds the `Picker` toggle and the `F12` close hint, and the
footer holds a hint for the active tab. `o` toggles the operation outlines
and selects the Ops tab, and it does not type while the overlay is on.

## Picking

While the overlay is on, the pointer belongs to the overlay. The `Picker`
toggle in the title bar arms the element picker: moving the cursor outlines
the innermost box under it, and the page picture stays still. The window
sends no hover, so the pixels the numbers describe do not change. A click
pins the box, opens the Ops tab, selects the operation whose centre sits in
the box, and scrolls its row into view. A click on the same box or on page
space with no box clears the pin. With the picker off, a click inside the
page reaches the page control under the cursor. The Elements tab shows the
pinned box when there is one and the hovered box otherwise. Alt+click
forwards the press, the click, and the release to the page, so a control can
still be exercised while inspecting. Closing the overlay clears the hovered
and pinned boxes. A new generation drops either pick when its id is gone
from `Boxes()`, and clears the operation pick.

The dock hit-tests its own rows, never `Page.Boxes()`. A dock click switches
tabs, folds JSON, toggles the outlines, or picks an operation; it never
reaches the page.

## Elements

The Elements tab prints the picked element as JSON:

```json
{
  "tag": "button",
  "id": "send",
  "action": "submit",
  "text": "Send",
  "rect": {
    "x": 24,
    "y": 96,
    "width": 120,
    "height": 40
  },
  "ops": 2
}
```

Keys, strings, numbers, and booleans take the usual syntax colours. Click a
node opener to fold it to `{...}` and click again to unfold. Text longer than
120 runes is clipped with dots.

## Frame

The Frame tab groups counters under section headers and right-aligns each
value:

| Section | Rows |
|---|---|
| WINDOW | window and page size, scroll offset, stretched, fallback, sequence |
| RENDERING | fps, tps, frame time, draw time |
| PACING | idle, next deadline in ms, presents skipped while parked |
| PIPELINE | redraws, parses, cascades, layouts, repaints, relayouts, skipped, boxes, ops, last redraw, last draw |
| RELOAD | reloads and the last reload error |
| PERFORMANCE | frame average, p95, p99, long frame count, per stage times for template, layout, display list, and paint |
| DIRTY | dirty rects taken, full frame fallbacks, operations repainted, operations skipped |
| MEMORY | Go heap in use, total allocation, RSS, goroutine count, GC runs |

A zero duration reads as `0.0ms`, and an empty reload error reads as `-`.
Label widths are memoized and each value is a right-aligned span, so a frame
does not rebuild a padding string per row.
The PERFORMANCE, PIPELINE DETAIL, and MEMORY rows read `-` until the app
opts into Perf with `Config.Perf`, `ownframe.WithPerf(true)`, or
`WindowOptions.Perf`. The legacy WINDOW, RENDERING, PIPELINE, and RELOAD
rows always report.

## Operations

The first row of the Ops tab toggles the outlines, and `o` does the same
from any tab. The count section names each kind, its colour chip, and its
count, then the total. The colour names the kind:

| Kind | Colour |
|---|---|
| fill | `#1a56db` |
| stroke | `#e5484d` |
| line | `#d97706` |
| text | `#176b45` |
| image | `#7c3aed` |
| grid run | `#0891b2` |

The paint order list shows each operation as `#N`, its kind, its bounds, and
the first characters of a text run. Clicking a row outlines that operation
in its kind colour, three pixels wide. Clicking the same row clears the
outline. Clicking a row also writes its text to the OS clipboard, both when
it selects the operation and when it clears the pick. There is no setting to
turn the write off, so keep the dock closed when the rendered text or
coordinates are sensitive.

The bounds are the op's painted box in CSS pixels, from
`replay.PaintBounds`: the same box the replay's partial repaint filter uses,
without its one-pixel slack. Op coordinates are points: divide by
`Display.PointsPerPixel` to reach CSS pixels. A text operation carries its
baseline in `Y`, so its box runs from the face ascent above the baseline to
the ink descent below it, a stroke grows by half its width, a line uses its
inward geometry and stroke width, and a grid run is the union of its
segments. An operation with a transform or a rotation cannot be bounded, so
its row falls back to its nominal box. `DisplayOpNoop`,
`DisplayOpUnknown`, and `DisplayOpLinkURI` paint nothing and get no outline.
A fallback page has no display list, so the tab prints `bitmap fallback` and
offers no list.

The Ops tab is the costliest of the three: it costs milliseconds per frame to
draw, so an fps or frame time read while the tab is open includes the dock's
own work. The row list is cached per `Page.Generation()`, so a still page
rebuilds nothing and pays only to draw the rows and the outlines in view. Read
fps and frame time on the Frame tab.

## A custom screen

The window type-asserts `host.Inspector` on the screen it already holds:

```go
type Inspector interface {
    DevTools() bool
    SetDevTools(on bool)
    Stats() Stats
}
```

A screen without it gets no overlay and no compile change. `Stats` is the
snapshot the Frame tab prints. A screen that also implements
`SetDrawTime(time.Duration)` receives the window's draw time while the
overlay is on, and `Stats().LastDraw` reports it. That hook is separate from
`host.Inspector`, so the interface keeps its three methods.

## Web mode

`GET /debug/state` returns the same data as JSON when the server is started
with `ServeOptions{Debug: true}` or `ServeOptions{Perf: true}`: `width` and
`height`,
`generation`, `fallback`, `boxes`, `stats`, and `ops`. `stats` is omitted
when the screen has no inspector, and `ops` when there is no display list.
The `ops` map counts each kind and a `total`. The route is in
[web.md](web.md).

## Limits

There are no computed styles to show. The engine styles a document internally
and publishes no per-element reader, so the panel shows geometry and
attributes, not `font-size` or `margin`. Suppressing pointer input freezes
the page's hover and press state at their last value. A page that cannot
shrink to the dock-free strip, because of its own minimum size, scrolls
instead of scaling. F12 is a browser key in a wasm build, and
`Config.DevTools` or `SetDevTools` is the cross-platform way in.
