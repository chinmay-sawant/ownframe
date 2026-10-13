# Filter replay

Recorded 2026-10-13 against the tree after commit `679b7b9`, engine pin
`github.com/chinmay-sawant/blinkless v0.0.0-20261010174808-34d93862f9ed`.

CSS `filter` today reaches the replay painter in three different states
depending on the element and the function, and none of them is "a filter on
the op". This plan records what exists, including the functions that are
parsed and then dropped, and designs a layer-and-shader replay for the
functions the engine implements. It needs an engine change first, because the
display list carries no filter payload.

## What happens today

The engine parses the full CSS filter list:

- `filter.go:22-34` names every function: blur, opacity, drop-shadow,
  grayscale, brightness, contrast, invert, sepia, saturate, hue-rotate.
- `filter.go:42-106` parses them into `parsedFilter{kind, val, dropShadow}`.
- `filter.go:124-151` builds the gaussian kernel: `sigma = max(radius*0.57735, 0.5)`,
  kernel radius `min(25, ceil(3*sigma))`.

What the engine then does with a parsed list depends on the content:

1. Image content is baked. Background image layers
   (`internal/layout/background_image.go:120-124`), gradient background
   layers (`:174-178`), inline images
   (`internal/layout/inline_image.go:78-82`), and `<img>` elements
   (`internal/layout/layout_images.go:484-487`) decode the image bytes, run
   the filter list over the pixels, re-encode a PNG, and store those bytes on
   the same op. `filter.go:357-419` is the apply function, and
   `filter.go:414-418` is the PNG re-encode. The painter never sees a filter
   for images. It sees different pixels.
2. `filter: opacity(x)` folds into the element opacity. The value multiplies
   `ResolvedStyle.Opacity` at `internal/layout/style_properties.go:202-207`,
   and the op carries the product through `Op.Opacity()`
   (`internal/layout/op_read.go:59-68`, used by the replay in
   `internal/replay/color.go:13-16`).
3. Everything else is parsed and dropped. The apply switch in
   `filter.go:393-412` has arms only for blur, grayscale, invert, and
   opacity. brightness, contrast, sepia, saturate, and hue-rotate have no
   arm. drop-shadow is stored on `parsedFilter.dropShadow`
   (`filter.go:74-77`) and read nowhere outside `filter.go`. So on an image
   those five are silent no-ops today, and drop-shadow is a no-op everywhere.

For non-image content, an element filter has no paint effect at all. A div
with `filter: blur(4px)` over text emits the same ops as the same div without
the filter. There is no op field for a filter, no accessor, and no group
marker. `layout/displaylist.go:37-49` lists every rare payload accessor the
replay can call and filter is not one of them. Filter does not even create a
stacking context, unlike transform, opacity, blend, and isolation
(`internal/layout/layout.go:1011-1020`), and the blend group mechanism that
`mix-blend-mode` and `isolation: isolate` use
(`internal/layout/blend_group.go:13-35`,
`internal/layout/layout.go:1026-1058`) never opens for a filter.

How this reaches the ownframe painter:

- Replayable pages. `internal/render/replayable.go:26-62` tests each op for
  blend groups, transforms, and kinds. It has no filter test because there is
  nothing to test. Image ops pass `op.ImageBytes() != nil`
  (`replayable.go:49-52`) and are drawn from the baked bytes by
  `internal/replay/image.go:15-33`. A filtered image therefore replays
  correctly today, at the cost of a CPU gaussian and a PNG encode during every
  layout.
- Fallback pages. A non-replayable page paints through
  `internal/page/page_draw_fallback.go:14-39` and `internal/bitmap`, which
  consumes the same baked image bytes. A filter never causes a fallback,
  because no filter ever reaches an op.
- So today a filtered `div` full of text is simply unfiltered, on both paths.
  A filtered image is baked, on both paths.

That is the honest baseline: half of CSS filter is implemented, only for
image payloads, and the display list cannot express the rest.

## Design

Two changes, one in the engine and one here.

### Engine side: emit a filter group

The display list needs three things it does not have:

1. A `FilterList` payload on the op, with the raw values, not baked pixels.
   The engine already parses them. Expose them through an accessor the way
   every other rare payload is exposed (`op_read.go:19-104`,
   `layout/displaylist.go:37-49`), for example a `Filter` accessor returning
   `[]Filter{Kind, Value, X, Y, Blur, Color}`. A `string` kind keeps the
   boundary small, but an enum is cheaper and matches the internal
   `filterKind`.
2. Group markers around the filtered element's ops, reusing the
   `BlendGroup` shape: a begin marker, the element's ops, an end marker,
   with the filter list attached to the group. The marker plumbing already
   exists (`layout.go:1060-1104`, `blend_group.go:37-94`), so the cheapest
   engine change is to let a group carry a filter as well as a blend mode.
   The replay then knows which op range to rasterize into one layer.
