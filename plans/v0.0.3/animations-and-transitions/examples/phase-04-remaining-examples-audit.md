# Phase 04 - remaining examples audit

Goal: every remaining `SetTick` and every animated pixel in the example set has a
written reason, and one ordinary example proves the automatic transition path
outside the catalog.

## Where it lives

- `examples/spotify-player/player/animate.go:9-56` drives the seek fill,
  elapsed/remaining text, credit, and equalizer from `audio.Position()`.
- `examples/spotify-player/player/seek_animate.go` writes text ops and calls
  `Probe().Invalidate`.
- `examples/dino/dino/new.go:45` registers a tick; the tick moves game state.
- `examples/departure-board/board/board.go` registers a tick and uses
  `WantFrameIn` for the flap spinner.
- `examples/eyes`, `examples/flappy-bird`, `examples/tetris` run game loops.
- `examples/perf-benchmarks/benchutil/flap.go` is a benchmark fixture.
- `cmd/animation-preview/main.go` and `cmd/transition-preview/main.go` call the
  low-level samplers directly.

## Audit table

Fill the last column during the phase.

| Example | Tick purpose | CSS motion in its styles | Decision |
|---|---|---|---|
| spotify-player | Audio position and labels | None found in HTML | Keep; document |
| dino | Game state and sprites | None | Keep |
| eyes | Pointer-following | None | Keep |
| flappy-bird | Game state | None | Keep |
| tetris | Game state | None | Keep |
| departure-board | Flap timing | Audit | Keep unless a transition is declared |
| perf-benchmarks | Measurement | None | Keep |
| animation-preview | Manual sampler harness | Its own demo CSS | Keep, share code |
| transition-preview | Manual sampler harness | Its own demo CSS | Keep, share code |

## Design

1. Spotify is the case people will point at as "an animation". It is not CSS
   motion: every value comes from the audio decoder's position, and no
   declaration could produce it. Keep the driver and add a comment at the top of
   `animate.go`:
   `// Audio position, not CSS motion: the player's clock, not a timeline.`
   The same comment belongs in the track README so the next audit does not
   reopen it.

2. Dino, eyes, flappy-bird, and tetris are games. Their time source is the game
   loop, and their pixels are game state. Keep `SetTick`. Do not convert a sprite
   to a `@keyframes` block to satisfy a metric.

3. Departure-board is the borderline case. Its flap is data-driven, but if its
   stylesheet declares a colour or transform transition, the automatic runtime
   must handle it with no code. Audit the HTML and CSS, then either keep the
   driver or delete the CSS. Do not leave both.

4. Prove the automatic path outside the catalog. Add one class-toggle transition
   to a normal app example. `examples/wispr-flow-dashboard` is the best fit: its
   sidebar nav already has hover and active state, and a
   `transition: background-color 120ms` is a two-line change. The click handler
   keeps doing what it does; the visual change comes from CSS.
   Do not add a transition to the animations or css-preview examples for this
   proof; those are already covered.

5. The preview commands stay manual. They exercise `StepWithInput` with fixed
   progress values, which is the only way to inspect a range or a trigger without
   a window. Refactor them to call the same sampling function the runtime uses,
   so a preview cannot drift from production:
   - extract the sample call into an exported-to-internal function in
     `internal/motion`,
   - have both commands call it,
   - keep their own clock and progress sliders.

6. Benchmark fixtures stay. `perf-benchmarks` measures a tick, and a CSS-only
   fixture would measure a different path.

7. Each decision goes into `examples/readme.md` as one line. A future reviewer
   should not need to re-derive why Spotify has a tick.

## Checklist

### A. Spotify

- [ ] Confirm no `@keyframes` or `transition` in its HTML and CSS.
- [ ] Add the audio-position comment to `animate.go`.
- [ ] Add the reason to the readme table.
- [ ] Do not change the driver.

### B. Games

- [ ] Dino, eyes, flappy-bird, tetris: confirm the tick is game state.
- [ ] Add one-line reasons.
- [ ] Remove any dead animation helper found during the audit.

### C. Departure-board

- [ ] Audit the HTML for a transition or keyframe.
- [ ] Keep the flap driver or delete the CSS, not both.
- [ ] Test the flap after the audit.

### D. Real-app proof

- [ ] Add `transition: background-color 120ms` to the wispr dashboard sidebar.
- [ ] Toggle the active class through the existing click handler.
- [ ] Confirm the transition runs with no motion call in the example.
- [ ] Add a test that the sampled colour is between the two states mid-run.

### E. Preview commands

- [ ] Share the sample function with `internal/motion`.
- [ ] Keep the manual clock and progress input.
- [ ] Build both commands.
- [ ] Confirm the preview output matches the conformance page.

### F. Readme

- [ ] Fill the audit table in `examples/readme.md`.
- [ ] Add the audio-position and game-loop reasons.
- [ ] Link the wispr dashboard as the transition proof.

Gate: every remaining tick has a reason, the wispr transition runs from CSS, and
both preview commands build and sample through the shared function.

## Risks

- "Audit" can turn into a rewrite. Time-box it: one line per example, and only
  delete code that is provably dead.
- A hover transition in wispr may redraw the page on every pointer move, which
  would mask the transition cost. Measure a click-driven class change instead if
  hover is noisy.
- The preview commands are the only manual callers left. If they are refactored
  into the driver, the low-level path loses its test. Keep at least one command
  that steps a renderer explicitly.
