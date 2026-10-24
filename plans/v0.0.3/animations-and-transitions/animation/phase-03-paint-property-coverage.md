# Phase 03 - paint property coverage

Goal: every keyframed property that resolves to a paint field animates, on the
right operations, without touching text that should not change.

## Where it lives

- `internal/animation/animation.go:19-24` carries opacity, RGB, and the transform
  that phase 02 adds.
- `internal/animation/compose.go:11-16` applies all six fields to every target,
  with no kind test. `background-color` therefore recolours the label text of the
  same element. This is a visible bug in `examples/animations`.
- `internal/cssmotion/ownership.go:7` returns every op in `box.OperationIDs`.
  Blinkless documents that set as the element's whole subtree
  (`internal/layout/layout.go:15-21`), so a parent also owns its children's ops.
- `internal/replay/op.go:14-39` dispatches per kind: FillRect, StrokeRect, Line,
  Text, Image, GridRun, Path.
- `internal/masking/paint_color.go:27` folds `Alpha` into `Opacity()` for the
  painters. A write of `Alpha = 1` can change a translucent op.
- `internal/cssmotion/paint.go:19-23` rewrites W, H, and Alpha on every write,
  even when the keyframe never mentions them.

## Property tiers

The compatibility matrix lists more paint properties than a per-op write can
express. Split them by mechanism so each lands in the phase that can do it
correctly.

| Tier | Properties | Mechanism | Phase |
|---|---|---|---|
| Direct field | `background-color`, `color`, `border-*-color`, `outline-color`, `text-decoration-color`, `text-emphasis-color`, `column-rule-color`, `row-rule-color`, `rule-color`, `accent-color` | op RGB, nearest owner only | 03 |
| Direct field | `opacity` on a single op or non-overlapping content | paint opacity | 03 |
| Direct field | `transform`, `translate`, `rotate`, `scale` | op matrix | 02 |
| Layer | `opacity` on a subtree with overlapping children, `filter`, `clip-path`, `mask-position`, `mask-size` | group layer + shader | 08 |
| Re-serialize | `fill`, `fill-opacity`, `stroke`, `stroke-opacity`, `stroke-width`, `stroke-dasharray`, `stroke-dashoffset` | inline SVG is an image op; needs a new payload | 10, [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) |
| Layout | `width`, `height`, `inset`, `margin`, `padding`, `border-width`, `border-radius`, `font-size`, `line-height`, `letter-spacing`, `box-shadow`, `text-shadow`, `background-image` gradients, `background-position` | `RelayoutElement` or a paint rebuild | 04 |
| Discrete | `visibility`, `display`, `content-visibility`, `position`, `z-index`, `order`, `mix-blend-mode`, `isolation` | presence, mode, and integer rules | transition track, phase 02 |

Phase 03 must not fake a tier. A `filter` keyframe is counted as deferred and
left at its base value until phase 08, with the reason in the counter, not
dropped silently.

## Design

1. Selection is by nearest owner, not by subtree. Build the target list from
   `d.OpOwners`: an op belongs to the animated element when
   `OpOwners[i] == boxIndex`. Subtree membership is only correct for inherited
   or group properties. This one change fixes the label-recolour bug and keeps a
   parent's `background-color` off a child's card.
   - `transform` still targets the whole subtree, because a moved element moves
     its children.
   - `color` targets the nearest text op. Inheritance is a later refinement: a
     child that explicitly sets its own colour must not animate, and the engine
     does not expose that bit today. Record the gap.

2. Field mask. Each target records which fields its keyframes can write:

   ```go
   const (
       fieldX uint16 = 1 << iota
       fieldY
       fieldMatrix
       fieldOpacity
       fieldRGB
   )
   ```

   `compose` writes only masked fields. W, H, and Alpha are never written unless
   a property explicitly owns them. This fixes the translucent-op regression
   where `Alpha` was reset to 1 on the first sample.

3. Kind applicability:

   | Kind | RGB | Opacity | Matrix |
   |---|---|---|---|
   | FillRect | yes | yes | yes |
   | StrokeRect, Line | yes (stroke colour) | yes | yes |
   | Text | yes (ink colour) | yes | yes |
   | Image | no | yes | yes |
   | GridRun | no | yes | yes |
   | Path | no | yes | yes |

   A `background-color` keyframe never targets Text. A `color` keyframe never
   targets FillRect unless the element's own text is painted by a fill, which is
   not the case here.

