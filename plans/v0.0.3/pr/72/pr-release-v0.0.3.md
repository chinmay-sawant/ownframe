## Summary

Record what landed from tag `v0.0.2` through `3ff7fd4f`, and move the pull request records from that cycle into `plans/v0.0.3/pr/<number>/`. The release body is `plans/v0.0.3/pr/72/release-v0.0.3.md`. This pull request does not create a tag or a GitHub Release.

## Motivation / context

- Plans: `plans/v0.0.3/pr/72/release-v0.0.3.md`, indexed from `plans/v0.0.3/README.md` and `plans/README.md`.
- The records were split across `plans/PR/`, `plans/v0.0.3/pr/`, `plans/v0.0.3/examples/`, `plans/v0.0.3/css-optimization/`, and `plans/v0.0.3/prod-ready-task/PR/`. v0.0.1 and v0.0.2 records stay where they are.
- No issue is fully completed here.

## Changes

### Release body

- Add the v0.0.3 body: the `Page` wrapper migration, motion, patch, replay, Linux AT-SPI, serve and fetch limits, components, the 15 new example directories, and the limits that stay open.
- The note names pins from `go.mod` at `3ff7fd4f`. The examples module still requires `ownframe v0.0.2` and still has `replace => ../`.
- The earlier draft that came with pull request #40 stays at `plans/v0.0.3/pr/40/draft-release-v0.0.3.md`.

### Pull request records

- Move each post-v0.0.2 record into a folder named with its GitHub number: 36, 37, 38, 39, 40, 41, 42, 46, 49, 52, 55, 58, 61, 66, and 70.
- The code-view file-browser record was folded into pull request #49, so it sits in that folder.
- Update the "committed under" path in each moved record, and the two template links whose relative depth changed.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No runtime change. Measurements in the note are quoted from the pull request records and were not remeasured here. |
| **Memory** | None. |
| **Behavior / correctness** | Documentation and plan layout only. |
| **API / CLI** | None. |
| **Dependencies** | None. |
| **Binary size / build time** | None. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None in this pull request | The note describes the break that already landed in pull request #40. |

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`make internal-import-check`)
- [x] No synonym names (`PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [x] Library code panics only for invariants; new user-facing failure paths return `error`
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's ctx
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr

## Test plan

- [x] `make plans-check`
- [ ] `make test` (markdown only; not run for this pull request)
- [ ] `make lint`

### Commands

```sh
make plans-check
```

## Screenshots / sample output

None. The release body is `plans/v0.0.3/pr/72/release-v0.0.3.md`.

## Related issues

None. Open issues #44, #45, and #48 stay open. This pull request does not close them.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] Labels: `documentation`
- [x] Related issues checked; none are closed by this work.
- [x] Filled body committed under `plans/v0.0.3/pr/72/pr-release-v0.0.3.md`.

## Follow-ups (out of scope)

- Tag `v0.0.3` and paste the release body into the GitHub Release.
- After that tag, bump the examples module require and drop `replace => ../` before tagging `examples/v0.0.3`.
- Pull request #71 refreshes the published site. It is separate from this note.

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

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.md` | 21 | 293 | 16 |
| **Total** | **21** | **293** | **16** |
