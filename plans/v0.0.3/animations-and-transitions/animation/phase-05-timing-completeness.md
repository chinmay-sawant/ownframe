# Phase 05 - timing completeness

Goal: every timing declaration the compatibility matrix accepts produces the
sample a browser produces, including the awkward cases: per-keyframe easing,
auto duration, accumulate, end delay, fractional iterations, and pause/resume.

## Where it lives

- `internal/cssanimation/timing.go:10-17` maps `layout.Animation` to
  `animation.Timing`. `DurationAuto` is copied at `:14` and never read anywhere.
- `internal/animation/cycle.go:8-34` handles infinite and finite iteration,
  fill, direction, and the end delay.
- `internal/animation/state_position.go:20-28` computes elapsed position with
  start delay and a signed rate.
- `internal/cssanimation/keyframes.go:20-31` ignores an
  `animation-timing-function` declared inside a keyframe block, because the
  switch only reads three properties.
- `internal/animation/compose.go:41` treats every non-`replace` mode as `add`.
- `internal/cssmotion/easing_curve.go:25-39` parses `steps()`;
  `easing_curve.go:26-28` makes `jump-none` behave like linear.
- `internal/animation/player.go:10-20` starts and never pauses. A caller that
  wants a pause tears the player down and builds a new one.

## Current behaviour against the matrix

| Declaration | Today | Required |
|---|---|---|
| `animation-delay` signed | correct, including negative start | keep |
| `animation-delay-start` | alias of delay | keep |
| `animation-delay-end` | extends liveness only | hold the final value for the end delay, then release |
| `animation-duration` time | correct | keep |
| `animation-duration: auto` | `Duration == 0`, no motion | resolve to 0s for a time-based animation, and to the timeline duration for a scroll or view timeline |
| `animation-iteration-count` integer | correct | keep |
| `animation-iteration-count` float | ends at the count but the last partial cycle is suspect | sample the fractional tail correctly |
| `animation-direction` | normal, reverse, alternate, alternate-reverse | keep, add a test for alternate-reverse at a cycle boundary |
| `animation-fill-mode` | none, forwards, backwards, both | keep, add a test at a negative delay |
| `animation-play-state` | forced to running by both examples | honor `paused`, and resume without restarting |
| `animation-timing-function` per animation | correct, parsed per sample | cache at bind (phase 07) |
| `animation-timing-function` in a keyframe | ignored | use it for the segment that starts at that keyframe |
| `animation-composition` replace | correct | keep |
| `animation-composition` add | sums onto the underlying value | keep, test against the matrix case |
| `animation-composition` accumulate | identical to add | scale the underlying value by the iteration progress, then add |
| `steps(..., jump-none)` | linear | hold the correct end values |
| `linear(...)` stops | parsed per sample | keep, cache |
| zero duration | undefined here | apply the end value immediately, honor fill |
| comma-separated list | one `Effect` per item | keep, test list repetition when the keyframe-name list is shorter |

## Design

1. Per-keyframe easing. Add `TimingFunction string` to `animation.Keyframe` and
   `KeyframeDeclaration` handling for `animation-timing-function` inside a
   `@keyframes` block. Blinkless already forwards the raw declaration
   (`layout.KeyframeDeclaration`), so this is parser work in
   `internal/cssanimation/keyframes.go`, not an engine change. At sample time,
   the easing of the segment starting at keyframe `i` applies to the interval
   `[i, i+1]`.

2. Auto duration. A time-based animation with `DurationAuto` resolves to 0 and
   applies its end value at once. A scroll or view timeline resolves to the
   timeline's range length. Put the resolution in one function,
   `resolveDuration(timing, input)`, and call it from the position code. Do not
   special-case it in three places.

3. Accumulate. Keep the current `add` path for `add`. For `accumulate`, the
   underlying value is multiplied by the iteration index before the add. For
   `replace`, unchanged. The difference only shows after the first iteration, so
   the test must run at least two cycles.

4. End delay. `animation-delay-end` holds the last sampled value after the
   active interval, then releases to the base when the end delay expires and the
   fill mode does not keep it. Today it only keeps `active` true. Implement it as
   a third state, `holding`, with the value fixed at the final sample.

5. Pause and resume. Add `Player.Pause(now)` and `Player.Resume(now)` that adjust
   the start offset instead of tearing down state. The driver calls them when the
   computed play state changes, and on rebind it restores the last phase for a
   paused animation. This fixes the example bug where clicking Pause snaps every
   chip to its base pose and Resume restarts at 0%.

6. Zero and fractional durations. A zero duration is not a divide-by-zero: the
   sample is the end keyframe. A fractional iteration count ends mid-cycle at the
   fractional position, not at the next keyframe.

7. List repetition. CSS repeats the shorter list (`animation-name: a` with four
   durations means all four use `a`). This is already handled by blinkless's
   `cssListAtOr`, so test it rather than re-implement it.

8. Composition order. Multiple effects on one element fold in list order with
   `replace` clobbering earlier effects and `add` summing. Keep the existing
   order and add a three-item test where replace, add, and accumulate mix.

9. Timing values are parsed once at bind. This phase may leave the cache to phase
   07, but it must not add a new per-sample string parse.

## Checklist

### A. Keyframe easing

- [ ] Add `TimingFunction` to `animation.Keyframe`.
- [ ] Parse `animation-timing-function` inside a keyframe block.
- [ ] Apply the segment easing between the current and next keyframe.
- [ ] Test a three-stop animation with a different easing on each segment.

### B. Duration and iteration

- [ ] `resolveDuration` for auto, time, scroll, and view.
- [ ] Zero duration applies the end value.
- [ ] Fractional iteration ends at the fractional position.
- [ ] Iteration count 0 applies only the fill value, if any.
- [ ] Test alternate-reverse at cycles 1, 2, and 3.

### C. Composition

- [ ] Implement accumulate with iteration-scaled add.
- [ ] Keep add and replace unchanged.
- [ ] Test a mixed list over three iterations.
- [ ] Test two animations on one element with disjoint fields.

### D. End delay

- [ ] Add the holding state.
- [ ] Hold the final value for the declared end delay.
- [ ] Release to base when the fill mode allows.
- [ ] Test a positive end delay with forwards, none, and both.

### E. Pause and resume

- [ ] Add `Pause` and `Resume` with a start offset.
- [ ] Restore a paused phase on rebind.
- [ ] Delete the example workaround that restarts players on resume.
- [ ] Test pause for 500 ms, resume, and compare with an uninterrupted run.

### F. Easing edge cases

- [ ] `steps(1, jump-none)` and `steps(n, jump-both)`.
- [ ] `linear()` with three stops and with duplicate positions.
- [ ] cubic-bezier with control points outside [0,1].

Gate: every timing row in `documentation/compatibility-matrix.md:97-114` has a
test that fails before the change and passes after it.

## Risks

- Pause semantics interact with `animation-trigger`. A trigger that pauses must
  use the same offset path, not a second one.
- End delay is easy to confuse with fill forwards. Fill decides the value after
  the end delay, not during it.