4. Colour parsing. Replace the hex-only `values.go:21-34` with a parser that
   accepts `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`, `rgb()`, `rgba()`, `hsl()`,
   `hsla()`, the named colours in the CSS list, `transparent`, `currentColor`,
   and `inherit` (as an error for a keyframe). Keep the parsed value in sRGB
   floats to match the op fields. Interpolate in sRGB; document that `oklab` and
   `color()` interpolation is out of scope until the engine exposes it.

   The rule-colour properties (`column-rule-color`, `row-rule-color`,
   `rule-color`) and `accent-color` are implemented and paint real ops. Treat
   them as ordinary colour targets with the nearest-owner rule; do not leave
   them out because they are less common. `fill` and `stroke` are a different
   class: inline SVG serializes to an `OpImage`, so they cannot be written into
   an op RGB field. They belong to [phase 11](phase-11-engine-named-timelines-and-svg-paint.md)'s
   `RestyleInlineSVG` and phase 10's integration, and must be counted as deferred
   until both land.

5. Opacity. A single op fades exactly. A subtree with overlapping children does
   not: per-op alpha lets the overlap show through. Apply per-op in this phase,
   detect overlap when the element owns more than one op, and count it as a
   fidelity gap. Phase 08 closes it with a group layer. Do not claim parity
   before then.

6. `currentColor`. A keyframe that animates a colour to or from
   `currentColor` resolves against the element's computed `color`. The engine
   does not export the computed colour on the box, so resolve it from the
   element's own text op when one exists; otherwise treat it as a bind error and
   count it.

7. Damage. A colour or opacity change damages the op's painted box, not the
   element's whole box. Reuse `internal/page/dirty/op.go` bounds. A transform
   change damages the old and new transformed boxes.

8. Static style must win on rebind. Capturing base paint at bind is correct for
   `PatchElement` and hover changes, but a rebind after a theme change must
   re-capture the base, not reuse the previous sampled value. Clear the captured
   base on generation change and capture again from the new display.

## Checklist

### A. Targeting

- [ ] Add `targetsFor(box, property)` using `OpOwners` nearest-owner equality.
- [ ] Keep transform on the subtree and colours on the nearest owner.
- [ ] Exclude Text from background and border properties.
- [ ] Exclude Image and GridRun from colour properties.
- [ ] Test a parent card with a label: the label colour does not change.
- [ ] Test a child card inside a parent: the parent colour does not leak.

### B. Field masks

- [ ] Add the mask to the target and to `compose`.
- [ ] Stop writing W, H, and Alpha unless their property is animated.
- [ ] Test a translucent op: animating its transform leaves Alpha alone.
- [ ] Test that two animations on one element with disjoint fields compose.

### C. Colour

- [ ] New parser with hex, functional, named, transparent, currentColor.
- [ ] Rule colours: `column-rule-color`, `row-rule-color`, `rule-color`.
- [ ] `text-emphasis-color` and `accent-color`.
- [ ] `fill` and `stroke` counted as deferred to phase 10, not dropped.
- [ ] Unit tests for each form and for malformed values.
- [ ] Interpolate in sRGB and clamp to [0,1].
- [ ] Count an unsupported colour space once per element, not per frame.

### D. Opacity and deferred properties

- [ ] Apply opacity per op for the single-op case.
- [ ] Count multi-op opacity as a fidelity gap until phase 08.
- [ ] Count `filter`, `clip-path`, and blend keyframes as deferred with a reason.
- [ ] Leave deferred properties at base paint rather than writing a partial value.

### E. Examples

- [ ] Restore the animations example label colours.
- [ ] Add cards for `color`, `border-color`, `outline-color`, `opacity`,
      `currentColor`, and `rgb()`.
- [ ] Add one card with overlapping children and opacity, marked as the phase 08
      parity case.

Gate: catalog cards for colours and single-op opacity match a browser; the
label-recolour regression has a test; a deferred filter card reports its reason.

## Risks

- Nearest-owner targeting depends on `OpOwners` being populated for every op.
  Assert it at bind and fall back to geometric containment with a counter.
- HSL interpolation in sRGB is what browsers do for legacy `hsl()`, so do not
  premultiply or convert to another space.
- `transparent` is `rgba(0,0,0,0)`. Interpolating to it darkens the midpoint in
  sRGB, which matches CSS.
