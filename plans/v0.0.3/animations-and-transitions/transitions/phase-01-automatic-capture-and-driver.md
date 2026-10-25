# Phase 01 - automatic capture and driver

Goal: when a style change lands, every element whose new style declares a
matching transition animates from the value it had before the change. The
application does nothing.

## Where it lives

- `layout.Transition` carries `Property`, `DurationMS`, `DelayMS`,
  `TimingFunction`, and `Behavior` (`layout/layout.go:59-65`).
- `internal/csstransition/bind.go:12-52` `Bind` and `BindTo` connect a property,
  operation indices, a from-paint, a to-paint, and a start time to a
  `transition.Renderer`.
- `internal/transition/step.go:10-44` samples a renderer with one time value.
- `internal/page/page_draw.go:174-185` `commitDisplay` has the previous display
  (`p.draw.Display`) and the new one in the same function, and calls
  `dirtyFromDisplay(prev, display)` before `CommitList`.
- `internal/page/patch_wire.go:127-136` `spliceElement` swaps the list first;
  capture must happen before the swap.
- `internal/page/page_tick.go:34-65` is the sample point.
- `examples/css-preview/bind.go:36-40` and `player.go:36-41` are the manual
  transition path, with a generation rebind and a queue.

## What Chrome does

A style change produces a new computed style. For each property with a transition
declaration, Blink compares the old and new computed values, creates a transition
with the old value as the start, and hands the animation to the timeline. The old
value is the currently presented value, so a change during a running transition
reverses from where it is, not from the declared base.

## Design

1. Capture at commit, bind after commit. `commitDisplay` and `spliceElement` call
   one helper:

   ```go
   cap := p.motion.CaptureTransitions(prev, next) // old paints, matched by id
   p.draw.CommitList(next)
   p.generation++
   p.motion.BindTransitions(next, cap, time.Now())
   ```

   `CaptureTransitions` reads the transition records from the **new** box, because
   CSS starts a transition when the new style declares one. It matches the old box
   by `box.ID`. No id means no match; count it and skip. Do not guess by index.

2. Changed property detection. Compare the new box's record list with the old
   paint per op:
   - `opacity`: the paint opacity field,
   - `background-color`, `color`, `border-*-color`, `outline-color`,
     `text-decoration-color`: the RGB fields on the ops selected by the nearest
     owner and kind rules from animation phase 03,
   - `transform`: the op matrix once animation phase 02 is merged,
   - `width`, `height`, `top`, `left`, and the other geometry properties: defer to
     transition phase 02, which needs the relayout override.
   A record with `Property: all` matches every property that actually changed.

3. Same-op-count rule. `BindTo` needs the operation indices of the new display and
   a from-paint for each. If an element's op count or op kinds changed, skip the
   transition with a counter and leave the new paint. Phase 02 removes this limit
   for geometry. Do not bind a partial set; a half-interpolated element looks
   worse than a snap.

4. From-value is the presented value. The old display holds the last sampled
   paint, because the runtime writes into the display each tick. Capturing the old
   op paint therefore gives the correct reversal start for free. Do not capture
   the element's declared base style.

5. Step order. Transitions step before animations, in the same `runTick`:

   ```text
   transitions -> animations -> user callback
   ```

   CSS gives animations precedence over transitions on the same property. Writing
   the transition first and the animation second makes the animation win without
   a conflict table. If only a transition touches the field, it is the only
   writer.

6. Registry. One `entry` per element, property, and renderer, in the same
   `Runtime` as the animation entries, with its own slice. A rebind after an
   unrelated generation change keeps the entry and its renderer; it only
   re-resolves indices if the element's ops moved.

7. Pacing. A running transition asks `WantFrameIn(0)`. A delayed transition asks
   for the delay end. A finished transition stops asking. A page with only
   transitions parks exactly like a page with only animations.

8. Cancel. A property that leaves the new style, an element that leaves the
   display, or a `Redraw` that removes the element cancels its transition. Cancel
   writes the final value once and marks the element's box; it never leaves a
   half-painted op.

9. Off switch. `Window().SetTransitions(false)` stops capture and clears the
   registry. The animation switch is separate; a page can run one without the
   other.

## Checklist

### A. Capture

- [ ] `CaptureTransitions(prev, next)` with id matching.
- [ ] Read transition records from the new box.
- [ ] Compare per property using the phase 03 selection and field mask.
- [ ] Store a pending record with old paints and the target op indices.
- [ ] Called from `commitDisplay`, `paintFallback`, and `spliceElement`.
- [ ] Capture before `CommitList` and before `SwapList`.

### B. Bind

- [ ] `BindTransitions` after the generation bump.
- [ ] `csstransition.BindTo` with from, to, and the commit time.
- [ ] Skip on an op count or kind mismatch, with a counter.
- [ ] `all` expands to the changed properties only.
- [ ] A paused or reduced-motion page does not bind.

### C. Step

- [ ] Transitions before animations in `runTick`.
- [ ] Damage from the transition's old and new paint.
- [ ] Skip unchanged writes.
- [ ] Cancel writes the final value once.

### D. Pacing and switch

- [ ] `WantFrameIn(0)` while running.
- [ ] Delay deadline while waiting.
- [ ] No request when idle.
- [ ] `SetTransitions(false)` clears and stops capture.
- [ ] Toggle mid-transition leaves the page at the current value, not the base.

### E. Migration proof

- [ ] Remove `csstransition` from `examples/css-preview/bind.go`.
- [ ] A hover colour change in an example transitions with no Go.
- [ ] A `SetData` change to a width animates only after phase 02; assert the
      phase 01 snap behaviour in the meantime.

Gate: a class change on colour, opacity, and transform transitions with no
application code, and an interrupted change reverses from the presented value.

## Risks

- Capturing the old paint after `CommitList` loses it. The order in design item 1
  is mandatory.
- An element with no id cannot be matched. Count it, and do not fall back to a
  positional match: a splice can move indices.
- `bind` on every commit can bind a transition for a change the user cannot see,
  such as a hidden tab. That is still correct CSS; do not filter by visibility in
  this phase.
