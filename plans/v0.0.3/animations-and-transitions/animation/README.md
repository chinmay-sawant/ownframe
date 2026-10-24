# v0.0.3 - CSS animation track

> **Parent:** [animations and transitions](../README.md)
> **Status:** planning
> **Coverage:** [coverage.md](../coverage.md) maps every implemented animatable
> property to a phase, an exclusion, or a blocker.
> **Goal:** every `@keyframes` animation in a page runs, at the right time and on
> the right properties, without application code.

## Definition of done

Open any example, delete every call to `cssanimation.Bind`, `Player.Step`,
`Probe().MarkRect`, and `WantFrameIn` that exists only to drive a CSS animation,
and the page looks the same. That is the acceptance test for this track.

## Phase index

| # | Phase | Status | Depends on | Phase file |
|---|---|---|---|---|
| 01 | Automatic driver | Planned | none | `phase-01-automatic-driver.md` |
| 02 | Keyframe transform model | Planned | 01, [priority 0](../priority-0-engine-transform-and-perspective.md) | `phase-02-keyframe-transform-model.md` |
| 03 | Paint property coverage | Planned | 02 | `phase-03-paint-property-coverage.md` |
| 04 | Layout property coverage | Planned | 03 | `phase-04-layout-property-coverage.md` |
| 05 | Timing completeness | Planned | 01 | `phase-05-timing-completeness.md` |
| 06 | Timelines, ranges, triggers | Planned | 01, 05, [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) | `phase-06-timelines-ranges-triggers.md` |
| 07 | Damage, allocation, pacing | Planned | 01 | `phase-07-damage-allocation-pacing.md` |
| 08 | Compositor layers | Planned | 07, optimizations 02/09 | `phase-08-compositor-layers.md` |
| 09 | Public API, docs, conformance | Planned | all | `phase-09-public-api-docs-conformance.md` |
| 10 | Path, SVG paint, bitmap fallback | Planned | 09, [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) for SVG | `phase-10-path-svg-bitmap-fallback.md` |
| 11 | Engine named timelines and SVG paint | Planned | parallel, gates 06 and 10 | `phase-11-engine-named-timelines-and-svg-paint.md` |

## Master checklist

### Phase 01 - automatic driver
Status: not started
- [ ] Create `internal/motion` with a runtime that binds a display and samples it
- [ ] Bind at every display commit, not only at `New`
- [ ] Run the driver in `Page` tick before the user callback
- [ ] Skip the whole-display diff when no user tick function exists
- [ ] Add a generation guard inside the runtime and rebind on change
- [ ] Honor `animation-play-state` from CSS instead of forcing `running`
- [ ] Default on, with `Page.Window().SetAnimations(false)` as the off switch
- [ ] Add `Probe()` counters for bound, sampled, and rebound players
- [ ] Delete the driver from `examples/animations` and prove parity
Gate: `make test`; the animations example runs with no motion code in `page.go`

### Phase 02 - keyframe transform model
Status: not started
- [ ] Replace scalar X/Y with a decomposed 2D transform plus a 3D function list
- [ ] Parse the full `transform` list: translate, scale, rotate, skew, matrix, matrix3d, and 3D functions
- [ ] Support px, %, deg, rad, turn, and unitless zero
- [ ] Apply 2D through `SetXform`; send 3D and perspective through `RestampTransforms`
- [ ] Implement 2D and 3D matrix decomposition for mismatched function lists
- [ ] Never promote a 3D element to a layer
- [ ] Get `transform-origin`, `transform-box`, and the matrices exported (priority 0)
Gate: rotate, scale, skew, rotateY, translateZ, and perspective cards match the CSS reference

### Phase 03 - paint property coverage
Status: not started
- [ ] Field masks per target so colour keyframes never touch text
- [ ] Animate opacity, background-color, color, border colours, outline colour
- [ ] Parse rgb(), rgba(), hsl(), hsla(), named colours, transparent, currentColor
- [ ] Animate filter through the layer path when phase 08 lands
- [ ] Animate box-shadow and text-shadow, or record the fallback explicitly
- [ ] Per-kind op targeting: fills, text, strokes, images
Gate: one catalog card per property, compared against the same CSS in a browser

### Phase 04 - layout property coverage
Status: not started
- [ ] Animate width, height, min/max size, inset, margin, padding, border-width
- [ ] Animate border-radius, font-size, line-height, letter-spacing
- [ ] Drive `layout.RelayoutElement` and adopt its display patch
- [ ] Preserve every other player's start time across a relayout
- [ ] Require `RetainState`, or enable it automatically on first bind
- [ ] Measure per-frame relayout cost and cap the work per frame
Gate: geometry cards animate without a full-page `Redraw`

### Phase 05 - timing completeness
Status: not started
- [ ] Per-keyframe `animation-timing-function`
- [ ] Resolve `animation-duration: auto`
- [ ] Real `accumulate` semantics, distinct from `add`
- [ ] End delay holds the final value instead of only extending liveness
- [ ] Fractional iteration counts sample mid-cycle
- [ ] Pause and resume without restarting from zero
- [ ] `alternate-reverse`, zero duration, negative delay at the boundary
- [ ] `steps(jump-none)` and `linear()` stop lists match the spec
Gate: every timing card in `examples/css-preview` passes under the auto driver

