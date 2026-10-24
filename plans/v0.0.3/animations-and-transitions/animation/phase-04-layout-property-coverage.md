# Phase 04 - layout property coverage

Goal: keyframes that change geometry animate, through the engine's incremental
relayout path, without a full-page redraw and without restarting other
animations.

## Where it lives

- `layout.RelayoutElement(ctx, doc, prev, elementID, options)` re-styles one
  subtree, places it, and splices its ops over the range the element owned. It
  returns `ElementPatch{Ops, OpOwners, Boxes, Old, New, Display}`. It requires the
  previous display to carry retained state, otherwise `ErrNoRetainedState`.
- `layout.DisplayListOptions.RetainState` is surfaced as `ownframe.Config.RetainState`
  (`internal/page/page_config.go:24-27`) and defaults to off.
- `internal/page/patch_wire.go:127-136` is the existing adoption path for a
  spliced display: swap the list, rebuild boxes and hits, diff damage, bump the
  generation.
- `documentation/optimizations.md:415-430` records that a click no longer
  re-lays out every card because `RelayoutElement` splices one subtree.
- `blinkless@v0.0.3` exports `html.Document.SetText` and `SetAttribute`, and both
  land on the parsed tree before the next cascade.

## What Chrome does

Layout properties are not compositable. Blink animates them on the main thread,
re-running style and layout for the affected subtree each frame, and paints the
result. It is slower than a transform animation and still correct. The compositor
never sees `width` or `top`.

## The missing engine hook

There is no way today to say "lay this element out as if `width` were 123px"
without mutating the authored tree. `SetAttribute("style", ...)` would work, but
it clobbers the page's own inline style and needs a restore path. That is a guess
dressed as an implementation, and it will corrupt pages that already use inline
styles.

The clean seam is an override list on the relayout call:

```go
type Override struct {
    ElementID string
    Property  string
    Value     string
}

func RelayoutElementWithOverrides(ctx, doc, prev, elementID string,
    overrides []Override, options Options) (*ElementPatch, error)
```

Overrides apply after the cascade and before placement, are not written to the
tree, and disappear when the call returns. This is a small upstream change in the
cascade inputs, not a layout rewrite.

Blocked row: this phase cannot finish geometry animation until that override
exists in the blinkless pin. Until then, geometry keyframes are counted as
deferred, left at their base value, and listed in the compatibility matrix. Do
not ship the inline-style workaround.

## Engine status

`RelayoutElementWithOverrides` is implemented on the blinkless feature branch,
on top of the pushed `72e595d` commit. The pin still points at `72e595d`; bump it
when the host wiring lands.

```go
type Override struct {
    ElementID string
    Property  string
    Value     string
}
func RelayoutElementWithOverrides(ctx, doc, prev, elementID string,
    overrides []Override, options Options) (*ElementPatch, error)
```

The pass targets the containing block, not only the element, so siblings reflow.
It clones the element style and custom properties, applies the override, resolves
descendants with the clone as the inherited parent, builds the containing block,
then restores every base style pointer. `Old` and `New` are the element's border
box. Sibling operations are re-emitted and get new op IDs, so the phase 01 rebind
must keep their timelines. Tile damage should come from `dirtyFromDisplay`, not
`Old`/`New` alone, because a moved sibling is outside those two rectangles.

Not overridable this way: transforms and perspective, SVG paint, paint-only
properties, flow and box-kind properties, many shorthands, table cells and rows,
sticky, float, anchored, named-flow, and viewport-fixed elements. A reflow that
resizes the containing block returns `ErrOverrideReflow`; the caller falls back to
a full display build for that frame.

## Design

1. Retain state on demand. At first bind, scan `display.Keyframes` for a property
   in the layout tier, or for `perspective`, `perspective-origin`, or a 3D
   transform function. If one exists and `RetainState` is off, turn it on and
   schedule one full redraw before the first sample. A page with no geometry, 3D,
   or perspective motion never pays the memory. Record the switch in the probe
   counters.

2. One relayout per element per frame, not one per property. Group the keyframed
   layout properties of an element into one override list and one
   `RelayoutElementWithOverrides` call. A card animating `width` and `height`
   relayouts once.

3. Adopt through the existing splice. Reuse the body of `spliceElement`: swap the
   list, rebuild boxes and hits, diff damage with `dirtyFromDisplay(prev, d)`,
   bump the generation. Do not write a second adoption path. The difference from
   a click patch is that the driver owns the edit list.

4. Preserve every other player. A splice bumps the generation, which triggers the
   phase 01 rebind. The runtime must carry each entry's original start time and
   re-resolve its op indices against the new display. An element that relayouts
   every frame must not restart its own animation, and must not restart the
   neighbours'.

5. Order within a frame: sample layout overrides first, relayout, adopt, then
   sample paint and transform against the new display. A transform animation on a
   child of a relaid-out parent uses the new box. Never sample a transform
   against a display that is about to be replaced.

6. Damage comes from `ElementPatch.Old` and `ElementPatch.New`. Do not recompute
   the element box from the keyframe value; the engine's patch is authoritative,
   and it accounts for text reflow and min/max clamping.

