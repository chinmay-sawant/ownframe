# v0.0.3 - CSS transition track

> **Parent:** [animations and transitions](../README.md)
> **Status:** landed on `features/animations-and-transitions`; geometry and
> discrete rules are implemented, the display cross-fade is blocked.
> **Coverage:** [coverage.md](../coverage.md) maps every implemented animatable
> property to a phase, an exclusion, or a blocker.
> **Goal:** a style change that matches a `transition` declaration animates from
> the old paint to the new paint, without application code.

## Definition of done

Change a class, a data value, a hover state, or a theme. Every element whose
`transition` covers a changed property animates from where it was, including
mid-flight reversals. No example calls `csstransition.Bind`, `BindTo`, or
`Renderer.Step`.

## Dependency on the animation track

Transitions reuse the driver, the target model, the damage set, and the layer
cache built by the animation track. They are a second registry in the same
`internal/motion` package, not a second tick path.

| Needs | From |
|---|---|
| Registry, bind-at-commit hook, tick step, damage set | animation phase 01 |
| Transform keyframe model and matrix apply | animation phase 02 |
| Field masks, colour parsing, nearest-owner targeting | animation phase 03 |
| Retained layers for subtree transform and opacity | animation phase 08 |
| Counters and `Window`/`Probe` doors | animation phase 09 |

Do not start phase 01 of this track before animation phase 01 is merged.

## Phase index

| # | Phase | Status | Depends on | Phase file |
|---|---|---|---|---|
| 01 | Automatic capture and driver | Planned | animation 01 | `phase-01-automatic-capture-and-driver.md` |
| 02 | Property coverage and discrete behaviour | Planned | 01, animation 03 | `phase-02-property-coverage-and-discrete.md` |
| 03 | Interruption, reversal, and precedence | Planned | 01, 02 | `phase-03-interruption-reversal-precedence.md` |
| 04 | Layer integration | Planned | 03, animation 08 | `phase-04-layer-integration.md` |
| 05 | Public API, docs, conformance | Planned | all | `phase-05-public-api-docs-conformance.md` |

## Master checklist

### Phase 01 - automatic capture and driver
- [ ] Capture old and new paint at every display commit
- [ ] Match elements by id, with a counted fallback
- [ ] Bind one renderer per changed property with a matching transition record
- [ ] Step renderers in the page tick after the animation registry
- [ ] Preserve an in-flight sampled value as the next `from`
- [ ] Ask for frames only while a transition runs
- [ ] Delete the manual transition path from `examples/css-preview`
Gate: a class change on a colour, opacity, and transform transitions with no Go

### Phase 02 - property coverage and discrete behaviour
- [ ] Paint properties from the matrix
- [ ] Transform through the animation matrix model
- [ ] Geometry through the relayout override
- [ ] `allow-discrete` for `display`, `visibility`, `content-visibility`, `position`
- [ ] Integer swaps for `z-index`, `order`, `grid-*-start/end`, `line-clamp`
- [ ] Mode swaps for the no-interpolation properties in the phase table
- [ ] No `@starting-style`; the engine does not implement it and the plan does not claim it
Gate: one card per transitionable property, browser-compared

### Phase 03 - interruption, reversal, and precedence
- [ ] A new change mid-flight reverses from the current value
- [ ] Reversing shortens the duration by progress, per spec
- [ ] Animation wins over transition on the same property
- [ ] Removing the declaration cancels cleanly
- [ ] Delay and negative delay at the boundary
Gate: interrupt at 25%, 50%, and 75%, compare with a browser

### Phase 04 - layer integration
- [ ] Transform transitions ride the retained layer
- [ ] Opacity transitions composite with group alpha
- [ ] Discrete `display` cross-fades through two layers
- [ ] No re-bake of the subtree during the transition
Gate: a transitioned card does not rebake descendants

### Phase 05 - public API, docs, conformance
- [ ] `Window().SetTransitions(bool)` and `Probe().Motion()` fields
- [ ] Rewrite the transition section of `css-motion.md`
- [ ] Update `features.md`, `repaint.md`, `compatibility-matrix.md`
- [ ] Remove the css-preview manual binding
- [ ] Conformance cards for every transition row
Gate: `make test`, `make lint`, `make graphics-check`, docs and plans checks

## Non-goals

- View transitions. `internal/viewtransition` keeps its explicit old/new display
  model and is not wired to `Page` in this track.
- Transitioning a property the engine cannot interpolate. Those are discrete and
  belong to phase 02, not to a guessed curve.
- A second driver thread.
