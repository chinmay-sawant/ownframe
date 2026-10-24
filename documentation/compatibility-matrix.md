# blinkless - HTML/CSS compatibility matrix

> **Parent:** [plans/v0.0.1/phase-wise-checklist.md](../plans/v0.0.1/phase-wise-checklist.md) and [plans/v0.0.1/html-css-json-compatibility-checklist.md](../plans/v0.0.1/html-css-json-compatibility-checklist.md)  
> **Status:** living contract - amendments go through plan review  
> **Target:** authored HTML and CSS to a `layout.DisplayList` (the drawing list). **Not** a browser. **Not** a PDF writer. **Not** a page rasterizer.  
> **Catalog:** `testdata/css/catalog/properties.json` (schema v2), measured 2026-10-20: 785 rows - 555 Implemented / 3 Partial / 219 Unsupported / 8 intentionally ignored. Motion rows with scope `blinkless+ownframe` include sampled behavior evidence from OwnFrame. Upstream pin: webref `ed/css` revision `1f2ec8f74a80c14066b4c7d6822cee59f69fa03b` (821 properties). Fidelity guide: [fidelity.md](fidelity.md).
> **Limitations:** Implemented means a handler, a consumer outside the style layer, and a resolvable behavior test. Browser rendering parity, HTML parser conformance, and this property catalog are separate claims with separate evidence. Section 1 is a rendering allowlist for HTML tags, not a parser conformance claim.

This document is the contract for the layout engine. Its output is a drawing
list; the PDF writer, the page rasterizer, and the CLI from earlier revisions
are not part of this tree. Property statuses come from
`testdata/css/catalog/properties.json` and are reproduced in section 2.
Anything not in that catalog is unsupported; unsupported input must degrade
gracefully (ignored declaration, skipped node, or documented error), never
crash. Product framing: [fidelity.md](fidelity.md). **Still not full CSS.**

---

## 1. Supported HTML tags

MVP renders these elements. Everything else is stripped, ignored, or rendered
as its inline text (per the note column).

| Tag | Behavior note |
|-----|---------------|
| `html`, `head`, `body` | Document shell |
| `title`, `meta`, `style`, `link` | Metadata; `link rel=stylesheet` fetched (Phase 2) |
| `div`, `span` | Generic block / inline boxes |
| `p` | Block, default margins |
| `br` | Forced line break |
| `hr` | Block-level horizontal rule |
| `h1`-`h6` | Heading levels; UA size and weight rules |
| `ul`, `ol`, `li` | Lists. UA stylesheet: `ul`/`menu` → `disc`, `ol` → `decimal`. `markerText` implements `disc` / `circle` / `square` / `decimal` / `decimal-leading-zero` / `lower-alpha` / `upper-alpha` / `lower-roman` / `upper-roman` |
| `table`, `thead`, `tbody`, `tfoot`, `tr`, `th`, `td`, `caption` | Table subset; see §4 and the table rows in §2 (`colspan` and `rowspan` Implemented; `<caption>` / `table-caption` rendered above the table) |
| `img` | Replaced element; **PNG/JPEG/SVG subset**. The image op keeps the encoded bytes and re-encodes them as PNG only when EXIF orientation or a clip needs new pixels (`layout_images.go:365`, `image_exif.go:497`). SVG images and inline `<svg>` are rasterized through `internal/svg` (`layout_svg.go:23`). Layout maps CSS px to points at 0.75 (96 px/in). `web.images=false` skips fetch and paint (`settings.ResolveImages`). |
| `a` | Link box in the drawing list (`OpLinkURI`): external `http/https/mailto` URIs (`isExternalHref`, `inline_paint.go:1876`) and same-document `#id` fragments (`isInternalHref`, `inline_paint.go:1906`). Relative references are retained for the caller to resolve. `layout.Result.HasFragmentLinks` reports whether any link targets a fragment (`layout.go:197`). |
| `strong`, `em`, `b`, `i`, `u`, `small` | `b`/`strong` → bold face; `em`/`i` → italic face (Liberation family; see the `font-family` row in §2); `u` underline; `small` smaller; fake stroke bold only if a bold face is missing |
| `pre`, `code` | `pre` honors `white-space: pre`; `code` follows the author’s `font-family` (generic `monospace` → bundled Liberation Mono; see [fonts.md](fonts.md)) |
| `blockquote` | Block-level only, no indent margins (UA rule `style_values.go:1628-1630`) |
| `header`, `footer`, `main`, `section`, `article`, `aside`, `nav` | Treated as `div` (semantic aliases) |

## 2. Supported CSS properties

The catalog in `testdata/css/catalog/properties.json` (schema v2) is the authoritative property record. Measured 2026-10-20: 785 rows - 555 implemented, 3 partial, 219 unsupported, 8 intentionally ignored. The pinned upstream inventory is webref `ed/css` revision `1f2ec8f74a80c14066b4c7d6822cee59f69fa03b` (821 properties: 671 draft, 70 vendor, 52 svg, 28 browser-ui). Regenerate the tables below with `python3 scripts/css-catalog-map.py --matrix`; `make catalog-check` fails on drift between the catalog and the code. Rows scoped to `blinkless+ownframe` have sampled runtime evidence in OwnFrame; Blinkless supplies their CSS metadata.

Statuses describe the stated layout pipeline scope, not browsers:

- **Implemented** - a handler parses the value, a consumer outside the style layer reads it, and at least one behavior test resolves. Motion rows scoped to `blinkless+ownframe` are sampled by OwnFrame.
- **Partial** - a handler and a consumer exist, but a named behavior is missing or unverified in the stated scope; the limitation column names it.
- **Unsupported** - no consumer and no observable behavior. The declaration may be parsed and stored or dropped entirely.
- **Intentionally ignored** - deliberately ignored print-noop UI chrome.

An implemented status is not a browser-parity or parser-conformance claim. Rendering comparisons and HTML parsing carry separate evidence in `plans/v0.0.1/html-css-json-compatibility-checklist.md`.

### 2.1 Implemented (555)

