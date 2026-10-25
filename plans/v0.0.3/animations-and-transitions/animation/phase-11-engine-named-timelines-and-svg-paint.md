# Phase 11 - engine named timelines and SVG paint

> **Status:** planning
> **Track:** [animations and transitions](../README.md)
> **Upstream:** `github.com/chinmay-sawant/blinkless`
> **Runs in parallel:** does not block phases 02 to 05
> **Gates:** phase 06 (named timelines), phase 10 (SVG paint)

This phase is one number because both items are engine-side and can land in
parallel with the host work. They are independent, so they are two parts with
separate checklists and gates. Do not couple their release.

## Why an engine phase

Neither item is an export of a value the engine already computes:

- The named timeline declarations do not exist in the engine at all.
- Inline SVG is serialized and rasterized into an image at layout, so a paint
  change needs a new serialization and a new raster.

Priority 0 remains the transform critical path. Phase 11 is required for full
coverage but does not block the automatic driver, the property model, or the
layer work.

---

# Part A - named timelines

## Current state

`documentation/compatibility-matrix.md` records the declaring properties as
unimplemented:

- `scroll-timeline`, `scroll-timeline-name`, `scroll-timeline-axis` at `:784-786`
- `view-timeline`, `view-timeline-name`, `view-timeline-axis`,
  `view-timeline-inset` at `:828-831`
- `timeline-scope` at `:1133`
- `:1131` adds that this engine has no scroll viewport, scroll snapping, or
  scroll timeline driver.

`layout.Animation.Timeline` already accepts `auto`, `none`, `scroll()`, `view()`,
and a dashed ident. `internal/animation/timeline.go:11-25` resolves `scroll()` and
`view()`; `:27-68` resolves named ranges and edges. `FrameInput`
(`internal/animation/timing.go:62-70`) already has a named-timeline map. The host
already knows the page scroll offset and viewport height. Only the declarations
and the definition export are missing.

## Scope

In scope:

- Named scroll timelines on the page scroller.
- Named view timelines for an element in the page scroller.
- `scroll-timeline-axis` (`block`, `inline`, `x`, `y`) and
  `view-timeline-axis`.
- `view-timeline-inset` start and end.
- Nearest ancestor wins when names repeat.
- `animation-range*` resolved against the named timeline through the existing
  range code.
- A missing or inactive name leaves the animation inactive. No wall-clock
  fallback.

Out of scope:

- Nested element scroll containers. The engine has no scroll viewport, so a name
  on a non-scrolling element is inactive.
- `timeline-scope`. It needs a timeline reference outside the ancestor chain.
- `animation-timeline: scroll()` and `view()`. They already work and do not
  change.

## Design

1. Parse and cascade the timeline declarations. Add to `ResolvedStyle`:

   ```go
   ScrollTimelineName string
   ScrollTimelineAxis string // block | inline | x | y
   ViewTimelineName   string
   ViewTimelineAxis   string
   ViewTimelineInsetStart, ViewTimelineInsetEnd float64
   ViewTimelineInsetStartSet, ViewTimelineInsetEndSet bool
   ```

   `scroll-timeline` and `view-timeline` are shorthands; the longhands are the
   source of truth. Neither name property is inherited.

2. Export a timeline definition on the box:

   ```go
   type TimelineDef struct {
       Name       string
       Kind       string // scroll | view
       Axis       string // block | inline | x | y
       SubjectID  string // element the view timeline observes
       ScrollerID string // "" means the page scroller
       InsetStart, InsetEnd float64
   }
   ```

   Add `Box.Timelines []TimelineDef`. A box can declare both kinds. The subject
   is the declaring element for a view timeline. The scroller is the page
   scroller in this phase.

3. Resolve on the host. The engine owns no clock and no scroll position, so it
   must not compute progress. OwnFrame resolves `animation-timeline: --name`
   against the nearest ancestor's `Timelines`, then computes:
   - scroll: `offset / (content - viewport)`, clamped, on the declared axis;
   - view: the subject's entry and exit position in the scrollport, adjusted by
     the insets.
   Feed the result through the existing `FrameInput` named-timeline map.

4. Inactive rules. If the name does not resolve, the element does not scroll, or
   the range is zero, the animation is inactive. Do not fall back to `auto`.
   This replaces the current phase 06 placeholder, which falls back to time.

5. Duplicate names. The nearest ancestor wins. Two declarations on one element
   with the same name are a parse error for the second; count it.

6. Axis. `block` and `inline` follow the writing mode. Reuse the engine's logical
   property mapping; do not assume `y` and `x`.

7. Ranges. `animation-range` already parses edges and percentages. Feed it the
   named timeline's progress and the subject's size. Do not write a second parser.

8. `timeline-scope` stays unimplemented. A name that would resolve only through a
   scope ancestor is inactive and counted.

## Checklist A

### Parse and cascade

