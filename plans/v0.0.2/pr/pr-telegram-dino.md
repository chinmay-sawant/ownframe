# feat(examples): add the Telegram and dino phone demos with touch, IME, and pinned bars

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md), saved before opening the PR.

- Base: `master` (`2747461`)
- Head: `store/example-telegram` (`39befdf`, pushed to `origin`)
- Commits: 8
- Diff: 199 files, +7092 / -159 (158 added, 41 modified)

---

## Summary

The branch ships two phone demos and the library pieces they needed. `examples/telegram` is a chat list with search, contacts, a settings tab with a dark theme, one open conversation, camera and gallery attachments, reactions, and a gift bubble, packaged as an Android project. `examples/dino` gains tap and swipe controls, a phone layout, and its own Android project. The library adds a touch gesture path (tap, swipe, hold, long press), an Android and iOS soft keyboard session through Ebiten's `exp/textinput`, and viewport pinning in the replay so the chat header and composer stay at the viewport edges while messages scroll. On a Pixel 7 a drag with the bars pinned ran at 55-85 fps, up from a full-page redraw of 70-200 ms per scroll frame.

## Motivation / context

- The library had no way to keep chrome pinned during a scroll, no soft keyboard on phones, and no long-press or swipe hooks for a screen. A phone build could only scroll and type through the window's own key path.
- The two demos are the proownframe: they exercise the bind, the AAR, Gradle, adb, the system insets, the camera and picker, and the IME in real UI.
- Plans: none for the phone set. This extends the v0.0.2 example suite (`plans/v0.0.2/README.md`) and the platform notes in `documentation/platforms.md`.
- Issues: none. The repo has no tracker entry for this work.

---

## What was there, what is now

Everything below compares `master` (`2747461`) with this branch (`39befdf`).

| Area | What was there | What is now |
|---|---|---|
| Touch input | Pointer and wheel only. A touch reached the page as a click and nothing else. | Tap, swipe, hold, and long press are separate gestures with fixed thresholds. A claimed long press routes movement to the selection drag instead of scrolling. |
| Soft keyboard | Absent, desktop included. | Focusing a field on Android or iOS opens an `exp/textinput` session. The window paints the preedit as field text, a commit goes through `Type` or `IMEReplace`, and an IME-consumed tick skips the window's own typing. Desktop and wasm keep key input. |
| Pinned chrome | A scroll translated every operation, so a header or composer could not stay at the viewport edge. | `Page.SetViewportPinZ` marks a z layer, the window records the offset each display was built with, and the replay draws that layer at the build offset while the rest blits. On the phone one debounced redraw at 120 ms bakes the final positions. |
| Replay and buffers | One `Draw` fill for all operations. | `DrawUnfixed` fills the scrolled buffer, `DrawFixed` paints fixed operations after the blit, and `DrawRect` leaves them out of dirty rects. |
| Host and page API | No input or pinning hooks. | `host.IME`, `host.LongPresser`, `host.Swiper`, `host.ViewportPinner`; `Page.IMEContext`, `IMEReplace`, `LongPress`, `Swipe`, `SetViewportPinZ`; `Handlers.LongPress` and `Handlers.Swipe`. All optional and additive. |
| Telegram example | Did not exist. | A phone-first demo with a Gradle project, a script, and a README. |
| Dino example | A keyboard-only desktop and wasm page. | Tap and swipe controls, a phone layout that scales the 900x300 scene to the width, and an Android project. |
| Docs | Desktop and wasm only. | README, AGENTS.md, and six documentation files cover touch, long press, IME, insets, and position limits; `showcase.md` renders seven Pixel 7 screenshots. |
| Tests | None for any of this. | 37 new test files, 68 `func Test`, across `internal/page`, `internal/window`, `examples/dino/dino`, and `examples/telegram/telegram`. |

---

## Changes

### Window input layer

