## Summary

Rebuild the generated GitHub Pages site from the corrected frontend source. The published bundle was still using the CSS from before the homepage demo alignment fix.

## Motivation / context

The merged homepage fix changed `frontend/src`, while GitHub Pages serves the committed `docs/` output. The live page still referenced `index-DHyUJznb.css`, which predates the fix.

## Changes

### Published site

- Regenerate the static `docs/` site from the current `master` source.
- Replace the stale JavaScript and CSS bundles and update `docs/index.html` to reference them.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No runtime code changed. |
| **Memory** | No change. |
| **Behavior / correctness** | The published homepage uses the aligned 1.85:1 demo frames after Pages deploys this commit. |
| **API / CLI** | No change. |
| **Dependencies** | No change. |
| **Binary size / build time** | No change to application binaries. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

## Test plan

- [x] `npm run build` from `frontend/`.
- [x] `git diff --check`.
- [x] Confirmed the generated CSS sets the shared media frame ratio and keeps the Spotify image contained.
- Automated tests were not run. This change refreshes generated static assets.

### Commands

```sh
cd frontend && npm run build
git diff --check
```

## Screenshots / sample output

The rebuilt `docs/index.html` references `index-CR5s1zBB.css` and `index-C8IsGTNP.js`. The stylesheet includes the 1.85:1 shared media frame and `object-fit: contain` for demo images.

## Related issues

- No issue ID applies. This completes the Pages output for merged PR #22.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`).
- [x] Label applied.
- [x] No related issue applies.
- [x] Filled body committed under `plans/PR/pr-pages-demo-preview-layout.md`.

## Follow-ups (out of scope)

- None.

## Reviewer checklist

- [ ] Behavior matches summary and test plan.
- [ ] No unrelated changes in diff.
- [ ] Public API / CLI changes documented.
- [ ] PR has assignee and label.
- [ ] Related issues use correct keywords.
- [ ] No secrets or generated artifacts outside the Pages output are committed.
- [ ] Diff-stat-by-extension table is included below.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 2 | 1 | 1 |
| `.html` | 1 | 2 | 2 |
| `.js` | 1 | 9 | 9 |
| `.md` | 1 | 78 | 0 |
| **Total** | **5** | **90** | **12** |
