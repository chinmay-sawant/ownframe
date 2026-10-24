# Phase 03 - interruption, reversal, and precedence

Goal: changing a property again while it is still transitioning behaves like a
browser: reverse from the presented value, shorten the reverse by the progress
already spent, and let an animation win over a transition on the same field.

## Where it lives

- `internal/transition/step.go:10-44` samples a renderer with a fixed start,
  duration, delay, and easing. It has no notion of a previous run.
- `internal/csstransition/bind.go:12-52` builds a renderer from a from-paint and a
  to-paint captured at bind.
- `internal/page/page_draw.go:174-185` commits a new display with the old one in
  hand, which is the moment an interruption is observable.
- `internal/motion` (animation phase 01) is the registry that must hold the live
  transition and its progress.

## What a browser does

CSS Transitions Level 1, reversing:

1. If the new end value equals the reversing-adjusted start value of the running
   transition, the new transition is a reversal.
2. The reversing shortening factor is the running transition's current progress.
3. The new duration is the declared duration multiplied by that factor.
4. The new start value is the currently presented value, not the original base.
5. Otherwise the change is a normal transition from the presented value with the
   full declared duration.

An animation in the CSS cascade overrides a transition for the same property,
because animations apply after transitions in the cascade origin order.

## Design

1. Keep the live entry addressable by element id and property. Before binding a
   new capture, look up the existing entry:
   - if none, bind normally,
   - if one exists and the new target equals its original from-value, it is a
     reversal: start from the current sampled paint, scale the duration by the
     progress, reverse the easing direction,
   - if one exists and the target is a third value, start from the current sample
     with the full duration.

2. Progress needs the same clock the renderer uses. Add
   `Renderer.Progress(now) float64` rather than recomputing it from the entry's
   fields in the driver. One source of truth for delay and duration.

3. The presented value is already in the old display. Capture reads it through
   the same code path as phase 01, so a reversal starts from exactly what the
   user sees, including a partial delay or a paused transition.

4. Precedence. Step transitions first, then animations, as phase 01 established.
   Additionally, skip a transition entry while an animation with the same element
   and an overlapping field mask is active. Skipping avoids writing a value that
   the animation immediately overwrites, and keeps the transition from consuming
   a frame request it cannot affect. When the animation finishes, the transition
   is already cancelled; do not revive it.

5. Multiple records. If two transition records match one property, the later
   declaration wins, which matches the cascade. `transition-property: all` plus a
   specific longhand is a normal page pattern; resolve it at capture.

6. Cancel. A property removed from the style, an element removed from the
   display, a redraw that replaces the element's ops, or a new run with no
   matching declaration cancels. Cancel writes the final value once, damages the
   box, and releases the entry. A cancelled transition never resumes on the next
   commit.

7. Delay boundaries. An interruption during the delay has progress 0, so the
   reversal duration is 0 and the value snaps to the new target. That is what the
   spec produces; do not special-case a "nicer" duration.

8. Zero duration. A zero declared duration snaps at commit. A reversal of a zero
   duration also snaps.

9. No `composite` for transitions. CSS transitions replace values. Do not reuse
   the animation `add` or `accumulate` path.

## Checklist

### A. Reversal

- [ ] `Renderer.Progress(now)` as the only progress source.
- [ ] Detect a reversal when the new target equals the original from-value.
- [ ] Scale the duration by the reversing shortening factor.
- [ ] Start from the presented paint.
- [ ] Third-value changes use the full duration.
- [ ] Test an interrupt at 25%, 50%, and 75% against a browser.

### B. Precedence

- [ ] Transitions step before animations.
- [ ] Skip a transition whose field is owned by an active animation.
- [ ] Do not revive a skipped transition when the animation ends.
- [ ] Two records on one property resolve to the later declaration.
- [ ] Test animation plus transition on `opacity`.

### C. Cancel

- [ ] Property removal cancels with the final value.
- [ ] Element removal cancels and releases.
- [ ] Op replacement cancels.
- [ ] A cancelled entry does not re-bind on the next commit.
- [ ] Damage on cancel is marked once.

### D. Boundary cases

- [ ] Interrupt during the delay snaps.
- [ ] Zero duration snaps.
- [ ] Interrupt at progress 1 cancels, no reverse run.
- [ ] Negative delay starts mid-progress and reverses from there.

Gate: three interrupt cards match a browser frame for frame at the same
timestamps, and an animation/transition conflict resolves to the animation.

## Risks

- Reversal math is easy to invert. Write the direction in a comment and test the
  two directions separately, not with one symmetric case.
- A transition that is skipped for precedence still needs a cancel on removal.
  Do not leave a zombie entry that fires when the animation ends.
- Progress must include the start delay. A renderer that reports progress before
  the delay as a negative number must be clamped at the call site, not inside the
  easing.
