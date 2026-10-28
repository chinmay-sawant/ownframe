# feat(code-view): folder browser drives level, scrolling list, clipped path

Base: `example/code-view-and-performance`. Head: `example/code-view-file-browser`.
Range: `origin/example/code-view-and-performance..HEAD`, 1 commit (`8ea4b434`). Stacked on PR #49, which targets `master`.

## Summary

The Open folder browser gains a This PC drives level, a ten-row scrolling list, and clipped long paths. Up at a drive root steps out to every available drive instead of stalling. Follow-up to the Code Review example and perf pass in PR #49.

## Motivation / context

- Plans: `plans/v0.0.3/examples/code-view/`, `plans/v0.0.3/performance/`, `plans/v0.0.3/examples/pr-code-view-and-performance.md` (parent PR #49 description)
- Issues: see **Related issues**
- The browser could not step above a drive root, the folder list had no window, and long paths painted past the panel. This is the small stacked fix on top of PR #49.

## Changes

### File browser drives level

- Up at a drive root steps out to a This PC level listing every available drive (`browse_project.go`, `browseUpTarget`).
- New `browse_drives_test.go` (188 lines) covers the drive-listing level.
- `actions.go` routes `browse-up` through `browseUpTarget` and resets the list window on every root change.

### Scrolling list window

- New `browse_list_window.go` (106 lines): the folder list windows ten rows and scrolls with the wheel (`browseWheelAt` in `actions.go`, wired in `onWheel` before the editor path).
- `app.go` carries a `browseList browseListWindow` and renders `browseWindowedEntries()`.

### Clipped path

- New `browse_path_clip.go` (48 lines) plus `browse_path_clip_test.go` (42 lines): long browser paths clip with an ellipsis, full path stays in the tooltip.
- `app.go` clips via `clipBrowsePath(winW, ...)`; `components/chrome.css` keeps the head row and path inside the panel (`.browse-head` overflow hidden, `.browse-path` ellipsis trio, `.browse-list` overflow hidden, input `box-sizing: border-box`).
- `README.md` (+8 lines) documents the WSL handoff note plus the new browser behavior.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Negligible. One windowed slice per frame, clip math on width change only. |
| **Memory** | One small list-window struct on `App`. No new caches. |
| **Behavior / correctness** | Up at drive root now leaves the root instead of no-op. Wheel over the browser list scrolls the list, not the page. Long paths no longer paint past the panel. |
| **API / CLI** | None. Example-only change. |
| **Dependencies** | None. |
| **Binary size / build time** | Not measured. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`scripts/example-internal-allowlist.txt` untouched; grep finds none in the new files)
- [x] No synonym names (`browseUpTarget`, `browseListWindow`, `clipBrowsePath` are new names, no `PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR (no new exported methods)
- [x] Library code panics only for invariants (no library code touched)
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline (example `onClick`/`onWheel` keep `ctx` first; no new I/O paths)
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr (edits the existing Code Review example only)

## Test plan

- [ ] `make test` (not run)
- [x] `make lint` / `go vet` — PASS on `8ea4b434` (`make lint`: all 8 gates PASS, including Go library and examples v2.15.0, docs snippets, plan states, example inventory, internal imports, website ESLint)
- [ ] `make test-race` (library module; when shared mutable state changes) — FAIL, unrelated to this branch: `TestDiffArena_Allocs` in `internal/page/dirty` expects 0 allocs per run from the pooled arena, got 1 (`diff_test.go:172`). This branch touches only `examples/code-view`, so the failure is pre-existing on the parent branch, not from this diff. All other race packages passed.
- [x] `make build` (when binary output is part of the change) — PASS on `8ea4b434` (`go vet -p 1 ./... ./examples/...` clean)
- [ ] `make run` wall time vs baseline (no `run` target in this repo; N/A, template belongs to the other repo)
- [ ] `make reference-metrics` / gopdfsuit hard metrics if detector surface changed (no such target here; N/A)

### Commands

```sh
make lint
make test-race
make build
```

New tests: `browse_drives_test.go` (drive level), `browse_path_clip_test.go` (ellipsis clip), plus existing `browse_list_window` coverage via the windowed entries path.

## Screenshots / sample output

`make lint` on `8ea4b434`:

```
PASS maintenance tooling
PASS linter version regression
PASS Go library and examples (v2.15.0)
PASS documentation snippets
PASS plan states
PASS example maintenance inventory
PASS example internal imports
PASS website ESLint
```

`make build` (`go vet -p 1 ./... ./examples/...`): clean, no findings.
`make test-race`: FAIL in `internal/page/dirty` (`TestDiffArena_Allocs`: expected 0 allocs, got 1). Pre-existing, outside this diff.

## Related issues

- Relates to #49 (parent PR: Code Review example and hot-path perf pass, `example/code-view-and-performance` -> `master`).
- Relates to #48. No issue is fully completed here. #44 and #45 stay open per the parent record.

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied
- [ ] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/PR/pr-code-view-file-browser.md`

## Follow-ups (out of scope)

- Window capture of the This PC level and clipped path.
- Re-run benches only if the list-window slice shows up in profiles (not expected).

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] No new root aliases, synonym methods, or example `internal/` imports
- [ ] New exported methods return named `ownframe` types (not `host.Stats` / `*state.Store` unless re-exported here)
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh origin/example/code-view-and-performance` on `8ea4b434`:

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 1 | 10 | 5 |
| `.go` | 7 | 522 | 12 |
| `.md` | 1 | 8 | 0 |
| **Total** | **9** | **540** | **17** |
