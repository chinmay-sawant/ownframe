## Summary

Refresh the published site output so `master` matches a rebuild. PR #66 and #70 changed the embedded frontend examples and the demo wasm after the last `docs/` commit, and the push-to-master site check rejected that drift.

## Motivation / context

- Plans: `documentation/testing.md` says the freshness comparison runs on pushes to `master`, and the published output is refreshed once after a merge.
- The last generated commit is `3f88c3d8`. Head `3ff7fd4f` still had the previous bundles.
- No issue is linked.

## Changes

### Generated site

- Rebuild `docs/` with `npm --prefix frontend run build`.
- The page script moves from `index-C5MHthts.js` to `index-BPa1b1NE.js` because the bundled starter sources changed (`examples/frontend`, including the builder sample).
- `Documentation-BYveH4tz.js` replaces `Documentation-B1CDBcx9.js` because that chunk imports the page script by hash.
- The seven demo wasm binaries are rebuilt from current `master`. Spotify stays gzipped.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No runtime change. Published demo wasm is rebuilt from the same sources CI compiles. |
| **Memory** | None measured. |
| **Behavior / correctness** | GitHub Pages serves the demos and docs that match current `master`. |
| **API / CLI** | None. |
| **Dependencies** | None. |
| **Binary size / build time** | Demo wasm sizes move by a few tens of kilobytes. No library source change. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | No migration needed. |

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

- [x] `python3 scripts/check-generated-site.py` after the rebuild. A second build matched this output (`site-check: generated output matches the source`).
- [x] `npm --prefix frontend test`
- [ ] `npm --prefix frontend run check:site` (browser layouts). The site job runs that on the PR.

### Commands

```sh
npm --prefix frontend test
python3 scripts/check-generated-site.py
```

## Screenshots / sample output

```
site-check: generated output matches the source
```

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] Labels: `documentation`
- [x] Related issues checked; none are linked.
- [x] Filled body committed under `plans/v0.0.3/pr/71/pr-rebuild-site-output.md`.

## Follow-ups (out of scope)

- The v0.0.3 release note is a separate change.

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
| `.gz` | 1 | Binary | Binary |
| `.html` | 1 | 1 | 1 |
| `.js` | 2 | 55 | 37 |
| `.wasm` | 6 | Binary | Binary |
| **Total** | **10** | **56** | **38** |
