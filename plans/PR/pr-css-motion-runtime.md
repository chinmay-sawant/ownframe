## Summary

Add caller-driven runtimes for CSS animations, transitions, triggers, and view transitions. Bind Blinkless motion records to retained paint operations and sample them from caller-provided time, timeline, and trigger inputs. Turn the CSS property catalog into a runnable example with 43 Start buttons and document the runtime.

## Motivation / context

- Plans: No separate implementation plan. The inputs and limits are documented in `documentation/css-motion.md`.
- Issues: No issue is linked to this work.

## Changes

### CSS motion runtimes

- Add `internal/cssanimation` and `internal/animation` for binding animation records to operation indexes and sampling keyframes with timing, fill, direction, composition, range, and trigger support.
- Add scroll, view, and named timelines plus range boundaries. Keep each sampled animation attached to the paint operations of its target element so nested elements animate independently.
- Add `internal/csstransition` and `internal/transition` for binding transition records and stepping their paint values. Support `allow-discrete` changes to `position` at the midpoint.
- Add `internal/csstrigger` for trigger-scope resolution and dispatch of entry and exit actions.
- Add `internal/viewtransition` for pairing named elements across old and new displays, respecting scope, group, and class values. Build old and new paint layers with `Frame.Display`.
- Add `internal/cssmotion` for shared paint reads, writes, capture, and operation ownership.

### CSS catalog example

- Move the CSS property catalog into `examples/css-preview` with 43 numbered cards and WebKit aliases.
- Give every property its own Start button, value, action description, and visual sample.
- Bind the selected animation or transition from the page tick and stop ticking when no sample is active.
- Add `-perf` to record pipeline timings and open the devtools dock.
- Add `scripts/css-preview.sh` for focused preview runs and tests.

### Documentation and frontend

- Add `documentation/css-motion.md` and update the feature index, frames guide, theming guide, and documentation home.
- Add the guide to the frontend Styling navigation and extend the site check to open it.
- Rebuild the generated GitHub Pages output.
- Add animation and transition fixtures under `testdata/`.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Motion runs only when the caller steps it. The Page window does not start animations or transitions automatically. The example advances samples from `Page.SetTick` and stops when idle. No renderer-wide performance claim. |
| **Memory** | Each bound sample holds paint values for its target operations. There is no page-wide animation cache. |
| **Behavior / correctness** | Adds animation, transition, trigger, timeline, composition, range, and view-transition sampling. The catalog preview exercises each property. |
| **API / CLI** | No root library API changes. The runtime packages are internal. Adds `-perf` to `examples/css-preview` and two internal preview commands. |
| **Dependencies** | Moves the Blinkless pin to `caaff1db`. No new external modules. |
| **Binary size / build time** | Adds one example package and two small preview commands. No library binary output. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | No migration needed. |

## Test plan

- [x] `make test`
- [x] `make build`
- [x] `go test -p 1 ./examples/css-preview`
- [x] `bash scripts/pr-diff-stat.sh master`
- [x] `npm run build`, `npm test`, and `npm run check:site` in `frontend/` passed on the docs commit.

### Commands

```sh
make test
make build
go test -p 1 ./examples/css-preview
bash scripts/pr-diff-stat.sh master
```

## Screenshots / sample output

`go run ./examples/css-preview` opens the catalog. Each of the 43 motion properties has a numbered card and a Start button. The card shows the property value and the sample action, and the page readout shows the active property.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] `enhancement` label applied.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/PR/pr-css-motion-runtime.md`.

## Follow-ups (out of scope)

- View transitions still need caller-provided old and new displays. The Page window does not capture them automatically.
- `allow-discrete` transitions still need paint operations to appear or disappear for `visibility` and `display`. The current renderer handles `position` only.

## Reviewer checklist

- [ ] Behavior matches the summary and test plan.
- [ ] No unrelated changes in the diff.
- [ ] Public API changes are documented.
- [ ] PR has assignee and labels.
- [ ] Related issues use the correct closing keywords.
- [ ] No secrets or generated artifacts were committed.
- [ ] Diff statistics were generated from the merge base.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 2 | 573 | 0 |
| `.go` | 166 | 6281 | 3 |
| `.html` | 4 | 404 | 1 |
| `.js` | 3 | 105 | 30 |
| `.md` | 6 | 209 | 9 |
| `.mjs` | 1 | 5 | 1 |
| `.mod` | 1 | 1 | 1 |
| `.sh` | 1 | 31 | 0 |
| `.sum` | 2 | 4 | 2 |
| **Total** | **186** | **7613** | **47** |