7. Budget. A relayout is tens to hundreds of microseconds, and a page can declare
   many geometry animations. Add a per-frame budget in the runtime, default one
   relayout per frame per page, and a queue ordered by element id for the rest.
   A queued element samples at the frame it is relaid out, so its time stays
   correct. Expose the queue depth and the budget hit in `Probe()`. Do not drop a
   relayout silently.

   Perspective and 3D transform restamps share this budget. `RestampTransforms`
   is cheaper than a full relayout but not free; it walks the subtree and
   re-stamps every op. Give each 3D or perspective card its own queue entry and
   the same budget hit counter.

8. Interaction with the tile grid. A geometry change damages a real page rect;
   the window repaints the covered tiles exactly as it does for a click. No new
   window path is needed.

9. Properties:
   - width, height, inline-size, block-size, min/max in physical and logical
     forms
   - top, right, bottom, left, inset and logical inset longhands
   - margin, padding, gap, flex-basis, flex-grow, flex-shrink
   - grid-template-columns, grid-template-rows, grid-auto-columns,
     grid-auto-rows when the track lists have the same shape
   - aspect-ratio
   - border-width longhands, border-radius longhands, border-spacing
   - border-image-slice, border-image-width, border-image-outset
   - outline-width, outline-offset
   - column-width, column-height, column-rule-width, row-rule-width, rule-width
   - column-rule-inset, row-rule-inset, rule-inset and their longhands
   - font-size, font-size-adjust, line-height, letter-spacing, word-spacing
   - text-indent, tab-size, vertical-align
   - text-decoration-thickness, text-decoration-inset, text-underline-offset
   - object-position, object-view-box
   - shape-outside, shape-margin, shape-padding, shape-inside, float-offset
   - contain-intrinsic-size and its logical and axis longhands
   - overflow-clip-margin and its longhands, clip
   - box-shadow and text-shadow, which are baked into shadow ops at layout
   - background-position and background-size, which move or scale an image op
   - background-image gradients, only where the stop list has the same shape
   - perspective and perspective-origin, through `RestampTransforms` under this
     phase's budget
   - 3D transform functions (`rotateX`, `rotateY`, `rotate3d`, `translateZ`,
     `scaleZ`, `matrix3d`, `perspective()`), through the same restamp path
   - fill and stroke, only through [phase 11](phase-11-engine-named-timelines-and-svg-paint.md)'s
     `RestyleInlineSVG` and phase 10's integration

   A property whose value form cannot interpolate (a keyword, a different
   gradient shape, a different shadow count) is discrete, not continuous, and
   belongs to the transition track's discrete rules.

## Checklist

### A. Engine

- [ ] Blinkless adds `RelayoutElementWithOverrides` with an override list applied
      after the cascade and before placement.
- [ ] Overrides do not mutate the parsed tree and do not leak into the cache.
- [ ] Percentages and relative units resolve against the element's containing
      block as usual.
- [ ] OwnFrame bumps the pin and adopts the API.

### B. Runtime

- [ ] Detect layout-tier keyframes at bind and enable `RetainState` on demand.
- [ ] Group overrides per element and call relayout once per element per frame.
- [ ] Adopt through the `spliceElement` body, not a copy.
- [ ] Preserve start times across the generation bump.
- [ ] Sample layout before paint and matrix in the same frame.
- [ ] Use `ElementPatch.Old/New` for damage.
- [ ] Add the one-relayout-per-frame budget and the queue.

### C. Properties

- [ ] Size, min/max, and their logical forms.
- [ ] Inset and logical insets.
- [ ] Margin, padding, gap, flex-basis, flex-grow, flex-shrink.
- [ ] Grid templates and implicit tracks with a matching shape.
- [ ] Aspect ratio.
- [ ] Border width, radius, spacing, and border-image lengths.
- [ ] Outline width and offset.
- [ ] Column and rule widths and insets.
- [ ] Font size, size adjust, line height, letter and word spacing.
- [ ] Text indent, tab size, vertical align.
- [ ] Text decoration thickness, inset, and underline offset.
- [ ] Object position and object view box.
- [ ] Shape outside, margin, padding, inside, and float offset.
- [ ] Contain intrinsic sizes.
- [ ] Overflow clip margin and legacy clip.
- [ ] Box-shadow and text-shadow: same count, interpolable fields.
- [ ] Background position and size.
- [ ] Gradient stop interpolation with a matching function shape.
- [ ] Perspective, 3D transforms, and perspective-origin through
      `RestampTransforms`.
- [ ] SVG fill and stroke, through phase 10 only.

### D. Tests

- [ ] Width animation changes the painted box and reflows a sibling.
- [ ] A neighbour transform animation does not restart when the width element
      relayouts.
- [ ] A queued relayout samples at the frame it runs, not at the queue time.
- [ ] Retain state turns on only when a geometry keyframe exists.
- [ ] A 20-element geometry page stays under the frame budget.

### E. Example

- [ ] Add width, height, top/left, border-radius, font-size, and box-shadow cards.
- [ ] Measure each card against a browser at three timestamps.
- [ ] Record the cost per card in `documentation/performance-baseline.md`.

Gate: geometry cards animate without a full-page redraw, the neighbour-restart
test passes, and the budget is reported when hit.

## Risks

- Per-frame relayout is the most expensive thing in this track. If the budget is
  hit on a realistic page, lower the sample rate for layout animations before
  raising the budget.
- `RetainState` increases memory for the whole page. Turning it on for one card
  costs the page, so measure before recommending it as a default.
- A relayout that changes op count invalidates every target index. The generation
  guard is mandatory here, not an optimisation.
