# Screen paint

`Redraw` in `internal/page/page_draw.go` does three steps.

1. It executes the page template with the data set by `SetData`.
2. It asks `internal/render.DisplayListState` for the placement as vector operations.
3. If `render.Replayable` accepts every operation, it stores the `layout.Display` and its `Boxes`, and no bitmap. Otherwise it calls `render.PaintState` and stores the `image.Image` and the `Boxes` from `layout.LayOptions`.

`internal/render/display.go` runs `html.Parse` and `css.Apply` with the frame size, media `screen`, and any page theme as an extra sheet, then calls `layout.DisplayListOptions` instead of `layout.LayOptions`. The engine stops before the paint step, so no picture exists.

`internal/render/paint.go` runs the same parse and cascade and calls `layout.LayOptions`, which paints through `imageout.RenderLayout` and returns a placement; `paint.go` takes its picture and boxes. go-gpui never calls `Document.WritePDF`, `Document.PDF`, or `ImageDocument`, so no PDF is built.

## Images

`Page.SetImage` stores encoded bytes (PNG, JPEG, or SVG) under a source name. `Redraw` passes a resolver through `render.State.Images`; `display.go` and `paint.go` hand it to the engine as `layout.Options.Images`. A template source that spells the name, such as `background-image: url("logo")` or `<img src="logo">`, then paints as an `OpImage` on the display-list path or into the bitmap on the fallback path. A source with no entry resolves to nothing, so a page that never calls `SetImage` paints as before. `SetImage(src, nil)` removes the entry. The fetch example registers a fetched PNG or JPEG under the name `fetched` and points the body background at it.

The library that does the layout is still named gowkhtmltopdf. Its image painter uses `pdf.Font` and `pdf.Registry` as font tables. Those types are not a PDF file, and this window does not rasterize one.

## Replay

`internal/render/replayable.go` decides whether `internal/replay` can draw every operation the way the engine's bitmap painter would. It accepts:

- `OpFillRect` with circular corners. `render.FillRadii` resolves the four corners and rejects elliptical ones.
- `OpStrokeRect` when every set bit is a known side (`StrokeMask` zero, or any combination of the four side bits). Masked strokes and elliptical corner radii replay; an unknown mask bit falls back.
- `OpImage` when `ImageBytes` returns a payload. The replay decodes it and applies the op's transform, so a rotated or scaled image replays; an op with no payload falls back.
- `OpLine`. `internal/replay/line.go` rebuilds the engine's centered, square-capped stroke as one filled rectangle from `PaintLineGeometry`, which matches a horizontal or vertical segment; a diagonal one, such as a checkbox tick or a wavy text-decoration, draws as an axis-aligned rectangle. A CSS outline is these same line ops, or one rounded `OpStrokeRect`, carrying the engine's outline flag; `Display.Order` moves them to the outline paint layer, and `replay.Draw` iterates that order.
- `OpText` and `OpBullet`. `internal/replay/text.go` shapes with Ebiten `text/v2` from `op.Font.Bytes()`, places the baseline with the face ascent, applies `text-transform`, and double-strikes fake bold by one pixel, as the engine does. Letter-spacing replays; a missing font, rotation, fake oblique, font features, and autospacing fall back.
- `OpGridRun` (one table row's collapsed border grid), replayed segment by segment.

`OpNoop` and `OpLinkURI` paint nothing and are accepted. Anything else falls back to the bitmap: elliptical fill corners, an unknown stroke mask bit, an image with no payload, a non-normal `mix-blend-mode`, a blend or isolation group, a non-identity transform on a non-image op, a text run with no font, rotation, fake oblique, font features, and text autospacing. A page with any of those keeps the `image.Image` path, so the window never shows a half-replayed page.

The window picks the mode in `internal/window/sync.go`. When `Screen.Display()` is non-nil it keeps the display list and disposes any bitmap; otherwise it builds one Ebiten image from `Screen.Image()`. `internal/window/draw.go` calls `replay.Draw` with the scroll offset, or blits the image. First it fills the frame with a background color: the first fill in paint order that starts at the canvas top-left and spans its width, which is the html or body background, or the bitmap's top-left pixel on a fallback page, or white. Hit testing and scrolling use `Display.Width` and `Display.Height` for a replayed page. While a fallback frame is on screen, Draw paints a small bitmap fallback badge in the top-right corner, so the active mode is visible while running an example.

Ebiten text replay needs Ebiten v2.10.4 or newer. Ebiten v2.9.8 requires `go-text/typesetting` v0.3.0 and builds a font face without the cmap cache v0.3.4 added; v0.3.4, which gowkhtmltopdf requires, then maps every codepoint in U+0000-U+00FF to glyph 0. Ebiten v2.10.4 requires v0.3.5 and builds the face through `font.NewFace`, which clears the cache.

`Display` holds `Ops`, the operations in source order, `Order`, the same operations as indices in paint order, and `Boxes`, the element border boxes in CSS pixels, matching the boxes `layout.LayOptions` returns. `Order` is the order to iterate, because it applies z-index, the outline paint layer, and the chrome-below-content rule. `Width` and `Height` are the canvas in CSS pixels. Op coordinates are points with y down, and for `OpText` and `OpBullet` the `Y` field is the baseline. Multiply a coordinate by `Display.PixelPerPoint` to reach CSS pixels.

The kinds are `gpui.DisplayOpFillRect`, `OpStrokeRect`, `OpLine`, `gpui.DisplayOpText`, `OpImage`, `OpLinkURI`, `OpBullet`, `OpGridRun`, `OpUnknown`, and `OpNoop`; the bare names are `internal/render` constants. Two kinds paint nothing and must be skipped: `OpNoop`, left behind when overflow clipping deactivates an operation, and `OpUnknown`, the boundary marker of a blend or isolation group. Treat any kind outside the list as inert, so a kind added later cannot be mistaken for a fill.

`DisplayOp` is the engine's own operation type under a local name, so callers of `render` do not import the engine to name it. Read a rare payload through its accessor methods, `LinkURI`, `ImageBytes`, `ImageAlt`, `Transform`, `BlendModeName`, `Opacity`, `Outline`, `FontFeatures`, `TextLanguage`, `TextAutospace`, `TextTransformValue`, and `NoFakeBoldValue`. The plain fields such as `Kind`, `X`, `Y`, `W`, `H`, `Text`, and `Font` are always safe. A blend group is read through `Group`, `GroupBoundary`, `IsGroupBegin`, and `IsGroupEnd`, which are nil-safe too.

## PNG and the web page

`Page.PNG` encodes the last picture. On a replayed page there is no picture, so it calls `render.PaintState` once with the stored source and the current state and caches the bytes until the next `Redraw`. `Page.Image` returns nil on a replayed page. `Page.PNG` returns nil when there is nothing to encode: no cached bytes, no picture, and no stored source, or a paint or encode error. `GET /frame.png` and the tests still get a PNG, and a registered image reaches it through the same resolver.

`Prepare` draws the page when both `Image()` and `Display()` are nil. It no longer encodes a PNG just to test for that.

`internal/web/page_http.go` reads `Display().Width` and `Display().Height` when a display exists, and falls back to the image bounds otherwise.

A blank template never reaches the parse. `New` rejects it with `ErrEmptyHTML`.
