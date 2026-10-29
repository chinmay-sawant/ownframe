# Screen paint

`Redraw` in `internal/page/page_draw.go` does three steps.

1. It executes the page template with the data set by `SetData`.
2. It asks `internal/render.DisplayListDocument` for the placement as vector operations, on the styled document from `internal/render/cache.go`.
3. If `render.Replayable` accepts every operation, it stores the `layout.Display` and its `Boxes`, and no bitmap. Otherwise it stores the `bitmap.Paint` result of that list and the `Boxes`.

`internal/render/cache.go` parses and cascades once with the frame size, media `screen`, and any page theme as an extra sheet. A later draw of the same executed source calls `css.Relayout` at the current viewport and pointer state, so a size or state change never parses or recollects.

`internal/render/display.go` holds the one-shot `DisplayListState` entry: it runs the same parse and cascade and calls `layout.DisplayListOptions`. blinkless stops at the drawing list, so no picture exists until ownframe replays it.

`internal/render/paint.go` holds the one-shot `PaintState` entry: it runs that same display list and returns a canvas-sized image. `bitmap.Paint` fills pixels for `page.Probe().PNG` and for a fallback page, returning an error the caller handles.

## Images

`Page.SetImage` stores encoded bytes (PNG, JPEG, or SVG) under a source name. `Redraw` passes a resolver through `render.State.Images`; the render entries hand it to the engine as `layout.Options.Images`. A template source that spells the name, such as `background-image: url("logo")` or `<img src="logo">`, then paints as an `OpImage` on the display-list path or into the bitmap on the fallback path. A source with no entry resolves to nothing, so a page that never calls `SetImage` paints as before. `SetImage(src, nil)` removes the entry. The fetch example registers a fetched PNG or JPEG under the name `fetched` and points the body background at it.

The library that does the layout is blinkless. A text operation carries a font face.

## Replay

`internal/render/replayable.go` decides whether `internal/replay` can draw every operation the way the engine's bitmap painter would. It accepts:

