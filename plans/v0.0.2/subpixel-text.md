# Subpixel and LCD text

Recorded 2026-10-13 against Ebiten v2.10.4 (`go.mod`), go-text/typesetting
v0.3.5, and the ownframe replay in `internal/replay`.

The question is what Ebiten's text stack does for glyph quality today and
what a Chrome-like subpixel or LCD mode would cost. The short answer: Ebiten
already does subpixel positioning with grayscale antialiasing, it does not
hint and does not gamma correct, and LCD text is not reachable through
Ebiten's public API. The rasterizer half can live in ownframe because
`AppendVectorPath` is public; the per-channel blend is the blocker.

## What Ebiten v2.10.4 does today

### Grayscale antialiasing

Glyph images are single-channel coverage. `text.go:146-153` documents the
returned `Glyph.Image` as "a grayscale image i.e. RGBA values are the same,
except for color glyphs like emojis". The outline path rasterizes glyph
segments with `golang.org/x/image/vector` into an RGBA image with the
coverage in every channel, `gotextseg.go:98-120`, and draws with
`draw.Src`. Color fonts take other paths, SVG, COLRv0, or bitmap strikes
(`gotext.go:400-411`), but ordinary text is one alpha mask.

### Subpixel positioning, quantized

Subpixel positioning is on, but it is quantized to a small set of variations
per pixel. `glyphVariationCount` returns 8 variations below a 20 px line
height, 4 below 40, 2 below 80, and 1 above that (`text.go:107-133`). For a
horizontal `GoTextFace`, `goTextGlyphImageInfo` snaps the origin X to that
grid and truncates Y to whole pixels (`gotext.go:456-468`), then folds the
remaining 1/64 px offset into the rasterization as `subpixelOffset`
(`gotextseg.go:66-96`). The glyph image cache is keyed on the glyph and the
offset variation (`gox.go:48-66`, `gotextfacesource.go:211-245`), so a run
at fractional positions creates several cached bitmaps per glyph. Horizontal
subpixel positioning is real. Vertical positioning for horizontal text is
pixel-aligned.

`GoXFace`, the `golang.org/x/image/font` path, has the same X-only scheme
with a smaller cache (`gox.go:305-330`, `:335-352`).

### Gamma

None. The rasterizer emits coverage, the coverage becomes alpha, and
`DrawImage` blends in the 8-bit sRGB framebuffer with the normal source-over
equation. There is no gamma table, no linear-light blending, and no
per-face option. The Kage recipes note that premultiplied alpha means any
operation that does not commute with premultiplication needs an explicit
straight-alpha round trip (`skills/writing-kage-shaders/reference/recipes.md`,
"Premultiplied alpha"); text drawing never does one.

### Hinting

None. `gotextfacesource.go:216-227` says the outline path assumes the
typesetting library does not apply hinting at glyph data time, because it
reads raw outline points. The go-text/typesetting module has no hinting
interpreter, only font parsing and shaping. So every size is an unhinted
scale of the outline, which is Chrome-like only in the sense that macOS and
modern Chrome dropped hinting too.

### Drawing

`Draw` appends glyphs, realizes each image, and issues one `dst.DrawImage`
per glyph (`layout.go:141-207`). The text color rides
`DrawImageOptions.ColorScale` (`layout.go:39-49`). The `Face` interface has
a private method (`text.go:33-56`), so no third-party face can be plugged
into `text.Draw`.

### ownframe today

`internal/replay/text.go:15-50` draws each op with `text.Draw` at the
baseline, in points scaled by `pxPerPt`, and re-draws one pixel to the right
for fake bold. Faces are cached per font value and size in
`internal/replay/text_cache.go:32-69`. When the window zooms or stretches,
the whole page buffer is scaled with `ebiten.FilterLinear`
(`internal/window/replay_scale.go:36-47`, `internal/window/view_fit.go:56-60`),
which blurs finished glyphs instead of rasterizing them at the new size.
That is the most visible text-quality gap in this repository, and it is
independent of LCD.

## What a Chrome-like LCD mode needs

Chrome's LCD path, FreeType's `FT_RENDER_MODE_LCD`, renders three horizontal
samples per output pixel and filters them into separate R, G, and B coverage
values. The compositor then blends

    out.rgb = fg.rgb * cov.rgb + dst.rgb * (1 - cov.rgb)

per channel. That is the whole mechanism, and each half is a separate
problem here.

1. Per-channel coverage. `golang.org/x/image/vector` produces one coverage
value per pixel. Ebiten cannot be asked for three. A rasterizer that emits
RGB coverage has to be new code: either a fork of the rasterizer in
`text/v2`, or a 3x horizontal supersample followed by an LCD filter, done in
ownframe. ownframe can reach the outlines and positions to do the
supersample: `text.AppendGlyphs` exposes `OriginX`, `AdvanceX`, and the
glyph list (`layout.go:302-332`), and `text.AppendVectorPath` appends outline
paths (`layout.go:348-356`). So the rasterizer half is buildable without
forking Ebiten.

