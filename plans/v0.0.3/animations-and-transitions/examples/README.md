# v0.0.3 - examples migration track

> **Parent:** [animations and transitions](../README.md)
> **Status:** landed on `features/animations-and-transitions`; animations and
> css-preview are migrated, the manual exceptions are listed in the phase files.
> **Goal:** no example carries a hand-written animation driver, and every example
> that declares CSS motion renders it from the automatic runtime.

## The rule

An example may contain HTML, CSS, template data, and handlers. It may not
contain `cssanimation.Bind`, `csstransition.Bind`, `Player.Step`,
`Renderer.Step`, `Probe().MarkRect`, or `WantFrameIn` calls that exist only to
drive a CSS animation or transition.

`internal/frame` stays legitimate for motion with no CSS source: an audio
position, a game sprite, a data-driven gauge. The rule is about CSS, not about
every moving pixel.

## Inventory, from the tree on 2026-10-24

| Example | What it has | Verdict |
|---|---|---|
| `examples/animations` | 24 chips, 5 `@keyframes` blocks, hand-written bind, step, damage, and frame requests in `page.go`; 3 internal imports | Rewrite. Delete the driver, fix the CSS, keep the teaching cards |
| `examples/css-preview` | 43 property cards; `bind.go`, `player.go`, a queue, a drain clock, range/timeline/trigger tests; imports `cssanimation`, `cssmotion`, `csstransition`, `animation`, `transition`, `csstrigger` | Migrate the motion to the auto driver; keep the cost windows and the property cards |
| `examples/spotify-player` | Audio-position paint in `player/animate.go`: seek fill width, elapsed/remaining text, credit, equalizer. No `@keyframes` in its HTML | Keep. It is data-driven, not CSS motion. Audit only |
| `examples/dino` | Game loop in `dino/tick.go`, sprite state | Keep. Game motion |
| `examples/departure-board` | Flap spinner driven by `WantFrameIn` and Go state | Keep, unless its CSS declares a transition after the audit |
| `examples/eyes`, `examples/flappy-bird`, `examples/tetris` | Game loops | Keep |
| `examples/perf-benchmarks` | `benchutil/flap.go` tick fixture | Keep as a benchmark, not a motion driver |
| `cmd/animation-preview` | Low-level animation sampling harness | Keep as a manual harness; share code with `internal/motion` where it fits |
| `cmd/transition-preview` | Low-level transition sampling harness | Keep as a manual harness |

Only `examples/animations/animations.html` declares `@keyframes` today.
`examples/css-preview` builds its cards in Go, so a file search does not see them.
Neither fact excuses a driver in a third example later.

## Phase index

| # | Phase | Status | Depends on | Phase file |
|---|---|---|---|---|
| 01 | Inventory and guardrails | Planned | none | `phase-01-inventory-and-guardrails.md` |
| 02 | Animations example rewrite | Planned | animation 01, 02, 03 | `phase-02-animations-example.md` |
| 03 | css-preview migration | Planned | animation 06, transition 01 | `phase-03-css-preview-migration.md` |
| 04 | Remaining examples audit | Planned | 02, 03 | `phase-04-remaining-examples-audit.md` |
| 05 | Docs, maintenance, and gates | Planned | all | `phase-05-docs-maintenance-gates.md` |

## Master checklist

### Phase 01 - inventory and guardrails
- [ ] Record the inventory above in `examples/readme.md`
- [ ] Add a check that fails on a new `examples/**` import of a motion internal package
- [ ] Freeze `scripts/example-internal-allowlist.txt`: no new rows
- [ ] Note the existing rows that this track removes
- [ ] Confirm `cmd/*-preview` is exempt and say why
Gate: `make examples-check` and `make internal-import-check` pass with the new check

### Phase 02 - animations example rewrite
- [ ] Delete `bind`, `Tick`, `chipBox`, `chunk` damage, `WantFrameIn`, and `running` from `page.go`
- [ ] Restore the `.chip` rule lost at `animations.html:281-287`
- [ ] Stop `background-color` from recolouring labels
- [ ] Honor `animation-play-state` instead of forcing running
- [ ] Make the timeline/range/trigger cards express their demo in CSS
- [ ] Add a geometry assertion test: `chip-q1-0` is 76x32 and positioned
- [ ] Add a label-colour test
- [ ] Leave the file as `New`, `SetData`, `Redraw`, `Handle`
Gate: the page renders identically with no motion Go code

### Phase 03 - css-preview migration
- [ ] Delete `bind.go`, `player.go`, `player_step.go`, and the queue
- [ ] Keep the 43 cards and their property declarations
- [ ] Replace Start buttons with a class toggle that sets `animation-play-state`
- [ ] Keep the cost windows, or delete them with a note if the driver makes them dead
- [ ] Port the range, timeline, trigger, composition, and transition tests to the auto driver
- [ ] Remove the internal motion imports from the example
Gate: every catalog case passes with the auto driver, no queue, no manual step

### Phase 04 - remaining examples audit
- [ ] Spotify: confirm no CSS motion is hidden in its styles; keep the audio driver
- [ ] Dino, eyes, flappy-bird, tetris: confirm their ticks are game motion
- [ ] Departure-board: keep the flap driver, add a CSS transition if the audit finds one
- [ ] `cmd/animation-preview` and `cmd/transition-preview` use the shared sampling code
- [ ] Add one small class-toggle transition to a non-game example to prove the auto path outside the catalog
Gate: the audit table is filled and every remaining tick is justified in one line

### Phase 05 - docs, maintenance, and gates
- [ ] Move `examples/animations` out of `experimental` if it is now a maintained showcase
- [ ] Remove every motion row from the allowlist
- [ ] Update `examples/readme.md` for animations and css-preview
- [ ] Update the AGENTS.md examples paragraph
- [ ] Run `make examples-check`, `make internal-import-check`, and `make test`
Gate: no example imports a motion internal package

## Dependency notes

Phase 02 can start as soon as animation phase 01 lands; the CSS fixes do not
depend on the full property set, but the scale and rotate cards do need phase 02.

Phase 03 is the last migration because css-preview is the behaviour reference for
ranges, timelines, and triggers. Do not delete its manual path before the auto
path passes those tests.

Phase 04 is an audit, not a rewrite. The temptation is to replace every `SetTick`
with CSS. Do not: a game loop and an audio position are not declarations.

## Non-goals

- Removing `SetTick` from examples that use it for data or game state.
- Converting example markup to a new component system.
- Adding new examples beyond the conformance page in the animation track.