3. Filter must create a stacking context and a containing block, the way
   `enterStackingContext` handles transform and opacity
   (`layout.go:1011-1020`). Chrome and the CSS spec require both.

The engine should keep baking image filters for now. Removing the bake is a
separate memory and CPU win, but it changes the bytes on every image op and
makes the fallback painter wrong until it learns filters too. Phase 1 keeps
the bake and applies the new group path only to non-image content.

Cost of keeping the bake: every filtered image is decoded, blurred, and
PNG-encoded again at layout time. A 1000px wide background with `blur(8px)`
is tens of milliseconds of pure CPU per layout, on the UI goroutine, and the
op is untracked by `render.Replayable` because it still looks like an image.
The plan records this as a known cost, not a plan item.

### ownframe side: cached layers plus Kage passes

New package `internal/filter` with four pieces.

Layer. One offscreen `*ebiten.Image` plus a page-space rectangle. Build it
by replaying the group's ops, and only the group's ops, into the image. The
existing replay entry points are op-order walks (`internal/replay/replay.go:15-27`),
so a layer build wants the `DrawRect`-style variant that agent b and agent c
are adding, filtered to the group's op range. Layer pixels are page points
scaled by `pxPerPt` (`internal/replay/color.go:11`) at the window's device
scale, so 1 CSS px of layout becomes 1 device px of layer. Bounds are the
union of the group ops' paint bounds inflated by the blur radius, in layer
pixels: `ceil(3*sigma)` for blur, plus offset and radius for drop-shadow.

Cap. One layer must fit `ebiten.MaxImageSize()` (`image.go:1494`). The
window already has an `oversized` helper for this
(`internal/window/replay_oversize.go:14-15`,
`internal/window/replay_viewport_paint.go:16`). A filtered group wider or
taller than the cap has two honest choices: tile the layer into
`MaxImageSize` squares and run the passes per tile with one-pixel overlap,
or decline the page back to the bitmap painter. Phase 1 declines, phase 3
tiles.

Passes. Each filter function is one full-layer draw with a Kage shader
(`ebiten.NewShader`, `shader.go:45-52`, drawn with
`(*Image).DrawRectShader`, `image.go:1039`). The chain runs left to right,
and each pass reads the previous result. Ping-pong between two layer images
of the same size. Cached final image per group, not the raw layer: for a
static page the whole chain runs once and every frame then draws one image.

Color matrices. grayscale, brightness, contrast, invert, sepia, saturate,
and hue-rotate are all affine transforms of RGB, so one shader covers them
with a 3x3 matrix plus a bias and a working-space flag. Build the matrix on
the CPU from the same formulas the engine uses (`filter.go:424-457` for
grayscale, `:462-496` for invert; brightness and contrast are scale and
scale-plus-bias; sepia, saturate, and hue-rotate are the standard CSS
matrices). The engine applies opacity at the op level, not in this shader,
because `filter.go:407-411` explicitly delegates it.

Premultiplied alpha. Ebiten images are premultiplied. A linear transform
commutes with premultiplication, but a bias does not: in straight space the
engine computes `out = M*c + b`, and in premultiplied space that is
`out = M*c' + b*a` with alpha unchanged. The shader can implement the second
form exactly, with no unpremultiply divide and no 8-bit loss. Do that.

