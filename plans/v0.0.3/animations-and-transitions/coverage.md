# Coverage against the compatibility matrix

This file is the audit trail for the claim "every implemented animation property
is covered". It maps the implemented properties in
[`documentation/compatibility-matrix.md`](../../../documentation/compatibility-matrix.md)
to a phase, or records why there is no work.

Method:

1. The matrix lists 555 implemented properties (§2.1). Vendored `-webkit-`
   aliases are included.
2. Of those, only the ones CSS defines as animatable need motion work. A
   property that is not animatable, or whose change never reaches paint, is
   excluded with a reason.
3. Every other property belongs to exactly one class below and has a phase.

A new matrix row that is animatable must add a row here in the same change that
implements it. A row without a phase and without a reason is a gap.

## 1. Control properties

### Animation

| Property | Phase |
|---|---|
| `animation`, `animation-name`, `-duration`, `-timing-function`, `-delay`, `-delay-start`, `-delay-end`, `-iteration-count`, `-direction`, `-fill-mode`, `-play-state` | 05 |
| `animation-composition` | 05 |
| `animation-timeline`, `-range`, `-range-start`, `-range-end`, `-range-center` | 06 |
| `animation-trigger`, `trigger-scope` | 06 |
| `@keyframes`, `@-webkit-keyframes` | 02, 05 |
| `-webkit-animation*` (nine longhands) | 05, through the alias table |

### Transition

| Property | Phase |
|---|---|
| `transition`, `-property`, `-duration`, `-timing-function`, `-delay` | transition 01, 03 |
| `transition-behavior` | transition 02 |
| `-webkit-transition*` (four longhands) | transition 01, through the alias table |

## 2. Target properties

### Transform and 3D

| Property | Phase | Mechanism |
|---|---|---|
| `transform` | 02 | matrix |
| `translate`, `rotate`, `scale` | 02 | matrix, individual-property order |
| `translate3d`, `translateZ`, `scale3d`, `scaleZ`, `rotateX`, `rotateY`, `rotateZ`, `rotate3d`, `matrix3d`, `perspective()` | 02, 04, priority 0 | 4x4, `RestampTransforms` |
| `transform-origin`, `transform-box` | 02, priority 0 | matrix, blocked on the engine export |
| `perspective`, `perspective-origin` | 02, 04, priority 0 | projective, needs `RestampTransforms` |
| `transform-style`, `backface-visibility` | 02, discrete | no interpolation; `preserve-3d` flattens as the engine does statically |

### Opacity and colour

| Property | Phase | Mechanism |
|---|---|---|
| `opacity` | 03, 08 | per-op, group alpha when overlapping |
| `color` | 03 | nearest text op |
| `background-color` | 03 | fill ops |
| `border-color`, `border-*-color`, logical border colours | 03 | stroke ops |
| `outline-color` | 03 | stroke ops |
| `text-decoration-color`, `text-emphasis-color` | 03 | text decor ops |
| `column-rule-color`, `row-rule-color`, `rule-color` | 03 | line ops |
| `accent-color` | 03 | meter fill ops |
| `fill`, `fill-opacity`, `stroke`, `stroke-opacity`, `stroke-width`, `stroke-dasharray`, `stroke-dashoffset` | 04, 10, [phase 11](animation/phase-11-engine-named-timelines-and-svg-paint.md) | re-serialized SVG image, not a per-op field |

### Backgrounds

| Property | Phase | Mechanism |
|---|---|---|
| `background-image` (url and gradients) | 04, 10 | repaint; gradient shape must match to interpolate |
| `background-position`, `-x`, `-y`, `-inline`, `-block` | 04 | image op offset |
| `background-size` | 04 | image op geometry |
| `background-repeat`, `-x`, `-y`, `-inline`, `-block` | discrete | no interpolation |
| `background-clip`, `background-origin`, `background-blend-mode` | discrete | no interpolation |

### Borders, outlines, and rules

| Property | Phase | Mechanism |
|---|---|---|
| `border-width`, logical border widths | 04 | layout |
| `border-radius`, corner radii | 04 | layout |
| `border-style`, logical border styles | discrete | no interpolation |
| `border-image-slice`, `-width`, `-outset` | 04 | border image rebuild |
| `border-image-source`, `-repeat` | discrete | no interpolation |
| `border-spacing` | 04 | table layout |
| `outline-width`, `outline-offset` | 04 | stroke rebuild |
| `outline-style` | discrete | no interpolation |
| `column-rule-width`, `row-rule-width`, `rule-width` | 04 | line rebuild |
| `column-rule-style`, `row-rule-style`, `rule-style` | discrete | no interpolation |