- `internal/window/touch_gesture.go` tracks fingers across frames. A finger that moves past 8 px latches to drag, and `dx`/`dy` accumulate only while exactly one finger is down. A second finger latches `multi` until every finger lifts, cancels the hold, and suppresses tap and swipe.
- `internal/window/long_press.go` fires after a press held within 8 px for about 450 ms, once, at the arm point. Movement past the slop cancels it while unclaimed. The same watch runs for the mouse, so a 450 ms mouse press also fires `LongPress`.
- `internal/window/touch_swipe.go` emits one swipe at lift when a single finger moved at least 24 px on either axis, measured from first sighting to lift, with no claim and no tap. Mouse drags never swipe.
- `internal/window/hold.go` bridges the watch into the frame loop, calls `host.LongPresser` once, and sets the drag flag so a claimed hold extends the selection through `dragAt` instead of scrolling.
- Commit `ae2fd3b` fixes a tap that left the watch armed. Android touches use `tapAt` alone, and `tapAt` never disarmed the watch, so 450 ms after a still tap the frame loop fired `LongPress`. In Telegram a press opened the reaction bar and the stale hold reopened it. The fix calls `hold.release()` in `tapAt` (`internal/window/tap_at.go:9`); `touch_tap_hold_test.go` fails before it and passes after. The function moved out of `pointer_touch.go` because that file was at the 2000-character limit.
- Tests pin the thresholds and the win rules: a 12 px move is neither tap nor swipe, a 4 px move still fires the hold, a 9 px move cancels it, a claimed drag scrolls nothing, and an unclaimed move scrolls and sends no drag.

### Soft keyboard and IME

- The window IME files are build-tagged `android || ios`. Everywhere else `imeInit`, `imeUpdate`, and `imeHandled` are no-ops (`internal/window/ime_other.go`).
- `ime_mobile.go` installs the four `textinput.Composer` callbacks and reads `host.Focuser.FocusID()` each tick. A changed id cancels the old session and adopts the new one; `imeNewSession` returns nil unless the screen implements `host.IME` and a field is focused.
- The preedit is painted into the page: `ime_show.go` backspaces one rune per rune of the previous preedit and then types the new text. A commit appends with `Type`, or calls `IMEReplace` when the IME sends surrounding text.
- `internal/page/ime_replace.go` snaps byte offsets to rune starts, goes through `editField` so `BeforeEdit` and `Change` run, and places the caret inside the inserted text rather than after a suffix.
- The caret box subtracts the window scroll offset (`internal/page/ime.go:28-30`). Without that, a scrolled chat reported a document position far below the screen, Ebiten panned the canvas, and the header left the screen when the keyboard opened.
- `ime_watch.go` restarts the session when the focused caret box moves more than 4 px. The restart commits the live composition instead of dropping it.
- Android `MainActivity` unions system bars, cutout, and IME insets, converts pixels to dp, and calls `Mobile.setInsets`. The page pads for them, keeps the composer above the keyboard, and scrolls the newest message up when the bottom inset grows. `refocusComposer` returns focus after send when the bottom inset is at least 100 px, so the keyboard stays up.

### Viewport pinning in the replay

- `host.ViewportPinner.ViewportPinZ` and `Page.SetViewportPinZ` select the layer. Telegram pins z 2, which covers the 60 px header, the composer, and the inset pads.
- `setDisplay` records the scroll offset the display was built with (`internal/window/sync.go:68-72`). `directReplay` passes that offset and the pin threshold to `replay.DrawVisiblePinned`.
- `internal/replay/visible.go` draws pinned operations at the pin offset and culls them with the pinned rect; everything else keeps the scroll translation. `drawOp` ignores translation for engine `op.Fixed` operations.
- The buffered path excludes fixed operations from the scrolled fill and paints them after the blit, so a scrolled message cannot cover the bars.
- Telegram's `Pin` returns false on the phone, so a scroll frame blits and one 120 ms settle redraw bakes the final positions. Hit boxes for the bars match after that redraw.

### Telegram example

