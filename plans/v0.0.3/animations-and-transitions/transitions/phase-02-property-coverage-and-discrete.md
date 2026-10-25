# Phase 02 - property coverage and discrete behaviour

Goal: a transition runs for every property the compatibility matrix accepts,
including the discrete ones, and snaps only when CSS says it snaps.

## Where it lives

- `internal/csstransition/bind.go:12-52` selects a transition record and
  connects paints.
- `internal/csstransition/discrete.go:3-7` states the current limit: only
  `position` changes at the halfway point; visibility and display need
  operations to appear or disappear, which the renderer does not handle.
- `internal/transition/step.go:10-44` interpolates one paint field set.
- `internal/cssmotion/paint.go:5-11` carries X, Y, W, H, RGB, Alpha, and opacity.
  It has no matrix, so a transform transition needs animation phase 02's model.
- `documentation/compatibility-matrix.md:585-590` lists the transition
  longhands; `:120-224`, `:303`, `:435-442`, `:550-571`, `:599` list the
  properties a transition can target.

## Property classes

| Class | Properties | Mechanism | Depends on |
|---|---|---|---|
| Paint, continuous | `color`, `background-color`, `border-*-color`, `outline-color`, `text-decoration-color`, `caret-color` | per-op RGB | animation 03 |
| Paint, continuous | `opacity` | per-op alpha, group alpha when a subtree overlaps | animation 03, 08 |
| Matrix | `transform`, `translate`, `rotate`, `scale` | `SetXform` | animation 02 |
| Layer | `filter`, `clip-path`, `mix-blend-mode` | layer + shader | animation 08 |
| Layout | `width`, `height`, `min/max`, `inset`, `margin`, `padding`, `border-width`, `border-radius`, `font-size`, `line-height`, `letter-spacing`, `gap`, `flex-basis`, `flex-grow`, `flex-shrink`, `grid-template-*`, `aspect-ratio`, `object-position`, `vertical-align`, `text-indent`, `text-decoration-thickness`, `text-underline-offset`, `column-width`, `shape-*`, `contain-intrinsic-*` | relayout override | animation 04 |
| Baked paint | `box-shadow`, `text-shadow`, `background-position`, `background-size`, `outline-width`, `outline-offset`, `border-image-*`, gradient stops | paint rebuild | animation 04 |
| Discrete | `display`, `visibility`, `content-visibility`, `position`, `z-index`, `order`, `grid-column-start/end`, `grid-row-start/end`, `line-clamp`, `column-count`, `font-weight` | presence, integer, or mode swap at 50% | this phase, layer phase 04 |
| Discrete, no interpolation | `background-repeat`, `background-clip`, `background-origin`, `background-blend-mode`, `border-style`, `border-image-source`, `border-image-repeat`, `object-fit`, `mask-repeat`, `mask-origin`, `mask-clip`, `mask-composite`, `mask-mode`, `text-decoration-style`, `text-emphasis-style` | switch at 50% | this phase |

A property not in this table is not claimed. Snap it, count it, and add a matrix
row only when it is implemented.

## Design

1. `all` expands to the changed properties. `transition-property: all` is the
   common declaration, so the expansion must happen at capture, not at bind. A
   property that did not change produces no renderer and no frame request.

2. Transform transitions reuse the animation matrix model. Capture the old
   matrix and the new matrix at commit, interpolate through the same
   decomposition used for keyframes, and write `SetXform`. Do not interpolate
   X/Y for a transform; that is the bug animation phase 02 removes.

3. Geometry transitions re-run the relayout override from animation phase 04 for
   the element, with the interpolated length values. The element reflows as it
   moves, which is what a browser does. This is the expensive class; keep the
   per-frame budget and the queue from phase 04, and give transitions priority
   over animations on the same element.

