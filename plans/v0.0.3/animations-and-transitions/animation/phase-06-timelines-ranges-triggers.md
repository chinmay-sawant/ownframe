# Phase 06 - timelines, ranges, triggers

Goal: an animation driven by scroll position, view progress, or a page event runs
from the page state, not from a host-supplied progress value. The caller stops
being responsible for `FrameInput`.

Depends on [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) for named
timelines. The `scroll()` and `view()` forms do not.

## Where it lives

- `internal/animation/player.go:22-26` `StepWithInput` accepts a `FrameInput`.
- `internal/animation/timing.go:62-70` defines the input: named timelines, scroll
  progress, view progress, and range values.
- `internal/animation/position.go:13-37` takes the time path when `Timeline` is
  empty or `auto`.
- `internal/animation/timeline.go:11-25` parses `scroll()` and `view()`, and
  `:27-68` resolves named ranges and edges.
- `internal/animation/center.go:8-50` resolves `animation-range-center`.
- `internal/csstrigger/dispatch.go:11-23` applies trigger actions to players.
- `examples/css-preview/catalog_triggers.go` and the range, timeline, and trigger
  tests are the behaviour reference; they call the runtime by hand today.
- `internal/page` already knows the scroll offset, the viewport height, and the
  element boxes at the same moment it samples.

## What is missing in the engine

`layout.Animation.Timeline` is a string, and `animation-timeline: scroll()` and
`view()` carry enough information to resolve on their own. A dashed-ident timeline
(`animation-timeline: --progress`) needs a matching `scroll-timeline-name` or
`view-timeline-name` declaration on an ancestor. The matrix does not implement
either declaring property today.

[Phase 11](phase-11-engine-named-timelines-and-svg-paint.md) adds the declarations and
exports `Box.Timelines`. Phase 06 consumes them; it does not parse CSS.

Two limits remain after that plan:
- nested element scroll containers are not implemented, so a named timeline on a
  non-scrolling element is inactive;
- `timeline-scope` is not implemented, so a name outside the ancestor chain is
  inactive.

## Design

1. One input builder. Add `func (r *Runtime) input(d *Display, box Box, now
   time.Time) animation.FrameInput` that fills scroll progress, view progress,
   named timelines where available, and the element's box. The sample loop never
   constructs a `FrameInput` inline.

2. Scroll progress. `0` at the top of the scroll container, `1` at the bottom,
   computed from the container's scroll offset and content height. The page
   already tracks the offset for drawing; reuse it, do not re-measure text.

3. View progress. `0` when the element's leading edge enters the viewport, `1`
   when its trailing edge leaves. Use the element's border box, the viewport
   height, and the scroll offset. Match the existing `view()` resolver so the
   preview tests keep passing.

4. Ranges. Resolve `entry`, `exit`, `cover`, `contain`, and percentage edges
   against the same geometry. `animation-range-start/end/center` all go through
   one `rangeFor(box, container)` function. Keep the existing edge parser; feed it
   real numbers.

5. Named timelines. Resolve `animation-timeline: --name` against the nearest
   ancestor's `Box.Timelines` and feed the progress into the existing
   `FrameInput` map. If the name does not resolve, the element does not scroll, or
   the range is zero, the animation is inactive. Do not fall back to wall-clock
   time; that is not what a browser does and it hides a broken name.

6. Triggers. A trigger declaration names a scope and a list of actions
   (`play`, `play-once`, `play-forwards`, `play-backwards`, `pause`, `reset`,
   `replay`). Dispatch when the element is activated:
   - click and tap on the element or a descendant,
   - keyboard activation when the element is the focused control,
   - the existing `Page.Press`/`Page.Release` path for pointer activation.
   Use `csstrigger.Dispatch` with the same action mapping the preview uses. Do
   not add a public "fire trigger" method; an app that wants a programmatic
   trigger changes the element's class or attribute, which is CSS.

7. Re-sample on scroll. A scroll or view timeline changes without a tick. The
   page's scroll path marks damage and requests a frame; the runtime recomputes
   the input when it samples. Do not run a tick just because a timeline exists
   while the page is still: a static scroll position is a static sample.

8. Resize and reload. A resize changes view progress for every element. The
   commit hook from phase 01 rebinds, and the next sample uses the new geometry.
   No separate path.

9. Trigger state survives rebind. `reset` clears the start time; `pause` keeps
   the offset; `replay` restarts without a rebind. Store trigger state on the
   entry, not on the player, so a generation change does not lose it.

## Checklist

### A. Input

- [ ] One `input()` builder used by every sample.
- [ ] Scroll progress from the page's scroll state.
- [ ] View progress matching `internal/animation`'s view resolver.
- [ ] Element box in page space with the current scroll offset applied.
- [ ] No allocation per sample: reuse a value, do not build maps.

### B. Ranges

- [ ] `entry`, `exit`, `cover`, `contain`.
- [ ] Percentage edges.
- [ ] `animation-range-start`, `-end`, `-center`.
- [ ] A range outside `[0,1]` clamps like the spec.
- [ ] Port the three catalog range tests to the auto driver.

### C. Timeline

- [ ] `scroll()` and `view()` resolve without host input.
- [ ] `none` disables the animation without error.
- [ ] `animation-timeline: --name` resolves through the nearest ancestor's
      `Box.Timelines`.
- [ ] An unknown, non-scrolling, or zero-range name leaves the animation inactive,
      with a counter.
- [ ] Named progress feeds `FrameInput` without a per-sample allocation.
- [ ] Pin [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) before claiming
      the named case.
- [ ] Nested scrollers and `timeline-scope` are reported as out of scope, not
      faked.

### D. Triggers

- [ ] Map click, tap, keyboard activation, and pointer press to the trigger.
- [ ] Apply every action in the list, in order.
- [ ] Keep trigger state across rebind.
- [ ] Respect the trigger scope element id.
- [ ] Test a reset followed by a replay at the same timestamp.

### E. Pacing

- [ ] A scroll timeline samples on the scroll that changed it.
- [ ] A still page with a scroll animation does not tick.
- [ ] A resize recomputes view progress on the next sample.

Gate: the catalog timeline, range, range-end, range-start, range-center, and
trigger cases pass with the example's manual dispatch removed.

## Risks

- View progress and scroll progress must use the same container metrics the
  window draws with, or an animation will be a frame or a few pixels off. Reuse
  the scroll state rather than reading the DOM-like tree again.
- A trigger that pauses and a class that pauses are two writers to the same
  player. Define precedence: the later event wins, and a rebind restores the
  stored state.