- One embedded HTML template assembled from per-screen fragments (`chats`, `contacts`, `settings`, `thread`, `tabs`) with four CSS sheets. `New` seeds seven chats, twenty unread, contacts, and Anna Petrova's 35-message thread, then wires one `onClick` switch over action prefixes.
- Search filters by chat name or preview; tabs switch between Chats, Contacts, and Settings; the dark toggle rebuilds the theme, and the Android bar icons follow it.
- Opening a chat clears unread and scrolls to the end. Send appends an own bubble with read ticks, updates the preview, and scrolls. The composer keeps the keyboard up after send.
- The paperclip opens a 150 px sheet with Camera and Gallery. `RequestAttach` asks the activity, `SetPhoto` returns bytes, and the tick decodes, scales to 240 px, and rounds the corners before appending a photo bubble.
- Long-pressing a message opens a six-emoji reaction row; the picked emoji shows as a badge. The gift button appears in Anna's chat only.
- Android back calls `mobile.Back`. It leaves a chat from the thread and closes the app from a list, with the IME-visible case deferred to the system.

### Android build path

- `scripts/android.sh` binds `examples/telegram/mobile` with `ebitenmobile`, builds `:app:assembleDebug`, and with `install` runs `adb install -r`. It requires `ANDROID_HOME` (or `$HOME/Android/Sdk`) and warns that `GOWORK` must not be exported, because gomobile builds in a temporary module that cannot join the workspace.
- Commit `3a3f085` took the debug APK from 188 MB to 29 MB: `ANDROID_TARGET=android/arm64` binds one ABI, `-s -w` drops Go symbols, and `max-page-size`/`common-page-size` 16384 keep 16 KB page devices happy. A Pixel 7 showed the system compatibility dialog without the page-size flags. AGP moved to 8.5.1 on Gradle 8.7 because older AGP does not zip-align uncompressed shared libraries to 16 KB, verified with `zipalign -c -P 16` and `llvm-objdump`.
- `examples/telegram/android/README.md` covers Java 17, the SDK and NDK packages, USB debugging, and the WSL2 routes (wireless debugging or usbipd-win).

### Dino example

- `BindTouch` installs `Handlers{KeyDown, KeyUp, Click, Swipe}` and keeps the keyboard working. A tap jumps and holds until the rise ends, so each tap reaches the high jump without bouncing. Swipe up jumps; swipe down ducks for 600 ms; sideways swipes are ignored.
- `fitView` scales the 900x300 scene by `min(pageW/900, pageH/300, 1)` and drops it to the bottom, so a taller phone screen becomes sky above the ground. It never upscales.
- The ground is now a display operation bound by id, and the CSS switches the scene to `100%`, which puts the FPS and score in the page's top-right corner.
- `examples/dino/mobile` calls `ownframe.BindMobile`, and `examples/dino/android` wraps the AAR in a landscape-locked Gradle app (`com.chinmaysawant.dino`). `scripts/android-dino.sh` binds, builds, and installs; from WSL it can name the Windows adb and converts the APK path with `wslpath -w`.

### Docs, showcase, and skill

- `README.md`, `AGENTS.md`, `documentation/editing.md`, `features.md`, `interaction.md`, `platforms.md`, `pointer.md`, and `screen.md` describe the gestures, the thresholds, the IME session, the insets, and the engine's position limits with the Telegram pin as the worked example.
- `showcase.md` renders the seven 540 px webp captures from a Pixel 7: chats, thread, contacts, settings, and the chat list and thread in dark. The settings capture predates the IME commit and still says typing needs a hardware keyboard, while the page at HEAD says the soft keyboard is wired. A recapture is a follow-up.
- `.agents/skills/explain-output/SKILL.md` is a new skill for rendering explanations, with rungs for prose, diagrams, an HTML page, and narrated video.
- `.gitattributes` marks `*.webp` as binary and linguist-vendored so the screenshots stay out of diffs and language stats. `go.work.sum` is new; it is the checksum file Go keeps for the committed two-module workspace, and no `go.mod` changed.