### Shadows, filters, masks, and clips

| Property | Phase | Mechanism |
|---|---|---|
| `box-shadow` and its longhands | 04 | shadow op rebuild |
| `text-shadow` | 04 | text paint rebuild |
| `filter` | 08 | layer + shader |
| `clip-path` | 08 | mask texture |
| `clip` | 04 | clip rect rebuild |
| `mask-position`, `mask-size` | 08 | mask texture transform |
| `mask-image` | 08, 10 | same-shape gradient or url cross-fade; otherwise discrete |
| `mask-border`, `mask-border-slice`, `-width`, `-outset` | 08 | mask rebuild |
| `mask-repeat`, `mask-origin`, `mask-clip`, `mask-composite`, `mask-mode`, `mask-border-source`, `mask-border-repeat`, `mask-border-mode` | discrete | no interpolation |
| `overflow-clip-margin` and its longhands | 04 | clip geometry |
| `mix-blend-mode`, `isolation` | discrete | mode switch, no interpolation |

### Layout and text metrics

| Property | Phase | Mechanism |
|---|---|---|
| `width`, `height`, `inline-size`, `block-size` | 04 | relayout |
| `min-width`, `min-height`, `max-width`, `max-height`, logical min/max | 04 | relayout |
| `top`, `right`, `bottom`, `left`, `inset`, logical insets | 04 | relayout |
| `margin`, `padding` and all longhands | 04 | relayout |
| `gap`, `row-gap`, `column-gap` | 04 | relayout |
| `flex-basis`, `flex-grow`, `flex-shrink` | 04 | relayout |
| `grid-template-columns`, `grid-template-rows`, `grid-auto-columns`, `grid-auto-rows` | 04 | relayout; track lists must match |
| `aspect-ratio` | 04 | relayout |
| `object-position`, `object-view-box` | 04 | image geometry |
| `vertical-align` | 04 | table and inline layout |
| `text-indent` | 04 | text layout |
| `tab-size` | 04 | text layout |
| `text-decoration-thickness`, `text-underline-offset`, `text-decoration-inset` | 04 | text paint rebuild |
| `text-emphasis-position` | discrete | keywords |
| `font-size`, `font-size-adjust` | 04 | text layout |
| `font-weight` | discrete | face selection |
| `font-variation-settings` | excluded | the engine picks a static face, no axis interpolation |
| `line-height`, `letter-spacing`, `word-spacing` | 04 | text layout |
| `column-count` | discrete | integer |
| `column-width`, `column-height` | 04 | multicol layout |
| `column-rule-inset*`, `row-rule-inset*`, `rule-inset*` | 04 | rule paint geometry |
| `shape-outside`, `shape-margin`, `shape-padding`, `shape-inside` | 04 | float layout |
| `contain-intrinsic-size`, `-width`, `-height`, `-inline-size`, `-block-size` | 04 | containment layout |
| `float-offset` | 04 | float layout |

### Discrete and presence

| Property | Phase | Rule |
|---|---|---|
| `display` | transition 02 | entry at 0%, exit at 100% with a retained layer |
| `visibility`, `content-visibility` | transition 02 | visible through the run, flip at the end |
| `position` | transition 02 | halfway swap, existing behaviour |
| `z-index` | transition 02 | integer, flip at 50% |
| `order` | transition 02 | integer, flip at 50% |
| `grid-column-start/end`, `grid-row-start/end` | transition 02 | integer, flip at 50% |
| `line-clamp` | transition 02 | integer, flip at 50% |
| `color-scheme` | excluded | a theme switch, not an interpolable value; the page repaints |

### Mode and layout switches

These implement a discrete change. A transition switches them at 50% and a
keyframe holds the previous value until its offset. They never produce a
per-frame value, and they are not gaps.

