## Summary

Expand the Eyes example into a 21-expression animated demo with automatic and static modes, gaze and spiral motion, touch controls, and Android support. This branch updates the docs and showcase, adds a signed arm64 release APK, and refines bored, happy, excited, and nervous expressions.

## Motivation / context

- Plans: `examples/eyes/README.md` documents the controls and Android build; no separate implementation plan exists.
- Issues: No issue is linked to this work.

## Changes

### Eyes demo

- Add animated expression poses, blinking, gaze motion, eased transitions, auto/static modes, and touch, key, drag, and device-motion controls.
- Refine bored with slow side glances, happy with bright-to-grin motion, excited with a quick shake, and nervous with a lower gaze and softer eye movement. Keep worried and shy unchanged.

### Android build and distribution

- Add the Android Eyes binding and support `eyes` in the signed arm64 release script.
- Check in the signed APK at `examples/eyes/android/artifacts/eyes-arm64-release.apk` and keep its digest in `examples/android-host/APK-SHA256SUMS`.

### Showcase and documentation

- Add the Eyes demo recording and preview image, update the showcase and published docs assets, and render raw HTML in the documentation markdown.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Changes stay in the Eyes example and update retained eye operations from the frame tick. No renderer-wide performance claim. |
| **Memory** | Adds a signed 29,405,846-byte APK and showcase media. No library memory claim. |
| **Behavior / correctness** | Adds 21 selectable expressions, animated gaze and spiral behavior, touch controls, and Android motion input. |
| **API / CLI** | No root library API changes. Adds `eyes` to `scripts/android-release.sh`. |
| **Dependencies** | Adds the frontend `rehype-raw` package. No Go module changes. |
| **Binary size / build time** | Checks in a signed arm64 APK of 29,405,846 bytes. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | No migration needed. |

## Test plan

- [x] `GOCACHE=/tmp/ownframe-eyes-gocache go test -p 1 ./examples/eyes/eyes -count=1`
- [x] `GOCACHE=/tmp/ownframe-eyes-gocache sh scripts/android-release.sh eyes`; `apksigner verify` passed for v1, v2, and v3 signatures.
- [x] Install and launch the updated debug build on the Pixel 7 using Windows ADB.
- [ ] Full `make test` was interrupted at the user's request; the suite was not completed or rerun.

### Commands

```sh
GOCACHE=/tmp/ownframe-eyes-gocache go test -p 1 ./examples/eyes/eyes -count=1
GOCACHE=/tmp/ownframe-eyes-gocache sh scripts/android-release.sh eyes
ADB=/mnt/c/Users/acer/platform-tools/adb.exe sh scripts/android-eyes.sh install
```

## Screenshots / sample output

The branch includes `assets/screenshots/android/eyes/eyes2-60fps.mp4` and its WebP preview. The updated debug build was installed and launched on the Pixel 7; no new recording was captured for the final pose refinements.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] `enhancement` label applied when opened.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/PR/pr-eyes-expression-animation.md`.

## Follow-ups (out of scope)

None known.

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
| `.apk` | 1 | Binary | Binary |
| `.css` | 1 | 12 | 11 |
| `.go` | 46 | 1170 | 187 |
| `.html` | 2 | 5 | 3 |
| `.js` | 3 | 2692 | 2677 |
| `.json` | 2 | 169 | 0 |
| `.jsx` | 2 | 17 | 2 |
| `.md` | 4 | 204 | 89 |
| `.mp4` | 1 | Binary | Binary |
| `.py` | 1 | 149 | 0 |
| `.sh` | 1 | 6 | 1 |
| `.webp` | 2 | Binary | Binary |
| No extension | 1 | 1 | 0 |
| **Total** | **67** | **4425** | **2970** |
