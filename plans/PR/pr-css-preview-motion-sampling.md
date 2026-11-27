# fix(css-preview): sample catalog motion across animation phases

## Summary

The post-merge `master` CI run for PR #119 failed in `TestCatalogCardsRenderDifferentMidframes`. The test started each catalog card in a separate redraw, then compared every card at one fixed time; a looping, stepped animation can legitimately return to its base pose at that instant. Keep the visual assertion and sample each card until it reaches a changed frame.

## Motivation / context

- CI run: [46041947864](https://github.com/chinmay-sawant/ownframe/actions/runs/46041947864).
- Failing assertion: `animation-iteration-count did not move off its declared paint`.
- The catalog card repeats every 700 ms with `steps(5, end)`. The old 600 ms shared wait could land on the first step of a later iteration after the cards had already started at different times.

## Changes

- Replace the one-time paint comparison with a bounded sampling loop that ticks the real page runtime for up to one second.
- Keep the per-card assertion against its original declared paint and keep the requirement for 43 distinct painted samples.
- Leave the renderer, CSS fixture, and runtime behavior unchanged.

## Impact

| Area | Impact |
| --- | --- |
| **Performance** | No production-path changes. The test samples for at most one second. |
| **Memory** | No production memory changes. |
| **Behavior / correctness** | The test observes motion across repeated animation phases without treating a valid repeated pose as a failure. |
| **API / CLI** | No changes. |
| **Dependencies** | No changes. |
| **Binary size / build time** | No changes. |

## Breaking changes / migration

| Item | Migration |
| --- | --- |
| None | - |

## Test plan

- [x] `xvfb-run -a go test -count=20 -v ./examples/css-preview -run '^TestCatalogCardsRenderDifferentMidframes$'`
- [x] `xvfb-run -a make test`
- [x] `make docs-check plans-check examples-check tools-check repo-size-check`

## Related issues

None.

## PR metadata checklist

- [x] Self-assigned with `--assignee @me`.
- [x] Labeled `bug`.
- [x] No related issue applies.
- [x] This PR record is committed under `plans/PR/`.

## Follow-ups

None.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 1 | 35 | 14 |
| `.md` | 1 | 63 | 0 |
| **Total** | **2** | **98** | **14** |
