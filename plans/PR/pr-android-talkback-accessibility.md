# feat(android): add TalkBack accessibility bridge and wire login app

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `feature/android-talkback-accessibility`

---

## Summary

Adds an Android TalkBack accessibility bridge (`androidaccess`, `internal/a11ybridge`, `internal/window/a11y_android.go`, and `dev.ownframe.android.AccessibilityBridge` in `examples/android-host`) and wires it in the Android Login app (`examples/login/mobile`, `examples/login/android`). The Android window loop publishes the page's accessibility snapshot as JSON and drains queued focus, activation, and text actions on the owning page loop, while the Java `AccessibilityNodeProvider` exposes virtual nodes to TalkBack.

---

## Motivation / context

- An `ownframe` Android app draws into a single GPU surface (`EbitenView`), so TalkBack sees no controls without a virtual accessibility tree.
- `ebitenmobile bind` cannot export Go structs across the language boundary, so the bridge crosses the boundary via two string/primitive entry points in `androidaccess` (`Tree` and `Perform`).
- Issues: none.

---

## Changes

### Go accessibility hand-off (`internal/a11ybridge`, `androidaccess`, `internal/window/a11y_android.go`)

- `internal/a11ybridge` encodes `host.AccessibilityTree` snapshots to JSON (omitting password values) and forwards `Perform(document, node, action, value)` requests with a 3-second timeout to the page's action queue.
- `androidaccess` exposes `Tree() string` and `Perform(document int64, node string, action int, value string) string` for `ebitenmobile`-bound packages.
- `internal/window/a11y_android.go` attaches the action queue (`atspi.NewActions`), drains actions on each frame in `runA11yActions`, and publishes updated snapshots in `syncA11y` whenever the page generation advances.

### Shared Android host (`examples/android-host`)

- `dev.ownframe.android.AccessibilityBridge` attaches an `AccessibilityNodeProvider` virtual tree to an `EbitenView`, converts page CSS-pixel bounds to view/screen pixels using display density, emits `TYPE_WINDOW_CONTENT_CHANGED` and `TYPE_VIEW_FOCUSED` events on tree updates, and dispatches `ACTION_FOCUS`, `ACTION_CLICK`, and `ACTION_SET_TEXT` on a background worker thread.
- Nodes without geometry of their own inherit ancestor bounds so Android does not prune their subtree, while remaining out of swipe navigation unless focusable.

### Login Android app (`examples/login/mobile`, `examples/login/android`)

- `examples/login/mobile/access.go` exports `Accessibility` and `AccessibilityPerform`.
- `examples/login/android` attaches `AccessibilityBridge` to the `EbitenView` in `MainActivity` and adds `.debug` `applicationIdSuffix` on debug builds so debug APKs install alongside signed release builds.

---

## Impact

| Area | Impact |
| --- | --- |
| **Performance** | Snapshot JSON encoding runs only when the page generation advances on Android; Java polling runs at 250 ms only while `AccessibilityManager.isEnabled()` is true |
| **Memory** | One atomic string pointer and one performer function pointer in `internal/a11ybridge` |
| **Behavior / correctness** | TalkBack detects and announces page controls, focus, and actions on Android; password values are never serialized or editable via `ACTION_SET_TEXT` |
| **API / CLI** | New public package `github.com/chinmay-sawant/ownframe/androidaccess` (`Tree`, `Perform`, action constants) for `ebitenmobile` bindings |
| **Dependencies** | None |
| **Binary size / build time** | Negligible outside Android builds |

---

## Breaking changes / migration

| Item | Migration |
| --- | --- |
| None | - |

---

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`make internal-import-check`)
- [x] No synonym names (`PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [x] Library code panics only for invariants; new user-facing failure paths return `error`
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's ctx
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr

---

## Test plan

- [x] `make test` (`./androidaccess`, `./internal/a11ybridge`, `./internal/window`, `./examples/login/...`)
- [x] `make lint`
- [x] `make test-race` (`go test -race ./androidaccess ./internal/a11ybridge ./internal/window`)
- [x] `make graphics-check`
- [x] Android device validation: built `examples/login/android` debug APK via `ebitenmobile bind` + `./gradlew :app:assembleDebug` and verified TalkBack detection on an Android device.

### Commands

```sh
make lint
make graphics-check
go test -race ./androidaccess ./internal/a11ybridge ./internal/window
go test ./examples/login/...
```

---

## Screenshots / sample output

```
ok  	github.com/chinmay-sawant/ownframe/androidaccess	1.023s
ok  	github.com/chinmay-sawant/ownframe/internal/a11ybridge	1.064s
ok  	github.com/chinmay-sawant/ownframe/internal/window	1.810s
ok  	github.com/chinmay-sawant/ownframe/examples/login/login	0.081s
```

---

## Related issues

- None.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled
- [x] Filled body committed under `plans/PR/pr-android-talkback-accessibility.md`

---

## Follow-ups (out of scope)

- Wire `AccessibilityBridge` and `androidaccess` wrappers into the remaining Android example apps (`telegram`, `platform`, `dino`, `eyes`, `depth-alert`, `digital-wellbeing`, `pet-care`).

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 7 | 617 | 3 |
| `.gradle` | 1 | 7 | 0 |
| `.java` | 2 | 434 | 0 |
| `.md` | 4 | 45 | 16 |
| `.yml` | 1 | 4 | 2 |
| **Total** | **15** | **1107** | **21** |