- [ ] `scroll-timeline`, `-name`, `-axis` with the dashed-ident grammar.
- [ ] `view-timeline`, `-name`, `-axis`, `-inset`.
- [ ] Longhands override the shorthand in declaration order.
- [ ] Names are not inherited.
- [ ] Logical axis resolves through the writing mode.
- [ ] Invalid names and axes are rejected with a count.

### Export

- [ ] `TimelineDef` and `Box.Timelines`.
- [ ] Subject and scroller ids are stable across a redraw.
- [ ] A box with no declarations carries a nil slice.
- [ ] No output change for a page without timelines.

### Host resolution

- [ ] Resolve a dashed ident to the nearest ancestor's definition.
- [ ] Page-scroller progress for a scroll timeline.
- [ ] View progress with start and end insets.
- [ ] Feed the existing `FrameInput` named-timeline map.
- [ ] An unresolved or inactive name keeps the animation inactive.
- [ ] Remove the phase 06 time fallback for a dashed ident.

### Tests

- [ ] Named scroll timeline at 0%, 50%, 100% matches Chrome.
- [ ] Named view timeline at entry, cover, and exit matches Chrome.
- [ ] `view-timeline-inset` shifts the range.
- [ ] Duplicate names: nearest ancestor wins.
- [ ] Missing name: inactive, no time motion.
- [ ] Zero-range scroller: inactive, no divide by zero.
- [ ] `scroll-timeline-axis: x` on a horizontal page.
- [ ] A nested scroll container is reported out of scope, not faked.

## Gate A

`animation-timeline: --name` runs from a declared timeline with no host input
beyond the page scroll state. A missing name is inactive with a counter. A page
with no timeline declarations produces the same display list as today.

---

# Part B - SVG paint

## Current state

`buildInlineSVG` (`internal/layout/layout_svg.go:20-60`) serializes the element
with `serializeInlineSVG`, rasterizes it through `svg.RasterizeContext`, and emits
one `OpImage` with PNG bytes. The op stores the raster, not the SVG. Child shapes
are serialized, not laid out as boxes. `fill`, `fill-opacity`, `stroke`,
`stroke-opacity`, `stroke-width`, `stroke-dasharray`, and `stroke-dashoffset` are
inherited paint properties baked into the serialized document
(`css_behavior_svg_paint_test.go`).

A paint change therefore needs a new serialization and a new raster. There is no
public API for that today.

## Scope

In scope:

- Animate the paint properties on the inline SVG root. They inherit into shapes
  that do not set their own paint.
- A per-part override by SVG element id, if part metadata can be exported.
- Re-serialize and re-rasterize one element per changed frame.
- Return an `ElementPatch` for the existing splice path.

Out of scope:

- Animating `d` (path geometry).
- Replacing the rasterizer with a vector command list.
- SVG filters as a separate animation target. A filter change rebuilds the whole
  image like a paint change.

## Design

1. Audit part boxes first. Today one box exists for the `<svg>` root. A keyframe
   on a child `<rect>` or `<path>` has no box, so the host cannot bind it.
   - Root-only first: a root `fill` or `stroke` inherits into the shapes.
   - Per-part second: export SVG part metadata (id, element, animation and
     transition records) and accept a part override.
   Do not assume child boxes exist.

2. API:

   ```go
   // layout/svg_paint.go
   type SVGPaintOverride struct {
       ElementID string // the <svg> box id
       PartID    string // optional SVG child id; "" means the root

       Fill             string
       FillOpacity      float64
       Stroke           string
       StrokeOpacity    float64
       StrokeWidth      float64
       StrokeDashArray  string
       StrokeDashOffset float64

       HasFill, HasFillOpacity     bool
       HasStroke, HasStrokeOpacity bool
       HasStrokeWidth              bool
       HasStrokeDashArray          bool
       HasStrokeDashOffset         bool
   }

   func RestyleInlineSVG(ctx context.Context, doc *css.Document, prev *Display,
       overrides []SVGPaintOverride, options Options) (*ElementPatch, error)
   ```

   A string form keeps the colour and dash parsers in the engine. Do not copy
   them into the host.

3. Semantics:
   - Find the SVG box by id in the retained tree.
   - Apply the override to the root or the named part for this pass only.
   - Re-serialize, rasterize, and replace the image payload.
   - Do not re-cascade CSS or re-place the element. The box geometry does not
     change unless the intrinsic size does, which a paint change does not.
   - Return an `ElementPatch` with the old and new image bounds.
   - Require `RetainState`. No overrides returns an empty patch.

4. Cost. Serialize, rasterize, and PNG encode per frame is the expensive part.
   Cap it: one restyle per element per frame, a per-page budget shared with the
   phase 04 queue, and a size cap. A large SVG snaps with a counter. Measure a
   24 px icon and a 512 px illustration before setting the cap.

5. Caching. Keep the previous raster keyed by the serialized string. If the
   string is unchanged, skip the raster. During an animation it changes every
   frame; the key helps at the endpoints and on pause.