### Tests

- `examples/telegram/telegram`: 23 files, 37 tests. List, tabs, search, contacts, send, backspace with `IMEReplace`, attach, photo, reactions, gift, back, insets, and the pinned-bar geometry, pixels, cover behavior, inset strips, and replayability.
- `internal/page`: 4 files, 11 tests. IME context and replacement, long press word selection and handler routing, swipe deltas.
- `internal/window`: 7 files, 11 tests. Hold claim, slop, swipe thresholds, tap disarm, and finger start state.
- `examples/dino/dino`: 3 files, 9 tests. Tap and swipe behavior, duck timing, the fitted view, and touch hints.

---

## Impact

| Area | Impact |
|---|---|
| **Performance** | A phone drag with the bars pinned ran at 55-85 fps on a Pixel 7; before the pin path every scroll frame redrew the page at 70-200 ms. Desktop pointer and wheel paths are unchanged. |
| **Memory** | No new retained buffers. The content-sized replay buffer is unchanged, and the pinned layer is a draw-time filter. |
| **Behavior / correctness** | Desktop and wasm keep key input and the old scroll behavior. The optional interfaces are consumed by type assertion, so a screen without them is untouched. The `tapAt` fix removes a delayed long press after a tap on Android. |
| **API / CLI** | Four new optional `host` interfaces and five new `Page` methods, all additive. `Handlers` gains `LongPress` and `Swipe`. Two `BindMobile` examples. No root alias changes. |
| **Dependencies** | None. `ebitenmobile` matches the `go.mod` Ebiten v2.10.4. The examples move to AGP 8.5.1 and Gradle 8.7. |
| **Binary size / build time** | Telegram arm64 debug APK 188 MB → 29 MB. Two Gradle wrappers and two Android projects are added to the examples module. |

---

## Breaking changes / migration

| Item | Migration |
|---|---|
| None | The new interfaces and methods are optional and additive; existing screens compile and behave as before. |

---

## Test plan

The branch commits record `make build` and `make test` green at `e4d86d9` and `make test` green at `ae2fd3b` and `727b7c1`. Device checks ran on a Pixel 7 with the rebuilt APK.

- [x] `make build` (`go vet -p 1 ./... ./examples/...`)
- [x] `make test` (37 new test files, 68 test functions)
- [x] Pixel 7: 200 ms taps do nothing, a 700 ms hold opens the bar under the right message, a tap closes it, and a picked emoji lands on the message (`ae2fd3b`)
- [x] Pixel 7: drag at 55-85 fps with bars pinned, send keeps the keyboard up, back leaves no white band (`727b7c1`)
- [ ] CI is not configured in this repo.

### Commands

```sh
make build
make test
sh scripts/android.sh          # bind, build, and print the APK path
sh scripts/android.sh install  # the above plus adb install -r
sh scripts/android-dino.sh install
```

---

## Screenshots / sample output

- `showcase.md` renders the Pixel 7 captures under `assets/screenshots/android/telegram/`, including the chat list, thread, contacts, settings, and dark-theme captures.
- Device numbers from the commit messages: Telegram arm64 debug APK 29 MB (188 MB with all four ABIs and unstripped symbols), phone drag 55-85 fps.

---

## Related issues

None. No ticket exists for this work.

---

## PR metadata checklist (author)

- [ ] Self-assigned with `--assignee "@me"`.
- [ ] Labels applied (`enhancement`, `documentation`).
- [ ] Body kept at `plans/v0.0.2/pr/pr-telegram-dino.md`.
- [ ] Related issues: none exist for this work.

Suggested open command (not run):

```sh
gh pr create \
  --base master \
  --head store/example-telegram \
  --title "feat(examples): add the Telegram and dino phone demos with touch, IME, and pinned bars" \
  --body-file plans/v0.0.2/pr/pr-telegram-dino.md \
  --assignee "@me" \
  --label enhancement \
  --label documentation
```

---

## Follow-ups (out of scope)