- `OpFillRect` with circular corners. `render.FillRadii` resolves the four corners and rejects elliptical ones.
- `OpStrokeRect` when every set bit is a known side (`StrokeMask` zero, or any combination of the four side bits). Masked strokes and elliptical corner radii replay; an unknown mask bit falls back.
- `OpImage` when `ImageBytes` returns a payload. The replay decodes it and applies the op's transform, so a rotated or scaled image replays; an op with no payload falls back.
- `OpLine`. `internal/replay/line.go` rebuilds the engine's centered, square-capped stroke as one filled rectangle from `PaintLineGeometry`, which matches a horizontal or vertical segment. A diagonal one, such as a checkbox tick or a wavy text-decoration, goes through `internal/replay/segment.go`, which strokes the centerline with square caps. A CSS outline is these same line ops, or one rounded `OpStrokeRect`, carrying the engine's outline flag; `Display.Order` moves them to the outline paint layer, and `replay.Draw` iterates that order.
- `OpText` and `OpBullet`. `internal/replay/text.go` shapes with Ebiten `text/v2` from `op.Font.Bytes()`, places the baseline with the face ascent, applies `text-transform`, and double-strikes fake bold by one pixel, as the engine does. Letter-spacing replays; a missing font, rotation, fake oblique, font features, and autospacing fall back.
- `OpGridRun` (one table row's collapsed border grid), replayed segment by segment.

`OpNoop` and `OpLinkURI` paint nothing and are accepted. Anything else falls back to the bitmap: elliptical fill corners, an unknown stroke mask bit, an image with no payload, a non-normal `mix-blend-mode`, a blend or isolation group, a non-identity transform on a non-image op, a text run with no font, rotation, fake oblique, font features, and text autospacing. A page with any of those keeps the `image.Image` path, so the window never shows a half-replayed page.

The window picks the mode in `internal/window/sync.go`. When `Screen.Display()` is non-nil it keeps the display list and disposes any bitmap; otherwise it builds one Ebiten image from `Screen.Image()`. `internal/window/draw.go` draws a display-list page through a persistent content-sized buffer and blits it. A tick that patches operations dirties tiles. `page.Probe().UseFrameDirty` is an alias. The window calls the inner screen. Pinned chrome paints after the tile blit. A stretched or zoomed page replays into a canvas-sized buffer and scales that to the window, and a fallback page draws its image. First it fills the frame with a background color: the first fill in paint order that starts at the canvas top-left and spans its width, which is the html or body background, or the bitmap's top-left pixel on a fallback page, or white. For a replayed page, hit testing uses the display's `Boxes`, the pointer mapping takes `Display.Width` and `Display.Height` as the page size, and the scroll clamp and the replay buffer grow that canvas to cover every box that overflows it. While a fallback frame is on screen, Draw paints a small bitmap fallback badge in the top-right corner, so the active mode is visible while running an example.

Ebiten text replay needs Ebiten v2.10.4 or newer. Ebiten v2.9.8 requires `go-text/typesetting` v0.3.0 and builds a font face without the cmap cache v0.3.4 added; v0.3.4, which blinkless requires, then maps every codepoint in U+0000-U+00FF to glyph 0. Ebiten v2.10.4 requires v0.3.5 and builds the face through `font.NewFace`, which clears the cache.

`Display` holds `Ops`, the operations in source order, `Order`, the same operations as indices in paint order, and `Boxes`, the element border boxes in CSS pixels, matching the boxes `layout.DisplayList` returns. `Order` is the order to iterate, because it applies z-index, the outline paint layer, and the chrome-below-content rule. `Width` and `Height` are the canvas in CSS pixels. Op coordinates are points with y down, and for `OpText` and `OpBullet` the `Y` field is the baseline. Divide a coordinate by `Display.PointsPerPixel` to reach CSS pixels; multiply a box by it to reach points.

The kinds are `ownframe.DisplayOpFillRect`, `OpStrokeRect`, `OpLine`, `ownframe.DisplayOpText`, `OpImage`, `OpLinkURI`, `OpBullet`, `OpGridRun`, `OpUnknown`, and `OpNoop`; the bare names are `internal/render` constants. Two kinds paint nothing and must be skipped: `OpNoop`, left behind when overflow clipping deactivates an operation, and `OpUnknown`, the boundary marker of a blend or isolation group. Treat any kind outside the list as inert, so a kind added later cannot be mistaken for a fill.

`DisplayOp` is the engine's own operation type under a local name, so callers of `render` do not import the engine to name it. Read a rare payload through its accessor methods, `LinkURI`, `ImageBytes`, `ImageAlt`, `Transform`, `BlendModeName`, `Opacity`, `Outline`, `FontFeatures`, `TextLanguage`, `TextAutospace`, `TextTransformValue`, and `NoFakeBoldValue`. The plain fields such as `Kind`, `X`, `Y`, `W`, `H`, `Text`, and `Font` are always safe. A blend group is read through `Group`, `GroupBoundary`, `IsGroupBegin`, and `IsGroupEnd`, which are nil-safe too.

## Position limits

These are blinkless holes, not ownframe bugs. 1.0 does not implement CSS sticky or flex-absolute. Pin chrome with `SetScrollWindow` and `SetViewportPinZ`. See the CSS matrix in [features.md](features.md).

- `position: absolute` on a flex child stays in flow, so `top` and `left` do nothing. Put a block container around the screen. Telegram's `app-thread` class sets `display: block` on the app div.
- `position: sticky` stays in flow and has no `Fixed` ops. The window translates it with the page. `position: fixed` ops stay untranslated. `Page.Scroll().SetViewportPinZ(z)` draws operations at z-index `z` or above at the offset the display was built with. `Page.Scroll().SetWindowing(true)` plus `Page.SetScrollWindow` counter-moves each pinned element's `top` by the scroll offset, so the window's translation lands it back at the viewport edge. Telegram uses both.
- An absolutely positioned element resolves against its containing block's content origin. Padding on the container shifts it. Subtract that padding in the computed `top`.
- A single-line text box narrower than its content wraps the text anyway and drops spaces at the span boundaries, which reads as overlapped text. Size the box to its content or lay out only the visible runes: the code viewer slices unwrapped rows to their viewport window in Go and paints the slice at the sub-column remainder.

`examples/telegram` is the worked example. Windowing is `Page.Scroll().SetWindowing(true)`, and the pin is `Page.Scroll().SetViewportPinZ(2)`. `pin.go` computes the bar positions from the offset and the insets, and `components/thread.css` carries the `app-thread` class.

## PNG and the web page

`page.Probe().PNG()` encodes a software picture. On a replayed page it paints the retained display list, including operation patches, and caches the bytes until the next paint change. Operation patches, display patches, mutating ticks, and explicit invalidation clear that cache; no-op display patches and ticks keep it. `page.Probe().Image()` returns nil on a replayed page. PNG returns nil when the retained operations cannot be painted, when nothing has been drawn, or when encoding fails. It does not silently replace an unsupported retained display with a source repaint. `GET /frame.png` uses this software path; it does not capture the window's GPU tiles.

`Prepare` draws the page when both the inner `Image()` and `Display()` are nil. It no longer encodes a PNG just to test for that. The public reads are `page.Probe().Image` and `page.Probe().Display`.

`internal/web/page_http.go` reads the inner `Display().Width` and `Display().Height` when a display exists, and falls back to the image bounds otherwise. The window calls the inner screen.

A blank template never reaches the parse. `New` rejects it with `ErrEmptyHTML`.
