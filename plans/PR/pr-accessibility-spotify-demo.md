## Summary

Adds accessible page semantics and actions, improves desktop IME, pointer, and scrollbar behavior, and expands the accessibility example and documentation. The Spotify Player now bundles its soundtrack and artwork, supports paginated data and navigation history, and appears as a compressed animated preview on the website.

## Motivation / context

- Plans: `plans/v0.0.2/accessibility-desktop-ime.md`
- Plans: `plans/v0.0.2/browser-apis-and-updater-proposal.md`
- Issues: no linked issue

## Changes

### Accessibility and interaction

- Add a page accessibility tree with roles, names, states, geometry, and actions.
- Improve IME composition and caret positioning, pointer routing, and scrollbar options.
- Expand the accessibility example and document native accessibility and desktop integration.

### Spotify Player

- Bundle 30 compressed MP3 tracks and local cover art with catalog descriptions and source/license records.
- Add pagination, navigation history, and playback controls, with tests for the local audio catalog and player behavior.
- Add the Spotify Player demo and showcase entry with a 657 KB animated WebP preview.

### Website and module packaging

- Limit documentation tables to a readable width across the site.
- Add nested Go module markers so website, assets, and development folders stay outside the library module archive.
- Refresh the generated Pages bundle and document how its module marker is restored after builds.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Spotify playback and artwork use bundled local files rather than runtime API requests. |
| **Memory** | Adds bundled audio and image assets to the examples module. |
| **Behavior / correctness** | Adds accessibility queries/actions, IME composition handling, pointer and scrollbar behavior, and Spotify pagination/history. |
| **API / CLI** | Adds public accessibility APIs and scrollbar configuration. |
| **Dependencies** | No new Go dependencies. |
| **Binary size / build time** | Adds the soundtrack and artwork to the Spotify example; the website preview is 657 KB. |

## Breaking changes / migration

None.

## Test plan

- [x] `make test`
- [x] `cd frontend && npm run build`
- [ ] Manual desktop audio playback smoke test was not repeated during PR preparation.

### Commands

```sh
make test
cd frontend && npm run build
```

## Screenshots / sample output

The homepage and showcase use the animated preview at `frontend/public/demos/videos/spotify.webp`.

## Related issues

- None identified.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues checked; no linked ticket applies
- [x] Filled body committed under `plans/PR/pr-accessibility-spotify-demo.md`

## Follow-ups (out of scope)

- A manual desktop audio playback smoke test remains useful after merge.

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API changes documented
- [ ] New rules have fixture coverage when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table is included below

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 11 | 80 | 31 |
| `.go` | 155 | 3766 | 517 |
| `.html` | 4 | 125 | 6 |
| `.jpg` | 40 | Binary | Binary |
| `.js` | 3 | 471 | 81 |
| `.json` | 1 | 892 | 0 |
| `.jsx` | 1 | 15 | 3 |
| `.md` | 32 | 1196 | 64 |
| `.mod` | 8 | 24 | 0 |
| `.mp3` | 30 | Binary | Binary |
| `.png` | 36 | Binary | Binary |
| `.txt` | 1 | 116 | 0 |
| `.webp` | 2 | Binary | Binary |
| **Total** | **324** | **6685** | **702** |
