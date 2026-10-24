# Phase 05 - docs, maintenance, and gates

Goal: the example catalog tells the truth about which examples drive motion and
which do not, the allowlist has no motion rows, and the gates catch a regression.

## Where it lives

- `examples/maintenance.json` lists `animations` and `css-preview` under
  `experimental`, and the maintained map has six entries.
- `examples/readme.md:3-7` says the maintained set is six examples and the rest
  are experiments.
- `examples/readme.md` has no row for `animations` in the catalog table.
- `scripts/example-internal-allowlist.txt` has 40 pairs; the three animations
  rows are the only motion rows in the examples module.
- `documentation/features-examples.md` is the feature list the examples follow.

## Design

1. Decide the classification. A CSS motion showcase is a distinct user-facing
   capability, and the readme says to add a focused example when a capability has
   none. Recommend moving `animations` to `maintained` with the description
   "CSS animation and transition showcase: every property the engine samples".
   Keep `css-preview` experimental: it is a property lab, not a teaching page.

2. Add the readme rows:
   - `animations`: the automatic driver, the property cards, and what is not
     supported.
   - `css-preview`: the 43-card lab and the conformance reference.
   - one line in the dev-loop section: CSS motion needs no code.

3. Remove the three allowlist rows. After phases 02 and 03, the examples module
   imports no motion internal package. The allowlist keeps its other rows.

4. Add the ratchet from phase 01 to CI. Once the baseline is empty, delete the
   baseline mode so the check is absolute.

5. Update `documentation/features-examples.md` if it lists the animation example
   or the CSS motion capability. Search for both before editing.

6. Gates:
   - `make examples-check` passes with an empty motion baseline,
   - `make internal-import-check` reports the reduced pair set,
   - `make test` runs both modules,
   - `make lint` covers the examples module,
   - `make build` compiles every example without linking binaries.

7. Release note. The automatic behaviour is a user-visible change: a page that
   used to need a driver now animates on its own, and a page that drives the same
   ops by hand must call `SetAnimations(false)` or stop its tick. Write that in
   the release notes and in the migration section of `css-motion.md`.

## Checklist

- [ ] Move `animations` to the maintained map with a description, or record why
      it stays experimental.
- [ ] Add the animations row to the readme table.
- [ ] Add the css-preview row with its conformance role.
- [ ] Add the "CSS motion needs no code" line.
- [ ] Remove the three motion rows from the allowlist.
- [ ] Empty the ratchet baseline and delete the baseline mode.
- [ ] Update `features-examples.md` if it mentions the example or the capability.
- [ ] Write the release note and the migration note.
- [ ] Run every gate in design item 6.

Gate: `make examples-check` passes with no motion baseline, the allowlist has no
motion row, and the readme explains every remaining tick.

## Risks

- Moving an example into the maintained set is a support promise. If the track
  is not ready, leave it experimental and say so; do not promote a page that
  still has a blocked property.
- The migration note matters more than the feature note: an app with its own
  driver will fight the automatic one. Call that out with the exact off switch.
- A readme table drifts. Link to the conformance page rather than duplicating the
  property list.
