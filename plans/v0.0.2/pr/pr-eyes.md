## Summary

Adds an animated Eyes demo for desktop, web, and Android. The eyes blink, look around, cycle expressions, and respond to touch, tilt, and shaking. The showcase now includes captures and a 60 fps recording, and the devtools dock accepts touch input.

## Motivation / context

- No separate design plan. The controls and Android setup are documented in `examples/eyes/README.md`.
- No linked issue exists for this work.

## Changes

### Eyes example

- Add the standalone `examples/eyes` page with expression changes, gaze movement, blinking, and a rotating spiral mode.
- Add the Android binding and app. Accelerometer readings are aligned with display rotation; shaking triggers the spiral expression.
- Add `scripts/android-eyes.sh` to bind the app, build its debug APK, and optionally install it.

### Showcase and devtools

- Add Eyes screenshots and a 60 fps recording to `showcase.md`, then rebuild the published `docs/` site.
- Add an Eyes showcase section and an image lightbox. Adjust the sticky showcase tabs and lightbox layout, and use the Octicon star in the site header.
- Route phone touches to the open devtools dock so tabs, rows, the picker, and scrolling work by touch.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No library performance claim. The demo animates through frame ticks; devtools touch handling runs while the dock is open. |
| **Memory** | The Eyes example embeds one 25 KB spiral image. No library memory change is measured. |
| **Behavior / correctness** | Adds the Eyes demo and touch handling for the devtools dock. |
| **API / CLI** | No public library API changes. Adds `go run ./examples/eyes`, its `-web` mode, and an Android build script. |
| **Dependencies** | No Go module changes. Adds an Android Gradle project and wrapper for the example. |
| **Binary size / build time** | Adds Android project files and showcase media. No library release binary is produced. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | No migration needed. |

## Test plan

- [x] `go test ./examples/eyes/eyes`
- [x] `go test ./internal/replay`
- [ ] Full `make test` was stopped after it exposed the pixel-read panic in `internal/replay`. The check now runs inside the opt-in GPU game loop, and the affected package passes. The full suite was not rerun.

### Commands

```sh
go test ./examples/eyes/eyes
go test ./internal/replay
```

## Screenshots / sample output

`showcase.md` includes the Eyes mood captures and `assets/screenshots/android/eyes/eyes-60fps.mp4`.

## Related issues

None. No issue exists for this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"`.
- [x] Labels applied (`enhancement`, `documentation`).
- [x] Related issues checked; none exist for this work.
- [x] This PR record is committed under `plans/v0.0.2/pr/pr-eyes.md`.

## Follow-ups (out of scope)

None known.

## Reviewer checklist

- [ ] Behavior matches the summary and test plan.
- [ ] No unrelated changes in the diff.
- [ ] Public API changes are documented.
- [ ] PR has an assignee and labels.
- [ ] Related issues use the correct closing keywords.
- [ ] No secrets or generated artifacts were committed.
- [ ] Diff statistics were generated from the merge base.

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.bat` | 1 | 92 | 0 |
| `.css` | 4 | 100 | 5 |
| `.go` | 52 | 1465 | 11 |
| `.gradle` | 4 | 50 | 0 |
| `.html` | 2 | 26 | 2 |
| `.jar` | 1 | Binary | Binary |
| `.java` | 1 | 137 | 0 |
| `.js` | 2 | 82 | 52 |
| `.jsx` | 2 | 77 | 13 |
| `.md` | 3 | 168 | 0 |
| `.mp4` | 1 | Binary | Binary |
| `.png` | 1 | Binary | Binary |
| `.properties` | 2 | 9 | 0 |
| `.sh` | 1 | 76 | 0 |
| `.webp` | 10 | Binary | Binary |
| `.xml` | 1 | 21 | 0 |
| No extension | 2 | 257 | 0 |
| **Total** | **90** | **2560** | **83** |
