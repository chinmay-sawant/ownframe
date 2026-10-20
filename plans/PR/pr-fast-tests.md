## Summary

Cut the slow example test suites from minutes to seconds without changing existing rendering behaviour. `Page.RedrawIfChanged` adds an opt-in path that skips layout when the frame inputs are unchanged, and the examples stop doing redundant PNG checks, full fixture seeds, and repeated SVG rasterization in tests.

## Motivation / context

- Plans: no separate plan; the work is test and render-pipeline optimization.
- Issues: no issue is linked to this work.
- Baseline package times included Spotify at 355s, corebackend at 32s, Teams at 26.7s, and LG remote at 26.3s.

## Changes

### Page rendering

- Add `Page.RedrawIfChanged`, keyed on source, viewport, focus, hover, active state, theme, and image revision. `Redraw` keeps its always-render behaviour.
- Track an image revision so replacing an image invalidates the render key.
- Move display commit bookkeeping next to the render key without changing existing redraw semantics.

### Example tests

- Remove full-page PNG sanity checks from Spotify, accessibility, Telegram, and Wispr tests.
- Use `RedrawIfChanged` for repeated test redraws in Spotify, accessibility, and Wispr.
- Seed only the demo spreadsheet workbook in corebackend tests instead of the demo plus a 100,000-row stress workbook.
- Substitute tiny PNG icons for SVGs in Teams and LG remote tests, avoiding repeated rasterization and PNG encoding.
- Avoid redundant redraws and use `t.Parallel()` in the Teams tests.

### Teams example

- Load only the active menu stylesheet through the page theme instead of matching all six menu stylesheets on every layout.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Spotify 355s to 7.6s, corebackend 32s to 1.84s, LG remote 26.3s to 2.34s, Wispr 10.8s to 8.2s, and Teams 26.7s to roughly 9-13s depending on load (10.05s in the full suite). |
| **Memory** | `Page` grows by the stored render key. No new retained frame buffers. |
| **Behavior / correctness** | Existing `Redraw` callers are unchanged. Teams renders the same active menu with only its own stylesheet loaded. |
| **API / CLI** | Adds the public `Page.RedrawIfChanged` method. No CLI changes. |
| **Dependencies** | None. |
| **Binary size / build time** | No meaningful change; test-only fixtures are smaller. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing `Redraw` calls keep their behavior. Use `RedrawIfChanged` only where skipping an unchanged frame is desired. |

## Test plan

- [x] `make test`
- [x] `go test -count=1 ./internal/page ./examples/teams/app`
- [x] Targeted tests for every changed example package
- [ ] `make build` was not run separately; the full test run compiles every package.

### Commands

```sh
make test
go test -count=1 ./internal/page ./examples/teams/app
```

## Screenshots / sample output

```text
spotify-player/player        355s -> 7.60s
spreadsheet/ui/corebackend    32s -> 1.84s
lg-remote/remote            26.3s -> 2.34s
teams/app                   26.7s -> 9-13s
wispr-flow-dashboard/app    10.8s -> 8.20s
telegram/telegram                  7.15s
accessibility/demo                 0.33s
```

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] `enhancement` label applied when opened.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/PR/pr-fast-tests.md`.

## Follow-ups (out of scope)

- `examples/css-preview` remains slow and is being handled separately.

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
|-----------|-------|------------|-----------|
| `.go` | 61 | 386 | 146 |
| `.md` | 1 | 105 | 0 |
| **Total** | **62** | **491** | **146** |
