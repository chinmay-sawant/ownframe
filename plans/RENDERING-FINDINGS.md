# Historical record (not live)

This file records 2026-10 rendering research. It is not a live compositor diary. Current paint behavior is in `documentation/` and `AGENTS.md`.

# Rendering findings: rasterize-to-image vs. real rendering

Ten sub-agents surveyed the pipeline, the dependency tree, and the 2026
landscape of HTML rendering in Go. This file is the record. Numbers come from
real benchmarks and real builds run on this machine (i7-13700HX, Go 1.26.0,
Linux/XWayland), not from documentation.

## Summary

Two findings decide this.

1. ownframe already has a vector renderer. `gowkhtmltopdf` builds a full
   retained display list (`internal/layout`, `[]Op`) and then flattens it to
   an `image.Image` at the last possible moment. The display list is what we
   actually want; it was only unreachable because of Go's `internal/` rule.
   **That export is now done and reachable** (see "Status" below). The replay
   that draws it is the remaining work.
2. The blur is not in the rasterizer. It is the final window blit. Fixing the
   blur is cheap and independent of the display-list work.

Nothing draws the display list yet at this point in the record. (2026-10-02:
`internal/replay` now does; see the status update below.) `internal/window` still blits one bitmap,
`internal/web` still encodes it for `GET /frame.png`, and the phone bind still
draws it. The export is a read path over the same placement, and no paint
behavior changed.

## The current pipeline

```
internal/render/paint.go
  html.Parse -> css.Apply -> layout.Lay -> placed.Image()
internal/window/draw.go
  screen.DrawImage(one bitmap)          // per frame, one call
```

`layout.Result` exposes exactly three methods:

```go
func (r *Result) Image() image.Image
func (r *Result) Boxes() []Box
func (r *Result) Size() (int, int)
```

The `[]Op` display list is built at `internal/layout/layout.go:152` with kinds
`OpFillRect`, `OpStrokeRect` (8 corner radii, stroke masks), `OpLine`,
`OpText` (font, size, baseline, letter-spacing), `OpImage`, `OpLinkURI`,
`OpBullet`, `OpGridRun`. It carries 2D transforms (`Matrix2D`), a blend-group
tree, and `PaintOrder` for z-index-correct sequencing. It has two sinks today:
`internal/pdf` and `internal/imageout` (hardwired to `*image.NRGBA`, no
injection point).

Element geometry available today: `[]layout.Box` = `{ID, Tag, Action, Text,
X, Y, W, H}` in CSS pixels, document order. Enough for hit testing. No
per-element style, no per-glyph geometry, no clip rects.

## Measurement: where the time goes

`Redraw` is 94-98% rasterization. Layout itself is 0.5-0.9 ms.

| viewport | layout | rasterize | ratio |
|---|---|---|---|
| 800x600 | 0.88 ms | 14.4 ms | 94% raster |
| 1280x800 | 0.50 ms | 22.8 ms | 98% raster |
| 1600x1200 | 0.60 ms | 35.2 ms | 98% raster |

CPU profile at 1x: `downscaleBox2` 22.76% (a hidden supersample buffer being
thrown away), `rasterGlyphAlpha` 31.57%, `paintFillRect` 13.22%.

### The supersample threshold cliff

`internal/imageout/imageout.go` paints at 2x and box-filters down, but only
below a hard pixel budget:

```go
const rasterSS = 2
const directRasterPixels = maxPooledRasterBytes / (4 * rasterSS * rasterSS)
                          // = 32MiB / 16 = 2,097,152 px
```

Canvases under 2,097,152 final pixels take the supersample path (4x the
pixels, then downscale). At or above it, supersampling is silently dropped.

| canvas | px | vs threshold | ns/op |
|---|---|---|---|
| 1448x1448 | 2,096,704 | 448 below | 41.9-58.6 ms |
| 1449x1449 | 2,099,601 | 2449 above | 6.9-7.4 ms |
| 2048x1024 | 2,097,152 | exactly at | 8.2-8.6 ms |

0.14% more pixels, 6-8x less time. Consequence: a naive HiDPI fix can make
typing much worse. A 2x canvas for an 800x600 window is 1.92 MP, below the
threshold, so it re-enters the expensive branch:

| page size | MP | keys/s (today) | keys/s if naively doubled |
|---|---|---|---|
| 480x640 | 0.31 | 140 | - |
| 800x600 | 0.48 | 99-108 | - |
| 1600x1200 | 1.92 | 26-27 | **113 -> 25** |
| 2400x1800 | 4.32 | 65-69 | - |

