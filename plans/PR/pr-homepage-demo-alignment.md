## Summary

The homepage now places Spotify and the neighboring demo in the same landscape frame. The Spotify interface stays fully visible, and the previews align within each row.

## Motivation / context

The Spotify image's intrinsic 780×520 size made it much taller than the video frame. Its longer description also pushed the preview down.

## Changes

### Demo media

- Wrap each homepage demo in a shared 1.85:1 frame.
- Fit images inside the frame and keep video playback controls within it.

### Card alignment

- Let descriptions use the available row height so preview frames begin at the same vertical position.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No runtime work added. |
| **Memory** | No change. |
| **Behavior / correctness** | Homepage demo frames share a size; the Spotify image remains uncropped. |
| **API / CLI** | No change. |
| **Dependencies** | No change. |
| **Binary size / build time** | No change. |

## Breaking changes / migration

None.

## Test plan

- [x] Chrome render at 1440×1200; Spotify and Wispr frames align at 522×282 px.
- [x] `git diff --check`
- [ ] Automated tests were not run; this is a frontend layout change.

### Commands

```sh
git diff --check
```

## Screenshots / sample output

The local Chrome render confirmed the Spotify and Wispr preview frames share the same top and bottom edges.

## Related issues

- None identified.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`bug`)
- [x] No related ticket applies
- [x] Filled body committed under `plans/PR/pr-homepage-demo-alignment.md`

## Follow-ups (out of scope)

- None.

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API / CLI changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table is included below

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 1 | 6 | 7 |
| `.jsx` | 1 | 8 | 7 |
| `.md` | 1 | 84 | 0 |
| **Total** | **3** | **98** | **14** |
