# Phase 03 - css-preview migration

Goal: the 43-card property catalog runs on the automatic driver. Its tests stay
the behaviour reference, but the example stops binding, stepping, and queueing.

## Where it lives

- `examples/css-preview/bind.go:14-61` captures ops, builds paints, calls
  `cssanimation.Bind` or `csstransition.BindTo`, and rewrites `PlayState`.
- `examples/css-preview/player.go:8-65` is the tick: capture pending, redraw on a
  dirty tree, drain a queue, rebind on generation, step every sample, and ask for
  frames.
- `player_queue.go:57-59` drains queued samples; `player_step.go` advances them.
- `motion.go` and `catalog_triggers.go` carry the sample model and the trigger
  dispatch.
- The catalog tests (`catalog_animation_*_test.go`,
  `catalog_transition_*_test.go`, `catalog_trigger_*_test.go`) call the low-level
  packages directly and assert the sampled values. They are the best behaviour
  specification in the tree.

## Design

1. Cards stay. Each card's HTML, CSS, and Start button stay. The Go that binds a
   card is deleted.

2. Start and restart without a driver call. The catalog's own stylesheet already
   pauses every sample. Add a `running` class that sets
   `animation-play-state: running`. A Start click toggles the class through
   `SetData`, which redraws and rebinds. The runtime starts a newly running
   animation at the commit time.
   Restart needs a new animation instance. Use the catalog's existing trick of
   switching the animation name between two identical keyframe blocks
   (`kf-a`/`kf-b`) when the user presses Start again. That is CSS, and it is
   visible to a reader. Do not add a restart API for one example.

3. Queue and drain. Delete `drain`, `capturePending`, and the pending queue. The
   auto driver starts every visible card at commit. A card that scrolls into view
   starts when it is bound, which is the behaviour the queue was emulating.

4. Cost windows. The catalog measures bind and step cost. Replace the manual
   timers with `Probe().Motion()` counters and the existing frame timing. Keep a
   small "cost" panel that shows sampled players, writes, damage rects, and
   rebinds. The numbers must come from the runtime, not from a second clock.

5. Generation rebind. Delete the manual `gen` field and the `rebind` call. Phase
   01 owns it. Keep the catalog's resize settle logic only if the auto driver
   does not already handle resize through commit.

6. Trigger cards. Replace `catalog_triggers.go` dispatch with real clicks. The
   trigger scope is the card element; the driver dispatches on activation.
   Delete the manual `csstrigger` import.

7. Tests. Keep every assertion; change the harness:
   - build the page, bind through `New`/`Redraw`,
   - set the class that starts the card,
   - advance the page clock through the testing hook, not a manual `Step`,
   - assert the same sampled value at the same timestamp.

   The `Testing()` door already drives the tick and caret for tests. Add a motion
   seam only if the existing hook cannot advance the sample clock; prefer a
   deterministic clock on the runtime over a test-only export.

8. Determinism. The auto driver reads `time.Now()`. The tests need a fixed clock.
   Add an internal clock seam on `Runtime` (`now func() time.Time`) set by the
   page's testing hook. Do not add a public time override.

9. Cost comparison. Record one before and one after number for the 43-card page:
   bind time, sample time, damage rects, and frame time. Put them in
   `documentation/performance-baseline.md` with the phase note.

## Checklist

### A. Removal

- [ ] Delete `bind.go`.
- [ ] Delete `player.go`, `player_queue.go`, and `player_step.go`.
- [ ] Delete the manual generation field and rebind call.
- [ ] Delete `running` and the `PlayState` rewrite.
- [ ] Delete `catalog_triggers.go` dispatch.
- [ ] Remove `cssanimation`, `cssmotion`, `csstransition`, `animation`,
      `transition`, and `csstrigger` imports.

### B. CSS start and restart

- [ ] `running` class sets `animation-play-state: running`.
- [ ] Start click toggles the class through `SetData`.
- [ ] Restart switches between two identical keyframe names.
- [ ] A paused card keeps its pose across a quiz click.
- [ ] A card that scrolls in starts at bind.

### C. Cost panel

- [ ] Read counters from `Probe().Motion()`.
- [ ] Show sampled players, writes, damage rects, rebinds.
- [ ] Remove the second clock.
- [ ] Record before and after numbers.

### D. Tests

- [ ] Add the deterministic clock seam.
- [ ] Port every animation value test.
- [ ] Port every transition value test.
- [ ] Port every trigger test.
- [ ] Port the range, timeline, range-center, and composition tests.
- [ ] Keep the cost window tests working or delete them with a note.

### E. Gates

- [ ] No internal motion import in the example.
- [ ] `make examples-check`.
- [ ] `make test`.
- [ ] A manual pass over all 43 cards.

Gate: every catalog test passes through the automatic driver, and the example has
no bind, step, queue, or trigger dispatch.

## Risks

- The catalog is the only coverage for timelines, ranges, and triggers. Do not
  delete its manual path until the auto path passes those tests; run the two
  paths side by side for one release if needed.
- `time.Now()` in tests is flaky. The clock seam is mandatory.
- The Start/restart CSS trick is visible in the markup. Write a comment in the
  stylesheet so the next reader does not "simplify" it back into a Go call.