Cost is linear at ~5.5 ms/MP from 2x on. No super-linear term.

## Measurement: where the blur is

The rasterizer is already good. `rasterSS = 2` gives a whole-canvas 2x
supersample plus a 2x box down, and `ttfraster.go` does 4x or 8x per-glyph
supersampling for greyscale AA. At 1x output that is measurably *sharper*
than a genuine 2x raster reduced to the same size (gradient energy 28.31 vs
24.98).

The blur is the last step. The app never gets a 2x canvas:

- `shell.Layout` returns dip-sized values, so Ebiten's `DrawOffscreen` canvas
  is dip-sized.
- `DefaultDrawFinalScreen` upscales to device pixels using **`FilterPixelated`**
  (nearest neighbour, 2x2 block duplication). That is the blurry text.

Cost of magnifying a 1x raster to 2x device pixels, versus a native 2x raster
of the same page:

| route | MAD on ink | gradient energy kept |
|---|---|---|
| native 2x (reference) | - | 100% |
| 1x -> bilinear 2x | 10.82/255 | 50% |
| 1x -> nearest 2x | 10.78/255 | 64% |

Roughly 1-2 px of that residual is baseline rounding rather than blur
(best-fit alignment search moved the error from 10.82 to 9.93).

A native 2x render of the same page costs **23.5-25.2 ms vs 34.0-36.2 ms at
1x**. It is cheaper, because 1.01 MP falls under the threshold and pays for a
4x-pixel buffer.

Reaching a 2x canvas is not a one-line change. `render.Paint` has no DPR knob,
`css.Options` has none, and `layout.Lay` hardcodes `Zoom: 0`, so
`imageout.RenderOptions.Zoom` is unreachable through the public path. `html
{zoom: 2}` and `body {transform: scale(2)}` are both silently ignored
(measured: identical `#document` box). The reachable fix is in
`internal/window`: have `Layout` return `dip * dpr`, divide
`CursorPosition()` by dpr before `contentPoint`, and untangle `stretched()`,
which will then take the `FilterLinear` downscale path.

Note: HiDPI is not active on this machine (`XWAYLAND0`, dsf = 1), so the blur
cannot be reproduced here as-is. `RunGameOptions.DisableHiDPI` is not set
anywhere in the repo.

## Dirty regions: sound idea, blocked by API shape

A single keystroke changes 40-58 pixels, 0.008-0.013% of the canvas, in a
10x12 dirty box. Marginal fill rate is 2-4 ms/MP direct, so a clipped repaint
would cost ~0.0003 ms against 35 ms today.

It does not work because `rasterizeContext` is strictly whole-canvas.
`RenderOptions.Crop` exists but `frame.go` applies it *after* rasterizing.
There is no clip, region, or tile parameter. `PlacedElement` carries no pixel
layer, so ownframe cannot even compute a damage rect.

This is worth more than the DPR work, but it is a performance project, not a
rendering project.

## The options, and why the obvious ones are dead

| option | verdict | reason |
|---|---|---|
| CEF / Chromium embedded | dead | `nicedoc/chromium` and `jchv/go-cef` do not exist. Every Go CEF binding is window-owning by design; none wraps offscreen rendering. The live one (`energye/cef`) needs a 311 MB CEF download plus `liblcl`. |
| WebView2 / WKWebView / WebKitGTK in-process | dead | Ebiten exposes **no native window handle** (`ebiten/v2` has zero exported functions returning `uintptr`/`unsafe.Pointer`; the handle is an unexported GLFW field). `webview_go` requires a `GtkWindow`, never a raw X11 window, which is what Ebiten has on Linux. |
| Wails | dead | Owns the process lifecycle, the window, the asset server, the JS bridge, and the main loop. `New` is a hard singleton. Ebiten wants all five. |
| Sidecar process | rejected | PNG encode alone is 4.1 ms against a 7 ms repaint; raw RGBA is a full memcpy per frame. Illegal on iOS. Compiles on wasm but fails at *runtime*, the worst failure mode for a library. |
| `modernc.org/webview` | not as a public dependency | Renders well and is cgo-free and wasm-capable, but: one contributor, 4 stars, zero issues ever filed, 1246 commits in three months, a checked-in 1 MB `devlog.md`, `CLAUDE.md` + `HANDOFF-CLAUDE.md`, and a shipped spec bug (`position:fixed` ignores z-index against absolutely-positioned siblings, a CSS 2.1 Appendix E violation that survived 1244 tests). Fine behind our own interface. Reckless as the engine our users depend on. |
| chromedp headless | viable, heavy | Real Chromium, one-shot bitmaps only. Needs a ~115 MB `chrome-headless-shell` download. Not continuous. |
| **GPU replay of the display list** | **recommended** | ~60 lines upstream, ~300 lines here. Zero new modules, no cgo, all platforms, no contract break. |