### Phase 06 - timelines, ranges, triggers
Status: not started
- [ ] Feed scroll and view timelines from the page into `FrameInput`
- [ ] Named timelines land in phase 11; phase 06 resolves them, no time fallback
- [ ] Resolve `animation-range*` against the element box and scroll container
- [ ] Dispatch page interactions to `animation-trigger` actions
- [ ] Respect trigger scope and element ids
- [ ] Keep time-driven animations on the same path
Gate: `TestCatalogAnimationTimelineSamplesHostProgress` and the range and trigger cases pass with no example code

### Phase 07 - damage, allocation, pacing
Status: not started
- [ ] Return per-target old ∪ new damage from the runtime
- [ ] Mark damage once, without a second whole-display diff
- [ ] Precompute keyframe stop tables at bind time
- [ ] Cache parsed easing functions and timelines at bind time
- [ ] Zero allocations per sampled frame
- [ ] Ask for the next frame only while motion is live
- [ ] Expose a TPS option or document `ebiten.SetTPS` for 120 Hz hosts
Gate: benchmark shows 0 allocs and no O(page) work on an animation-only frame

### Phase 08 - compositor layers
Status: not started
- [ ] Retain one texture per promoted animated element
- [ ] Composite it with a matrix and a group alpha
- [ ] Skip owned ops during tile repaint and blit the layer instead
- [ ] Invalidate on content signature, generation, and size change
- [ ] Cap layer pixels and count; demote cleanly
- [ ] Preserve group, clip, fixed, and z-index rules
- [ ] Report promoted, rebaked, and demoted layers in `Probe()`
Gate: rotating a card with text and an image does not rebake its descendants

### Phase 09 - public API, docs, conformance
Status: not started
- [ ] `Window().SetAnimations(bool)` and the matching `Probe()` counters
- [ ] Rewrite `documentation/css-motion.md` for automatic driving
- [ ] Update `features.md`, `frames.md`, `repaint.md`, `window.md`, `optimizations.md`
- [ ] Update the compatibility matrix notes and the 1.0 exclusions `OUT-12`, `OUT-21`
- [ ] Add a conformance page with one card per matrix row
- [ ] Remove the `examples/animations` allowlist rows
Gate: `make lint`, `make test`, `make graphics-check`, `make docs-check plans-check`

### Phase 10 - path, SVG paint, bitmap fallback
Status: not started
- [ ] Add `DisplayOpPath` to the render kind table and the replay painter
- [ ] Make a corner-shape page stay on the list path
- [ ] Consume `RestyleInlineSVG` from phase 11, or defer with a counter
- [ ] Detect a bitmap-fallback page and report why motion is static
- [ ] Measure a per-frame re-render from the retained list before adopting it
Gate: `make graphics-check`; corner shapes animate; fallback pages report a reason

### Phase 11 - engine named timelines and SVG paint
Status: not started
- [ ] Parse `scroll-timeline-*` and `view-timeline-*` and export `Box.Timelines`
- [ ] Resolve the nearest ancestor name and feed the existing `FrameInput`
- [ ] A missing or inactive name leaves the animation inactive, no time fallback
- [ ] Add `SVGPaintOverride` and `RestyleInlineSVG`
- [ ] Re-serialize and rasterize under the phase 04 budget
- [ ] Pin blinkless and bump both `go.mod` files
Gate: both parts match Chrome; no override leaves the display list unchanged

## Dependency notes

Phase 01 is the keystone. Phases 02 to 04 widen what a keyframe can address and
are independent of each other in code, but they share the target model from 01,
so land 01 first and do not fork it.

Phase 02 is blocked on [priority 0](../priority-0-engine-transform-and-perspective.md)
until blinkless exports the element matrix and the 3D and perspective
re-projection pass. Do not implement a center-origin fallback and call the phase
done.

Phase 05 and phase 06 change sampling only and can run beside 02 to 04.

Phase 06 also needs [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) for the
named timeline export. `scroll()` and `view()` do not.

Phase 07 must follow 01 because it removes the diff that 01 currently relies on.
Phase 08 depends on 07 for honest damage: a layer that is not damaged correctly
is worse than no layer.

Phase 09 waits for the counters in 07 and 08; do not document limits before they
are measured.

Phase 10 is last because it changes the paint path, not the timeline. It closes
the blocked rows in the coverage ledger: corner-shape paths, SVG paint through
[phase 11](phase-11-engine-named-timelines-and-svg-paint.md), and bitmap-fallback pages.

Phase 11 is the engine phase. It runs in parallel with phases 02 to 05 and does
not block them. Do not start phase 06 or the SVG cards in phase 10 before the
matching part of phase 11 is pinned.

## Non-goals

- A second draw thread. Ebiten draws on one goroutine.
- Out-of-process raster.
- View transitions. They have their own capture model and stay in
  `internal/viewtransition` until a later plan.
- `will-change` consumption until blinkless exports it. Layer promotion uses the
  animated property set instead.
