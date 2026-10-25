# v0.0.3 - CSS animations, transitions, and compositor layers

> **Status:** planning. No behavior changes with these files.
> **Tracks:** [animation](animation/README.md), [transitions](transitions/README.md), [examples](examples/README.md)
> **Depends on:** [optimizations phase 02](../optimizations/phase-02-tiling-and-sparse-rasterization.md) (tile grid), [phase 09](../optimizations/phase-09-directly-composited-images-and-solid-color.md) (direct paint seams)

## The invariant

A page declares CSS. OwnFrame samples and composites it. Application code never
binds, steps, or marks damage for a CSS animation or transition.

This is the contract Chrome has: the style engine resolves declarations, the
compositor samples the timeline and moves layers. The app writes HTML and CSS
and nothing else. Every phase in these three tracks moves ownframe toward that
contract, and no phase may add a per-animation call to an example that a CSS
declaration could express.

[Coverage](coverage.md) is the ledger that maps every implemented animatable
property in the compatibility matrix to a phase, an exclusion, or a named
blocker. A property outside that ledger is not claimed by this track.

## Priority 0

[Engine transform, 3D, and perspective support](priority-0-engine-transform-and-perspective.md)
is the only upstream prerequisite that blocks phase 02. It exports the element's
own matrix, `transform-origin`, `transform-box`, and a 3D and perspective
re-projection pass from blinkless. Animation phase 02 cannot be correct without
the first three, and no 3D transform or `perspective` animation is possible
without the fourth. It lands upstream first, then the pin bumps.

## Phase 11 engine work

[Phase 11](animation/phase-11-engine-named-timelines-and-svg-paint.md) covers the
two engine features required for full coverage: named timelines and SVG paint. It
runs in parallel with animation phases 02 to 05. Named timelines gate phase 06;
SVG paint gates the SVG cards in phase 10. Neither blocks phase 02.

The manual packages (`internal/cssanimation`, `internal/csstransition`,
`internal/animation`, `internal/transition`, `internal/frame`) stay. They become
the implementation under the automatic driver, not the API applications use.
`internal/frame` is explicitly **not** merged into the CSS runtime. It remains
the Go-controlled helper for motion that has no CSS declaration, such as an
audio position or a game sprite ([frames.md](../../../documentation/frames.md)).

## What already ships

- Blinkless exports `box.Animations []layout.Animation`,
  `box.Transitions []layout.Transition`, and `display.Keyframes
  []layout.KeyframesRule` with raw author declarations
  (`layout/layout.go:37-84` in the pinned `v0.0.3`).
- `internal/cssanimation.Bind` turns animation records and keyframes into an
  `animation.Player` with timing, direction, fill, delay, iteration, and
  composition (`internal/cssanimation/bind.go:12-44`).
- `internal/animation` samples time, scroll, view, named timelines, ranges, and
  triggers, and writes paint values into retained ops.
- `internal/csstransition` and `internal/transition` select a transition record
  and interpolate between two paints, including the discrete `position` case.
- `internal/replay` has per-op 2D matrices, an element chunk cache with content
  signatures, a 256 px damage tile grid, and a per-group texture composite.
- `internal/page` owns the loop, generation counter, damage tracker, and the
  `Window()` and `Probe()` doors.

## What is broken today

1. Nothing starts the runtime. No file under `internal/page` or
   `internal/window` imports `cssanimation` or `csstransition`. The only callers
   are `cmd/*-preview` and two examples.
2. The keyframe model carries three properties. `animation.Keyframe` holds X, Y,
   opacity, and RGB (`internal/animation/animation.go:19-24`). The parser accepts
   bare `translate*()` in px, unitless opacity, and hex colours
   (`internal/cssanimation/keyframes.go:20-31`). Scale, rotate, matrix, transform
   lists, `rgb()`, `hsl()`, and geometry are dropped silently.
3. Every op the box owns is a target, including text. Background-colour
   keyframes therefore recolour labels.
4. `Player.Step` returns no damage. Every frame pays an O(all ops) snapshot and
   `SameOp` diff (`internal/page/page_tick.go:44-90`), plus `bytes.Equal` on image
   payloads.
5. `animation.Player` holds raw op pointers with no generation guard. Callers
   compare `Probe().Generation()` by hand.
6. No retained layer. A moving element re-walks the plan for every dirty tile and
   re-bakes any op whose signature changed. `paintGroup` renders to a texture and
   frees it inside one call (`internal/replay/group_layer.go`). `compositeLayer`
   translates only (`internal/replay/group_composite.go`). CSS `opacity` is
   flattened into per-op alpha at layout, so there is no group alpha.

## How the three tracks fit

| Track | Owns | Does not own |
|---|---|---|
| Animation | `@keyframes`, animation longhands, timelines, ranges, triggers, layer promotion for animation | transition capture, view transitions |
| Transitions | style-change capture, transition longhands, discrete behavior, interruption | keyframe parsing, view transitions |
| Examples | removing hand-written drivers, CSS fixes, allowlist cleanup | runtime behavior |

Both runtime tracks share one package shape: `internal/motion` with a registry,
a bind step at display commit, a sample step in the page tick, and a damage
return. The animation plan builds that package. The transition plan adds a second
registry to it. They must not grow two independent tick paths.

## Cross-track dependency order

0. Priority 0 lands in blinkless first, then the pin bumps. It blocks animation
   phase 02 and the perspective part of phase 04.
0b. Phase 11 engine work lands in parallel. Named timelines gate phase 06; SVG
   paint gates the SVG cards in phase 10. Neither blocks phase 02.
1. Animation phase 01 (automatic driver) creates `internal/motion` and the page
   hook. Everything else builds on it.
2. Animation phases 02 to 07 make the runtime correct and cheap for the full
   property set in the compatibility matrix.
3. Animation phase 08 adds the retained compositor layer. Transition phase 04
   reuses it; do not duplicate the cache.
4. Transition phases 01 to 03 can land in parallel with animation phases 02 to 07
   once phase 01 exists, except that capture needs the commit hook from phase 01.
5. Animation phase 09 and transition phase 05 change public API and docs last,
   after the counters and limits are stable.
6. Animation phase 10 closes the path, SVG, and bitmap-fallback gaps. It changes
   the paint path, so it lands after the layers and before the final docs pass.
7. The examples track runs beside all of it. It may land example fixes as soon as
   the matching runtime phase is green.

## Scope rules

- Only implemented properties. `coverage.md` is the authority; a matrix row that
  the engine does not implement is not in scope, even if CSS defines it.

- No engine change without an upstream blinkless commit first, then a pin bump in
  `go.mod` and `examples/go.mod`. The pin is the only engine require.
- No new root alias and no new exported `Page` method beyond door access. Policy
  goes on `Window()`, counters on `Probe()`.
- No example imports `internal/`. Existing allowlist rows are removed by the
  examples track, not extended.
- The compositor thread is out of scope. Ebiten draws on the game goroutine only.
  These plans keep sampling in `Update` and compositing in `Draw`, and get the
  Chrome win from not repainting content under a moving layer.
- `internal/frame` is not unified and not deprecated.

## Gate for every phase

`make test`, `make lint`, `make docs-check plans-check examples-check`. Any phase
that touches paint, layers, or the window also runs `make graphics-check` and
records before and after numbers from the matching benchmark or paced dump.
