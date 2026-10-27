# Scrolling

A page taller or wider than the window scrolls on the mouse wheel, a touch drag, or a request from the page. The window moves the page and owns the offset.

## Wheel

`internal/window/fit.go` moves the offset 48 px per wheel notch. A positive wheel moves toward the start of the page, a negative wheel toward the end, and the horizontal wheel pans sideways. The offset is clamped to the content, so a wheel or a drag cannot pass an edge. Holding Ctrl (or Command) while turning the wheel zooms the page instead of scrolling it, unless the page sets `Config.DisableWheelZoom`, in which case that wheel scrolls. [keys.md](keys.md) has the zoom keys and the range.

The window measures the canvas and grows it to cover every box that overflows it. A fixed-width child wider than the window scrolls too. The replay buffer uses the same measure, so scrolled content stays painted.

`Config.Scrollbar` controls the scrollbar thumbs. `ownframe.ScrollbarAuto` is
the default: it shows thumbs on desktop and web, and hides them on Android and
iOS. Set `ownframe.ScrollbarVisible` or `ownframe.ScrollbarHidden` to override
that choice for one page. Hovering a visible thumb shows the matching resize
cursor; dragging it moves that scroll axis without sending a page click.

Thumb visibility does not control scrolling. Touch drags, mouse wheels, and
page requests continue to work when thumbs are hidden. `Page.Scroll().SetAllowScroll`
controls whether the page can scroll at all. Stretched pages do not show thumbs.

## Coordinates

The window maps a cursor before the page sees it ([pointer.md](pointer.md)). A scrolling page adds the offset, so page coordinates start at the content's top left. A stretched picture scales the cursor instead, so a point halfway across the window is halfway across the picture. Both mappings are in `internal/window/fit.go`.

## Inner regions

A page can claim a wheel event before the window pans the page. Set
`Handlers.Wheel` to `func(ctx context.Context, x, y, dx, dy float64) (bool, error)`.
The point is in page coordinates, including the current offset. Return true to
consume the event; return false to let the window scroll the page. The handler
owns any redraw. The code-view explorer uses this to scroll its file tree while
the editor keeps the page offset.

## Page requests

`Page.ScrollTo(x, y)` and `Page.ScrollBy(dx, dy)` store one request. The window consumes it through `host.ScrollRequester`, clamps it to the content, and clears it. The offset stays owned by the shell, so a request, the wheel, and a touch drag all move the same value. [interaction.md](interaction.md) has the calls.

## Touch

A finger that moves past 8 px drags the page, clamped to the content ends. After lift, the page continues with a decaying fling. Mobile updates follow the display refresh. `Page.Window().SetTouchScrollOptions` tunes finger sensitivity and how quickly the fling slows; both settings have bounded defaults. Two fingers pinch a zoom that stays between 0.25 and 4. [interaction.md](interaction.md) has the gesture details.

## Limits

- The offset survives `Load`, `Back`, and `Forward` ([navigation.md](navigation.md)).
  The window clamps it to the content after any relayout or redraw, so a
  shorter page cannot show empty space below it.
- Desktop and wheel scrolling keep their existing step behavior. Touch scroll
  inertia applies only to touch gestures.

The [scrolling example](../examples/scrolling) is a column of 40 rows in a 360x480 window.