| Family | Properties |
|---|---|
| Fragmentation | `break-before`, `break-after`, `wrap-before`, `wrap-after`, `wrap-inside`, `region-fragment`, `flow-into`, `flow-from`, `flow-tolerance`, `float-reference` |
| Tables | `border-collapse`, `caption-side`, `empty-cells`, `table-layout` |
| Text layout | `text-align`, `text-align-all`, `text-align-last`, `text-justify`, `text-transform`, `white-space`, `white-space-collapse`, `white-space-trim`, `word-break`, `overflow-wrap`, `hyphens`, `line-break`, `text-wrap`, `text-wrap-mode`, `text-wrap-style`, `text-spacing`, `text-spacing-trim`, `text-autospace`, `text-box`, `text-box-edge`, `text-box-trim`, `text-combine-upright`, `text-orientation`, `text-group-align`, `ruby-align`, `ruby-merge`, `ruby-overhang`, `ruby-position`, `hanging-punctuation`, `initial-letter`, `initial-letter-align`, `initial-letter-wrap`, `dominant-baseline`, `alignment-baseline`, `text-anchor`, `max-lines` |
| Fonts and faces | `font-style`, `font-width`, `font-feature-settings`, `font-kerning`, `font-variant*`, `font-synthesis*`, `font-palette`, `font-optical-sizing`, `font-language-override` |
| Lists and generated content | `list-style`, `list-style-image`, `list-style-position`, `list-style-type`, `content`, `quotes`, `counter-increment`, `counter-reset`, `counter-set` |
| Positioning and anchoring | `position`, `float`, `clear`, `container`, `container-type`, `container-name`, `anchor-name`, `position-anchor`, `position-area`, `position-try`, `position-try-fallbacks`, `position-try-order`, `position-visibility` |
| Images | `object-fit`, `image-orientation`, `image-resolution` |
| Misc | `contain`, `dynamic-range-limit` |
| Separate track | `view-transition-*` (view transitions are not part of this plan) |

### No visual effect

| Property | Reason |
|---|---|
| `scroll-margin` and its longhands | the matrix records no paint effect |
| `orphans`, `widows` | paged media, no screen paint |
| `print-color-adjust`, `forced-color-adjust` | print and forced-colors only |
| `color-adjust` | alias of `print-color-adjust` |
| `image-orientation`, `image-resolution` | decode-time, not a frame value |
| `shape-rendering`, `color-interpolation`, `color-interpolation-filters` | renderer hint, discrete |
| `font-optical-sizing`, `font-synthesis*`, `font-palette` | face selection, discrete |
| `text-decoration-style`, `text-emphasis-style` | discrete |
| `clip-rule`, `fill-rule` | discrete |
| `box-sizing`, `table-layout`, `caption-side`, `empty-cells`, `direction`, `unicode-bidi` | discrete or no interpolable value |

## 3. Blocked

These are implemented by the engine but cannot animate correctly until a named
dependency lands. They are not silently dropped; each has a counter and a matrix
note.

| Item | Blocker | Phase |
|---|---|---|
| `transform-origin`, `transform-box`, the element's own matrix | `layout.Box` does not export them; [priority 0](priority-0-engine-transform-and-perspective.md) adds `TransformInfo` | 02, priority 0 |
| 3D transform functions, `perspective`, `perspective-origin` | projective; [priority 0](priority-0-engine-transform-and-perspective.md) adds `RestampTransforms`, because a 2D layer matrix cannot represent them | 02, 04, priority 0 |
| Named timelines (`animation-timeline: --x`) | [phase 11](animation/phase-11-engine-named-timelines-and-svg-paint.md) adds `scroll-timeline-*` and `view-timeline-*`; nested scrollers and `timeline-scope` stay out of scope | 06, phase 11 |
| `scroll()`, `view()` timelines | host input is available; not blocked | 06 |
| Corner shapes (`corner-*` and `DisplayOpPath`) | `internal/replay` has no painter for `DisplayOpPath`, so the page falls back to a bitmap and no op patch is visible | 10 |
| SVG `fill`/`stroke` paint | inline SVG serializes to an `OpImage`; [phase 11](animation/phase-11-engine-named-timelines-and-svg-paint.md) adds `RestyleInlineSVG`, and phase 10 integrates it | 10, phase 11 |
| Motion on a bitmap-fallback page | the window draws the picture, not the display list, so patched ops are invisible | 10 |
| `will-change` | parsed and ignored by the engine (§2.3); promotion uses the animated property instead | 08 |

## 4. What "all" means

The track claims every implemented property that is animatable, not every
implemented property. The two excluded groups are stated above: no visual effect,
and blocked with a named dependency. CSS that no plan implements, such as
`offset-path` and `offset-distance`, is out of scope. Named timelines and SVG
paint are implemented by the phase 11 plan, not excluded.

The conformance page in animation phase 09 renders one card per class and one card
per blocked case, so a regression is a failing test rather than a missing row.