Blur. Two passes, horizontal then vertical, with a constant loop bound of 25
taps and a `break` on the uniform radius. Kage loop bounds must be constants
(`skills/writing-kage-shaders/reference/recipes.md`, "A uniform-controlled
loop count"). Weights are `exp(-i*i/(2*sigma*sigma))` normalized, the same
as `buildGaussianKernel` (`filter.go:125-151`). Sampling clamps to the layer
edge, matching the engine's index clamp (`filter.go:194-200`), not
transparent black. Reproduce the engine's radius gate exactly: `r <= 0` or
`r > 20` is a no-op (`filter.go:172-175`). The two passes are separate draws
into an RGBA8 target, so the intermediate quantization matches the engine's
two uint8 rounding passes (`filter.go:209-214`, `:237-242`).

Linear light. When the style carries `color-interpolation-filters: linearRGB`
(`style.go:512`, `filter.go:153-163`), the shader decodes sRGB to linear,
runs the pass, and encodes back, matching `blurInLinearLight`
(`filter.go:253-325`) and `applyGrayscaleInSpace`/`applyInvertInSpace`. Phase 1
ships the sRGB path only and gates the linear flag behind the same code so
adding it is a shader edit.

Drop shadow. The most complex pass, because it changes the layer's bounds and
composites under the element. Take the current layer's alpha, blur it with
the same two-pass blur, tint it with the shadow color, draw it at the offset
into a padded result, then draw the current layer over it. This is also the
one filter the engine has never applied to anything, so there is no engine
output to match. Parity target is Chrome's rendering, and Chrome's exact
gaussian is not ours, so parity here means tolerance, not bytes.

Where it plugs in. `internal/window/draw.go:46-82` chooses between replay,
bitmap, and fitted paths. A filter group becomes one more draw: the cached
filtered layer at its page rect, through `drawReplayPartial` or the scaled
paths like any image. `internal/render/replayable.go` must accept the group:
today a leftover group marker makes `replayableOp` return false
(`replayable.go:31-33`), which is the correct conservative behavior for a
blend group but would send every filtered page to the bitmap path. Accept a
group when it carries a filter and its layer fits, and keep rejecting blend
groups.

Cache. Key on group ID, page generation, and device scale. A fixed op inside
a group sits at a viewport position that changes with scroll, so either keep
fixed ops out of a group's layer or add the scroll offset to the key. Store
an LRU capped in bytes, and dispose `ebiten.Image`s on eviction. Invalidate
on generation change the way the window already invalidates `bitmapView`
(`internal/window/bitmap_viewport.go:31-33`).
A `SetTick` animation inside a filtered element rebuilds the chain per
frame. That is expensive and should be documented, not hidden: the tick API
exists to avoid a relayout, and a filter chain is a full-layer GPU pass set,
not a relayout.

## Pixel parity strategy

The repository already compares replay pixels against the bitmap painter
(`internal/replay/gpu_parity_*_test.go`). Filters extend that.

1. Reference for color filters: run the same filter list through the engine's
   CPU code and through the shader chain on a generated test image, then
   compare. The engine operates on straight NRGBA. The shader operates on
   premultiplied RGBA8. Assert max channel delta 0 on fully opaque pixels and
   at most 1 on pixels with alpha.
2. Reference for blur. Same comparison, on a hard-edged fixture, with clamp
   at the edges. Expect delta at most 1 per channel. A radius exactly at the
   gate boundaries, 0, 1, and 20, gets its own test, plus 20.5 to pin the
   engine's no-op drop.
3. linearRGB runs the same comparisons against the engine's linear arms.
4. Non-goal: byte equality with Chrome. Chrome's blur kernel, edge mode, and
   filter region differ from this engine and from each other. The parity
   target is ownframe replay against ownframe fallback, so a page that
   changes paths does not change pixels.
5. The fallback painter must learn filters before `Replayable` starts
   returning true for filtered pages that still need the fallback for other
   reasons. Otherwise the same page differs by path. This is the main
   sequencing constraint: `internal/bitmap` needs the same filter chain, or
   the replay must decline group filters whenever anything else on the page
   forces the fallback. Phase 1 takes the second option because it is one
   check in `Replayable`.

## What Ebiten's public API cannot do

- No destination sampling in Kage. A filter that reads the backdrop under
  the element, `backdrop-filter`, cannot be implemented without drawing the
  backdrop into a separate texture first. Not planned.
- No per-channel coverage blending for text. A filter group containing text
  replays through the grayscale glyph images Ebiten produces
  (`internal/replay/text.go:15-50`). That is already what the replay draws,
  so filters do not make text worse. It does mean an LCD text mode, see
  [subpixel-text.md](subpixel-text.md), would need its own layer story.
- Constant loop bounds. The blur shader hardcodes 25 taps. A larger kernel
  needs a second shader or a compile-time variant, and the shader must still
  no-op above radius 20 to match the engine.
- Nearest sampling in shaders, unless bilinear is hand written. The blur
  samples integer texels, so this does not bite. A scaled pass would need the
  hand-rolled bilinear from the Kage recipes.
- `MaxImageSize()` varies by driver, 4096 on some Metal devices, 16384 on
  others. Layer tiling or the fallback is mandatory for tall pages.
- Offscreen layers are VRAM. A page with twenty filtered groups holds twenty
  layer pairs unless the cache evicts. The cache budget is not optional.

## Phases

1. Engine: filter payload and group markers, stacking context, tests. Sibling
   checkout, 3 to 5 days.
2. ownframe: `internal/filter` layer build, cache, color-matrix shader,
   `Replayable` acceptance, parity tests. 1.5 to 2 weeks.
3. Blur, including the two-pass quantization match and the radius gate. 3 to
   5 days.
4. Drop shadow, with Chrome-tolerance parity rather than byte parity. 2 to 3
   days.
5. Tiling above `MaxImageSize` and the scaled or high-DPI layer path. 3 to 5
   days.

About 3 to 4 weeks for one developer after the engine lands, with the usual
caveat that parity tests decide the schedule, not the shader count.

## Limits to write down when this ships

- `backdrop-filter` stays unimplemented.
- A filtered group attached to the bitmap path needs the fallback painter to
  learn the same chain, or the page stays on the vector path by construction.
- Animated filtered content rebuilds the chain per frame, so it costs GPU
  passes proportional to the group size, not the changed pixels.