4. Discrete rules:
   - `display: none` to a visible value: the element appears at 0% and
     transitions normally.
   - `display: visible` to `none`: keep the element visible for the whole
     transition and remove it at 100%. To do that, capture the outgoing element
     into a layer at commit and composite that layer until the end. Without the
     layer, the new display has already dropped its ops.
   - `visibility: hidden` to visible: visible at 0%. Visible to hidden: stay
     visible and flip at 100%. This is the current `position` rule applied to a
     presence property.
   - `content-visibility` follows `visibility`.
   - `z-index` and `order` are integers: flip at 50% and damage the affected
     stacking area.
   - `grid-column-start/end` and `grid-row-start/end` are integers: flip at 50%,
     through the relayout path because the item moves.
   - `line-clamp` and `column-count` are integers: flip at 50%, through layout.
   - `font-weight` selects a face in this engine, so it is discrete; flip at 50%
     and accept the text reflow.
   - The no-interpolation properties in the table switch at 50% and damage the
     element. They never produce a per-frame value.
   - `position` keeps the existing halfway swap.

5. `transition-behavior: allow-discrete` is what makes a discrete property
   transition at all. Without it, the value snaps at the end of the delay. Keep
   the existing gate and extend it to the presence properties.

6. `@starting-style` is not implemented by the engine and no matrix row claims
   it. Keep it out of the plan. A page that needs an entry animation declares a
   keyframe animation instead, and phase 01 of the animation track covers it.

7. Shadow and gradient transitions need a paint rebuild, not a scalar write. Route
   them through the relayout override with the interpolated declaration. A
   gradient with a different stop count or function shape is discrete; snap it.

8. Cancel rules. A property removed from the new style snaps to the new value and
   cancels. A property whose duration is 0 snaps at the commit. A property with a
   delay keeps the old value until the delay ends; the old display already holds
   it, so no extra capture is needed.

9. Damage. A paint transition damages the op box. A geometry transition uses
   `ElementPatch.Old/New`. A disappearing element damages its layer bounds until
   the layer is released.

## Checklist

### A. Expansion and selection

- [ ] Expand `all` to the changed property set at capture.
- [ ] Match longhands and shorthand records to the same property.
- [ ] No renderer for an unchanged property.
- [ ] `transition-property: none` binds nothing.
- [ ] Comma lists keep declaration order.

### B. Paint and matrix

- [ ] Colour transitions through the phase 03 target selection.
- [ ] Opacity transition per op; group alpha when a layer is promoted.
- [ ] Transform transition through the phase 02 decomposition.
- [ ] A transform transition does not write X or Y.
- [ ] Test an interrupt on a transform at 30%.

### C. Layout

- [ ] Width, height, and inset through the relayout override.
- [ ] Margin, padding, border width, radius.
- [ ] Font size, line height, letter spacing.
- [ ] Box-shadow and text-shadow with a matching layer count.
- [ ] Gradient stops with a matching shape.
- [ ] Budget and queue shared with animation phase 04.

### D. Discrete

- [ ] `display` in and out with a retained outgoing layer.
- [ ] `visibility` and `content-visibility` flip at the correct end.
- [ ] `overlay` is not implemented by the engine and is not claimed.
- [ ] `z-index` and `order` flip at 50%.
- [ ] Grid line integers flip at 50% through relayout.
- [ ] `line-clamp`, `column-count`, and `font-weight` flip at 50%.
- [ ] The no-interpolation mode properties switch at 50%.
- [ ] `allow-discrete` gate preserved.
- [ ] `position` keeps the existing behaviour.
- [ ] A snap still damages the element and paints the final value.

### E. Examples

- [ ] Add cards for colour, opacity, transform, width, shadow, and display.
- [ ] Each card starts from a class toggle, not a Go call.
- [ ] Compare with a browser at 0%, 50%, and 100%.

Gate: every property class has a card, the display exit keeps its pixels until
the end, and a geometry transition reflows without a page redraw.

## Risks

- Display exits are the hardest case in this track. A layer is the only correct
  way to keep pixels that the new display no longer contains. Do not fake it by
  leaving the old display mounted.
- Geometry transitions can trigger a relayout cascade. Measure the reflow cost,
  not just the sample cost.
- A shadow transition with a different layer count is discrete. Snapping is
  correct; interpolating the first layer only is not.