## Recommended: export the display list, replay it on the GPU

The trick is a type alias. `Op` is already an exported struct whose rare
payload sits on an embedded `*opExtra`, so an exported alias in a non-internal
package hands external code the exact struct the PDF and PNG adapters read.

Verified by compiling a probe against the real module: promoted exported
fields **are** readable cross-package through the unexported embedded pointer,
exported setters work, and exported methods are callable. The only two
unnameable types are `*pdf.Font` and `*pdf.StructElem`, and neither must be
named by a consumer (cache with `map[any]*text.GoTextFace` keyed on
`op.Font`).

### The font handoff, solved

`pdf.Font` keeps raw face bytes in the unexported `data` field with no
accessor. One 4-line method closes that. `DefaultFont()` is Liberation Sans
embedded in the binary via `internal/pdf/assets`, so there is no system-font
or filesystem dependency.

Variation axes are a non-issue: `resolveFontVariants` calls `face.Instance`
at layout time and `buildInstancedTTF` bakes the axes into a static TTF, so
`Op.Font` is already the post-instance face.

Baseline placement is exact and free. Measured against real serif-bold-italic
and mono ops:

```
text "Wij Affligé" size=16px  HAscent=14.2578  Ascent/upm*px=14.2578  delta=+0.0000
text "mono 123"    size=16px  HAscent=13.3203  Ascent/upm*px=13.3203  delta=+0.0000
```

So translate by `(X*pxPerPt, Y*pxPerPt - HAscent)` and the baseline lands
exactly. `pxPerPt = 4/3`. Note `GoTextFace.Size` is in **pixels**, so
`Size = op.Size * 4/3`, and it should be bucketed because Ebiten caches glyph
images per size.

### The one footgun

Reading a promoted field through a nil `opExtra` panics with a nil pointer
dereference. Confirmed by running a probe. The zero value of the alias is not
safe to touch. Any exported surface must either bind an extra or expose
nil-safe accessors.

### What the replay does not recover

