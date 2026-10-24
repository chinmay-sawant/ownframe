# Phase 05 - public API, docs, conformance

Goal: transitions are automatic and documented, the off switch and counters are
the only public additions, and every transition row in the matrix has a card.

## Where it lives

- `page_window_door.go` and `page_probe.go` are the only doors this track adds
  to, shared with the animation track.
- `documentation/css-motion.md:22-30` describes the manual transition path and
  the `position`-only discrete limit.
- `documentation/features.md:117-120` repeats that the caller supplies time.
- `examples/css-preview/bind.go` and `player.go` still bind transitions by hand;
  `catalog_transition_*_test.go` is the behaviour reference.
- `documentation/compatibility-matrix.md:585-590` lists the longhand rows.

## Design

1. Public surface, shared with animation phase 09:

   ```go
   func (w Window) SetTransitions(on bool)
   func (w Window) Transitions() bool
   ```

   `Probe().Motion()` gains transition counters: bound, active, reversed,
   cancelled, snapped, and skipped for precedence. Default on.

2. Manual path. `internal/csstransition` and `internal/transition` stay as the
   implementation and as the preview tools. The docs say so, and say that a page
   must not step a renderer for an element the automatic driver owns.

3. Docs to update:
   - `css-motion.md`: replace the transition section with automatic capture,
     reversal, discrete rules, and the cross-fade; keep the manual API described
     as the preview path.
   - `features.md`: replace the caller-supplies-time rows and add the layer and
     cross-fade rows.
   - `repaint.md`: the outgoing-layer rule for `display` transitions.
   - `compatibility-matrix.md`: a host-side note on each transition row this
     track claims. No `@starting-style` row exists and the plan does not claim
     one.
   - `AGENTS.md`: the transition half of the automatic contract.
   - `examples/readme.md`: the css-preview entry after migration.

4. Conformance. One card per transition row:
   - a colour change from a class toggle,
   - an opacity change with overlapping children,
   - a transform change,
   - a width change with reflow,
   - a shadow change,
   - a display none/block pair,
   - a visibility pair,
   - an `allow-discrete` case,
   - a reversal at 25%, 50%, and 75%.

   Each card asserts the sampled value at five timestamps and has a golden frame.
   The display and reversal cards also have a browser note.

5. Cross-repository check. `make catalog-ownframe-check` covers transition rows in
   the blinkless checkout. Run it after the pin bump and record the result.

6. Migration proof. No example imports `internal/csstransition` or
   `internal/transition`. The css-preview cost windows keep working through the
   automatic driver, or they are removed with a note.

## Checklist

### A. API

- [ ] `SetTransitions` and `Transitions` on `Window`.
- [ ] Transition counters in `Motion()`.
- [ ] Default on, toggle mid-run leaves the presented value.
- [ ] Off releases entries and layers.
- [ ] API tests for on, off, and toggle.

### B. Docs

- [ ] Rewrite the transition section of `css-motion.md`.
- [ ] Update `features.md` rows and the feature table.
- [ ] Add the outgoing-layer rule to `repaint.md`.
- [ ] Add matrix notes for every claimed row and no `@starting-style` claim.
- [ ] Update the AGENTS.md paragraph.
- [ ] Update `examples/readme.md`.

### C. Conformance

- [ ] One card per transition row.
- [ ] Values at five timestamps.
- [ ] Golden frames.
- [ ] Reversal cards at three interruption points.
- [ ] Display and visibility cards with browser notes.
- [ ] `make catalog-ownframe-check` passes.

### D. Migration

- [ ] Remove `csstransition` from `examples/css-preview/bind.go`.
- [ ] Remove the transition queue from `player.go` if the automatic driver makes
      it dead.
- [ ] Keep the cost windows working or delete them with a note.
- [ ] No example imports a transition internal package.

### E. Gates

- [ ] `make test`
- [ ] `make lint`
- [ ] `make graphics-check`
- [ ] `make docs-check plans-check examples-check`
- [ ] `make internal-import-check`
- [ ] `make test-race`

Gate: all of E passes and the reversal and display cards match the browser.

## Risks

- The transition plan can drift from the animation plan if the two phases update
  the same docs separately. Land the shared API paragraph once, from the
  animation track, and have this phase only add the transition-specific text.
- A conformance card that snaps because of a missing engine export must show the
  snap, not a fake animation. The matrix note is the honest record.