2. Per-channel blending. This is the blocker. Standard source-over has one
alpha, so it cannot express three coverages. `ebiten.Blend` has per-channel
RGB factors (`blend.go:42`), but the target formula needs the destination
color at each pixel, and Kage cannot sample the destination framebuffer. A
shader can output per-channel values, but the blend equation combines source
and destination with factors, not with a subtraction of the destination per
channel times coverage. Concretely, the only single-pass forms available are
`src*F + dst*Fd` with factors built from source and destination colors, and
none of them computes `fg*cov + dst*(1-cov)` when `fg` is arbitrary. The one
workaround is to keep a copy of the text-free background as a second
texture, draw glyphs with a shader that samples that copy and writes with a
one/zero blend, then discard the copy. That works, and it costs a full-layer
copy per text batch and breaks the incremental repaint path, which repaints
single rects into a persistent buffer.

3. Gamma correction. Chrome applies a gamma and contrast adjustment to the
coverage before blending so light glyphs on dark backgrounds do not look
thin. That is a per-face table, computable in the coverage stage. It only
matters once per-channel blending exists.

4. Opaque backgrounds only. Chrome disables LCD for text under transforms,
opacity, or non-opaque layers, because the math above assumes an opaque
destination. A filtered or blended layer containing LCD text would need to
fall back to grayscale. This interacts with
[filter-replay.md](filter-replay.md), where a text layer under a filter is
exactly that case.

## Options

Option A, do nothing. Grayscale AA with quantized subpixel positioning is
what the current stack produces. Chrome's text mode on Linux depends on the
system font settings, and macOS dropped subpixel text after Mojave. At
normal sizes the visual gap between this stack and Chrome is mostly
positioning and unhinted outlines, not the LCD channel separation.

Option B, device-scale text. Rasterize glyphs at the device scale instead of
scaling the finished page buffer. Today `drawReplayScaled` and `blitFitted`
scale whole-page buffers (`replay_scale.go:36-47`, `view_fit.go:56-60`).
Replaying text with a face size of `op.Size * pxPerPt * scale` and scaling
only the geometry around it keeps glyphs sharp at zoom and stretch. This is
ownframe-only work, it needs no new shader, and it is what a user actually
notices when they resize a window. It also interacts with the text raster
cache work: cache entries must key on the device scale, not just the CSS
size.

Option C, own LCD rasterizer plus the two-texture blend. Build the 3x
supersample in `internal/replay`, pack RGB coverage into one RGBA image,
and composite through a copied background. Correct for opaque static layers,
roughly two extra full-layer copies per text region, broken under partial
repaint unless the copied background is also maintained per dirty rect.
Weeks of work with a fragile result.

Option D, fork the text renderer. Add an LCD glyph format to `text/v2`
(RGB coverage image plus a draw path that lets the graphics backend blend
per channel). The backend seam is the graphics command blend state, which
Ebiten keeps internal (`internal/graphicscommand/`, `internal/atlas/`), so
the fork would not just be `text/v2`; it would include the render command
path. Estimate 2 to 4 weeks plus a pinned Ebiten fork to rebase on every
upgrade. This is the only path to Chrome-equivalent LCD without the
two-texture trick.

## Recommendation

Ship grayscale, and spend the effort on Option B. Ebiten's public API cannot
express per-channel LCD blending; the rasterizer half can be replaced in
ownframe but the blend half cannot. The only in-repo route is the
background-copy workaround, which buys a text-quality feature at the cost of
the partial repaint architecture. That trade is not worth it in 0.0.2.

Concrete 0.0.2 items:

1. Device-scale text: rasterize faces at `size * pxPerPt * windowScale` and
   scale positions, replacing the scaled-buffer path for pages that are not
   oversized. Keep the buffer scale as the fallback for pages above
   `ebiten.MaxImageSize()`.
2. Keep the current text raster cache honest about scale: the face key in
   `text_cache.go:15-20` has source, size, and language. A device-scale mode
   gives it a fourth dimension. Decide with the text cache work whether the
   scale multiplies into `size` or becomes its own field, and make sure a
   zoom change does not silently reuse the old face.
3. Record the LCD blocker in `documentation/` next to the other replay
   limits, with the exact seam: `text.Face` is closed (`text.go:33-56`),
   glyph images are grayscale (`text.go:146-153`), and Kage cannot sample
   the destination. Anyone who wants LCD later should start from that list.

If text quality becomes a headline later, the cheapest correct path is
Option D scoped to `text/v2` plus a per-glyph blend path, and it should
start as an upstream Ebiten proposal rather than a permanent fork.

## What cannot be done with Ebiten's public API

- No RGB coverage images. Every glyph image is grayscale except color
  glyphs.
- No per-channel coverage blend. The destination cannot be sampled, and
  source-over has one alpha.
- No face plug-in. `Face` has a private method, so a custom face cannot be
  passed to `text.Draw`.
- No gamma or hinting options. Neither exists in `DrawOptions`.
- Subpixel positioning granularity is an internal decision, 8 variations at
  small sizes down to 1 at large ones. Changing it for large display text
  means a `text/v2` change, not an ownframe one.