| behaviour | status |
|---|---|
| `mix-blend-mode` | unsupported. Ebiten's `Blend` is Porter-Duff only, `ColorScale` is `scale*x + add` and cannot express multiply. Only `normal` survives. |
| `letter-spacing`, `text-autospace` | need per-glyph blits via `text.AppendGlyphs` (whose `Glyph.Image`/`X`/`Y` are public). Works, costs a second shaping walk. |
| glyph rasterization | different rasterizer, better subpixel placement (1/64 px vs integer snapping), no TrueType hinting either way. |
| glyph cache lifetime | improves. Today's atlas is per-`rasterizeContext`, so every `Redraw` re-rasterizes every glyph. |
| `vector.FillPath` has no `GeoM` and no scissor | batch ops by colour/alpha class, or pay a stencil flush per op. Cull by bounds. |
| images | improves (linear filtering vs today's nearest). |
| `RotateDeg` | improves (`GoTextFace.Direction` handles vertical runs properly). |
| element `opacity < 1` on a container | already broken today; not a regression. |
| CSS `Xform` | bake `Matrix2D.Apply` into path vertices; `vector.Path` has no `Transform` and `ArcTo` cannot be transformed. |

### Effort

| milestone | lines |
|---|---|
| fill-rect only, no radius | ~40 |
| + rounded fill, paint opacity | ~90 |
| + text (face cache, baseline, faux bold/oblique, features, language, transform) | +150-200 |
| **first shippable (fills + text + lines + grid)** | **~250-300** |
| + full stroke rect, stroke masks | +120-150 |
| + images, transforms, isolation groups | +190 |
| full parity | ~800-900 |

Under ownframe's 2000-*character* rule that is roughly 20 files at ~45 lines
each. Calibrate against `internal/clipboard`: 19 files for one OS
integration. Plan the split up front.

### What it does not buy

JavaScript. No `<a href>`. No flexbox beyond what gowkhtmltopdf already
implements. None of the "still absent" list in `documentation/features.md`.

If the real requirement is JS, the only shape that works is a platform split:
on `js/wasm` you are already inside Blink so `document.body.innerHTML` is
free (~1.5k chars, one build tag); on mobile, WebView is the standard answer;
on desktop, a native webview. The cost is that `Boxes()` must die, because no
real engine hands you a CSS-pixel box list. That breaks
`Handlers.Click func(ctx, box Box)` for every downstream user on every
platform. That is a v0.0.2 with a deprecation shim, not a patch.

## Compatibility matrix

| option | linux | windows | darwin | js/wasm | android | ios |
|---|---|---|---|---|---|---|
| today (pure Go) | yes | yes | yes | yes | yes | yes |
| display-list replay | yes | yes | yes | yes | yes | yes |
| native webview | cgo | cgo | cgo | no | yes | yes |
| sidecar process | cgo | cgo | cgo | runtime fail | partial | forbidden |

`internal/window`, `internal/render`, and `internal/host` have **zero build
tags**. Every build tag in the repo lives in `internal/clipboard` and
`examples/login`. Any option except the display list forces that property to
end.

## Status: the export is implemented and merged

The display-list export described below has landed in `gowkhtmltopdf` and is
merged to its `master` (PR #84, merge commit `9b445ce`). `ownframe` pins that
commit, so no local checkout or `replace` directive is needed.

In `gowkhtmltopdf`:

| file | what it adds |
|---|---|
| `layout/displaylist.go` | public `DisplayList`, `Display`, `DisplayOp`, `DisplayGroup`, `DisplayKind`, `DisplayOrder`, `DisplayFakeBold`, `DisplayTransformText`, and the `DisplayOp*` kind constants |
| `internal/layout/op_read.go` | nil-safe accessors for the rare payload, which sits behind an embedded pointer to an unexported type |
| `internal/pdf/fonts.go` | `Font.Bytes()` returns the raw SFNT face |
| `internal/imageout/frame.go` | `LayoutResult` lays out without rasterizing |
| `internal/layout/layout_measure.go` | `OpKindNoop` names the deactivated-operation sentinel |

In `ownframe`:

| file | what it adds |
|---|---|
| `internal/render/display.go` | `render.DisplayList(ctx, source, width, height)` |
| `internal/render/display_ops.go` | the kind constants and `Kind`, re-stated locally |
| `internal/render/display_kinds_test.go`, `display_clip_test.go` | guards that every emitted kind is nameable and that clipped ops are skippable |

`go.mod` pins the merged commit. Once `gowkhtmltopdf` tags a release carrying
the export, that pin can move to the release version.

### Status update, 2026-10-02: the replay landed

`internal/replay` in ownframe now draws the display list: fills with circular
corners, axis-aligned border lines, grid runs, and shaped text. A page with an
operation it cannot reproduce keeps the bitmap path. `Display.Boxes` was added
upstream so a replayed page does not need `Lay` for hit testing. Ebiten moved
to v2.10.4: v2.9.8 builds its font face without the cmap cache that
`go-text/typesetting` v0.3.4 added, which maps every codepoint in
U+0000-U+00FF to glyph 0. `examples/login` replays fully. Because the new
`Display.Boxes` field still rides on the local `chore/changes-for-go-gpui`
branch upstream, ownframe carries a temporary `replace` to `../gowkhtmltopdf`
until that branch merges.

### Status update, 2026-10-02: full strokes and images replay

`internal/replay` now also draws two kinds that previously forced the bitmap
path. `OpStrokeRect` replays when `StrokeMask == 0` (the complete rounded
rectangle) and the corner radii are circular; a masked stroke or an elliptical
radius still keeps the bitmap. `OpImage` replays when its transform is the
identity, drawn from the op's encoded payload; a transformed image still keeps
the bitmap. The rejections live in `internal/render/replayable.go`.

### Corrections the review surfaced

Four things differed from the research above. All are fixed in code and docs.

**A third inert kind exists.** `opKindNoop` is `255`, written by the clipper
when overflow deactivates an operation. It is not removed, because the box tree
stores operation indices that must not shift. The reproducing case is an
absolutely positioned child clipped by `overflow: hidden` on a positioned
ancestor: the page yields one op of kind 255, sometimes alongside real content.
The public list now names it as `DisplayOpNoop`. Exporting it made the
`exhaustive` linter demand that two production switches stop ignoring it
implicitly, which is the right outcome.

**`OpUnknown` is emitted, not reserved.** It is the boundary marker that opens
or closes a blend or isolation group. A `mix-blend-mode: multiply` div produces
two of them, each with a non-zero `GroupBoundary()`. Treating it as inert is
still correct; claiming layout never emits it was not.

**Canvas height can differ by one pixel.** `Lay` reads its size back off the
painted picture, which rounds at the supersample factor, while `DisplayList`
converts the placement height straight from points and truncates. Measured
across roughly 310 documents: 61 mismatches, all delta `(0, 1)`, never in width.
Three reproductions: a 3x3 viewport, a fractional 777.77px div, and a 30-row
table. The public docs and the CHANGELOG now say so.

**The ops slice is safe to retain, with two caveats.** `layoutContextWithStyles`
always allocates, and the only reuse path is `WithWorkspace`, which `imageout`
never uses, so the caller holds the only reference. The slice carries growth
slack (`len=200` came back as `cap=601`, about 100 KB at 256 bytes per op), so
clip it once at ingest. `ImageBytes()` aliases the engine's image buffer: read
it, do not write it.

### A pre-existing race, not caused by this work

`go test -race ./internal/layout/` fails intermittently, and it failed before
this change too. Measured by moving the new files aside: 2 of 3 baseline runs
failed, 4 of 4 with the files removed in one later sample. The shared object is
the `*go-text/typesetting/font.Face` cached per `pdf.Font` under `gotOnce`; the
default faces are a process-wide singleton, so parallel tests shape through one
face, and typesetting v0.3.4's lazy caches (`font.extentsCache`, the cmap
`cache21_19_8`) are unsynchronized. Stack: `layout.Paint` -> `paintPages` ->
`drawText` -> `Content.TextShowLanguageFeatures` -> `HarfbuzzShaper.Shape`.
A correct fix is per-face locking around shaping or one face per concurrent
run; upgrading typesetting only helps if a later version locks those caches.
This is a live CI flake, since CI runs `-race` on `layout` and `pdf`.

## Corrections to earlier assumptions

- **The wasm path does not use PNG-over-HTTP.** `browser/index.html` is 20
  lines that load `wasm_exec.js` and run the module; Ebiten's wasm driver owns
  the WebGL canvas and `draw.go` blits into it, same as desktop. PNG-over-HTTP
  exists only in desktop `-web` mode (`GET /frame.png` in `internal/web`).
- **`layout.Lay` returns a content-height canvas**, not the requested height:
  `Paint(src, 800, 600)` returns an 800x1263 image. Any replay must apply
  `max(content, requested)` and convert points by 4/3 to match.
- **`cgo` is already committed** by Ebiten on linux and macos
  (`CGO_ENABLED=0 go build ./...` fails on linux, succeeds on windows). This
  softens the usual objection to cgo-requiring webviews, on two of three
  desktops.
- **`Op` has no path/bezier primitive.** Rects with elliptical corner radii
  only; the arcs are generated at paint time (`roundedArcSteps = 8`). Not a
  gap for a replay, since we build the arcs ourselves.

## Related observations, out of scope

- `internal/page/page_prepare.go` PNG-encodes the whole page just to test
  `len(page.PNG()) == 0`. (Fixed 2026-10-02: `Prepare` now checks `Image()`
  and `Display()` instead.)
- `internal/imageout` already has a `rasterPolicyDirect` branch reachable only
  internally. Exposing it is ~10 lines for a 4-5x `Redraw` win, but rounded
  corners and hairlines visibly degrade (4.03% of subpixels differ, corner
  arcs stair-step) because only the skipped canvas supersample was
  antialiasing them. The GPU replay improves both axes; this does not.
- `ebitenui` is irrelevant here. It is a retained widget library that shares
  no code path with HTML and would fight the `data-action` / `Boxes()`
  hit-testing model.
- There is no production-ready SDF/MSDF text renderer in Go. The best option
  (`gogpu/gg/text/msdf`) has zero importers and would need a fragment shader
  that Ebiten's stock pipeline has no place for. `text/v2.AppendVectorPath` +
  `vector.FillPath` already gives exact outlines on the GPU.
- `benoitkugler/textlayout` is superseded ("merged into go-text/typesetting").
  Its successor `benoitkugler/textprocessing` is **LGPL**; do not let the line
  breaker route through it.
