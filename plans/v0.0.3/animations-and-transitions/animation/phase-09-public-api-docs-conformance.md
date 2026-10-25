# Phase 09 - public API, docs, conformance

Goal: the automatic behaviour is the documented behaviour, the off switch and the
counters are the only public additions, and the property matrix has a test for
every row this track claims.

## Where it lives

- `page_window_door.go` exposes `Window()`; `page_probe.go` exposes `Probe()`.
  These are the only doors this track adds to.
- `documentation/css-motion.md:3-6` still says the runtime does not start by
  itself. `features.md:24` and `:117-120` repeat it. `frames.md:113-117` draws the
  boundary with `internal/frame`.
- `documentation/optimizations.md:432-439` records the layer gap this track
  closes.
- `plans/v0.0.3/prod-ready-task/out-of-1.0.md:37` (`OUT-12`) and `:52` (`OUT-21`)
  keep compositor layers out of 1.0. The user decision for this track is to bring
  them in; the rows must say so rather than contradict the plan.
- `examples/maintenance.json:14` classifies `animations` as experimental, and
  `scripts/example-internal-allowlist.txt:7-9` still lists its three internal
  imports.

## Design

1. Public surface, no more than this:

   ```go
   func (w Window) SetAnimations(on bool)
   func (w Window) Animations() bool
   func (w Window) SetTransitions(on bool)
   func (w Window) Transitions() bool
   func (p Probe) Motion() MotionStats
   ```

   `SetAnimations` and `SetTransitions` default to on. A zero `Config` behaves
   like Chrome: CSS motion runs. `MotionStats` reports bound, active, sampled,
   written, skipped, rebinds, damage rects, dropped rects, layers, layer bytes,
   rebakes, demotions, and the last bind error. It is a read-only value.

   No new `Page` method. No new root alias unless it is added to
   `plans/v0.0.3/prod-ready-task/phase-01-public-api.md` in the same change, per
   the rule in AGENTS.md.

2. Manual packages stay importable inside the module. Document them as the
   implementation and as preview tools, not as the application API. An
   application that needs a custom sample can still call `SetTick` and the manual
   packages, but it must turn the automatic driver off for that element, or it
   will fight the driver for the same ops. Document that interaction in
   `css-motion.md`.

3. Boundary with `internal/frame`. Write one paragraph in `frames.md`:
   - CSS declarations animate automatically.
   - `internal/frame` is for motion with no CSS source, such as an audio position
     or a game sprite.
   - Both write retained ops; do not drive the same op with both.

4. Docs list, in order:
   - `css-motion.md`: rewrite the opening, add "automatic by default", the off
     switch, the counters, the property tiers, the layer behaviour, and the
     blocked rows (`transform-origin`, perspective, geometry overrides, path).
     Document named timelines and SVG paint as phase 11 deliverables, with the
     nested-scroller and per-part limits.
   - `features.md`: replace rows 24 and 117-120, add the layer row, add the TPS
     option to the window table.
   - `frames.md`: the boundary paragraph and a link to the layer section.
   - `repaint.md`: the retained layer, its damage rule, and the caps.
   - `optimizations.md`: mark gap two closed and link phase 08.
   - `window.md`: `WindowOptions.TPS` and the 120 Hz sampling note.
   - `compatibility-matrix.md`: a short host-side section per row this track
     claims, plus the `transform-origin` and named-timeline gaps while they are
     real.
   - `AGENTS.md`: replace the paragraph that says callers bind and step with the
     automatic contract and the three doors.
   - `examples/readme.md`: the animations example entry, once the examples track
     lands.

5. Roadmap rows. Change `OUT-12` from "compositor layers out of 1.0" to
   "compositor layers move to the animation track, phase 08". Do the same for
   `OUT-21`. Do not leave two documents disagreeing about scope.

6. Conformance page. Add one page that renders one card per row in this track's
   claimed set, driven only by CSS. A table in the test file names the property,
   the declaration, and the expected sampled value at 0%, 25%, 50%, 75%, and
   100%. The page is the manual browser comparison and the golden source at the
   same time. It replaces the ad hoc cards in the animations example as the
   contract, while the example stays the teaching page.

7. Timing tests. For each row, assert the sampled value at fixed timestamps, not
   a screenshot only. A screenshot catches a wrong colour; a value catches a
   wrong curve.

8. Cross-repository check. `make catalog-ownframe-check` in the blinkless checkout
   already validates the 43 property cases. Run it after the pin bump and record
   the result in the phase note.

9. Migration proof. No example under `examples/` imports
   `internal/cssanimation`, `internal/csstransition`, `internal/animation`,
   `internal/transition`, or `internal/frame` when this phase closes. The
   allowlist has no rows for examples. `make internal-import-check` reports the
   reduced set.

## Checklist

### A. API

- [ ] `Window().SetAnimations`, `Animations`, `SetTransitions`, `Transitions`.
- [ ] `Probe().Motion()` returning `MotionStats`.
- [ ] Zero value keeps animations on.
- [ ] Turning a driver off releases its entries immediately.
- [ ] Add the alias to `phase-01-public-api.md` before exporting `MotionStats`.
- [ ] API tests for on, off, and toggle mid-animation.

### B. Docs

- [ ] Rewrite `css-motion.md`.
- [ ] Update `features.md` rows and the feature table.
- [ ] Add the `internal/frame` boundary to `frames.md`.
- [ ] Add the layer section to `repaint.md`.
- [ ] Mark gap two closed in `optimizations.md`.
- [ ] Document `WindowOptions.TPS` in `window.md`.
- [ ] Add the host-side notes to `compatibility-matrix.md`.
- [ ] Update the AGENTS.md paragraph.
- [ ] Update `examples/readme.md`.

### C. Plans

- [ ] Update `OUT-12` and `OUT-21` in `out-of-1.0.md`.
- [ ] Link this track from the optimizations README's phase 09 row.
- [ ] Record the before and after numbers in `performance-baseline.md`.
- [ ] Mark each phase's status in the two track READMEs.

### D. Conformance

- [ ] One card per claimed property row.
- [ ] Expected sampled values at five timestamps.
- [ ] Golden frames for the paint and layer cases.
- [ ] Browser comparison notes for the transform and origin cases.
- [ ] `make catalog-ownframe-check` passes on the pinned engine.
- [ ] A page with every card keeps 0 allocations in the sample path.

### E. Gates

- [ ] `make test`
- [ ] `make lint`
- [ ] `make graphics-check`
- [ ] `make docs-check plans-check examples-check`
- [ ] `make internal-import-check` shows no example rows for motion
- [ ] `make test-race` on the library module

Gate: all of E passes and the conformance page matches the browser at the five
timestamps.

## Risks

- Documentation drift is the main risk after a track this size. The phase is not
  done while any file still says the window does not start animations.
- A public counter that reports a guess is worse than no counter. Only report
  what is measured at the call site.
- The 1.0 scope change needs to be visible in the roadmap, not only in a plan
  folder. Update the exclusion rows in the same change as the API.