- Desktop IME stays unwired; only Android and iOS open a session. The route is recorded in `documentation/interaction.md`.
- `internal/window` IME files are build-tagged and have no unit tests; device checks cover them. The typing and backspace pass noted in `e4d86d9` was still pending on the device at that commit.
- `ViewportPinner` and `DrawVisiblePinned` have no direct unit tests; the Telegram example tests cover the pinned geometry and pixels.
- Z-pin applies in the direct replay path (ticking or oversized pages) and to operations with a z-index. The buffered path only special-cases engine `position: fixed`.
- Pinned bars' hit boxes follow the 120 ms settle redraw, so a tap during an active drag can miss until it settles.
- A claimed drag that moved skips `Release`; `dragActive` stays set until the next press. No test covers that lift.
- Recapture `assets/screenshots/android/telegram/telegram-settings.webp`; it predates the IME commit and shows the old hardware-keyboard note.
- Dino: one swipe per lift, a fixed 600 ms duck with no hold-to-crouch, no upscaling above 1x, and an arm64-only default bind.
- Telegram: state is in memory, a reaction is one emoji that overwrites, search excludes message text, and the photo queue drops shots beyond two.

---

## Reviewer checklist

- [ ] Behavior matches the summary and the test plan.
- [ ] No unrelated changes: the branch is the two phone demos plus the library pieces they call.
- [ ] Public API additions documented (`documentation/interaction.md`, `platforms.md`, `pointer.md`, `screen.md`, `AGENTS.md`).
- [ ] PR has assignee and labels.
- [ ] No secrets or generated artifacts committed. The AAR, `build/`, and `.gradle/` are gitignored; the Gradle wrapper jars are committed on purpose.
- [ ] Diff-stat table matches `bash scripts/pr-diff-stat.sh master`.

---

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh master` on 2026-10-06.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.bat` | 2 | 184 | 0 |
| `.css` | 6 | 468 | 4 |
| `.go` | 128 | 4713 | 137 |
| `.gradle` | 8 | 94 | 0 |
| `.html` | 6 | 129 | 1 |
| `.jar` | 2 | Binary | Binary |
| `.java` | 2 | 280 | 0 |
| `.md` | 13 | 431 | 17 |
| `.properties` | 4 | 18 | 0 |
| `.sh` | 2 | 142 | 0 |
| `.sum` | 1 | 60 | 0 |
| `.svg` | 11 | 11 | 0 |
| `.webp` | 7 | Binary | Binary |
| `.xml` | 2 | 51 | 0 |
| No extension | 5 | 511 | 0 |
| **Total** | **199** | **7092** | **159** |

---

## Commit list

Counts come from `git show <sha> --shortstat` on 2026-10-06. Commit sums overlap files, so the net row is the range diff.

| Commit | Subject | Files | Insertions | Deletions |
|---|---|---|---:|---:|
| `44027fb` | feat(examples/telegram): add the Telegram-like phone demo and its Android project | 64 | +2240 | -2 |
| `3a3f085` | fix(examples/telegram): 16 KB-align the Android build and shrink the APK | 4 | +63 | -6 |
| `b831b89` | docs: add Telegram screenshots from a Pixel 7 and a showcase page | 10 | +27 | -0 |
| `8d82820` | docs(skills): add the explain-output skill | 1 | +147 | -0 |
| `e4d86d9` | feat(telegram): soft keyboard, long-press gestures, pinned chat bars, phone features | 87 | +2957 | -149 |
| `ae2fd3b` | fix(window): disarm the held-press watch on a touch tap | 3 | +58 | -21 |
| `727b7c1` | feat(telegram): pin the phone bars in the replay, keep the keyboard up | 31 | +493 | -41 |
| `39befdf` | feat(dino): swipe controls, a phone layout, and an Android build | 43 | +1233 | -66 |
| **Sum (commits overlap)** |  | **243** | **+7218** | **-285** |
| **Net diff** |  | **199** | **+7092** | **-159** |
