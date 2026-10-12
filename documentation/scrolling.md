# Scrolling

A page taller or wider than the window scrolls on the mouse wheel. The window moves the page; a page calls nothing.

## Wheel

`internal/window/fit.go` moves the offset 48 px per wheel notch. A positive wheel moves toward the start of the page, a negative wheel toward the end, and the horizontal wheel pans sideways. The offset is clamped to the content, so a wheel or a drag cannot pass an edge.

`internal/window/scrollbar.go` measures the content for that clamp: the canvas size, grown to cover every box that overflows it. A fixed-width child wider than the window scrolls too.

While the picture is stretched to the window, `internal/window/view.go` ignores the wheel and `internal/window/scrollbar.go` and `internal/window/scrollbar_drag.go` hide the thumbs.

## Thumbs

A page that overflows gets a thumb overlay, 8 px thick and at least 24 px long, in `#BDBDBD`. The vertical thumb sits on the right edge and the horizontal one on the bottom edge. Only the thumb is painted, so the page background stays visible under it.

Press a thumb to start a drag. The drag keeps the distance between the cursor and the thumb's start and continues when the cursor leaves the thumb. The window consumes the drag, so it does not also click the page. A press on the track does not page-jump; it misses the thumb and falls through to the content. Thumb drag is mouse-only.

## Coordinates

The window maps a cursor before the page sees it ([pointer.md](pointer.md)). A scrolling page adds the offset, so page coordinates start at the content's top left. A stretched picture scales the cursor instead, so a point halfway across the window is halfway across the picture. Both mappings are in `internal/window/fit.go`.

## Limits

- There is no programmatic scroll API. The page type has no `SetScroll` method. Only the wheel and a thumb drag change the offset.
- The offset survives `Load`, `Back`, and `Forward` ([navigation.md](navigation.md)), and nothing clamps it again when a later page or window is smaller. A shorter page can show empty space until the next wheel event.
- Touch cannot scroll. `internal/window/pointer_touch.go` sends each fresh touch as a press, click, and release. There is no touch drag, pinch, or kinetic scroll.

The [scrolling example](../examples/scrolling) is a column of 40 rows in a 360x480 window.
