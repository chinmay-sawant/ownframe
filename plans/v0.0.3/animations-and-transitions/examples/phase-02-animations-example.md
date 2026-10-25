# Phase 02 - animations example rewrite

Goal: `examples/animations` becomes HTML, CSS, template data, and click handlers.
Delete every line that exists to drive the animations by hand.

## Where it lives

- `examples/animations/page.go:12-17` imports `layout`, `internal/animation`,
  `internal/cssanimation`, and `internal/cssmotion`.
- `:114-153` `newApp` calls `Redraw`, `bind`, `SetTick`, and `Handle`.
- `:178-228` is the tick: `player.Step(now)`, a union damage over every chip, and
  `WantFrameIn(frameInterval)`.
- `:231-272` is `bind`: `OperationIndices`, `Capture`, `cssanimation.Bind`.
- `:275-289` `running` forces `PlayState` and rewrites triggers.
- `animations.html:281-287` is the truncated rule that drops the whole `.chip`
  block; the chips have no position, size, or radius.
- `examples/animations/page_test.go:39-65` only asserts that a player exists, so
  the geometry regression passed CI.

## Design

1. Delete the driver. The final `page.go` has:
   - `newApp`: `ownframe.New(Config{...})`, `SetData`, `Redraw`, `Handle`.
   - `Tick`: gone.
   - `bind`, `chipBox`, `opX`, `running`, `chunk`: gone.
   - `handleClick`: stays, because it changes quiz state and restarts a track.
     A restart becomes a class or data change that the CSS expresses, not a
     player call.

2. Fix the stylesheet first, as a separate commit. Restore the missing selector
   before `.chip` or delete the orphan declarations at 281-287. Add a test that
   reads `page.Boxes()` and asserts a chip is 76x32 at its expected position.
   This test should fail before the fix.

3. Colour targeting. The label must not take the chip's animated background.
   Animation phase 03's nearest-owner rule fixes the runtime. Until it lands, the
   example can scope its `background-color` keyframes to a background element,
   but that is a workaround, not the fix. The phase is done when the runtime fix
   is in and the label keeps its colour with no markup change.

4. Play state. Delete `running`. A paused sample stays paused. The pause button
   sets the class that the CSS already declares
   (`body.is-paused .chip { animation-play-state: paused }`). Resume continues
   from the phase, not from zero; animation phase 05 provides the resume path.

5. Cards. Keep the timing, direction, delay, fill, and WebKit cards, and add:
   - rotate, scale, skew, matrix,
   - transform-origin: top left,
   - opacity with overlapping children,
   - colour, border colour, outline colour,
   - width, height, border-radius, font-size,
   - filter, marked as the layer case,
   - a scroll timeline and a range card,
   - a trigger card that uses a click, not a Go dispatch.

   Every card is selected by the conformance page in animation phase 09; the
   example stays the teaching page and does not need to be exhaustive.

6. Sources of truth. The HTML file carries the declaration. `page.go` carries the
   quiz data and handlers. A card's number, title, and explanation are data, not
   style.

7. Tests:
   - geometry: 76x32 and absolute position for `chip-q1-0`,
   - label colour: text op RGB differs from the chip background at 50%,
   - play state: paused page does not advance a sampled value,
   - resume: pause for 500 ms, resume, sample equals an uninterrupted run,
   - no driver: a test that greps `page.go` for `SetTick`, `Bind`, `MarkRect`, and
     `WantFrameIn`, or better, asserts `page.Ticking() == false`.

8. Documentation. The example README explains the property each card
   demonstrates and points at the conformance page for the normative list.

## Checklist

### A. Stylesheet

- [ ] Restore or delete the orphan rule at `animations.html:281-287`.
- [ ] Confirm `.chip` applies: position absolute, 76x32, radius, colour.
- [ ] Add the geometry test; watch it fail before the fix.
- [ ] Keep the 5 existing `@keyframes` blocks and their WebKit alias.

### B. Driver removal

- [ ] Delete `bind`, `running`, `chipBox`, `opX`, `chunk`.
- [ ] Delete `Tick` and `SetTick`.
- [ ] Delete `WantFrameIn`, `MarkRect`, and the union damage.
- [ ] Delete the `operation indices`, `base paint`, and `start` fields from the
      chip model.
- [ ] Remove the four internal imports.

### C. Behaviour

- [ ] Pause keeps the current pose.
- [ ] Resume continues; it does not restart.
- [ ] A quiz click does not snap the chips to base.
- [ ] A paused page parks; a running page asks for frames.
- [ ] Clicking a track restarts that track only.

### D. Cards

- [ ] Add the transform cards from animation phase 02.
- [ ] Add the colour and opacity cards from phase 03.
- [ ] Add the geometry cards from phase 04.
- [ ] Add the filter card, marked as the layer case.
- [ ] Add the scroll and trigger cards from phase 06.

### E. Tests and docs

- [ ] Geometry test.
- [ ] Label-colour test.
- [ ] Pause and resume tests.
- [ ] No-tick test.
- [ ] README per-card table.

Gate: `page.go` has no motion call and the page looks the same at the same
timestamps. The geometry and label tests pass.

## Risks

- The `.chip` rule drop is a parser behaviour, not a typo to "fix in the browser".
  The test must read `Boxes()`, not a screenshot, so the next parser change
  cannot hide it again.
- Removing the tick while the runtime is not merged leaves a static page. Land
  this phase after animation phase 01 is green, in the same release window.
- A trigger card driven by a click needs animation phase 06; keep it commented in
  the data table until then rather than adding a Go dispatch back.