6. Inheritance. A root override must enter the used style before serialization,
   so shapes without their own paint follow. Do not string-replace the serialized
   SVG; that misses gradients, inheritance, and escaped values.

7. Damage. Use the image op bounds. If the SVG overflows its box, use the painted
   union, not the element box.

8. Host integration. Phase 10 counts these properties as deferred until this
   part lands, then routes them here under the phase 04 budget. A missing engine
   hook means the property snaps and reports a reason.

## Checklist B

### Audit

- [ ] Confirm which boxes exist for an inline SVG and its children.
- [ ] Confirm how `fill` and `stroke` inheritance reaches each shape.
- [ ] Decide root-only versus per-part for the first release.
- [ ] Record the decision and the gap in `coverage.md`.

### API

- [ ] `SVGPaintOverride` and `RestyleInlineSVG`.
- [ ] A string form for every paint value, parsed by the engine.
- [ ] `ElementPatch` result with image bounds.
- [ ] `RetainState` requirement.
- [ ] Empty override returns the existing raster.

### Engine work

- [ ] Apply the override to the root's used style before serialization.
- [ ] Apply a part override by SVG element id.
- [ ] Re-serialize and rasterize through the existing functions.
- [ ] Replace the op image payload and bounds.
- [ ] Keep the previous raster when the serialized string is unchanged.
- [ ] Enforce the per-frame budget and size cap.
- [ ] No behavior change when the API is not called.

### Tests

- [ ] `fill` animation matches Chrome at 25%, 50%, and 75%.
- [ ] `stroke` and `stroke-width` match Chrome.
- [ ] `stroke-dashoffset` matches Chrome.
- [ ] Root inheritance: a child with no paint follows the root.
- [ ] A child with its own paint does not follow.
- [ ] Per-part override, if enabled.
- [ ] A large SVG hits the cap and reports a snap.
- [ ] No override: the display list is unchanged.

## Gate B

`fill` and `stroke` keyframes on an inline SVG animate at browser parity. A child
that inherits follows; one that sets its own paint does not. The image payload
rebuilds only when the paint changes. The cost is measured and capped.

## Known limits after the first implementation

- Inline-level SVG is retained now. The inline collector keeps the SVG box, the
  parent line box attributes the image op to it, and `boxForID` finds it. Block,
  inline, and per-part restyles all work.
- A raster cache keyed by the serialized string lives on the retained layout,
  bounded at 8 entries and 2 Mi pixels, oldest-first. It dies with the layout or
  its clone. A raster larger than the pixel budget is never stored, so it
  restyles and misses every time.
- The per-frame budget and the large-SVG size cap are still host-side. Phase 10
  must cap restyle calls until the engine adds a budget.
- `fill` and `stroke` accept CSS colours and `none`. A paint server such as
  `url(#gradient)` is rejected because `ResolvedStyle` stores paint as RGB.
  `currentcolor` resolves against the root element's used colour for inherited
  descendants, not each child's colour.
- The normal, non-restyled serializer still emits `fill-opacity="0"` for a
  cascade `fill:none`. The restyle path tracks no-fill separately. This is
  pre-existing behaviour and is left alone.
- A CSS transform on an inline `<svg>` now reaches its op because the icon has a
  box in the stamp walk. Before, the transform was ignored. This is closer to
  Chrome and is not a restyle-path change.

Checklist follow-ups:

- [x] Retain the inline SVG box and attribute its image op so inline icons work.
- [x] Cache the raster by the serialized string.
- [ ] Add the per-frame budget and size cap with a snap counter.
- [ ] Accept paint servers or reject them with a named error.
- [ ] Resolve `currentcolor` per element.

---

# Release and integration

- [ ] Merge both parts upstream, tag the next blinkless version.
- [ ] Bump `go.mod` and `examples/go.mod` to that tag. No replace.
- [ ] Run `make build` on both modules.
- [ ] Record the new pin in `AGENTS.md`.
- [ ] Update `compatibility-matrix.md` rows 784-786, 828-831, 300-301, and
      531-539 from ignored to implemented, with the limits noted.
- [ ] Update `coverage.md`: named timelines move to phase 06, SVG paint to
      phase 10.
- [ ] Update phase 06 and phase 10 dependencies and checklists.
- [ ] Add the named timeline and SVG paint cards to the animations example.

## Risks

- The engine has no nested scroll viewport, so a named inner scroller cannot
  work. Report it; do not pretend the page scroller is the inner one.
- `timeline-scope` is the escape hatch for sibling timelines. Without it, a
  reference outside the ancestor chain is inactive. Document it.
- Rasterizing a PNG per frame is the SVG cost limit. If a 512 px SVG cannot hold
  60 fps, cap it and document the snap.
- Root inheritance is not per-shape animation. Do not claim the latter until part
  metadata exists.
- A string replace on the serialized SVG is not an implementation. Apply the
  override in the used style before serialization.