| Property | Accepted values | Source | Behavior tests |
|---|---|---|---|
| `-webkit-animation` | comma-separated keyframe names with optional duration, easing, delay, iteration count, direction, fill mode, play state, and timeline | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation)` |
| `-webkit-animation-delay` | signed start <time> with optional non-negative end <time>; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-delay)` |
| `-webkit-animation-direction` | normal \| reverse \| alternate \| alternate-reverse; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-direction)` |
| `-webkit-animation-duration` | auto \| non-negative <time>; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-duration)` |
| `-webkit-animation-fill-mode` | none \| forwards \| backwards \| both; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-fill-mode)` |
| `-webkit-animation-iteration-count` | infinite \| non-negative <number>; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-iteration-count)` |
| `-webkit-animation-name` | none \| custom-ident \| quoted keyframes name; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-name)` |
| `-webkit-animation-play-state` | running \| paused; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-play-state)` |
| `-webkit-animation-timing-function` | linear \| ease \| ease-in \| ease-out \| ease-in-out \| step-start \| step-end \| cubic-bezier() \| steps() \| linear() | `internal/layout/animation_transition.go`, `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (-webkit-animation-timing-function)` |
| `-webkit-backface-visibility` | alias of the unprefixed property | `internal/layout/style_cascade.go`, `internal/layout/transform.go` | `TestBehaviorWebkitBackfaceAliasHides` |
| `-webkit-mask` | url(...) \| none; supported mask mode, position, size, repeat, origin, clip, and composite values | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-mask-box-image` | alias of mask-border | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderVendorAliasUsesCSSSourceOrder`, `TestBehaviorMaskBorderSupportsAcceptedValues`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-box-image-outset` | alias of mask-border-outset | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-box-image-repeat` | alias of mask-border-repeat | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-box-image-slice` | alias of mask-border-slice | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderInvalidValuesKeepEarlier`, `TestBehaviorMaskBorderSupportsAcceptedValues`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-box-image-source` | alias of mask-border-source | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderInvalidValuesKeepEarlier`, `TestBehaviorMaskBorderSupportsAcceptedValues`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-box-image-width` | alias of mask-border-width | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderAliasReachesGroup`, `TestComposeMaskBorderRasterOnGroup` |
| `-webkit-mask-clip` | border-box \| padding-box \| content-box | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-mask-image` | url(...) \| none | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup`, `TestBehaviorMaskGradientBytesOnGroup`, `TestBehaviorMaskImageSetPicks1x` |
| `-webkit-mask-origin` | border-box \| padding-box \| content-box | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-mask-position` | CSS position keywords, lengths, and percentages, plus 4-value edge-offset positions such as left 10px top 20px | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-mask-repeat` | repeat \| repeat-x \| repeat-y \| no-repeat \| space \| round \| one or two of repeat, space, round, no-repeat | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-mask-size` | auto \| cover \| contain \| one or two non-negative CSS lengths or percentages | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `-webkit-perspective` | alias of the unprefixed property | `internal/layout/style_cascade.go`, `internal/layout/transform.go` | `TestBehaviorWebkitPerspectiveAliasForeshortens` |
| `-webkit-perspective-origin` | alias of the unprefixed property | `internal/layout/style_cascade.go`, `internal/layout/transform.go` | `TestBehaviorWebkitPerspectiveOriginAlias` |
| `-webkit-transform-style` | alias of the unprefixed property | `internal/layout/style_cascade.go`, `internal/layout/transform.go` | `TestBehaviorWebkitTransformStyleAliasPreserve3D` |
| `-webkit-transition` | comma-separated property with optional duration, easing, delay, and behavior | `internal/layout/animation_transition.go`, `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (-webkit-transition)` |
| `-webkit-transition-delay` | signed <time>; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (-webkit-transition-delay)` |
| `-webkit-transition-duration` | non-negative <time>; comma-separated | `internal/layout/animation_transition.go`, `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (-webkit-transition-duration)` |
| `-webkit-transition-property` | all \| none \| custom-ident; comma-separated except none stands alone | `internal/layout/animation_transition.go`, `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (-webkit-transition-property)` |
| `-webkit-transition-timing-function` | linear \| ease \| ease-in \| ease-out \| ease-in-out \| step-start \| step-end \| cubic-bezier() \| steps() \| linear() | `internal/layout/animation_transition.go`, `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestWebkitAnimationTransitionAliasesMatchStandardProperties`, `TestWebkitAnimationTransitionAliasesRespectCascadeOrder`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (-webkit-transition-timing-function)` |
| `accent-color` | auto \| <color> | `internal/layout/style_properties.go` | `TestBehaviorAccentColorTintsMeterFill` |
| `align-content` | normal \| <baseline-position> \| <content-distribution> \| <overflow-position>? <content-position> | `internal/layout/style_properties.go` | `TestBehaviorAlignContentGridStartUsedOffset`, `TestBehaviorAlignContentGridCenterUsedOffset`, `TestBehaviorAlignContentGridEndUsedOffset`, `TestBehaviorAlignContentGridBetweenUsedOffset`, `TestBehaviorAlignContentGridAroundUsedOffset` |
| `align-items` | normal \| stretch \| <baseline-position> \| <overflow-position>? <self-position> | `internal/layout/style_properties.go` | `TestWebkitPrefixAliases`, `TestBehaviorAlignItemsCenterOffset` |
| `align-self` | auto \| <overflow-position>? [ normal \| <self-position> ]\| stretch \| <baseline-position> | `internal/layout/style_properties.go` | `TestBehaviorAlignSelfEndUsedOffset` |
| `alignment-baseline` | baseline \| <baseline-metric> | `internal/layout/style_paint_props.go` | `TestBehaviorAlignmentBaselineNoPaintEffect`, `TestSvgTextAlignmentBaselineCentralIdeographic`, `TestSvgTextAlignmentBaselineMiddleHanging` |
| `anchor-name` | none \| <anchor-name># | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestAnchorNameScope` |
| `anchor-scope` | none \| all \| <anchor-name># | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestAnchorScopeLookup` |
| `animation` | comma-separated keyframe names with optional duration, easing, delay, iteration count, direction, fill mode, play state, and timeline | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go`, `internal/css/keyframes.go`, `layout/displaylist.go` | `TestAnimationDeclarationValueAccepted`, `TestExpandAnimationDeclaration`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation)` |
| `animation-composition` | replace \| add \| accumulate | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationCompositionChangesOwnFrameSample (add)` |
| `animation-delay` | signed start <time> with optional non-negative end <time>; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestExpandAnimationDeclaration`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-delay)` |
| `animation-delay-end` | non-negative <time>; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-delay-end)` |
| `animation-delay-start` | signed <time>; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-delay-start)` |
| `animation-direction` | normal \| reverse \| alternate \| alternate-reverse; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-direction)` |
| `animation-duration` | auto \| non-negative <time>; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-duration)` |
| `animation-fill-mode` | none \| forwards \| backwards \| both; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-fill-mode)` |
| `animation-iteration-count` | infinite \| non-negative <number>; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-iteration-count)` |
| `animation-name` | none \| custom-ident \| quoted keyframes name; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go`, `internal/css/keyframes.go`, `layout/displaylist.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-name)` |
| `animation-play-state` | running \| paused; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-play-state)` |
| `animation-range` | start timeline-range edge with optional end edge; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestExpandAnimationDeclaration`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationRangeReachesOwnFrameFrame (entry 10px exit 90px)` |
| `animation-range-center` | normal \| source \| target \| <length-percentage> \| source/target with <length-percentage> | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationRangeCenterReachesOwnFrameFrame (source 50%)` |
| `animation-range-end` | normal \| timeline-range-name with optional <length-percentage> \| <length-percentage> | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationRangeEndReachesOwnFrameFrame (75%)` |
| `animation-range-start` | normal \| timeline-range-name with optional <length-percentage> \| <length-percentage> | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationRangeStartReachesOwnFrameFrame (25%)` |
| `animation-timeline` | auto \| none \| <dashed-ident> \| scroll() \| view() with accepted argument forms; comma-separated | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationTimelineSamplesHostProgress (scroll())` |
| `animation-timing-function` | linear \| ease \| ease-in \| ease-out \| ease-in-out \| step-start \| step-end \| cubic-bezier() \| steps() \| linear() | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationPropertiesReachOwnFrameFrames (animation-timing-function)` |
| `animation-trigger` | none \| <dashed-ident> followed by one or more none/play/play-once/play-forwards/play-backwards/pause/reset/replay actions | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogAnimationTriggerChangesOwnFrameAfterHostEvent (play reset)` |
| `aspect-ratio` | auto \|\| <ratio> | `internal/layout/style_aspect_ratio_props.go` | `TestBehaviorAspectRatioDerivesHeightFromWidth` |
| `backface-visibility` | visible \| hidden | `internal/layout/style_properties.go` | `TestBehaviorBackfaceVisibilityNoPaintEffect`, `TestBackfaceHiddenCullRotateY180`, `TestBackfaceVisiblePaintsRotateY180`, `TestBackfaceHiddenKeeps2DMirror` |
| `background` | [<'background-color'> \|\| <'background-image'> \|\| <'background-repeat'> \|\| <'background-attachment'> \|\| <'background-position'>] \| inherit | `internal/layout/style_properties.go` | `TestApplyImageKeyBackgroundAlias`, `TestAuthorBackgroundShorthandOverridesButtonUA`, `TestBackgroundImageParse`, `TestBackgroundSingleFieldNoWebMirror`, `TestGlobalGetSetRoundTripAndIgnored` |
| `background-blend-mode` | <'mix-blend-mode'># | `internal/layout/style_advanced_props.go` | `TestBehaviorBackgroundBlendModeMultiplyBlends` |
| `background-clip` | <bg-clip># | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundClipClipsImage` |
| `background-color` | <color> \| transparent \| inherit | `internal/layout/style_properties.go` | `TestBackgroundFill`, `TestBehaviorBackgroundColorPaintsFill` |
| `background-image` | <uri> \| none \| inherit | `internal/layout/style_properties.go` | `TestBackgroundImageLayoutPaints`, `TestBackgroundLonghands` |
| `background-origin` | <visual-box># | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundOriginInsetsImage` |
| `background-position` | [ [ <percentage> \| <length> \| left \| center \| right ] [ <percentage> \| <length> \| top \| center \| bottom ]? ] \| [ [ left \| center \| right ] \|\| [ top \| center \| bottom ] ] \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBackgroundPositionMovesImage` |
| `background-position-block` | [ center \| [ [ start \| end ]? <length-percentage>? ]! ]# | `internal/layout/style_properties.go` | `TestBehaviorBackgroundPositionBlockOffsetsImage` |
| `background-position-inline` | [ center \| [ [ start \| end ]? <length-percentage>? ]! ]# | `internal/layout/style_properties.go` | `TestBehaviorBackgroundPositionInlineOffsetsImage` |
| `background-position-x` | [ center \| [ [ left \| right \| x-start \| x-end ]? <length-percentage>? ]! ]# | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundPositionXOffsetsImage` |
| `background-position-y` | [ center \| [ [ top \| bottom \| y-start \| y-end ]? <length-percentage>? ]! ]# | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundPositionYOffsetsImage` |
| `background-repeat` | repeat \| repeat-x \| repeat-y \| no-repeat \| inherit | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundRepeatTilesImage` |
| `background-repeat-block` | <repetition># | `internal/layout/style_properties.go` | `TestBackgroundRepeatLonghandsKeepTheOtherAxisAtInitialRepeat`, `TestBehaviorBackgroundRepeatBlockTilesHorizontally` |
| `background-repeat-inline` | <repetition># | `internal/layout/style_properties.go` | `TestBackgroundRepeatLonghandsKeepTheOtherAxisAtInitialRepeat`, `TestBehaviorBackgroundRepeatInlineTilesVertically` |
| `background-repeat-x` | <repetition># | `internal/layout/style_properties.go` | `TestBackgroundRepeatLonghandsKeepTheOtherAxisAtInitialRepeat`, `TestBehaviorBackgroundRepeatXTilesVertically` |
| `background-repeat-y` | <repetition># | `internal/layout/style_properties.go` | `TestBackgroundRepeatLonghandsKeepTheOtherAxisAtInitialRepeat`, `TestBehaviorBackgroundRepeatYTilesHorizontally` |
| `background-size` | <bg-size># | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestBehaviorBackgroundSizeScalesImage` |
| `block-size` | <'width'> | `internal/layout/style_properties.go` | `TestLogicalSize` |
| `border` | [ <border-width> \|\| <border-style> \|\| <'border-top-color'> ] \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox`, `TestTransparentBorderPaintsNothing` |
| `border-block` | <'border-block-start'> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockPaintsTopBottomEdges` |
| `border-block-color` | <'border-top-color'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockColorPaintsBlockEdges` |
| `border-block-end` | <line-width> \|\| <line-style> \|\| <color> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockEndPaintsBottomEdge` |
| `border-block-end-color` | <color> \| <image-1D> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockEndColorPaintsBottomEdge` |
| `border-block-end-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestBehaviorBorderBlockEndRadiusRoundsBottomCorners` |
| `border-block-end-style` | <line-style> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockEndStyleExpandsBottomSegments` |
| `border-block-end-width` | <line-width> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockEndWidthSetsBottomStrokeWidth` |
| `border-block-start` | <line-width> \|\| <line-style> \|\| <color> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockStartPaintsTopEdge` |
| `border-block-start-color` | <color> \| <image-1D> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockStartColorPaintsTopEdge` |
| `border-block-start-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderBlockStartRadiusRoundsTopCorners` |
| `border-block-start-style` | <line-style> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockStartStyleExpandsTopSegments` |
| `border-block-start-width` | <line-width> | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockStartWidthSetsTopStrokeWidth` |
| `border-block-style` | <'border-top-style'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockStyleExpandsBlockSegments` |
| `border-block-width` | <'border-top-width'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderBlockWidthSetsBlockStrokeWidth` |
| `border-bottom` | [ <border-width> \|\| <border-style> \|\| <'border-top-color'> ] \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `border-bottom-color` | <color> \| transparent \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderBottomColorPaintsEdge` |
| `border-bottom-left-radius` | <length-percentage [0,∞]>{1,2} | `internal/layout/style_paint_props.go` | `TestBehaviorBorderBottomLeftRadiusPaintsCorner` |
| `border-bottom-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_paint_props.go` | `TestBehaviorBorderBottomRadiusRoundsBottomCorners` |
| `border-bottom-right-radius` | <length-percentage [0,∞]>{1,2} | `internal/layout/style_paint_props.go` | `TestBehaviorBorderBottomRightRadiusPaintsCorner` |
| `border-bottom-style` | <border-style> \| inherit | `internal/layout/style_properties.go` | `TestBorderSideStyles`, `TestBehaviorBorderBottomStyleDashedExpandsSegments` |
| `border-bottom-width` | <border-width> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderBottomWidthStrokesEdge` |
| `border-collapse` | collapse \| separate \| inherit | `internal/layout/style_properties.go` | `TestBorderSpacing`, `TestTableLayout`, `TestBehaviorBorderCollapseMergesAdjacency` |
| `border-color` | [ <color> \| transparent ]{1,4} \| inherit | `internal/layout/style_properties.go` | `TestBorderColorFourValues`, `TestTransparentBorderPaintsNothing`, `TestBehaviorBorderColorPaintsTopEdge` |
| `border-end-end-radius` | <border-radius> | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderEndEndRadiusMapsBottomRight` |
| `border-end-start-radius` | <border-radius> | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderEndStartRadiusMapsBottomLeft` |
| `border-image` | <'border-image-source'> \|\| <'border-image-slice'> [ / <'border-image-width'> \| / <'border-image-width'>? / <'border-image-outset'> ]? \|\| <'border-image-repeat'> | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBehaviorBorderImageShorthandPaintsFrame` |
| `border-image-outset` | [ <length [0,∞]> \| <number [0,∞]> ]{1,4} | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBorderImageProps`, `TestBehaviorBorderImageOutsetPaints` |
| `border-image-repeat` | [ stretch \| repeat \| round \| space ]{1,2} | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBorderImageProps`, `TestBehaviorBorderImageRepeatTilesEdges`, `TestBehaviorBorderImageRepeatClipsPartialTile` |
| `border-image-slice` | [<number [0,∞]> \| <percentage [0,∞]>]{1,4} && fill? | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBorderImageProps`, `TestBehaviorBorderImageSliceSelectsGeometry` |
| `border-image-source` | none \| <image> | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBorderImageProps`, `TestBehaviorBorderImageSourcePaintsSlices` |
| `border-image-width` | [ <length-percentage [0,∞]> \| <number [0,∞]> \| auto ]{1,4} | `internal/layout/border_image.go`, `internal/layout/style_properties.go` | `TestBorderImageProps`, `TestBehaviorBorderImageWidthThickensFrame` |
| `border-inline` | <'border-block-start'> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlinePaintsLeftRightEdges` |
| `border-inline-color` | <'border-top-color'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineColorPaintsInlineEdges` |
| `border-inline-end` | <line-width> \|\| <line-style> \|\| <color> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineEndPaintsRightEdge` |
| `border-inline-end-color` | <color> \| <image-1D> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineEndColorPaintsRightEdge` |
| `border-inline-end-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderInlineEndRadiusRoundsRightCorners` |
| `border-inline-end-style` | <line-style> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineEndStyleExpandsRightSegments` |
| `border-inline-end-width` | <line-width> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineEndWidthSetsRightStrokeWidth` |
| `border-inline-start` | <line-width> \|\| <line-style> \|\| <color> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineStartPaintsLeftEdge` |
| `border-inline-start-color` | <color> \| <image-1D> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineStartColorPaintsLeftEdge` |
| `border-inline-start-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestBehaviorBorderInlineStartRadiusRoundsLeftCorners` |
| `border-inline-start-style` | <line-style> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineStartStyleExpandsLeftSegments` |
| `border-inline-start-width` | <line-width> | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineStartWidthSetsLeftStrokeWidth` |
| `border-inline-style` | <'border-top-style'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineStyleExpandsInlineSegments` |
| `border-inline-width` | <'border-top-width'>{1,2} | `internal/layout/style_properties.go` | `TestBehaviorBorderInlineWidthSetsInlineStrokeWidth` |
| `border-left` | [ <border-width> \|\| <border-style> \|\| <'border-top-color'> ] \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `border-left-color` | <color> \| transparent \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderLeftColorPaintsEdge` |
| `border-left-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_paint_props.go` | `TestBehaviorBorderLeftRadiusRoundsLeftCorners` |
| `border-left-style` | <border-style> \| inherit | `internal/layout/style_properties.go` | `TestBorderSideStyles`, `TestBehaviorBorderLeftStyleDashedExpandsSegments` |
| `border-left-width` | <border-width> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderLeftWidthEmittedStrokeWidth` |
| `border-radius` | <length-percentage [0,∞]>{1,4} [ / <length-percentage [0,∞]>{1,4} ]? | `internal/layout/style_properties.go` | `TestRadiusEllipticalLonghand`, `TestRadiusLonghand`, `TestRadiusPercentAxes`, `TestRadiusSlash` |
| `border-right` | [ <border-width> \|\| <border-style> \|\| <'border-top-color'> ] \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `border-right-color` | <color> \| transparent \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderRightColorPaintsEdge` |
| `border-right-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_paint_props.go` | `TestBehaviorBorderRightRadiusRoundsRightCorners` |
| `border-right-style` | <border-style> \| inherit | `internal/layout/style_properties.go` | `TestBorderSideStyles`, `TestBehaviorBorderRightStyleDashedExpandsSegments` |
| `border-right-width` | <border-width> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderRightWidthStrokesEdge` |
| `border-spacing` | <length> <length>? \| inherit | `internal/layout/style_properties.go` | `TestBorderSpacing`, `TestBehaviorBorderSpacingWidensColumns` |
| `border-start-end-radius` | <border-radius> | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderStartEndRadiusMapsTopRight` |
| `border-start-start-radius` | <border-radius> | `internal/layout/style_logical_border.go`, `internal/layout/style_paint_props.go` | `TestLogicalCornerRadii`, `TestBehaviorBorderStartStartRadiusMapsTopLeft` |
| `border-style` | <border-style>{1,4} \| inherit | `internal/layout/style_properties.go` | `TestTransparentBorderPaintsNothing` |
| `border-top` | [ <border-width> \|\| <border-style> \|\| <'border-top-color'> ] \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `border-top-color` | <color> \| transparent \| inherit | `internal/layout/style_properties.go` | `TestTransparentBorderPaintsNothing`, `TestBehaviorBorderTopColorPaintsEdge` |
| `border-top-left-radius` | <length-percentage [0,∞]>{1,2} | `internal/layout/style_paint_props.go` | `TestBehaviorBorderTopLeftRadiusPaintsCorner` |
| `border-top-radius` | <length-percentage [0,∞]>{1,2} [ / <length-percentage [0,∞]>{1,2} ]? | `internal/layout/style_paint_props.go` | `TestBehaviorBorderTopRadiusRoundsTopCorners` |
| `border-top-right-radius` | <length-percentage [0,∞]>{1,2} | `internal/layout/style_paint_props.go` | `TestBehaviorBorderTopRightRadiusPaintsCorner` |
| `border-top-style` | <border-style> \| inherit | `internal/layout/style_properties.go` | `TestBorderSideStyles`, `TestBehaviorBorderTopStyleDashedExpandsSegments` |
| `border-top-width` | <border-width> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorBorderTopWidthEmittedStrokeWidth` |
| `border-width` | <border-width>{1,4} \| inherit | `internal/layout/style_properties.go` | `TestTransparentBorderPaintsNothing` |
| `bottom` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestBackgroundPositionUsesCSSPxIntrinsicSize`, `TestCaptionSideParse`, `TestBehaviorBottomShiftsRelativeBox` |
| `box-shadow` | none \| <shadow># | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowPaintsOffsetFill` |
| `box-shadow-blur` | <length [0,∞]># | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowBlurExpandsLayers` |
| `box-shadow-color` | <color># | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowColorPaintsLayer` |
| `box-shadow-inset` | inset | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowInsetPaintsInnerRim` |
| `box-shadow-offset` | [ none \| <length>{1,2} ]# | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowOffsetMovesLayer` |
| `box-shadow-position` | [ outset \| inset ]# | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowPositionTogglesInset` |
| `box-shadow-spread` | <length># | `internal/layout/style_paint_props.go` | `TestBehaviorBoxShadowSpreadGrowsLayer` |
| `box-sizing` | content-box \| border-box | `internal/layout/style_properties.go` | `TestBoxSizingBorderBox`, `TestWebkitPrefixAliases`, `TestBehaviorBoxSizingBorderBoxIncludesPadding` |
| `break-after` | auto \| avoid \| always \| all \| avoid-page \| page \| left \| right \| recto \| verso \| avoid-column \| column \| avoid-region \| region | `internal/layout/style_properties.go` | `TestPageBreakParsing`, `TestBehaviorBreakAfterColumnNoPageEffect`, `TestFragmentAssessBreakAfterColumnForcesColumn` |
| `break-before` | auto \| avoid \| always \| all \| avoid-page \| page \| left \| right \| recto \| verso \| avoid-column \| column \| avoid-region \| region | `internal/layout/style_properties.go` | `TestPageBreakParsing`, `TestBehaviorBreakBeforeColumnNoPageEffect`, `TestFragmentAssessBreakBeforeColumnForcesColumn`, `TestFragmentAssessBreakBeforeFirstChildNoColumnEffect` |
| `caption-side` | top \| bottom \| inherit | `internal/layout/style_properties.go` | `TestCaptionSideParse`, `TestBehaviorCaptionSideBottomBelowTable` |
| `clear` | none \| left \| right \| both \| inherit | `internal/layout/style_properties.go` | `TestBackgroundImageParse`, `TestFloatInsideTableCell`, `TestFloatLeftRightClear`, `TestBehaviorClearDropsBelowFloats` |
| `clip` | <shape> \| auto \| inherit | `internal/layout/style_paint_props.go` | `TestSvgTextClipRectCutsAbsposBox` |
| `clip-path` | <clip-source> \| [ <basic-shape> \|\| <geometry-box> ] \| none | `internal/layout/clip_path.go`, `internal/layout/style_paint_props.go` | `TestBehaviorClipPathEllipseMasksImage` |
| `clip-rule` | nonzero \| evenodd | `internal/layout/style_paint_props.go` | `TestBehaviorClipRuleEvenOddNoPaintEffect`, `TestBehaviorClipRuleInlineFillRuleMasksStar`, `TestBehaviorClipRuleWireEvenOddMasksStarCenter`, `TestBehaviorClipRuleWireNonzeroKeepsStarCenter`, `TestBehaviorClipRuleWireInlineFillRuleWins`, `TestBehaviorClipRuleWireInheritsFromParent`, `TestBehaviorClipRuleWireInvalidKeepsDefault`, `TestBehaviorClipRuleWireBackgroundMasksStarCenter` |
| `color` | <color> \| inherit | `internal/layout/style_properties.go` | `TestCascadeAndInline`, `TestCascadeEngineSupportsPropertyValues`, `TestColorAdjustPropsForeignProperty`, `TestColorModeSetGrayscale`, `TestOutlineParse`, `TestParseBasic`, `TestParseColorHsl`, `TestParseInline`, `TestWebkitPrefixAliases` |
| `color-adjust` | <'print-color-adjust'> | `internal/layout/style_color_adjust_props.go` | `TestColorAdjustPropsAcceptLegalKeywords`, `TestColorAdjustPropsCSSWideKeywords`, `TestColorAdjustPropsRejectIllegalKeywords` |
| `color-interpolation` | auto \| sRGB \| linearRGB | `internal/layout/style_paint_props.go` | `TestBehaviorColorInterpolationNoPaintEffect`, `TestColorInterpolationColorInterpolationLinearDiffersFromSRGB`, `TestColorInterpolationColorInterpolationPropertyChangesGradientPaint` |
| `color-interpolation-filters` | auto \| sRGB \| linearRGB | `internal/layout/style_paint_props.go` | `TestBehaviorColorInterpolationFiltersNoPaintEffect`, `TestColorInterpolationColorInterpolationFiltersLinearDiffersFromSRGB`, `TestColorInterpolationColorInterpolationFiltersPropertyChangesFilteredPaint` |
| `color-scheme` | normal \| [ light \| dark \| <custom-ident> ]+ && only? | `internal/layout/style_color_adjust_props.go` | `TestColorAdjustPropsAcceptLegalKeywords`, `TestColorAdjustPropsCSSWideKeywords`, `TestColorAdjustPropsReachRestPass`, `TestColorAdjustPropsRejectIllegalKeywords`, `TestBehaviorColorSchemeDarkDefaults`, `TestBehaviorColorSchemeDarkWireCanvasAndText`, `TestBehaviorColorSchemeDarkWireAuthorBGWins` |
| `column-count` | auto \| <integer [1,∞]> | `internal/layout/style_multicol_props.go` | `TestMulticolFlexStretchBalanceNoPageSnap`, `TestBehaviorColumnCountUsedWidths` |
| `column-fill` | auto \| balance \| balance-all | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnFillAutoStacksFirstColumn` |
| `column-gap` | normal \| <length-percentage [0,∞]> \| <line-width> | `internal/layout/style_gap_props.go`, `internal/layout/style_properties.go` | `TestGridRowGapVsColumnGap`, `TestBehaviorColumnGapUsedSpacing`, `TestBehaviorColumnGapFlexUsedSpacing` |
| `column-height` | auto \| <length [0,∞]> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnHeightOpensSecondRow` |
| `column-rule` | <gap-rule-list> \| <gap-auto-rule-list> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnRulePaintedBetweenColumns` |
| `column-rule-break` | none \| normal \| intersection | `internal/layout/style_multicol_props.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleBreakIntersectionSplitsAtRowGaps` |
| `column-rule-color` | <line-color-list> \| <auto-line-color-list> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnRuleColorBluePaintsRule` |
| `column-rule-inset` | <'column-rule-inset-cap'> [ / <'column-rule-inset-junction'> ]? | `internal/layout/style_multicol_props.go`, `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-inset-cap` | <inset-value> <inset-value>? | `internal/layout/style_multicol_props.go`, `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-inset-cap-end` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-inset-cap-start` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-inset-end` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-inset-junction` | <inset-value> <inset-value>? | `internal/layout/style_multicol_props.go`, `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `column-rule-inset-junction-end` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/gap_rules.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `column-rule-inset-junction-start` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/gap_rules.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `column-rule-inset-start` | <inset-value> | `internal/layout/style_multicol_props.go`, `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go` | `TestBehaviorColumnRuleJunctionAndCapInsets` |
| `column-rule-style` | <line-style-list> \| <auto-line-style-list> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnRuleStyleDashedSegmentsRule` |
| `column-rule-width` | <line-width-list> \| <auto-line-width-list> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnRuleWidthUsedThickness` |
| `column-span` | none \| <integer [1,∞]> \| all \| auto | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnSpanAllFullWidth` |
| `column-width` | auto \| <length [0,∞]> | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnWidthAutoCount` |
| `column-wrap` | auto \| nowrap \| wrap | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnWrapNowrapKeepsSingleRow` |
| `columns` | [ <'column-width'> \|\| <'column-count'> ] [ / <'column-height'> ]? | `internal/layout/style_multicol_props.go` | `TestBehaviorColumnsShorthandUsedCount` |
| `contain` | none \| strict \| content \| [ [size \| inline-size] \|\| layout \|\| style \|\| paint ] | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestMulticolAnonInTableNoOverlap`, `TestBehaviorContainSizeCollapsesToEmpty` |
| `contain-intrinsic-block-size` | auto? [ none \| <length [0,∞]> ] | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestBehaviorContainIntrinsicBlockSizeSetsContainedHeight` |
| `contain-intrinsic-height` | auto? [ none \| <length [0,∞]> ] | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestBehaviorContainIntrinsicHeightSetsContainedHeight` |
| `contain-intrinsic-inline-size` | auto? [ none \| <length [0,∞]> ] | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestBehaviorContainIntrinsicInlineSizeSetsInlineBlockWidth` |
| `contain-intrinsic-size` | [ auto? [ none \| <length [0,∞]> ] ]{1,2} | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestBehaviorContainIntrinsicSizeSetsContainedHeight` |
| `contain-intrinsic-width` | auto? [ none \| <length [0,∞]> ] | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestBehaviorContainIntrinsicWidthSetsInlineBlockWidth` |
| `container` | <'container-name'> [ / <'container-type'> ]? | `internal/layout/style_container_props.go`, `internal/layout/style_properties.go` | `TestBehaviorContainerShorthandNoLayoutEffect`, `TestBehaviorContainerShorthandNamedQuerySwitch` |
| `container-name` | none \| <custom-ident>+ | `internal/layout/style_container_props.go`, `internal/layout/style_properties.go` | `TestBehaviorContainerNameNoLayoutEffect`, `TestBehaviorContainerNameLonghandQuerySwitch` |
| `container-type` | normal \| [ [ size \| inline-size ] \|\| scroll-state ] | `internal/layout/style_container_props.go`, `internal/layout/style_properties.go` | `TestBehaviorContainerTypeNoSizeQueryEffect`, `TestBehaviorContainerTypeInlineSizeQuerySwitch` |
| `content` | normal \| none \| [ <string> \| <uri> \| <counter> \| attr(<identifier>) \| open-quote \| close-quote \| no-open-quote \| no-close-quote ]+ \| inherit | `internal/layout/style_paint_props.go` | `TestApplyContainmentPropsParsing`, `TestCSSPartialRemainingTextContent`, `TestContainmentKeywordHelpers`, `TestBehaviorContentBeforeEmitsGeneratedText` |
| `content-visibility` | visible \| auto \| hidden | `internal/layout/style_containment_props.go` | `TestApplyContainmentPropsParsing`, `TestContentVisibilityHiddenSkipsDescendantLayout`, `TestBehaviorContentVisibilityAutoPaintsDescendants` |
| `corner` | <'border-radius'> \|\| <'corner-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_shorthand.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBCornerShorthandRadiusShape` |
| `corner-block-end` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-block-end-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBBlockEndPairNotch` |
| `corner-block-start` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-block-start-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBBlockStartPairNotch` |
| `corner-bottom` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-bottom-left` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-bottom-left-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeASingleBottomLeftBevel` |
| `corner-bottom-right` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-bottom-right-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeASingleBottomRightBevel` |
| `corner-bottom-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBBottomPairScoopBevel` |
| `corner-end-end` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-end-end-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBLogicalSinglesMatchPhysical` |
| `corner-end-start` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-end-start-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBLogicalSinglesMatchPhysical` |
| `corner-inline-end` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-inline-end-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBInlineEndPairNotch` |
| `corner-inline-start` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-inline-start-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBInlineStartPairNotch` |
| `corner-left` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-left-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBLeftPairNotchScoop` |
| `corner-right` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-right-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBRightPairSquircleSquare` |
| `corner-shape` | 1 to 4 <corner-shape-value> tokens assigned to the top, right, bottom, and left sides | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_shorthand.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeABevelEmitsPath`, `TestCornerShapeASuperellipseMatchesKeywords` |
| `corner-start-end` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-start-end-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBLogicalSinglesMatchPhysical` |
| `corner-start-start` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-start-start-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBLogicalSinglesMatchPhysical` |
| `corner-top` | <'border-radius'> \|\| <'corner-shape'> side form (radius piece, shape piece, or both) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_side.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSideShorthands` |
| `corner-top-left` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-top-left-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeASingleTopLeftBevel`, `TestDrawingCornerShapeScoopPath` |
| `corner-top-right` | <'border-top-left-radius'> \|\| <'corner-top-left-shape'> | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_corners.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBSingleCornerShorthands` |
| `corner-top-right-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeASingleTopRightBevel` |
| `corner-top-shape` | round \| scoop \| bevel \| notch \| square \| squircle \| superellipse(<number> \| infinity \| -infinity) | `testdata/css/catalog/upstream/webref-ed-css-1f2ec8f.json`, `internal/layout/style_corner_shape_pairs.go`, `internal/layout/corner_shape_paint.go`, `internal/layout/corner_path.go`, `internal/layout/op_extra.go`, `internal/layout/layout_chrome.go`, `layout/displaylist.go` | `TestCornerShapeBTopPairBevelScoop` |
| `counter-increment` | [ <identifier> <integer>? ]+ \| none \| inherit | `internal/layout/style_paint_props.go` | `TestCounterInBefore`, `TestCounterResetIncrementLayout`, `TestQuotes` |
| `counter-reset` | [ <identifier> <integer>? ]+ \| none \| inherit | `internal/layout/style_paint_props.go` | `TestCounterInBefore`, `TestCounterResetIncrementLayout`, `TestQuotes` |
| `counter-set` | [ <counter-name> <integer>? ]+ \| none | `internal/layout/style_paint_props.go` | `TestCounterInBefore`, `TestCounterResetIncrementLayout`, `TestQuotes` |
| `direction` | ltr \| rtl \| inherit | `internal/layout/style_properties.go` | `TestBehaviorDirectionRtlRightAlignsLine` |
| `display` | inline \| block \| list-item \| inline-block \| table \| inline-table \| table-row-group \| table-header-group \| table-footer-group \| table-row \| table-column-group \| table-column \| table-cell \| table-caption \| none \| inherit | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestDisplayNone`, `TestTableLayout` |
| `dominant-baseline` | auto \| <baseline-metric> | `internal/layout/style_paint_props.go` | `TestBehaviorDominantBaselineNoPaintEffect`, `TestPerspDominantBaselineBeyondHanging`, `TestPerspDominantBaselineInitialAndBogus`, `TestPerspDominantBaselineInherits` |
| `dynamic-range-limit` | standard \| no-limit \| constrained \| <dynamic-range-limit-mix()> | `internal/layout/style_color_adjust_props.go` | `TestColorAdjustPropsAcceptLegalKeywords`, `TestColorAdjustPropsCSSWideKeywords`, `TestColorAdjustPropsReachRestPass`, `TestColorAdjustPropsRejectIllegalKeywords`, `TestBehaviorDynamicRangeLimitStandardNoClampEffect`, `TestBehaviorDynamicRangeLimitClampsBrightColors`, `TestBehaviorDynamicRangeLimitWireClamps`, `TestBehaviorDynamicRangeLimitPaintClampDiverges`, `TestBehaviorDynamicRangeLimitParserYieldsOnlyInGamut`, `TestBehaviorDynamicRangeLimitParserClampsRGB`, `TestBehaviorDynamicRangeLimitParserRejectsWideGamut`, `TestBehaviorDynamicRangeLimitParserClampsMixAndOklch`, `TestBehaviorDynamicRangeLimitParserClampsOklch`, `TestBehaviorWideGamutDynamicRangeDiverges`, `TestBehaviorWideGamutTextDiverges`, `TestBehaviorWideGamutInGamutAgrees`, `TestBehaviorWideGamutBridgeBoundary`, `TestBehaviorWideGamutCascadeEndToEnd`, `TestBehaviorWideGamutTextCascadeEndToEnd` |
| `empty-cells` | show \| hide \| inherit | `internal/layout/style_advanced_props.go` | `TestBehaviorEmptyCellsHideOmitsBackground` |
| `fill` | <paint> | `internal/layout/style_paint_props.go` | `TestBehaviorFillBakesPaint` |
| `fill-opacity` | <'opacity'> | `internal/layout/style_paint_props.go` | `TestBehaviorFillOpacityBakesPaint` |
| `fill-rule` | nonzero \| evenodd | `internal/layout/style_paint_props.go` | `TestBehaviorFillRuleNoPaintEffect`, `TestBehaviorSvgBakeFillRuleGap`, `TestPaintOutputFillRuleEvenOddHolesRaster` |
| `filter` | none \| <filter-value-list> | `internal/layout/style_properties.go` | `TestBehaviorFilterOpacityFoldsIntoPaint` |
| `flex` | none \| [ <'flex-grow'> <'flex-shrink'>? \|\| <'flex-basis'> ] | `internal/layout/style_properties.go` | `TestWebkitBoxFlexGrows`, `TestChromeFlexCase01LegacyAlgorithm`, `TestWebkitPrefixAliases` |
| `flex-basis` | content \| <'width'> | `internal/layout/style_properties.go` | `TestBehaviorFlexBasisUsedWidth` |
| `flex-direction` | row \| row-reverse \| column \| column-reverse | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestWebkitPrefixAliases`, `TestWebkitBoxOrientVerticalStacks`, `TestBehaviorFlexDirectionRowUsedOrder`, `TestBehaviorFlexDirectionColumnUsedStack` |
| `flex-flow` | <'flex-direction'> \|\| <'flex-wrap'> | `internal/layout/style_properties.go` | `TestFlexFlowShorthand`, `TestFlexFlowDirectionWrapLayout` |
| `flex-grow` | <number [0,∞]> | `internal/layout/style_properties.go` | `TestWebkitPrefixAliases`, `TestChromeFlexCase01LegacyAlgorithm`, `TestBehaviorFlexGrowWeightedUsedWidths` |
| `flex-shrink` | <number [0,∞]> | `internal/layout/style_properties.go` | `TestBehaviorFlexShrinkNoShrinkUsedWidth` |
| `flex-wrap` | nowrap \| [ wrap \| wrap-reverse ] \|\| balance | `internal/layout/style_properties.go` | `TestBehaviorFlexWrapWrapUsedPosition` |
| `float` | left \| right \| none \| inherit | `internal/layout/style_properties.go` | `TestFloatInsideTableCell`, `TestFloatLeftRightClear`, `TestTableClearsFloat`, `TestBehaviorFloatLeftSharesBandWithSibling` |
| `float-offset` | <length-percentage> | `internal/layout/style_float_page_props.go` | `TestBehaviorFloatOffsetNudgesFloat` |
| `float-reference` | inline \| column \| region \| page | `internal/layout/style_float_page_props.go` | `TestBehaviorFloatReferencePinsToPage` |
| `flow-from` | none \| <custom-ident> | `internal/layout/named_flow.go`, `internal/layout/layout_flow.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorNamedFlowMovesTextIntoSecondBox`, `TestBehaviorNamedFlowContinuesInNextRegion`, `TestBehaviorNamedFlowNoRegionDropsSourceContent`, `TestBehaviorNamedFlowPropertyStorage` |
| `flow-into` | none \| <custom-ident> [element \| content]? | `internal/layout/named_flow.go`, `internal/layout/layout_flow.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorNamedFlowPropertyStorage`, `TestBehaviorNamedFlowMovesTextIntoSecondBox`, `TestBehaviorNamedFlowElementMovesWholeSourceBox`, `TestBehaviorNamedFlowContinuesInNextRegion` |
| `flow-tolerance` | normal \| <length [0,∞]> \| infinite | `internal/layout/masonry_tolerance.go`, `internal/layout/grid_masonry.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorFlowToleranceChangesMasonryColumn`, `TestBehaviorFlowToleranceNonMasonryGridIgnored`, `TestBehaviorNamedFlowPropertyStorage` |
| `font` | [ [ <'font-style'> \|\| <'font-variant'> \|\| <'font-weight'> ]? <'font-size'> [ / <'line-height'> ]? <'font-family'> ] \| caption \| icon \| menu \| message-box \| small-caption \| status-bar \| inherit | `internal/layout/style.go`, `internal/layout/style_cascade.go` | `TestFontShorthand`, `TestFontLonghandAfterShorthandWins` |
| `font-family` | [ [ <family-name> \| <generic-family> ] [, <family-name> \| <generic-family>]* ] \| inherit | `internal/layout/style_cascade.go` | `TestBehaviorFontFamilyFallbackSelectsFace` |
| `font-feature-settings` | normal \| <feature-tag-value># | `internal/layout/style_font_feature_props.go` | `TestApplyFontVariantPropsIgnoresOtherFontProps`, `TestFontPropsWave4`, `TestBehaviorFontFeatureSettingsSmallCapsTag` |
| `font-kerning` | auto \| normal \| none | `internal/layout/style_font_feature_props.go` | `TestFontPropsWave4`, `TestBehaviorFontKerningNoneDisablesKern` |
| `font-language-override` | normal \| <string> | `internal/layout/style_font_variant_props.go` | `TestApplyFontVariantProps`, `TestFontVariantPropsDispatchAndInheritance`, `TestBehaviorFontLanguageOverrideCarriesTag` |
| `font-optical-sizing` | auto \| none | `internal/layout/style_font_variant_props.go` | `TestApplyFontVariantProps`, `TestBehaviorFontOpticalSizingStaticKeepsFace` |
| `font-palette` | normal \| light \| dark \| <palette-identifier> \| <palette-mix()> | `internal/layout/style_font_variant_props.go` | `TestApplyFontVariantProps`, `TestBehaviorFontPaletteStaticKeepsColor` |
| `font-size` | <absolute-size> \| <relative-size> \| <length> \| <percentage> \| inherit | `internal/layout/style_cascade.go` | `TestApplyTextSupportPropsUnicodeBidi`, `TestFontSizeEmInherit`, `TestParseBasic`, `TestParseInline`, `TestBehaviorFontSizeScalesUsedSize` |
| `font-size-adjust` | none \| [ ex-height \| cap-height \| ch-width \| ic-width \| ic-height ]? [ from-font \| <number [0,∞]> ] | `internal/layout/style_font_size_adjust_props.go` | `TestFontPropsWave4`, `TestBehaviorFontSizeAdjustScalesUsedSize` |
| `font-style` | normal \| italic \| oblique \| inherit | `internal/layout/style_cascade.go` | `TestRealBoldFaceOps`, `TestBehaviorFontStyleItalicSelectsFace` |
| `font-synthesis` | none \| [ weight \|\| style \|\| small-caps \|\| position] | `internal/layout/style_font_synthesis_props.go` | `TestBehaviorFontSynthesisNoneDisablesFakeBold` |
| `font-synthesis-position` | auto \| none | `internal/layout/style_font_synthesis_props.go` | `TestFontSynthesisPositionSubScales`, `TestFontSynthesisSmallCapsUppercases`, `TestFontSynthesisStyleFakeOblique`, `TestBehaviorFontSynthesisPositionScalesSub` |
| `font-synthesis-small-caps` | auto \| none | `internal/layout/style_font_synthesis_props.go` | `TestFontSynthesisPositionSubScales`, `TestFontSynthesisSmallCapsUppercases`, `TestFontSynthesisStyleFakeOblique`, `TestBehaviorFontSynthesisSmallCapsGatesUppercase` |
| `font-synthesis-style` | auto \| none \| oblique-only | `internal/layout/style_font_synthesis_props.go` | `TestFontSynthesisPositionSubScales`, `TestFontSynthesisSmallCapsUppercases`, `TestFontSynthesisStyleFakeOblique`, `TestBehaviorFontSynthesisStyleGatesOblique` |
| `font-synthesis-weight` | auto \| none | `internal/layout/style_font_synthesis_props.go` | `TestBehaviorFontSynthesisWeightGatesFakeBold` |
| `font-variant` | normal \| small-caps \| inherit | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantSmallCapsUppercases` |
| `font-variant-alternates` | normal \| [ stylistic(<font-feature-value-name>) \|\| historical-forms \|\| styleset(<font-feature-value-name>#) \|\| character-variant(<font-feature-value-name>#) \|\| swash(<font-feature-value-name>) \|\| ornaments(<font-feature-value-name>) \|\| annotation(<font-feature-value-name>) ] | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantAlternatesEmitsHistTag` |
| `font-variant-caps` | normal \| small-caps \| all-small-caps \| petite-caps \| all-petite-caps \| unicase \| titling-caps | `internal/layout/style_font_feature_props.go` | `TestFontPropsWave4`, `TestBehaviorFontVariantCapsEmitsSmcpTag` |
| `font-variant-east-asian` | normal \| [ <east-asian-variant-values> \|\| <east-asian-width-values> \|\| ruby ] | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantEastAsianEmitsJp78Tag` |
| `font-variant-emoji` | normal \| text \| emoji \| unicode | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantEmojiAppliesFill` |
| `font-variant-ligatures` | normal \| none \| [ <common-lig-values> \|\| <discretionary-lig-values> \|\| <historical-lig-values> \|\| <contextual-alt-values> ] | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantLigaturesDisablesLiga` |
| `font-variant-numeric` | normal \| [ <numeric-figure-values> \|\| <numeric-spacing-values> \|\| <numeric-fraction-values> \|\| ordinal \|\| slashed-zero ] | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantNumericEmitsTnumTag` |
| `font-variant-position` | normal \| sub \| super | `internal/layout/style_font_feature_props.go` | `TestBehaviorFontVariantPositionEmitsSubsTag` |
| `font-variation-settings` | normal \| [ <opentype-tag> <number> ]# | `internal/layout/style_font_variant_props.go` | `TestApplyFontVariantProps`, `TestBehaviorFontVariationSettingsStaticKeepsFace` |
| `font-weight` | normal \| bold \| bolder \| lighter \| 100 \| 200 \| 300 \| 400 \| 500 \| 600 \| 700 \| 800 \| 900 \| inherit | `internal/layout/style_cascade.go` | `TestRealBoldFaceOps`, `TestBehaviorFontWeightBoldSelectsFace` |
| `font-width` | normal \| <percentage [0,∞]> \| ultra-condensed \| extra-condensed \| condensed \| semi-condensed \| semi-expanded \| expanded \| extra-expanded \| ultra-expanded | `internal/layout/style_font_width_props.go` | `TestFontPropsWave4`, `TestFontStretchAliasesToWidth`, `TestFontWidthCondensesAdvance`, `TestFontWidthKeywords`, `TestBehaviorFontWidthCondensesAdvance` |
| `forced-color-adjust` | auto \| none \| preserve-parent-color | `internal/layout/style_color_adjust_props.go` | `TestColorAdjustPropsAcceptLegalKeywords`, `TestColorAdjustPropsReachRestPass`, `TestColorAdjustPropsRejectIllegalKeywords`, `TestBehaviorForcedColorAdjustNoneNoPaintEffect`, `TestBehaviorForcedColorAdjustNonePreservesAuthor`, `TestBehaviorForcedColorAdjustWirePreservesAuthor`, `TestBehaviorForcedColorAdjustTestModeDiverges`, `TestBehaviorForcedColorAdjustPrintKeepsAuthor`, `TestBehaviorForcedColorAdjustOptionsModeDiverges`, `TestBehaviorForcedColorAdjustActiveHelpersDiverge` |
| `gap` | <'row-gap'> <'column-gap'>? | `internal/layout/style_gap_props.go`, `internal/layout/style_properties.go` | `TestGridRowGapVsColumnGap`, `TestBehaviorGapGridUsedSpacing` |
| `grid` | <'grid-template'> \| <'grid-template-rows'> / [ auto-flow && dense? ] <'grid-auto-columns'>? \| [ auto-flow && dense? ] <'grid-auto-rows'>? / <'grid-template-columns'> | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestGridPlaceItemsCenterShrinksItems`, `TestGridPlaceSelfEndShrinksItem`, `TestGridRowSpanStretchMatchesFixture32`, `TestGridStretchKeepsSiblingStyleIndependent`, `TestGridTemplateShorthand` |
| `grid-area` | <grid-line> [ / <grid-line> ]{0,3} | `internal/layout/style_properties.go` | `TestBehaviorGridAreaPlacesNamedItem` |
| `grid-auto-columns` | <track-size>+ | `internal/layout/style_properties.go` | `TestBehaviorGridAutoColumnsImplicitUsedWidth` |
| `grid-auto-flow` | [ row \| column ] \|\| dense | `internal/layout/style_properties.go` | `TestBehaviorGridAutoFlowColumnUsedPlacement` |
| `grid-auto-rows` | <track-size>+ | `internal/layout/style_properties.go` | `TestBehaviorGridAutoRowsUsedHeight` |
| `grid-column` | <grid-line> [ / <grid-line> ]? | `internal/layout/style_properties.go` | `TestBehaviorGridColumnSpanUsedWidth` |
| `grid-column-end` | <grid-line> | `internal/layout/style_properties.go` | `TestBehaviorGridColumnEndUsedWidth` |
| `grid-column-start` | <grid-line> | `internal/layout/style_properties.go` | `TestBehaviorGridColumnStartUsedOffset` |
| `grid-row` | <grid-line> [ / <grid-line> ]? | `internal/layout/style_properties.go` | `TestGridRowSpan`, `TestBehaviorGridRowSpanUsedHeight` |
| `grid-row-end` | <grid-line> | `internal/layout/style_properties.go` | `TestGridRowSpan` |
| `grid-row-start` | <grid-line> | `internal/layout/style_properties.go` | `TestGridRowSpan` |
| `grid-template` | none \| [ <'grid-template-rows'> / <'grid-template-columns'> ] \| [ <line-names>? <string> <track-size>? <line-names>? ]+ [ / <explicit-track-list> ]? | `internal/layout/style_properties.go` | `TestGridTemplateShorthand`, `TestDisplayGridTemplateRejectsInvalidDeclaration` |
| `grid-template-areas` | none \| <string>+ | `internal/layout/style_properties.go` | `TestBehaviorGridTemplateAreasUsedPlacement` |
| `grid-template-columns` | none \| <track-list> \| <auto-track-list> \| subgrid <line-name-list>? | `internal/layout/style_properties.go` | `TestBehaviorGridTemplateColumnsUsedTracks` |
| `grid-template-rows` | none \| <track-list> \| <auto-track-list> \| subgrid <line-name-list>? | `internal/layout/style_properties.go` | `TestBehaviorGridTemplateRowsUsedTracks` |
| `hanging-punctuation` | none \| [ first \|\| [ force-end \| allow-end ] \|\| last ] | `internal/layout/style_hyphenation_props.go` | `TestHangingPunctuationFirst`, `TestHyphenateLimitChars`, `TestBehaviorHangingPunctuationFirstHangsQuote` |
| `height` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestBehaviorHeightContentBoxUsedSize` |
| `hyphenate-character` | auto \| <string> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestHyphenateLimitChars`, `TestSoftHyphenUsesHyphenateCharacter`, `TestTextPropsWave3` |
| `hyphenate-limit-chars` | [ auto \| <integer [0,∞]> ]{1,3} | `internal/layout/style_hyphenation_props.go` | `TestHyphenateLimitChars` |
| `hyphenate-limit-last` | none \| always \| column \| page \| spread | `internal/layout/style_hyphenation_props.go` | `TestHyphenateLimitChars` |
| `hyphenate-limit-lines` | no-limit \| <integer [0,∞]> | `internal/layout/style_hyphenation_props.go` | `TestHyphenateLimitChars` |
| `hyphenate-limit-zone` | <length-percentage> | `internal/layout/style_hyphenation_props.go` | `TestHyphenateLimitChars` |
| `hyphens` | none \| manual \| auto | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestHyphenateLimitChars`, `TestSoftHyphenUsesHyphenateCharacter`, `TestTextPropsWave3`, `TestBehaviorHyphensNoneSuppressesSoftHyphenBreak` |
| `image-orientation` | from-image \| none \| [ <angle> \|\| flip ] | `internal/layout/style_image_adjust_props.go` | `TestBehaviorImageOrientationSwapsAxes` |
| `image-resolution` | [ from-image \|\| <resolution> ] && snap? | `internal/layout/style_image_adjust_props.go` | `TestBehaviorImageResolutionScalesIntrinsic` |
| `initial-letter` | normal \| <number [1,∞]> <integer [1,∞]> \| <number [1,∞]> && [ drop \| raise ]? | `internal/layout/style_initial_letter_props.go` | `TestInitialLetterSpansThreeLines`, `TestBehaviorInitialLetterExclusionWidth` |
| `initial-letter-align` | [ border-box? [ alphabetic \| ideographic \| hanging \| leading ]? ]! | `internal/layout/style_initial_letter_props.go` | `TestInitialLetterAlignHangingShiftsY`, `TestBehaviorInitialLetterAlignHangingRaisesLetter` |
| `initial-letter-wrap` | none \| first \| all \| grid \| <length-percentage> | `internal/layout/style_initial_letter_props.go` | `TestInitialLetterWrapNoneDoesNotExclude`, `TestBehaviorInitialLetterWrapAllExcludesLaterLines` |
| `inline-size` | auto \| <length-percentage> | `internal/layout/style_properties.go` | `TestContainerPropsParsed`, `TestLogicalSize`, `TestParseContainerRules`, `TestParseContainerShorthand` |
| `inset` | <'top'>{1,4} | `internal/layout/style_properties.go` | `TestLogicalInset`, `TestBehaviorInsetAnchorsAbsoluteBox` |
| `inset-block` | <'top'>{1,2} | `internal/layout/style_properties.go` | `TestLogicalInset` |
| `inset-block-end` | <'top'> | `internal/layout/style_properties.go` | `TestBehaviorInsetBlockEndAnchorsAbsoluteBox` |
| `inset-block-start` | <'top'> | `internal/layout/style_properties.go` | `TestBehaviorInsetBlockStartAnchorsAbsoluteBox` |
| `inset-inline` | <'top'>{1,2} | `internal/layout/style_properties.go` | `TestLogicalInset` |
| `inset-inline-end` | <'top'> | `internal/layout/style_properties.go` | `TestBehaviorInsetInlineEndAnchorsAbsoluteBox` |
| `inset-inline-start` | <'top'> | `internal/layout/style_properties.go` | `TestBehaviorInsetInlineStartAnchorsAbsoluteBox` |
| `isolation` | <isolation-mode> | `internal/layout/style_advanced_props.go` | `TestBehaviorIsolationOpensGroup` |
| `justify-content` | normal \| <content-distribution> \| <overflow-position>? [ <content-position> \| left \| right ] | `internal/layout/style_properties.go` | `TestWebkitPrefixAliases`, `TestFlexDistributedJustifyKeepsGap`, `TestBehaviorJustifyContentCenterUsedOffset` |
| `justify-items` | normal \| stretch \| <baseline-position> \| <overflow-position>? [ <self-position> \| left \| right ] \| legacy \| legacy && [ left \| right \| center ] | `internal/layout/style_properties.go` | `TestBehaviorJustifyItemsCenterUsedOffset` |
| `justify-self` | auto \| <overflow-position>? [ normal \| <self-position> \| left \| right ] \| stretch \| <baseline-position> | `internal/layout/style_properties.go` | `TestBehaviorJustifySelfCenterUsedOffset` |
| `left` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestCaptionSideParse`, `TestTableCellClipsTransformedContent`, `TestBehaviorLeftShiftsRelativeBox` |
| `letter-spacing` | normal \| <length> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorLetterSpacingWidensText` |
| `line-break` | auto \| loose \| normal \| strict \| anywhere | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorLineBreakAnywhereWrapsToken` |
| `line-clamp` | none \| [<'max-lines'> \|\| <'block-ellipsis'>] -webkit-legacy? | `internal/layout/style_advanced_props.go` | `TestLineClampClearsWebkitBoxNowrap`, `TestWaveBTextTruncationAndClamping`, `TestBehaviorLineClampTwoLimitsLines` |
| `line-height` | normal \| <number> \| <length> \| <percentage> \| inherit | `internal/layout/style.go`, `internal/layout/style_properties.go` | `TestMarginCollapse`, `TestExplicitLineHeightAllowsNegativeHalfLeading`, `TestNormalLineHeightUsesFaceMetrics`, `TestBehaviorLineHeightExplicitEnlargesBox` |
| `list-style` | [ <'list-style-type'> \|\| <'list-style-position'> \|\| <'list-style-image'> ] \| inherit | `internal/layout/style_properties.go` | `TestBehaviorListStyleNoneHidesMarker`, `TestBehaviorListStyleShorthandImageEmitsImageMarker` |
| `list-style-image` | <uri> \| none \| inherit | `internal/layout/style_paint_props.go` | `TestBehaviorListStyleImageBadURLFallsBackToMarker`, `TestBehaviorListStyleImageResolvedEmitsImageMarker` |
| `list-style-position` | inside \| outside \| inherit | `internal/layout/style_properties.go` | `TestBehaviorListStylePositionInsideVsOutside` |
| `list-style-type` | disc \| circle \| square \| decimal \| decimal-leading-zero \| lower-roman \| upper-roman \| lower-greek \| lower-latin \| upper-latin \| armenian \| georgian \| lower-alpha \| upper-alpha \| none \| inherit | `internal/layout/style_properties.go` | `TestBehaviorListStyleTypeDecimalEmitsNumbers` |
| `margin` | <margin-width>{1,4} \| inherit | `internal/layout/style_properties.go` | `TestBlockWidthsAndMargins`, `TestMarginCollapse` |
| `margin-block` | <'margin-top'>{1,2} | `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestLogicalMargin` |
| `margin-block-end` | <'margin-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorMarginBlockEndOffsetsSibling` |
| `margin-block-start` | <'margin-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorMarginBlockStartOffsetsSibling` |
| `margin-bottom` | <margin-width> \| inherit | `internal/layout/style_properties.go` | `TestBlockWidthsAndMargins`, `TestMarginCollapse` |
| `margin-inline` | <'margin-top'>{1,2} | `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestLogicalMargin` |
| `margin-inline-end` | <'margin-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorMarginInlineEndShrinksAutoWidth` |
| `margin-inline-start` | <'margin-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestCSSPartialRemainingPagedWritingMode`, `TestBehaviorMarginInlineStartOffsetsSibling` |
| `margin-left` | <margin-width> \| inherit | `internal/layout/style_properties.go` | `TestBlockWidthsAndMargins`, `TestMarginCollapse`, `TestBehaviorMarginLeftOffsetsSibling` |
| `margin-right` | <margin-width> \| inherit | `internal/layout/style_properties.go` | `TestBlockWidthsAndMargins`, `TestMarginCollapse` |
| `margin-top` | <margin-width> \| inherit | `internal/layout/style_properties.go` | `TestBlockWidthsAndMargins`, `TestCascadeEngineSupportsPropertyValues`, `TestMarginCollapse`, `TestBehaviorMarginTopOffsetsSibling` |
| `margin-trim` | none \| block \| [ block-start \|\| block-end ] | `internal/layout/style_advanced_props.go` | `TestBehaviorMarginTrimTrimsFirstChild` |
| `mask` | url(...) \| none; supported mask mode, position, size, repeat, origin, clip, and composite values | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskShorthandCreatesElementGroup`, `TestMaskShorthandCSSWideKeywords`, `TestMaskShorthandResolvesCustomPropertiesAndKeepsLonghandOrder` |
| `mask-border` | <'mask-border-source'> \|\| <'mask-border-slice'> [ / <'mask-border-width'>? [ / <'mask-border-outset'> ]? ]? \|\| <'mask-border-repeat'> \|\| <'mask-border-mode'> | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestMaskBorderShorthandExpansionAndSourceOrder`, `TestMaskBorderShorthandWideKeywordResetsToInitial`, `TestMaskBorderShorthandResolvesCustomProperties`, `TestBehaviorMaskBorderVendorAliasUsesCSSSourceOrder`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-mode` | alpha \| luminance | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-outset` | <length [0,∞]> \| <number [0,∞]>, one to four values | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderInvalidValuesKeepEarlier`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-repeat` | stretch \| repeat \| round \| space, one or two values | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestBehaviorMaskBorderAliasReachesGroup`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-slice` | one to four non-negative <number> or <percentage> values, plus the optional fill keyword | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestBehaviorMaskBorderAliasReachesGroup`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-source` | none \| url(...) \| image-set(...) whose chosen candidate is url(...) | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderInvalidValuesKeepEarlier`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-border-width` | auto \| <length [0,∞]> \| <number [0,∞]>, one to four values | `internal/layout/mask_border.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go`, `internal/layout/inline.go`, `internal/layout/mask_border_raster.go` | `TestBehaviorMaskBorderDescriptorNotPaint`, `TestBehaviorMaskBorderAliasReachesGroup`, `TestBehaviorMaskBorderInvalidValuesKeepEarlier`, `TestComposeMaskBorderRasterOnGroup` |
| `mask-clip` | border-box \| padding-box \| content-box \| fill-box \| stroke-box \| view-box \| no-clip | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint` |
| `mask-composite` | add \| subtract \| intersect \| exclude | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint`, `TestBehaviorMaskVendorAliasesReachMaskGroup` |
| `mask-image` | url(...) \| none | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint`, `TestBehaviorMaskMatchSourceUsesImageAlpha`, `TestBehaviorMaskGradientBytesOnGroup`, `TestBehaviorMaskImageSetPicks1x` |
| `mask-mode` | match-source \| alpha \| luminance | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint`, `TestRasterizeMaskUsesSourceLuminanceForMatchSource` |
| `mask-origin` | border-box \| padding-box \| content-box \| fill-box \| stroke-box \| view-box | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint` |
| `mask-position` | CSS position keywords, lengths, and percentages, plus 4-value edge-offset positions such as left 10px top 20px | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint` |
| `mask-repeat` | repeat \| repeat-x \| repeat-y \| no-repeat \| space \| round \| one or two of repeat, space, round, no-repeat | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint` |
| `mask-size` | auto \| cover \| contain \| one or two non-negative CSS lengths or percentages | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | `TestBehaviorMaskLayerPropertiesAttachToAllElementPaint` |
| `max-block-size` | <'max-width'> | `internal/layout/style_properties.go` | `TestLogicalSize`, `TestBehaviorMaxBlockSizeClampsUsedHeight` |
| `max-height` | <length> \| <percentage> \| none \| inherit | `internal/layout/style_properties.go` | `TestBehaviorMaxHeightClampsUsedHeight` |
| `max-inline-size` | <'max-width'> | `internal/layout/style_properties.go` | `TestLogicalSize`, `TestBehaviorMaxInlineSizeClampsUsedWidth` |
| `max-lines` | auto \|\| <integer [1,∞]> | `internal/layout/style_advanced_props.go` | `TestWaveBTextTruncationAndClamping`, `TestBehaviorMaxLinesTwoLimitsLines` |
| `max-width` | <length> \| <percentage> \| none \| inherit | `internal/layout/style_properties.go` | `TestBehaviorMaxWidthClampsUsedWidth` |
| `min-block-size` | <'min-width'> | `internal/layout/style_properties.go` | `TestLogicalSize`, `TestBehaviorMinBlockSizeClampsUsedHeight` |
| `min-height` | <length> \| <percentage> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorMinHeightClampsUsedHeight` |
| `min-inline-size` | <'min-width'> | `internal/layout/style_properties.go` | `TestLogicalSize`, `TestBehaviorMinInlineSizeClampsUsedWidth` |
| `min-width` | <length> \| <percentage> \| inherit | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestBehaviorMinWidthClampsUsedWidth`, `TestBehaviorMinWidthClampsUsedWidthReference` |
| `mix-blend-mode` | <blend-mode> \| plus-lighter | `internal/layout/style_advanced_props.go` | `TestBehaviorMixBlendModeOpensGroup` |
| `object-fit` | fill \| none \| [contain \| cover] \|\| scale-down | `internal/layout/style_image_adjust_props.go` | `TestBehaviorObjectFitContainKeepsRatio` |
| `object-position` | <position> | `internal/layout/style_image_adjust_props.go` | `TestBehaviorObjectPositionAlignsNoneImage` |
| `object-view-box` | none \| <basic-shape-rect> | `internal/layout/style_image_adjust_props.go` | `TestBehaviorObjectViewBoxXywhCropsImage` |
| `opacity` | <opacity-value> | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestBehaviorOpacityFoldsIntoPaintOps` |
| `order` | <integer> | `internal/layout/style_properties.go` | `TestWebkitPrefixAliases`, `TestBehaviorOrderReorderUsedPosition` |
| `orphans` | <integer> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorOrphansNoFragmentationEffect`, `TestFragmentAssessOrphansWidowsAnonStripNoEffect`, `TestOrphansKeepsLinesInFirstColumn`, `TestSnapAnonBandForOrphans` |
| `outline` | [ <'outline-color'> \|\| <'outline-style'> \|\| <'outline-width'> ] \| inherit | `internal/layout/style_paint_props.go` | `TestCurrentColor`, `TestOutlineDoesNotInherit`, `TestOutlineOmittedWidthUsesMedium`, `TestOutlineParse`, `TestOutlinePaintsInflatedRect`, `TestBehaviorOutlineShorthandPaintsOutsideBorderBox` |
| `outline-color` | <color> \| invert \| inherit | `internal/layout/style_paint_props.go` | `TestCurrentColor`, `TestOutlineParse`, `TestOutlinePaintsInflatedRect`, `TestBehaviorOutlineColorPaintsStrokeColor` |
| `outline-offset` | <length> | `internal/layout/style_paint_props.go` | `TestOutlineParse`, `TestOutlinePaintsInflatedRect`, `TestBehaviorOutlineOffsetInflatesRect` |
| `outline-style` | <border-style> \| inherit | `internal/layout/style_paint_props.go` | `TestOutlineParse`, `TestOutlinePaintsInflatedRect`, `TestBehaviorOutlineStyleDashedExpandsSegments` |
| `outline-width` | <border-width> \| inherit | `internal/layout/style_paint_props.go` | `TestOutlineParse`, `TestOutlineOmittedWidthUsesMedium`, `TestOutlinePaintsInflatedRect`, `TestBehaviorOutlineWidthSetsStrokeWidth` |
| `overflow` | visible \| hidden \| scroll \| auto \| inherit | `internal/layout/style_properties.go` | `TestBehaviorOverflowHiddenClipsTallChild`, `TestBehaviorOverflowVisibleKeepsTallChild` |
| `overflow-block` | visible \| hidden \| clip \| scroll \| auto | `internal/layout/style_overflow_logical.go` | `TestBehaviorOverflowBlockClipsTallChild` |
| `overflow-clip-margin` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go`, `internal/layout/style_paint_props.go` | `TestBehaviorOverflowClipMarginExpandsClip` |
| `overflow-clip-margin-block` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginBlockKeepsBelowChild` |
| `overflow-clip-margin-block-end` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginBlockEndKeepsBelowChild` |
| `overflow-clip-margin-block-start` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginBlockStartKeepsAboveChild` |
| `overflow-clip-margin-bottom` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginBottomKeepsBelowChild` |
| `overflow-clip-margin-inline` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginInlineKeepsWideChild` |
| `overflow-clip-margin-inline-end` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginInlineEndKeepsWideChild` |
| `overflow-clip-margin-inline-start` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginInlineStartKeepsLeftChild` |
| `overflow-clip-margin-left` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestBehaviorOverflowClipMarginLeftKeepsLeftChild` |
| `overflow-clip-margin-right` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestWaveDTextDecorationSkipInkAndOverflowClipMargin`, `TestBehaviorOverflowClipMarginRightKeepsWideChild` |
| `overflow-clip-margin-top` | <visual-box> \|\| <length> | `internal/layout/style_advanced_props.go` | `TestWaveDTextDecorationSkipInkAndOverflowClipMargin`, `TestBehaviorOverflowClipMarginTopKeepsAboveChild` |
| `overflow-inline` | visible \| hidden \| clip \| scroll \| auto | `internal/layout/style_overflow_logical.go` | `TestBehaviorOverflowInlineClipsWideChild` |
| `overflow-wrap` | normal \| break-word \| anywhere | `internal/layout/style_properties.go` | `TestBehaviorOverflowWrapBreakWordWrapsToken` |
| `overflow-x` | visible \| hidden \| clip \| scroll \| auto | `internal/layout/style_properties.go` | `TestBehaviorOverflowXHiddenClipsWideChild` |
| `overflow-y` | visible \| hidden \| clip \| scroll \| auto | `internal/layout/style_properties.go` | `TestBehaviorOverflowYScrollClipsTallChild` |
| `padding` | <padding-width>{1,4} \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `padding-block` | <'padding-top'>{1,2} | `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestLogicalPadding` |
| `padding-block-end` | <'padding-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorPaddingBlockEndExpandsBox` |
| `padding-block-start` | <'padding-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestCSSPartialRemainingPagedWritingMode`, `TestBehaviorPaddingBlockStartOffsetsContent` |
| `padding-bottom` | <padding-width> \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `padding-inline` | <'padding-top'>{1,2} | `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestLogicalPadding` |
| `padding-inline-end` | <'padding-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorPaddingInlineEndExpandsBox` |
| `padding-inline-start` | <'padding-top'> | `internal/layout/style_logical_axes.go`, `internal/layout/style_logical_box.go`, `internal/layout/style_properties.go` | `TestBehaviorPaddingInlineStartOffsetsContent` |
| `padding-left` | <padding-width> \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox`, `TestBehaviorPaddingLeftOffsetsContent` |
| `padding-right` | <padding-width> \| inherit | `internal/layout/style_properties.go` | `TestPaddingBorderBox` |
| `padding-top` | <padding-width> \| inherit | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestPaddingBorderBox`, `TestBehaviorPaddingTopOffsetsContent` |
| `perspective` | none \| <length [0,∞]> | `internal/layout/style_properties.go` | `TestBehaviorPerspectiveNoLayoutEffect`, `TestPerspBuildRotateYNarrowsOrtho`, `TestPerspBuildAncestorPerspectiveForeshortens`, `TestPerspBuildPerspectiveFunctionAppliesOwnDistance`, `TestPerspBuildPreserveComposedSkipsProjection`, `TestPerspPerspectiveNoEffectOn2DChild`, `TestPerspPerspectiveFunctionDepthless`, `TestPerspRotate3DOrthographic` |
| `perspective-origin` | <position> | `internal/layout/style_properties.go` | `TestBehaviorPerspectiveOriginNoLayoutEffect`, `TestOriginBuildDefaultIsCenter`, `TestOriginBuildPercentages`, `TestOriginBuildKeywords`, `TestOriginBuildLengths`, `TestOriginBuildVanishingPointShiftsProjection`, `TestOriginBuildShiftsProjectedOutput`, `TestPerspPerspectiveOriginNoEffect` |
| `place-content` | <'align-content'> <'justify-content'>? | `internal/layout/style_properties.go` | `TestPlaceShorthands`, `TestFlexPlaceContentDistributes` |
| `place-items` | <'align-items'> <'justify-items'>? | `internal/layout/style_properties.go` | `TestPlaceShorthands`, `TestGridPlaceItemsCenterShrinksItems` |
| `place-self` | <'align-self'> <'justify-self'>? | `internal/layout/style_properties.go` | `TestPlaceShorthands`, `TestGridPlaceSelfEndShrinksItem` |
| `position` | static \| relative \| absolute \| fixed \| inherit | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestPositionLiteFixtureReservesOverlaySpace`, `TestBehaviorPositionRelativeKeepsFlow` |
| `position-anchor` | auto \| <anchor-name> | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestAnchorInvalidLeavesPrevious`, `TestBehaviorPositionAreaPlacesAboveAnchor` |
| `position-area` | none \| <position-area> | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestBehaviorPositionAreaPlacesAboveAnchor` |
| `position-try` | <'position-try-order'>? <'position-try-fallbacks'> | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestBehaviorPositionTryShorthandExpands` |
| `position-try-fallbacks` | none \| [ [<dashed-ident> \|\| <try-tactic>] \| <position-area> ]# | `internal/layout/anchor_position.go`, `internal/css/at_position_try.go` | `TestBehaviorPositionTryAlwaysKeepsBase`, `TestPositionTryAtRule` |
| `position-try-order` | normal \| most-width \| most-height \| most-block-size \| most-inline-size | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestBehaviorPositionTryOrderMostInlineSizePicksWide` |
| `position-visibility` | always \| anchor-valid \| anchor-visible \| no-overflow | `internal/layout/anchor_position.go`, `internal/layout/style_cascade.go` | `TestBehaviorPositionVisibilityNoOverflow` |
| `print-color-adjust` | economy \| exact | `internal/layout/style_color_adjust_props.go` | `TestColorAdjustPropsAcceptLegalKeywords`, `TestColorAdjustPropsReachRestPass`, `TestColorAdjustPropsRejectIllegalKeywords` |
| `quotes` | [<string> <string>]+ \| none \| inherit | `internal/layout/style_paint_props.go` | `TestQuotes`, `TestBehaviorQuotesEmitGeneratedMarks` |
| `region-fragment` | auto \| break | `internal/layout/named_flow.go`, `internal/layout/layout_flow.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorNamedFlowRegionFragmentAutoVsBreak`, `TestBehaviorNamedFlowPropertyStorage` |
| `right` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestAuthorOverridesButtonUADefaults`, `TestBackgroundPositionUsesCSSPxIntrinsicSize`, `TestCaptionSideParse`, `TestTableCellClipsTransformedContent`, `TestBehaviorRightShiftsRelativeBox` |
| `rotate` | none \| <angle> \| [ x \| y \| z \| <number>{3} ] && <angle> | `internal/layout/style_leftovers.go`, `internal/layout/style_properties.go` | `TestBehaviorRotatePropertyBakesMatrix` |
| `row-gap` | normal \| <length-percentage [0,∞]> \| <line-width> | `internal/layout/style_gap_props.go`, `internal/layout/style_properties.go` | `TestGridRowGapVsColumnGap`, `TestBehaviorRowGapGridUsedSpacing` |
| `row-rule` | <gap-rule-list> \| <gap-auto-rule-list> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRulePaintedBetweenRows` |
| `row-rule-break` | none \| normal \| intersection | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleBreakNoneStillPaints` |
| `row-rule-color` | <line-color-list> \| <auto-line-color-list> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleColorBlue` |
| `row-rule-inset` | <'column-rule-inset-cap'> [ / <'column-rule-inset-junction'> ]? | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-inset-cap` | <inset-value> <inset-value>? | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-inset-cap-end` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-inset-cap-start` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-inset-end` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-inset-junction` | <inset-value> <inset-value>? | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `row-rule-inset-junction-end` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `row-rule-inset-junction-start` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `row-rule-inset-start` | <inset-value> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `row-rule-style` | <line-style-list> \| <auto-line-style-list> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleDashedSegments` |
| `row-rule-visibility-items` | all \| around \| between \| normal | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleVisibilityAround` |
| `row-rule-width` | <line-width-list> \| <auto-line-width-list> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleWidthCurrentColor` |
| `ruby-align` | start \| center \| space-between \| space-around | `internal/layout/style_paint_props.go` | `TestBehaviorRubyAlignNoPaintEffect`, `TestBehaviorRubyLayoutAlignStartVsCenter`, `TestBehaviorRubyAlignCentersNarrowAnnotation`, `TestBehaviorRubyAlignWideAnnotationFillsColumn` |
| `ruby-merge` | separate \| merge \| auto | `internal/layout/style_paint_props.go` | `TestBehaviorRubyMergeNoPaintEffect`, `TestBehaviorRubyLayoutMergeNarrowsSharedSpace`, `TestBehaviorRubyMergeSharesAnnotationSpace` |
| `ruby-overhang` | auto \| spaces | `internal/layout/style_paint_props.go` | `TestBehaviorRubyOverhangNoPaintEffect`, `TestBehaviorRubyLayoutOverhangSpacesWidensSegment`, `TestBehaviorRubyOverhangGrowsBaseWhenForbidden` |
| `ruby-position` | [ alternate \|\| [ over \| under ] ] \| inter-character | `internal/layout/style_paint_props.go` | `TestBehaviorRubyPositionNoPaintEffect`, `TestBehaviorRubyLayoutStacksAnnotationOverBase`, `TestBehaviorRubyLayoutPositionUnderStacksBelow`, `TestBehaviorRubyPositionStacksOverAndUnder` |
| `rule` | <'column-rule'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRulePaintsBothAxes` |
| `rule-break` | <'column-rule-break'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleBreakNoneStillPaints` |
| `rule-color` | <'column-rule-color'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRuleLonghandsBothAxes` |
| `rule-inset` | <'column-rule-inset'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `rule-inset-cap` | <'column-rule-inset-cap'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `rule-inset-end` | <'column-rule-inset-end'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `rule-inset-junction` | <'column-rule-inset-junction'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestRowRuleBadTokenKeepsEarlier` |
| `rule-inset-start` | <'column-rule-inset-start'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleInsetCap` |
| `rule-overlap` | row-over-column \| column-over-row | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRuleOverlapOrdersLines` |
| `rule-style` | <'column-rule-style'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRuleLonghandsBothAxes` |
| `rule-visibility-items` | <'column-rule-visibility-items'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRowRuleVisibilityAround` |
| `rule-width` | <'column-rule-width'> | `internal/layout/style_gap_rules.go`, `internal/layout/gap_rules.go`, `internal/layout/grid.go`, `internal/layout/flex.go` | `TestBehaviorRuleLonghandsBothAxes` |
| `scale` | none \| [ <number> \| <percentage> ]{1,3} | `internal/layout/style_leftovers.go`, `internal/layout/style_properties.go` | `TestTextSpacingTrimApply`, `TestBehaviorScalePropertyScalesPaint` |
| `scroll-margin` | <length>{1,4} | `internal/layout/style_paint_props.go` | `TestBehaviorScrollMarginNoPaintEffect` |
| `scroll-margin-bottom` | <length> | `internal/layout/style_paint_props.go` | `TestBehaviorScrollMarginBottomNoPaintEffect` |
| `scroll-margin-left` | <length> | `internal/layout/style_paint_props.go` | `TestBehaviorScrollMarginLeftNoPaintEffect` |
| `scroll-margin-right` | <length> | `internal/layout/style_paint_props.go` | `TestBehaviorScrollMarginRightNoPaintEffect` |
| `scroll-margin-top` | <length> | `internal/layout/style_paint_props.go` | `TestBehaviorScrollMarginTopNoPaintEffect` |
| `shape-inside` | auto \| outside-shape \| [ <basic-shape> \|\| <shape-box> ] (circle(), ellipse(), inset(); polygon(), path(), and url() are rejected) | `internal/layout/style_shape_inside.go`, `internal/layout/shape_inside.go`, `internal/layout/style_cascade.go` | `TestBehaviorShapeInsideCircleTapersLines`, `TestBehaviorShapeInsideShapeBoxMovesReference`, `TestBehaviorShapeInsideOutsideShapeCopiesOutside`, `TestBehaviorShapeInsideStacksWithFloatExclusion`, `TestBehaviorShapeInsideDoesNotInherit`, `TestBehaviorShapeInsideNoLayoutEffect` |
| `shape-margin` | <length-percentage [0,∞]> | `internal/layout/style_shape_props.go` | `TestBehaviorShapeMarginExpandsExclusion` |
| `shape-outside` | none \| [ <basic-shape> \|\| <shape-box> ] \| <image> | `internal/layout/style_shape_props.go` | `TestBehaviorShapeOutsideNarrowsWrap` |
| `shape-padding` | <length-percentage [0,∞]> | `internal/layout/style_shape_inside.go`, `internal/layout/shape_inside.go`, `internal/layout/style_cascade.go` | `TestBehaviorShapeInsidePaddingShrinksMiddleLine`, `TestBehaviorShapeInsidePaddingAloneNoEffect`, `TestBehaviorShapeInsideOversizePaddingKeepsZeroWidthLines` |
| `shape-rendering` | auto \| optimizeSpeed \| crispEdges \| geometricPrecision | `internal/layout/style_paint_props.go` | `TestBehaviorShapeRenderingNoPaintEffect`, `TestBehaviorSvgBakeShapeRenderingGap`, `TestPaintOutputShapeRenderingCrispSnapsRaster` |
| `stroke` | <paint> | `internal/layout/style_paint_props.go` | `TestBehaviorStrokeBakesPaint` |
| `stroke-dasharray` | none \| <dasharray> | `internal/layout/style_leftovers.go`, `internal/layout/style_paint_props.go` | `TestLeftoversPropsWave5`, `TestBehaviorSvgBakeDasharrayDashesStroke` |
| `stroke-dashoffset` | <length-percentage> \| <number> | `internal/layout/style_leftovers.go`, `internal/layout/style_paint_props.go` | `TestLeftoversPropsWave5`, `TestBehaviorSvgBakeDashoffsetShiftsPhase` |
| `stroke-linecap` | butt \| round \| square | `internal/layout/style_leftovers.go`, `internal/layout/style_paint_props.go` | `TestLeftoversPropsWave5`, `TestBehaviorSvgBakeLinecapRoundsEnds` |
| `stroke-linejoin` | miter \| round \| bevel | `internal/layout/style_leftovers.go`, `internal/layout/style_paint_props.go` | `TestLeftoversPropsWave5`, `TestBehaviorSvgBakeLinejoinRoundsCorners` |
| `stroke-miterlimit` | <number> | `internal/layout/style_leftovers.go`, `internal/layout/style_paint_props.go` | `TestLeftoversPropsWave5`, `TestBehaviorSvgBakeMiterlimitClipsSpike` |
| `stroke-opacity` | <'opacity'> | `internal/layout/style_paint_props.go` | `TestBehaviorStrokeOpacityBakesPaint` |
| `stroke-width` | <length-percentage> \| <number> | `internal/layout/style_paint_props.go` | `TestBehaviorStrokeWidthBakesPaint` |
| `tab-size` | <number [0,∞]> \| <length [0,∞]> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorTabSizeValueChangesWidth`, `TestBehaviorTabSizeSameValueDeterministic` |
| `table-layout` | auto \| fixed \| inherit | `internal/layout/style_properties.go` | `TestTableLayoutFixedIgnoresContentMax`, `TestBehaviorTableLayoutFixedEqualShare` |
| `text-align` | left \| right \| center \| justify \| inherit | `internal/layout/style_properties.go` | `TestTextAlignJustify`, `TestBehaviorTextAlignCentersLine` |
| `text-align-all` | start \| end \| left \| right \| center \| <string> \| justify \| match-parent | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestBehaviorTextAlignAllCentersLine` |
| `text-align-last` | auto \| start \| end \| left \| right \| center \| justify \| match-parent | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorTextAlignLastCentersLastLine` |
| `text-anchor` | start \| middle \| end | `internal/layout/style_paint_props.go` | `TestBehaviorTextAnchorNoPaintEffect`, `TestSvgTextTextAnchorMiddleAnchorsSvgText` |
| `text-autospace` | normal \| <autospace> \| auto | `internal/layout/style_text_spacing_props.go` | `TestTextAutospaceIdeographAlpha`, `TestBehaviorTextAutospaceWidensIdeographAlpha` |
| `text-box` | normal \| <'text-box-trim'> \|\| <'text-box-edge'> | `internal/layout/style_text_box_props.go` | `TestTextBoxTrimBothShrinksHalfLeading`, `TestBehaviorTextBoxShorthandMatchesLonghands` |
| `text-box-edge` | auto \| <text-edge> | `internal/layout/style_text_box_props.go` | `TestTextBoxTrimBothShrinksHalfLeading`, `TestBehaviorTextBoxEdgeRetargetsTrimmedMetrics` |
| `text-box-trim` | none \| trim-start \| trim-end \| trim-both | `internal/layout/style_text_box_props.go` | `TestTextBoxTrimBothShrinksHalfLeading`, `TestBehaviorTextBoxTrimsHalfLeading`, `TestBehaviorTextBoxTrimStartShrinksBox` |
| `text-combine-upright` | none \| all \| [ digits <integer [2,4]>? ] | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextCombineUpright`, `TestBehaviorTextCombineUprightCombinesDigits` |
| `text-decoration` | none \| [ underline \|\| overline \|\| line-through \|\| blink ] \| inherit | `internal/layout/style_properties.go` | `TestBoldUnderline` |
| `text-decoration-color` | <color> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextDecorationColorOverridesInk` |
| `text-decoration-inset` | <length-percentage>{1,2} \| auto | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextDecorationInset`, `TestBehaviorTextDecorationInsetTrimsStroke` |
| `text-decoration-line` | none \| [ underline \|\| overline \|\| line-through \|\| blink ] \| spelling-error \| grammar-error | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextDecorationLineUnderlineStrokes` |
| `text-decoration-skip` | none \| auto | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextDecorationSkipShorthand`, `TestBehaviorTextDecorationSkipShorthandSpacesSplits` |
| `text-decoration-skip-box` | none \| all | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextDecorationSkipLonghands`, `TestBehaviorTextDecorationSkipBoxBreaksAtChrome` |
| `text-decoration-skip-ink` | auto \| none \| all | `internal/layout/style_advanced_props.go` | `TestWaveDTextDecorationSkipInkAndOverflowClipMargin`, `TestBehaviorTextDecorationSkipInkSplitsDescenders` |
| `text-decoration-skip-self` | auto \| skip-all \| [ skip-underline \|\| skip-overline \|\| skip-line-through ] \| no-skip | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextDecorationSkipLonghands`, `TestBehaviorTextDecorationSkipSelfSuppressesStroke` |
| `text-decoration-skip-spaces` | none \| all \| [ start \|\| end ] | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextDecorationSkipLonghands`, `TestBehaviorTextDecorationSkipSpacesSplitsWords` |
| `text-decoration-style` | solid \| double \| dotted \| dashed \| wavy | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextDecorationStyleDashedSplits` |
| `text-decoration-thickness` | auto \| from-font \| <length-percentage> \| <line-width> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextDecorationThicknessSetsWidth` |
| `text-emphasis` | <'text-emphasis-style'> \|\| <'text-emphasis-color'> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextEmphasisPropsReachWave3`, `TestBehaviorTextEmphasisShorthandPaintsDots` |
| `text-emphasis-color` | <color> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextEmphasisPropsReachWave3`, `TestBehaviorTextEmphasisColorPaintsRedDots` |
| `text-emphasis-position` | [ over \| under ] && [ right \| left ]? | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextEmphasisPropsReachWave3`, `TestBehaviorTextEmphasisPositionMovesDotsBelow` |
| `text-emphasis-skip` | spaces \|\| punctuation \|\| symbols \|\| narrow | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextEmphasisPropsReachWave3`, `TestBehaviorTextEmphasisSkipNoLayoutEffect`, `TestEmphasisSkipPunctuationSkipsComma`, `TestEmphasisSkipNarrowSkipsNarrow` |
| `text-emphasis-style` | none \| [ [ filled \| open ] \|\| [ dot \| circle \| double-circle \| triangle \| sesame ] ] \| <string> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextEmphasisPropsReachWave3`, `TestBehaviorTextEmphasisStyleOpenStrokesMarks` |
| `text-group-align` | none \| start \| end \| left \| right \| center | `internal/layout/style_text_spacing_props.go` | `TestTextGroupAlignCenter`, `TestBehaviorTextGroupAlignEndMovesLine` |
| `text-indent` | <length> \| <percentage> \| inherit | `internal/layout/style_properties.go` | `TestTextIndentInheritsAndShiftsFirstLine`, `TestBehaviorTextIndentShiftsFirstLine` |
| `text-justify` | [ auto \| none \| inter-word \| inter-character \| ruby ] \|\| no-compress | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorTextJustifyNoneDisablesExpansion`, `TestBehaviorTextJustifyInterWordExpandsWordGaps`, `TestBehaviorTextJustifyInterCharacterExpandsWithinWord` |
| `text-orientation` | mixed \| upright \| sideways | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsTextOrientation`, `TestBehaviorTextOrientationSidewaysRotatesRun` |
| `text-overflow` | [ clip \| ellipsis \| <string> \| fade \| <fade()> ]{1,2} | `internal/layout/style_advanced_props.go` | `TestWaveBTextTruncationAndClamping`, `TestBehaviorTextOverflowEllipsisTruncatesLine` |
| `text-shadow` | none \| <shadow># | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextShadowPaintsOffsetCopy` |
| `text-spacing` | none \| auto \| <spacing-trim> \|\| <autospace> | `internal/layout/style_text_spacing_props.go` | `TestTextAutospaceIdeographAlpha`, `TestBehaviorTextSpacingShorthandMirrorsTrim` |
| `text-spacing-trim` | <spacing-trim> \| auto | `internal/layout/style_text_spacing_props.go` | `TestTextAutospaceIdeographAlpha`, `TestBehaviorTextSpacingTrimStartTrimsLead`, `TestBehaviorTextSpacingTrimLeadingTrimsPunct` |
| `text-transform` | capitalize \| uppercase \| lowercase \| none \| inherit | `internal/layout/style_properties.go` | `TestBehaviorTextTransformUppercaseEmitsCaps` |
| `text-underline-offset` | auto \| <length-percentage> | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestBehaviorTextUnderlineOffsetShiftsStroke` |
| `text-underline-position` | auto \| [ from-font \| under ] \|\| [ left \| right ] | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextDecorationPropsWave3`, `TestUnderlinePositionUnderDropsStroke`, `TestUnderlinePositionLeftRightMoveStroke` |
| `text-wrap` | wrap \| nowrap \| auto \| balance \| stable (at most one mode and one style) | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextWrapBalanceMatchesChromeLineBreaks`, `TestDisplaySupportsTextWrapStyleAcceptance`, `TestBehaviorTextWrapNowrapKeepsSingleLine` |
| `text-wrap-mode` | wrap \| nowrap | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorTextWrapModeNowrapFoldsOntoWhiteSpace` |
| `text-wrap-style` | auto \| balance \| stable | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go`, `internal/layout/inline_balance.go` | `TestTextWrapBalanceMatchesChromeLineBreaks`, `TestTextWrapBalanceMatchesChromeWideWidths`, `TestTextWrapBalanceMatchesChromeEdgeCases`, `TestTextWrapBalanceLeavesLongParagraphAlone`, `TestTextWrapBalanceLeavesNowrapAlone`, `TestTextWrapBalanceCenterAlignmentUsesFullWidth`, `TestTextWrapStyleAcceptanceGate`, `TestDisplaySupportsTextWrapStyleAcceptance`, `TestBehaviorTextWrapStyleAutoMatchesNormal`, `TestAvoidShortWireRebreaksOrphan`, `TestAvoidShortHelperRebreaksOrphan`, `TestAvoidShortKeepsGoodBreaks`, `TestAvoidShortApplyGates` |
| `top` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestBackgroundLonghands`, `TestCaptionSideParse`, `TestMarginEdgeUnknown`, `TestBehaviorTopShiftsRelativeBox` |
| `transform` | none \| <transform-list> | `internal/layout/style_properties.go` | `TestWebkitPrefixAliases`, `TestBehaviorTransformTranslateMovesPaint`, `TestBehaviorTranslate3dAndScale3d`, `TestBehaviorRotateXNarrowsHeight`, `TestBehaviorMatrix3DRotateY` |
| `transform-box` | content-box \| border-box \| fill-box \| stroke-box \| view-box | `internal/layout/style_properties.go` | `TestBehaviorTransformBoxFillBoxNoPaintEffect` |
| `transform-origin` | [ left \| center \| right \| top \| bottom \| <length-percentage> ] \| [ left \| center \| right \| <length-percentage> ] [ top \| center \| bottom \| <length-percentage> ] <length>? \| [ [ center \| left \| right ] && [ center \| top \| bottom ] ] <length>? | `internal/layout/style_properties.go` | `TestBehaviorTransformOriginShiftsRotation` |
| `transform-style` | flat \| preserve-3d | `internal/layout/style_properties.go` | `TestBehaviorTransformStylePreserve3DNoPaintEffect`, `TestPerspTransformStyleFlatteningIdentical`, `TestPreserveBuildNestedCounterRotationDiverges`, `TestPreserveBuildFlatBoundaryClearsChain` |
| `transition` | comma-separated property with optional duration, easing, delay, and behavior | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestExpandTransitionShorthand`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (transition)` |
| `transition-behavior` | normal \| allow-discrete; comma-separated | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionBehaviorReachesOwnFrameFrames (allow-discrete)` |
| `transition-delay` | signed <time>; comma-separated | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (transition-delay)` |
| `transition-duration` | non-negative <time>; comma-separated | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (transition-duration)` |
| `transition-property` | all \| none \| custom-ident; comma-separated except none stands alone | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (transition-property)` |
| `transition-timing-function` | linear \| ease \| ease-in \| ease-out \| ease-in-out \| step-start \| step-end \| cubic-bezier() \| steps() \| linear() | `internal/layout/transition_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestTransitionDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTransitionPropertiesReachOwnFrameFrames (transition-timing-function)` |
| `translate` | none \| <length-percentage> [ <length-percentage> <length>? ]? | `internal/layout/style_leftovers.go`, `internal/layout/style_properties.go` | `TestBehaviorTranslatePropertyMovesPaint` |
| `trigger-scope` | none \| all \| one or more <dashed-ident> values | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogTriggerScopeLimitsOwnFrameActions (--fire)` |
| `unicode-bidi` | normal \| embed \| bidi-override \| inherit | `internal/layout/style_text_support_props.go` | `TestApplyTextSupportPropsUnicodeBidi` |
| `vertical-align` | baseline \| sub \| super \| top \| text-top \| middle \| bottom \| text-bottom \| <percentage> \| <length> \| inherit | `internal/layout/style_properties.go` | `TestTableCellVerticalAlignMiddle`, `TestBehaviorVerticalAlignSuperRaisesText` |
| `view-transition-class` | none \| one or more custom-ident values | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogViewTransitionClassChangesOwnFrameFrame (promo)` |
| `view-transition-group` | normal \| contain \| nearest \| custom-ident | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogViewTransitionGroupBuildsOwnFrameHierarchy (nearest)` |
| `view-transition-name` | none \| custom-ident | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogViewTransitionNameReachesOwnFrameFrames (card)` |
| `view-transition-scope` | none \| all | `internal/layout/animation_style.go`, `internal/layout/placed.go`, `layout/layout.go` | `TestAnimationDeclarationValueAccepted`, `TestAnimationTransitionCascadeRoundTrips`, `TestDisplayAnimationAndTransitionMetadata`; OwnFrame cases: `TestCatalogViewTransitionScopeLimitsOwnFrameFrames (all)` |
| `visibility` | visible \| hidden \| collapse \| inherit | `internal/layout/style_properties.go` | `TestVisibilityHidden`, `TestBehaviorVisibilityHiddenKeepsGeometry` |
| `white-space` | normal \| pre \| nowrap \| pre-wrap \| pre-line \| inherit | `internal/layout/style_properties.go` | `TestWhiteSpacePre`, `TestWhiteSpacePreWrap`, `TestBehaviorWhiteSpaceNowrapKeepsSingleLine` |
| `white-space-collapse` | collapse \| discard \| preserve \| preserve-breaks \| preserve-spaces \| break-spaces | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorWhiteSpaceCollapsePreserveKeepsSpaces` |
| `white-space-trim` | none \| discard-before \|\| discard-after \|\| discard-inner | `internal/layout/style_properties.go`, `internal/layout/style_text_props.go` | `TestTextPropsWave3`, `TestBehaviorWhiteSpaceTrimDiscardBeforeTrimsLeading`, `TestBehaviorWhiteSpaceTrimDiscardAfterTrimsTrailing`, `TestBehaviorWhiteSpaceTrimDiscardInnerTrimsBothEdges` |
| `widows` | <integer> \| inherit | `internal/layout/style_properties.go` | `TestBehaviorWidowsNoFragmentationEffect`, `TestBehaviorWidowsKeepsMinLinesInNextColumn` |
| `width` | <length> \| <percentage> \| auto \| inherit | `internal/layout/style_properties.go` | `TestApplyImageKeyBackgroundAlias`, `TestBoxSizingBorderBox`, `TestCascadeEngineSupportsPropertyValues`, `TestImageSet`, `TestOutlineParse`, `TestPseudoElementSelectorDoesNotApplyToHost`, `TestBehaviorWidthContentBoxUsedSize` |
| `word-break` | normal \| break-all \| keep-all \| manual \| auto-phrase \| break-word | `internal/layout/style_properties.go` | `TestBehaviorWordBreakKeepAllKeepsToken` |
| `word-spacing` | normal \| <length> \| inherit | `internal/layout/style_properties.go` | `TestWordSpacingInherits`, `TestWordSpacingWidensRuns`, `TestBehaviorWordSpacingWidensText` |
| `wrap-after` | auto \| line \| avoid \| avoid-line \| flex \| avoid-flex | `internal/layout/wrap_break.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorWrapBeforeLineBreaks`, `TestBehaviorNamedFlowPropertyStorage` |
| `wrap-before` | auto \| line \| avoid \| avoid-line \| flex \| avoid-flex | `internal/layout/wrap_break.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorWrapBeforeLineBreaks`, `TestBehaviorWrapAvoidSkipsSoftWrap`, `TestBehaviorWrapBeforeFlexBreaksLine`, `TestBehaviorWrapBreakNotPageBreak`, `TestBehaviorNamedFlowPropertyStorage` |
| `wrap-flow` | auto \| both \| start \| end \| minimum \| maximum \| clear | `internal/layout/exclusion.go`, `internal/layout/float.go`, `internal/layout/layout_flow.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorWrapFlowBothShortensLine`, `TestBehaviorWrapFlowKeywords`, `TestBehaviorNamedFlowPropertyStorage` |
| `wrap-inside` | auto \| avoid \| avoid-line | `internal/layout/wrap_break.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorWrapInsideAvoidFitsAndOverflows`, `TestBehaviorNamedFlowPropertyStorage` |
| `wrap-through` | wrap \| none | `internal/layout/exclusion.go`, `internal/layout/float.go`, `internal/layout/layout_flow.go`, `internal/layout/style_cascade.go`, `internal/layout/style_value_accept.go` | `TestBehaviorWrapThroughNoneIgnoresExclusion`, `TestBehaviorNamedFlowPropertyStorage` |
| `writing-mode` | horizontal-tb \| vertical-rl \| vertical-lr \| sideways-rl \| sideways-lr | `internal/layout/style_properties.go` | `TestWritingModeInherits`, `TestBehaviorWritingModeVerticalRotatesRun` |
| `z-index` | auto \| <integer> \| inherit | `internal/layout/style_properties.go` | `TestCascadeEngineSupportsPropertyValues`, `TestBehaviorZIndexOrdersOverlappingPaint` |

### 2.2 Partial (3)

| Property | Accepted values | Source | Named missing behavior |
|---|---|---|---|
| `-webkit-mask-composite` | source-over \| source-out \| source-in \| xor | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go` | The engine accepts source-over, source-in, source-out, and xor only; Chrome 155 also accepts source-atop, destination-over, destination-in, destination-out, destination-atop, clear, and copy. Mask pixel replay stays a host contract. |
| `break-inside` | auto \| avoid \| avoid-page \| avoid-column \| avoid-region | `internal/layout/style_properties.go` | break-inside keywords parse but no layout or paint pass reads them; the engine returns a drawing list without paginating, so avoidance has no observable effect (TestBehaviorBreakInsideAvoidColumnNoPageEffect). break-inside:avoid parses but multicol advances whole items without a straddle test at the distribution call, so avoidance equals auto (TestBreakInsideAvoidBalancedStraddleGap). |
| `mask-type` | luminance \| alpha | `internal/layout/style_cascade.go`, `internal/layout/style_mask_accept.go`, `internal/layout/style_mask_props.go`, `internal/layout/mask_paint.go`, `internal/layout/layout.go`, `bindings/wasm/drawing_list.go`, `internal/svg/mask.go` | mask-type cascades into ResolvedStyle.MaskType but no consumer reads it; the SVG rasterizer uses only the mask-type on the SVG source element. Pixel replay stays a host contract. |

### 2.3 Unsupported (219)

| Property | Named missing behavior |
|---|---|
| `-webkit-appearance` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-background-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-background-origin` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-background-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-text-size-adjust` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-text-stroke` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-text-stroke-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-text-stroke-width` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `-webkit-user-select` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `all` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `backdrop-filter` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `background-attachment` | Parsed and stored (BackgroundAttachment, style_properties.go:1297) but no layout or paint consumer reads it: background_image.go:343 documents fixed as intentionally painting as scroll. The engine has no scrolling viewport, so fixed/local versus scroll is unobservable. Only test is a stored-string assertion (TestBackgroundLonghands). |
| `background-tbd` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `baseline-shift` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `baseline-source` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-ellipsis` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-step` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-step-align` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-step-insert` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-step-round` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `block-step-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `bookmark-label` | Standalone outline model in bookmark_outline.go builds used labels (BuildBookmarkOutline); no cascade handler feeds it yet. |
| `bookmark-level` | Standalone outline model in bookmark_outline.go builds used nesting depths; no cascade handler feeds it yet. |
| `bookmark-state` | Standalone outline model in bookmark_outline.go carries used open/closed flags; no cascade handler feeds it yet. |
| `border-block-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-block-end-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-block-start-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-bottom-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-boundary` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-inline-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-inline-end-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-inline-start-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-left-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-limit` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-right-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-shape` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `border-top-clip` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `box-decoration-break` | Parsed and stored (BoxDecorationBreak, style_advanced_props.go:64) but no fragmenter exists in the engine: single display list, multicol snaps whole lines, page-break props are no-ops, so slice versus clone is unobservable. Requires block splitting plus per-fragment chrome in prependChrome before this value can mean anything. Only test is a stored-string assertion (TestWaveCBoxDecorationBreak). |
| `box-snap` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `caret` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `caret-animation` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `caret-shape` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `column-rule-visibility-items` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `continue` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `copy-into` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `cue` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `cue-after` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `cue-before` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `cx` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `cy` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `d` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `event-trigger` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `event-trigger-name` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `event-trigger-source` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `field-sizing` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-break` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-image` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-origin` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-position` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-repeat` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `fill-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `flex-line-count` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `float-defer` | No page-float model exists in this engine. Decision recorded in the shapes ledger: support only none (the initial behavior); integer and last stay unsupported. |
| `flood-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `flood-opacity` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `footnote-display` | Standalone collection in footnote.go numbers and groups footnotes; no footnote area exists yet so values move no box. |
| `footnote-policy` | Standalone grouping in footnote.go (GroupFootnotesByPolicy); no page splitter exists yet so no forced break is emitted. |
| `frame-sizing` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `glyph-orientation-vertical` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `image-animation` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `image-rendering` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `inline-sizing` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `input-security` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `interactivity` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `interest-delay` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `interest-delay-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `interest-delay-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `interpolate-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `lighting-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `line-fit-edge` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `line-grid` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `line-height-step` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `line-padding` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `line-snap` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `link-parameters` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `margin-break` | Parsed and stored (MarginBreak, style_properties.go:1658) but no page-break consumer reads it. Implement path: honor keep/discard on adjoining margins at fragment breaks in the break logic beside BreakInside (independent_blocks.go:112). Only test is a stored-string assertion (TestMarginBreakProperty). |
| `marker` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `marker-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `marker-mid` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `marker-side` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `marker-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `math-depth` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `math-shift` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `math-style` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `max-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `min-intrinsic-sizing` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `min-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `nav-down` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `nav-left` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `nav-right` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `nav-up` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset-anchor` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset-distance` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset-path` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset-position` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `offset-rotate` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overflow-anchor` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overlay` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overscroll-behavior` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overscroll-behavior-block` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overscroll-behavior-inline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overscroll-behavior-x` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `overscroll-behavior-y` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `paint-order` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `path-length` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pause` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pause-after` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pause-before` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pointer-timeline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pointer-timeline-axis` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `pointer-timeline-name` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `r` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `reading-flow` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `reading-order` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `rest` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `rest-after` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `rest-before` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `rx` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `ry` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-axis-lock` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-behavior` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-initial-target` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-block` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-block-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-block-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-inline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-inline-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-margin-inline-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-marker-group` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-block` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-block-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-block-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-bottom` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-inline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-inline-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-inline-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-left` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-right` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-padding-top` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-snap-align` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-snap-stop` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-snap-type` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-target-group` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-timeline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-timeline-axis` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scroll-timeline-name` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scrollbar-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scrollbar-gutter` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `scrollbar-width` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `shape-image-threshold` | No image shape is parsed; shape-outside rejects url(), so no alpha contour can feed a threshold. Decision recorded in the shapes ledger: do not add url() shapes here. |
| `size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `slider-orientation` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `spatial-navigation-action` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `spatial-navigation-contain` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `spatial-navigation-function` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `speak` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `speak-as` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stop-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stop-opacity` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `string-set` | Standalone named-string table in string_set.go with set-then-read semantics; no page-margin pipeline consumes it yet. |
| `stroke-align` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-alignment` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-break` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-color` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-dash-corner` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-dash-justify` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-dashadjust` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-dashcorner` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-image` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-origin` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-position` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-repeat` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `stroke-size` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `text-fit` | Parsed and stored (TextFit, style_text_spacing_props.go:36) but no scale-search consumer reads it; the apply arm records the same gap. Missing behavior: grow/shrink scale-to-fit. TestTextSpacingTrimApply asserts the stored string only. |
| `text-rendering` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `text-size-adjust` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-scope` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-activation-range` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-activation-range-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-activation-range-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-active-range` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-active-range-end` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-active-range-start` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-name` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `timeline-trigger-source` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `vector-effect` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `view-timeline` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `view-timeline-axis` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `view-timeline-inset` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `view-timeline-name` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-balance` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-duration` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-family` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-pitch` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-range` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-rate` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-stress` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `voice-volume` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `will-change` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `window-drag` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `word-space-transform` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `x` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `y` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |
| `zoom` | No apply handler in internal/layout; declaration is ignored (graceful degrade). Pinned upstream revision: 1f2ec8f74a80c14066b4c7d6822cee59f69fa03b. |

### 2.4 Intentionally ignored (8)

| Property | Reason |
|---|---|
| `appearance` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `caret-color` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `cursor` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `page` | Engine returns a continuous display list with no paged renderer; named pages only select among @page rules at page breaks, so the declaration is ignored by design (paged-media noop). |
| `pointer-events` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `resize` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `touch-action` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |
| `user-select` | Print engine has no pointer, caret, or form chrome; declaration is ignored by design (print-noop). |

## 3. Supported units

Status legend: Implemented and Partial as in §2; Not implemented means the unit is rejected or ignored. Resolution sites: `LengthToPt`
`internal/css/container.go:116` (`ex`/`ch` as 0.5em, `exChToEmFactor` at
container.go:103), `lengthBox` `style_values.go:734` (width/height/min/max),
`marginLen` `style_values.go:819` (margins/padding/letter-spacing), parse gate
`ParseLength` `values.go:110`.

| Unit | Status | Notes |
|------|--------|-------|
| `px` | Implemented | 1 px = 0.75 pt (96 px/in reference) |
| `pt` | Implemented | |
| `mm`, `cm`, `in` | Implemented | |
| `pc` | Implemented | `internal/css/container.go:152` |
| `em` | Implemented | relative to element font-size (font-size, margins, lengths) |
| `rem` | Implemented | 16 px reference (`internal/css/container.go:102,127`; layout margin path at `style_values.go:903`) |
| `%` | Implemented | containing block for box/margins; parent font-size for `font-size` |
| `vw`, `vh` | Partial | resolved from the containing block in box, margin, and padding lengths (`lengthBoxFromUnit`, `style_values.go:767-770`; `marginLenFromUnit`, `style_values.go:901-902`); not resolved for `font-size` |
| `dvh`, `svh`, `lvh` | Partial | viewport-height variants in box lengths (`viewportHeightUnitPt`, `style_values.go:789-802`) |
| `ex` | Partial | Resolved as 0.5em (`internal/css/container.go:126-129`, `exChToEmFactor` at `container.go:103`). Not font-metric x-height. |
| `ch` | Partial | Resolved as 0.5em (`chLengthPt`, `style_ch.go:26-28`), the same fallback `css.LengthToPt` uses. Not the font's digit-zero advance. |
| `calc()`, `min()`, `max()`, `clamp()` | Partial | Evaluated by `css.EvalMath` (`internal/css/math_expr.go:29`) from `calcLength` (`style_values.go:914`, call at `:929`): `+ - * /`, parentheses, comma arguments, nesting, and mixed units. Other functions stay invalid so a fallback can win. Test `TestMathLengthMinMaxClamp`. |
| `vmin`, `vmax` | Partial | Layout `vminVmaxPt` (`style_values.go:693`) on box and margin lengths (`style_properties.go:598,651`). `css.ParseLength` rejects them; used values are parsed in layout. Test `TestVminVmax`. |
| `dpi`-style | Not implemented | rejected at parse |

## 4. Supported selector syntax (cascade)

Status legend: Implemented and Partial as in §2; Not implemented means the selector or at-rule is parsed without a matching or rendering consumer. Evidence in `internal/css/`.

| Selector | Status | Notes / verified by |
|----------|--------|---------------------|
| Element (`h1`, `p`, …), class (`.foo`), ID (`#bar`) | Implemented | `parseCompoundCtx` `selector_parser.go:165`; matching `Match` `match.go:95`; test `css_test.go::TestMatch` |
| Universal (`*`) | Implemented | `writeStarPrefix` `selector_parser.go:116` |
| Descendant (`div p`), child (`ul > li`) | Implemented | combinators `combinatorFor` `has.go:147`; matching `leftmostStep` `match.go:185` |
| Sibling (`a + b`, `a ~ b`) | Implemented | next-sibling `+` and subsequent-sibling `~` (`css.Match`); test `TestSiblingCombinators` |
| Attribute (`[href]`, `[href="…"]`) | Partial | presence, exact `=`, word `~=`, substring `*=`, prefix `^=`, suffix `$=`, dash `|=`; ASCII `i` flag on valued selectors (`[attr=value i]`); no `s` flag. Tests `TestAttrWordAndSubstring`, `TestAttrPrefixSuffixDash`, `TestAttrIFlag` |
| `:first-child`, `:last-child`, `:nth-child(n)` | Implemented | `odd`/`even`/`an+b`/integer; tests `TestMatch`, `TestNthChildZebraSheet` |
| `:first-of-type`, `:last-of-type`, `:nth-of-type()`, `:nth-last-of-type()` | Implemented | 1-based index among same-tag siblings; `odd`/`even`/`an+b`/integer via `parseNthArg`/`matchNth` (`match.go:632/664`); invalid an+b never matches. Tests `TestFirstOfType`, `TestLastOfType`, `TestNthOfType`, `TestNthLastOfType` |
| `:link`, `:visited` | Partial | Match any `a` with a non-empty `href` (no visit history; `:visited` equals `:link`). Specificity counts as a class-level pseudo. Proof: `TestLinkVisitedPseudos`, `TestLinkPseudoColor` |
| `:hover`, `:active`, `:focus` | Partial (caller state) | Parsed onto the compound and matched only against the caller-supplied `MatchState` ids (`matchStatePseudo`, `match.go:447`), so `a:hover` never degrades to bare `a`. Test `TestMatchStateFocusHoverActive` |
| `::before` / `::after` | Partial / Implemented | `MatchPseudo` plus generated content (`pseudo_content.go`): quoted strings and `attr()`. Host-element rules do not apply to the host |
| `!important` | Implemented | `isImportant` `values.go:75`; separate cascade tier `style_cascade.go:1207-1211`; test `css_test.go::TestParseImportant` |
| Specificity (ID > class > element), inline `style` wins, `!important` overrides | Implemented | `Specificity` `css.go:1009`; inline style priority `style_cascade.go:655`; test `css_test.go::TestSpecificity` |
| `@media print` / `screen` filtering | Implemented | `MediaMatches` (`css/media.go`); the caller passes `Media` (`settings.MediaScreen` or `settings.MediaPrint`, `internal/settings/settings.go:232-264`); only `print` and `screen` are evaluated (all other media types and unsupported feature queries evaluate to false); tests `TestParseMedia`, `TestMediaMatches*` |
| `@media` feature queries (`(min-width: …)`) | Partial | size features + orientation vs viewport; unknown features → false; `TestMediaMatchesSizeFeatures` |
| `:has()` | Partial | Relative selectors inside `:has(...)`; descendant/child/sibling + simple compounds; no forgiving-selector list / complex chrome edge cases. `has.go`; fixture-41 |
| `:not()` | Implemented | `appendFunctionalPseudo` (`selector_parser.go:353`); match `matchNone` (`match.go:515`). Argument list is strict (empty items fail). Specificity of the most specific argument (`has.go:343`). Tests `has_test.go` |
| `:is()` | Implemented | Strict selector-list arguments; nested `:is` allowed; `::` in args rejected. Match any argument. Specificity of the most specific argument. Tests `TestParseIs`, `TestIsPseudo`, `TestIsSpecificity` |
| `:where()` | Implemented | Same matching as `:is()`; specificity contribution 0. Test `TestWherePseudo` |
| `:root` | Implemented | Matches the document element (`<html>`), not the synthetic `#document` wrapper (`matchPseudo` `match.go:420`; `isRootElement` `match.go:539`). Test `TestRootPseudo` |
| `var()` / `--*` custom properties | Partial | `--*` inherit then overlay (`mergeCustomProps` `style_cascade.go:68`); `var()` expanded before apply (`resolveRawVars` `style_cascade.go:170`; `ResolveCustomProps` `values.go:696`). Cycles resolve empty. Tests `cssvar_font_test.go`, `TestResolveCustomProps*` |
| `@container` | Partial | Size queries only (`inline-size`/`width` + `and`/`or`/`not`); named containers; two-pass style after used inline size. No style/scroll-state queries; no `cq*` units. `internal/css/container.go`; fixture-42 |
| `container-type` / `container-name` / `container` | Implemented | Parsed `applyContainerProps` (`layout/style_container_props.go`). Size containers measured in `layout/container.go:44`. `container-type` honors `normal`/`size`/`inline-size`; CSS-wide keywords resolve (`inherit` copies the parent, `initial`/`unset`/`revert` reset to `normal` / none). Fixture-42; fixture-61 rows 39-41 |
| `@page` | Partial | Unnamed `margin`/`size` on every page. `:first` / `:left` / `:right` override **margin** (LTR page 1 is `:right`; `:first` wins on page 1). `page: ident` used-value inherit plus sibling break; named `@page` margin on pages that overlap that name. Size unnamed-only. Margin boxes: unnamed quoted `@top-*` / `@bottom-*` parse onto the stylesheet (`internal/css/page_margin.go`) with no consumer in this tree. See the `page` and `break-*` rows in §2. |
| `@font-face` | Partial | Parsed; `MergeFontFaces` loads TTF/OTF/WOFF1 via `FetchSub` (local **and** `https://`) under the same load ACL and network policy as other subresources. `.woff2` / `.eot` / `data:` skipped. See §5 |
| `@import` | Partial | Parsed onto `Stylesheet.Imports`; `CollectSheets` fetches under the same load policy as `<link>`, depth cap 8, cycle skip, failed fetch skipped. Media prelude uses `MediaMatches`. Tests `TestParseImport`, `TestCollectSheetsImportAppliesImportedRules` |

## 5. Explicitly unsupported (MVP)

| Feature | Handling |
|---------|----------|
| JavaScript / `<script>` / DOM APIs | **Not executed.** `<script>` content is captured as raw text and hidden by the UA sheet (`display: none`); no code path evaluates scripts |
| Full CSS Grid / full Flexbox | Stage A/B layout subset **shipped** (see the `flex-*` and `grid-*` rows in §2); Stage C lite + flex min-size polish + Partial subgrid/masonry span. **Not** Bootstrap/Tailwind / Chrome layout-test parity |
| `transform`, `filter`, `animation`, `transition` | **Transforms:** static 2D and the supported 3D-to-2D paint transform path, including perspective, preserve-3d composition, perspective origin, and hidden backface culling. **`filter`:** 2D image filter (opacity, blur, grayscale, invert, adjustments) on raster images (`filter.go`) plus CSS `opacity()` on elements; no CSS shader/SVG filter composition. **`animation`/`transition`/`@keyframes`:** valid declarations and keyframes are parsed, cascaded, and exported as host-facing metadata. Blinkless does not advance time, sample timelines, compare transition states, or interpolate frames. Fixture-40; `transform.go`, `transform_3d.go`, `filter.go`, `animation_style.go`, `transition_style.go` |
| `background-image` / gradients | **Implemented** multi-layer `url(...)` and pure-Go linear/radial gradient rasterization (`background_image.go`, `gradient.go`) |
| `@font-face` (remote / WOFF2) | **Partial:** local **and `https://`** TTF/OTF/WOFF1 via `FetchSub` (same load ACL and network policy as other subresources). **`.woff2` / `.eot` / `data:`** skipped. Missing faces fall back to registry / Liberation |
| Custom XSLT TOC (`--xsl-style-sheet`) | Not implemented (no XSLT in the standard library); this tree has no TOC pipeline |
| SVG-as-`<img>` / SVG presentation | **Implemented** SVG-as-`<img>` and inline `<svg>` rasterization via `internal/svg`; 5 CSS properties (`fill`, `stroke`, `stroke-width`, `fill-opacity`, `stroke-opacity`) parsed in style; the remaining SVG presentation properties are unsupported (§2.3) |
| Masking / clipping | **Implemented** `overflow-clip` for descendant box clipping (`overflow_clip.go`) and `clip-path` (§2.1). The 18 first-wave mask-family rows are registered as element-group mask data: 16 are implemented, `mask-type` and `-webkit-mask-composite` are Partial, and the 13 `mask-border*` rows are implemented: the frame composes into a nine-slice raster at the element box and the host applies it as a soft mask (§2.1). |
| CSS Regions & Exclusions | **Implemented** as four one-canvas contracts: named flows (`flow-into`, `flow-from`, `region-fragment`), masonry tolerance (`flow-tolerance`), exclusions (`wrap-flow`, `wrap-through`), and line/flex wrap breaks (`wrap-before`, `wrap-after`, `wrap-inside`). The engine returns one display list; paged fragmentation stays a non-goal (§2.1). |
| WebP, AVIF | Not implemented; broken-image placeholder or skip |
| Fixed CSS headers/footers via `position: fixed` alone | `position: fixed` is marked viewport-fixed (`markOpsFixed`, `layout_chrome.go:12`) when not under a transformed ancestor; not a full running-element model. There are no header/footer flags in this tree |
| Complex-script shaping (Indic, Arabic, CJK) | **Arabic OT** via `go-text/typesetting` when the face has GSUB (+ presentation-form `ShapeText` fallback); Hangul needs a Hangul face. `writing-mode` vertical keywords inherit and rotate glyphs (`RotateDeg == -90`); block/line layout stays horizontal. **Indic Partial** (OT when face/cmap allow; not production-claimed). Optional OT **`halt`/`palt`** for CJK punctuation via `ShapeTextFont` FontFeatures |
| PDF output (versions, profiles) | **Not part of this tree.** The engine returns a drawing list; PDF emission is out of scope. There are no `--pdf-version` / `--pdf-profile` flags, no `Document.PDFVersion` / `Document.PDFProfile` fields, and no PDF/A or PDF/UA claims |

### 5.1 Deferred niche and draft families (60 properties - Not implemented)

The following 60 properties are intentionally left **unsupported** (Not implemented). They have no layout or paint consumer and are out of scope for this engine. Declarations are parsed as valid property names where recognized and then ignored with graceful degrade, never claimed as Implemented. No Implemented claims are made for any of the 60.

| Family | Count | Status | Note | Examples (not exhaustive) |
|--------|-------|--------|------|---------------------------|
| Ruby/MathML/rhythmic niche (33 properties: block-ellipsis, block-step-*, box-snap, ruby-*, math-*, etc.) | 33 | Not implemented | Ruby/MathML/rhythmic niche; not implemented here. Left unsupported. | `block-ellipsis`, `block-step-*`, `box-snap`, `ruby-align`, `ruby-position`, `math-depth`, `math-style`, `line-snap`, etc. |
| Draft gap/row-rule decorations (27 properties: row-rule*, rule*) | 27 | Not implemented | Draft gap/row-rule decorations; no layout consumer. Left unsupported. `row-rule*` and `rule*` are not `column-rule` (see the `column-rule*` rows in §2). | `row-rule`, `row-rule-break`, `row-rule-color`, `row-rule-style`, `row-rule-width`, `rule`, `rule-break`, `rule-color`, `rule-style`, `rule-width`, etc. |

Total: 60 properties across two families. All are Not implemented and left unsupported with no layout consumer.

### 5.2 Phase 83 hard-defer categories (87 properties in the original scope)

The following 87 properties across three categories were grouped as hard-deferred or permanent non-goals in Phase 83. The catalog in section 2 records their current status:

| Category | Count | Status | Description and scope |
|----------|-------|--------|-----------------------|
| SVG presentation & geometry (`B_svg_presentation`) | 53 | Not implemented | Remaining SVG presentation and geometry properties. SVG-as-`<img>` is implemented via `internal/svg` rasterizer; 5 CSS properties (`fill`, `stroke`, `stroke-width`, `fill-opacity`, `stroke-opacity`) are parsed in style; remaining 53 SVG presentation properties are unsupported (`alignment-baseline`, `baseline-shift`, `color-interpolation`, `cx`, `cy`, `d`, `dominant-baseline`, `fill-break`, `fill-color`, `fill-image`, `fill-origin`, `fill-position`, `fill-repeat`, `fill-rule`, `fill-size`, `glyph-orientation-vertical`, `image-rendering`, `marker`, `marker-end`, `marker-mid`, `marker-side`, `marker-start`, `paint-order`, `path-length`, `r`, `rx`, `ry`, `shape-rendering`, `stop-color`, `stop-opacity`, `stroke-align`, `stroke-alignment`, `stroke-break`, `stroke-color`, `stroke-dash-corner`, `stroke-dash-justify`, `stroke-dashadjust`, `stroke-dasharray`, `stroke-dashcorner`, `stroke-dashoffset`, `stroke-image`, `stroke-linecap`, `stroke-linejoin`, `stroke-miterlimit`, `stroke-origin`, `stroke-position`, `stroke-repeat`, `stroke-size`, `text-anchor`, `text-rendering`, `vector-effect`, `x`, `y`). |
| Mask, clip, and filter effects (`B_mask_clip_filter_effects`) | 25 | Mixed | `clip-path` is implemented. The 10 standard mask properties are mostly implemented: `mask`, `mask-clip`, `mask-composite`, `mask-image`, `mask-mode`, `mask-origin`, `mask-position`, `mask-repeat`, and `mask-size`; `mask-type` stays Partial. They emit element-group mask data for a display-list consumer; this repository has no pixel replay backend. The 13 `mask-border*` properties are implemented: the descriptor rides the blend group, composes into a nine-slice frame raster at the element box, and the host applies it as a soft mask. The seven Unsupported names are `backdrop-filter`, `clip`, `clip-rule`, `color-interpolation-filters`, `flood-color`, `flood-opacity`, and `lighting-color`; the six `-webkit-mask-box-image*` aliases are implemented like their bases. |
| CSS Regions & Exclusions (`B_regions_exclusions`) | 9 | Implemented | The four one-canvas contracts are implemented: named flows (`flow-from`, `flow-into`, `region-fragment`), masonry tolerance (`flow-tolerance`), exclusions (`wrap-after`, `wrap-before`, `wrap-flow`, `wrap-inside`), and `wrap-through`. Paged fragmentation stays a non-goal; the engine returns one display list. |

Total: 87 names across three Phase 83 categories. Current statuses are in section 2; in the wider mask family (31 catalog rows: the 18 first-wave mask rows plus the 13 mask-border rows), 29 are implemented and 2 are Partial (`mask-type` and `-webkit-mask-composite`). See section 2 for the authoritative per-row status.

Catalog reconciliation: this phase-83 tally counts names with no print
consumer as unsupported, but several are parsed and stored in the style layer
and therefore appear as Partial in the catalog with unverified behavior:
`alignment-baseline`, `color-interpolation`, `dominant-baseline`, `fill-rule`,
`shape-rendering`, `stroke-dasharray`, `stroke-dashoffset`, `stroke-linecap`,
`stroke-linejoin`, `stroke-miterlimit`, and `text-anchor`. The 18 first-wave
mask properties are mostly implemented, with `mask-type` and
`-webkit-mask-composite` Partial because the display list leaves pixel replay
to its consumer; the 13 mask-border rows are Partial because the descriptor
rides the blend group and the host applies the frame. `clip-path` is
Implemented. The
five SVG properties consumed by SVG-as-img (`fill`, `stroke`, `stroke-width`,
`fill-opacity`, `stroke-opacity`) are Partial as well. The catalog in section 2
is authoritative for status; this table remains the phase-83 family record.

### 5.3 Vendor-prefix aliases (-webkit-) - Phase 69 and Phase 82

This section is the contract for vendor-prefixed alias handling. Alias mechanism is
`normalizeVendorPrefix` at `internal/layout/style_cascade.go:913` called from
`applyStyleProp`, plus `display` value remaps at `internal/layout/style_properties.go:81`
and value remaps in `remapWebkitValue` at `internal/layout/style_cascade.go:1000`.

- **Total tracked:** 70 `-webkit-` alias names.
- **Registered:** 36 aliases that map to an unprefixed base the engine consumes. The catalog records an alias as Partial when its base is Partial; see the reconciliation note below.
- **Unsupported:** 34 aliases whose bases are not Implemented or are print-noop. These are
  not Implemented and degrade gracefully (ignored declaration).

Do not treat Unsupported aliases as Implemented. Tests for Implemented aliases live in
`internal/layout/style_cascade_test.go:188` (`TestWebkitPrefixAliases`).

Catalog reconciliation: alias status follows the base property's status in
`testdata/css/catalog/properties.json`. Bases that are Partial make their
aliases Partial, so the catalog records `-webkit-box-shadow`, `-webkit-filter`,
`-webkit-transform`, `-webkit-transform-origin`, `-webkit-align-content`,
`-webkit-align-self`, `-webkit-flex-basis`, `-webkit-flex-shrink`,
`-webkit-flex-wrap`, the four `-webkit-border-*-radius` corner longhands, and
`-webkit-line-clamp` as Partial. The 28/42 split above is the phase-69/82
tally of registered aliases, not the catalog status count.

#### 5.3.1 Registered vendor aliases (36)

28 aliases were registered in Phases 69 and 82. Eight mask aliases were added
in the CSS masking first wave. Each folds onto its unprefixed base; aliases
with a Partial base remain Partial in the catalog.

| Vendor alias | Canonical property | Notes / base |
|--------------|--------------------|--------------|
| `-webkit-box-sizing` | `box-sizing` | `normalizeVendorPrefix` `style_cascade.go:913` |
| `-webkit-border-radius` | `border-radius` | same |
| `-webkit-border-top-left-radius` | `border-top-left-radius` | same |
| `-webkit-border-top-right-radius` | `border-top-right-radius` | same |
| `-webkit-border-bottom-left-radius` | `border-bottom-left-radius` | same |
| `-webkit-border-bottom-right-radius` | `border-bottom-right-radius` | same |
| `-webkit-transform` | `transform` | same |
| `-webkit-transform-origin` | `transform-origin` | same |
| `-webkit-flex` | `flex` | flex shorthand |
| `-webkit-flex-basis` | `flex-basis` | same |
| `-webkit-flex-direction` | `flex-direction` | same |
| `-webkit-flex-flow` | `flex-flow` | same |
| `-webkit-flex-grow` | `flex-grow` | same |
| `-webkit-flex-shrink` | `flex-shrink` | same |
| `-webkit-flex-wrap` | `flex-wrap` | same |
| `-webkit-justify-content` | `justify-content` | same |
| `-webkit-align-content` | `align-content` | same |
| `-webkit-align-items` | `align-items` | same |
| `-webkit-align-self` | `align-self` | same |
| `-webkit-order` | `order` | same |
| `-webkit-box-shadow` | `box-shadow` | alias to `box-shadow` (see `box-shadow` row in 2.4) |
| `-webkit-filter` | `filter` | alias to `filter` |
| `-webkit-box-align` | `align-items` | **New in Phase 82 slice A.** Value remap `start`->`flex-start`, `end`->`flex-end`, `center`->`center`, `stretch`->`stretch`, `baseline`->`baseline` via `remapWebkitValue` `style_cascade.go:1000` |
| `-webkit-box-flex` | `flex-grow` | **New in Phase 82 slice A.** Direct value pass-through to `flex-grow` |
| `-webkit-box-ordinal-group` | `order` | **New in Phase 82 slice A.** Value remap `N` -> `N-1` (1-based group to 0-based order) via `remapWebkitValue` `style_cascade.go:1000` |
| `-webkit-box-orient` | `flex-direction` | **New in Phase 82 slice A.** Value remap `horizontal`/`inline-axis`->`row`, `vertical`/`block-axis`->`column` via `remapWebkitValue` |
| `-webkit-box-pack` | `justify-content` | **New in Phase 82 slice A.** Value remap `start`->`flex-start`, `end`->`flex-end`, `center`->`center`, `justify`->`space-between` via `remapWebkitValue` |
| `-webkit-text-fill-color` | `color` | **New in Phase 82 slice A.** Alias to `color` |
| `-webkit-mask` | `mask` | **CSS masking first wave.** Shorthand alias; Partial |
| `-webkit-mask-image` | `mask-image` | **CSS masking first wave.** URL and `none` values; Partial |
| `-webkit-mask-size` | `mask-size` | **CSS masking first wave.** Accepted size tokens; Partial |
| `-webkit-mask-repeat` | `mask-repeat` | **CSS masking first wave.** Accepted repeat tokens; Partial |
| `-webkit-mask-position` | `mask-position` | **CSS masking first wave.** Position keywords, lengths, and percentages; Partial |
| `-webkit-mask-origin` | `mask-origin` | **CSS masking first wave.** Border, padding, or content box; Partial |
| `-webkit-mask-clip` | `mask-clip` | **CSS masking first wave.** Box values or `no-clip`; Partial |
| `-webkit-mask-composite` | `mask-composite` | **CSS masking first wave.** Remaps the four legacy operators; Partial |

Display value aliases (also Phase 82 slice A, at `internal/layout/style_properties.go:81`):

| Specified value | Used value |
|-----------------|------------|
| `display: -webkit-box` | `display: flex` |
| `display: -webkit-inline-box` | `display: inline-flex` |

#### 5.3.2 Still Unsupported vendor aliases (34)

34 aliases remain **Unsupported**. Each group notes the
blocking base. Do not claim Implemented for any of these 34.

**Group B - 3 background longhands:** bases `background-clip`,
`background-origin`, and `background-size` are Partial in the catalog (parsed,
behavior unverified). Their `-webkit-*` remaps are not registered, so the
aliases stay Unsupported
(`testdata/css/catalog/properties.json`).

| Vendor alias | Base | Reason |
|--------------|------|--------|
| `-webkit-background-clip` | `background-clip` | remap not registered |
| `-webkit-background-origin` | `background-origin` | remap not registered |
| `-webkit-background-size` | `background-size` | remap not registered |

**Group C - 6 mask-border aliases:** the eight selected `-webkit-mask*`
aliases are registered and mostly implemented (`mask-type` and
`-webkit-mask-composite` stay Partial). These six `-webkit-mask-box-image*`
aliases normalize onto the `mask-border*` bases, which are Partial: the
descriptor rides the element's blend group (`MaskBorder`) for the host to
apply.

| Vendor alias | Base | Reason |
|--------------|------|--------|
| `-webkit-mask-box-image` | `mask-border` | Partial: descriptor on the blend group, host applies the frame |
| `-webkit-mask-box-image-source` | `mask-border-source` | Partial: source bytes on the blend group, host applies the frame |
| `-webkit-mask-box-image-slice` | `mask-border-slice` | Partial: slice insets on the blend group, host applies the frame |
| `-webkit-mask-box-image-width` | `mask-border-width` | Partial: frame thickness on the blend group, host applies the frame |
| `-webkit-mask-box-image-outset` | `mask-border-outset` | Partial: outward offsets on the blend group, host applies the frame |
| `-webkit-mask-box-image-repeat` | `mask-border-repeat` | Partial: repeat pair on the blend group, host applies the frame |

**Group D - 20 animation / transition / 3D / UI aliases:** 14 animation and
transition aliases normalize to their base properties, then the paired
OwnFrame runtime samples those values when the caller supplies time and other
runtime inputs (`../ownframe/internal/animation/player.go:22-64`,
`../ownframe/internal/transition/step.go:10-44`). Four transform aliases are
Implemented through their base properties. Two UI aliases remain Unsupported.

| Vendor alias | Base property | Reason |
|--------------|---------------|--------|
| `-webkit-animation` | `animation` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-delay` | `animation-delay` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-direction` | `animation-direction` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-duration` | `animation-duration` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-fill-mode` | `animation-fill-mode` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-iteration-count` | `animation-iteration-count` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-name` | `animation-name` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-play-state` | `animation-play-state` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-animation-timing-function` | `animation-timing-function` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-transition` | `transition` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-transition-delay` | `transition-delay` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-transition-duration` | `transition-duration` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-transition-property` | `transition-property` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-transition-timing-function` | `transition-timing-function` | Implemented in the paired pipeline: maps to the base value sampled by OwnFrame |
| `-webkit-backface-visibility` | `backface-visibility` | Implemented: mapped into the supported 3D-to-2D paint path |
| `-webkit-perspective` | `perspective` | Implemented: mapped into the supported 3D-to-2D paint path |
| `-webkit-perspective-origin` | `perspective-origin` | Implemented: mapped into the supported 3D-to-2D paint path |
| `-webkit-transform-style` | `transform-style` | Implemented: mapped into the supported 3D-to-2D paint path |
| `-webkit-appearance` | `appearance` | Unsupported: no print consumer |
| `-webkit-user-select` | `user-select` | Unsupported: no interactive selection |

**Group E - 5 WebKit-native with no print consumer:** bases are WebKit extensions
with no consumer in this engine. Left Unsupported unless a base plus consumer
lands.

| Vendor alias | Blocking base | Reason |
|--------------|---------------|--------|
| `-webkit-line-clamp` | `line-clamp` | WebKit-native, no print consumer |
| `-webkit-text-size-adjust` | `text-size-adjust` | WebKit-native, no print consumer |
| `-webkit-text-stroke` | `text-stroke` | WebKit-native, no print consumer |
| `-webkit-text-stroke-color` | `text-stroke-color` | WebKit-native, no print consumer |
| `-webkit-text-stroke-width` | `text-stroke-width` | WebKit-native, no print consumer |

### 5.4 Phase 84 print-noop categories (114 names, mixed current status)

This list has 114 names across five print-noop families. Their current catalog statuses are 102 Unsupported, five Implemented, and seven Intentionally ignored. The five `scroll-margin` properties are accepted but do not change static paint. The CSS motion properties have separate statuses in section 2. Their `blinkless+ownframe` rows have sampled behavior in OwnFrame, which needs host-supplied time, timeline progress, trigger events, or display snapshots. Blinkless itself still exports data without advancing animation time. This engine also has no interactive scroll viewport, pointer/caret chrome, motion or view timeline drivers, or aural speech synthesis. The 3D transform properties have a tested 2D paint projection path in section 2. Anchor positioning is also excluded here because its one-canvas subset is implemented in section 2.

| Category | Count | Status | Description and scope |
|----------|-------|--------|-----------------------|
| Event and timeline trigger runtime (`A_time_animation_transition`) | 16 | Not implemented | These properties describe event-trigger definitions and pointer-driven timelines; this engine has no runtime for them (`event-trigger`, `event-trigger-name`, `event-trigger-source`, `image-animation`, `pointer-timeline`, `pointer-timeline-axis`, `pointer-timeline-name`, `timeline-trigger`, `timeline-trigger-activation-range`, `timeline-trigger-activation-range-end`, `timeline-trigger-activation-range-start`, `timeline-trigger-active-range`, `timeline-trigger-active-range-end`, `timeline-trigger-active-range-start`, `timeline-trigger-name`, `timeline-trigger-source`). They are separate from `animation-trigger` and `trigger-scope`, which have sampled OwnFrame cases in section 2. |
| Scroll snap, overscroll, and scrollbar (`A_scroll_snap_overscroll`) | 41 | Mixed: 36 Unsupported, 5 Implemented | This engine has no scroll viewport, scroll snapping, scroll timeline drivers, or scrollbar gutter/color/width chrome. The five `scroll-margin` properties are accepted but do not change static paint (`overscroll-behavior`, `overscroll-behavior-block`, `overscroll-behavior-inline`, `overscroll-behavior-x`, `overscroll-behavior-y`, `scroll-axis-lock`, `scroll-behavior`, `scroll-initial-target`, `scroll-margin`, `scroll-margin-block`, `scroll-margin-block-end`, `scroll-margin-block-start`, `scroll-margin-bottom`, `scroll-margin-inline`, `scroll-margin-inline-end`, `scroll-margin-inline-start`, `scroll-margin-left`, `scroll-margin-right`, `scroll-margin-top`, `scroll-marker-group`, `scroll-padding`, `scroll-padding-block`, `scroll-padding-block-end`, `scroll-padding-block-start`, `scroll-padding-bottom`, `scroll-padding-inline`, `scroll-padding-inline-end`, `scroll-padding-inline-start`, `scroll-padding-left`, `scroll-padding-right`, `scroll-padding-top`, `scroll-snap-align`, `scroll-snap-stop`, `scroll-snap-type`, `scroll-target-group`, `scroll-timeline`, `scroll-timeline-axis`, `scroll-timeline-name`, `scrollbar-color`, `scrollbar-gutter`, `scrollbar-width`). |
| Pointer, caret, and form UI (`A_pointer_form_ui`) | 25 | Mixed: 18 Unsupported, 7 Intentionally ignored | This engine has no mouse pointer, cursor styling, caret animation/color/shape, spatial navigation, dynamic field sizing, input security masking, touch gestures, user text selection, or window dragging. The seven intentionally ignored properties keep their catalog status (`appearance`, `caret`, `caret-animation`, `caret-color`, `caret-shape`, `cursor`, `field-sizing`, `input-security`, `interactivity`, `interest-delay`, `interest-delay-end`, `interest-delay-start`, `nav-down`, `nav-left`, `nav-right`, `nav-up`, `pointer-events`, `resize`, `slider-orientation`, `spatial-navigation-action`, `spatial-navigation-contain`, `spatial-navigation-function`, `touch-action`, `user-select`, `window-drag`). |
| Motion path, offset and view timelines (`A_anchor_timeline_motion`) | 13 | Not implemented | Motion path offset rotations and distances, view timelines, scroll anchoring, and paint invalidation hints have no layout consumer (`offset`, `offset-anchor`, `offset-distance`, `offset-path`, `offset-position`, `offset-rotate`, `overflow-anchor`, `timeline-scope`, `view-timeline`, `view-timeline-axis`, `view-timeline-inset`, `view-timeline-name`, `will-change`). Anchor positioning (`anchor-name`, `anchor-scope`, `position-anchor`, `position-area`, `position-try`, `position-try-fallbacks`, `position-try-order`, `position-visibility`) is implemented in section 2 and excluded from this count. |
| Speech and aural (`A_speech_aural`) | 19 | Not implemented | This engine has no aural speech synthesis, sound cues, pauses, rests, or speech voice properties (`cue`, `cue-after`, `cue-before`, `pause`, `pause-after`, `pause-before`, `rest`, `rest-after`, `rest-before`, `speak`, `speak-as`, `voice-balance`, `voice-duration`, `voice-family`, `voice-pitch`, `voice-range`, `voice-rate`, `voice-stress`, `voice-volume`). |
| 3D transforms (`A_3d_transforms`) | 4 | Implemented in section 2 | The supported 3D-to-2D paint path applies perspective, preserve-3d composition, perspective origin, and backface culling (`backface-visibility`, `perspective`, `perspective-origin`, `transform-style`). This row is not part of the unsupported count. |

Total: 114 names across five families: 102 Unsupported, five Implemented, and seven Intentionally ignored. The four 3D transform properties and the 29 standard CSS motion properties have separate statuses in section 2. The motion catalog also has 14 WebKit aliases, for 43 sampled property rows. Anchor positioning (8 properties) is implemented in section 2 and not counted here.

Catalog reconciliation: the counts above exclude the 29 standard CSS motion
properties, their 14 WebKit aliases, and four 3D transform properties now
described in section 2. The 43 motion rows have OwnFrame behavior tests; the
host must still supply runtime inputs. The five scroll margin properties are
accepted but do not change static paint.
`appearance`,
`caret-color`, `cursor`, `pointer-events`, `resize`, `touch-action`, and
`user-select` remain intentionally ignored. The catalog in section 2 is
authoritative for each property status.

### 5.5 v0.2.7 Borders-4 / Round Display drafts (14 properties - Unsupported / draft-not-ready)

Phase 87.7 product choice: **Defer all 14**. Spec text for these Borders-4 /
css-round-display-1 names is largely not ready for implementation; there is no
apply arm and no paint consumer in `internal/layout`. Do not claim Implemented
for parse-only stubs. Catalog rows stay `unsupported` in
`testdata/css/catalog/properties.json`. Fixture-64 Effect cells already warn
`Chrome: no render expected` where Chrome BCD has no render path.

| Property | Spec | Status | Note |
|----------|------|--------|------|
| `border-block-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-block-end-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-block-start-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-bottom-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-boundary` | css-round-display-1 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-inline-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-inline-end-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-inline-start-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-left-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-limit` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-right-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |
| `border-shape` | css-borders-4 | Unsupported / draft-not-ready | Chrome BCD yes since 147; still no consumer in this tree; v0.2.7 defer |
| `border-top-clip` | css-borders-4 | Unsupported / draft-not-ready | No apply/paint; v0.2.7 defer |

Total: 14 properties. All remain Unsupported / draft-not-ready for v0.2.7. The
phase-87.7 ledger is not part of this tree; the catalog rows in §2 are the
current record.

## 6. Security policy (frozen defaults)

| Rule | Value |
|------|-------|
| Local file access | **Blocked by default**; `settings.Load.Allow` is the path allowlist, enforced by the `internal/load` ACL (`ErrAccessDenied`, load.go:47) |
| Untrusted HTML | **Not supported**; use HTML you control only |
| Remote URL fetch | `net/http` defaults: connect + response timeouts, redirect limit. `CompatibleNetworkPolicy` allows `http://localhost` and RFC1918; `RestrictedNetworkPolicy` does not (load.go:135-152). `file://` is gated by the local-file ACL |
| SSRF posture | No automatic form submission; POST only when the caller configures `settings.Load.Post`; no cookies are forwarded |

## 7. CLI flags

The `bin/blinkless` CLI and its flag matrix from earlier revisions are not part
of this tree: there is no `cmd/` package and no `internal/cli`. The engine is
consumed through the Go packages (`html`, `css`, `layout`) and the WASM binding
in `bindings/wasm` (see [library-api.md](library-api.md)). A CLI would need its
own flag contract; this matrix makes no flag claims.

---

## Amendment process

Any change to this matrix is a catalog edit or a plan amendment, recorded
against [plans/v0.0.1/html-css-json-compatibility-checklist.md](../plans/v0.0.1/html-css-json-compatibility-checklist.md).
Regenerate section 2 with `python3 scripts/css-catalog-map.py --matrix` after a
catalog change; `make catalog-check` fails on drift between the catalog and the
code. The 0.0.1 compatibility work maps its rows to this matrix.
