# feat(platform): Windows UI Automation, Windows fit, IME fix, Wayland clipboard

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `feature/native-platform-uia-ime-wayland-clipboard`

---

## Summary

A Windows UI Automation provider so NVDA and Narrator announce ownframe controls, a Windows window-fit so a window opened on one monitor stays on screen on another, a fix for Windows IME composition being committed early, and a pure-Go Wayland data-control clipboard. These are the native-platform roadmap items for Windows accessibility, Windows IME, and a native Wayland clipboard.

---

## Motivation / context

- Roadmap: Windows UI Automation (UIA), real-host Windows IME validation, and a native Wayland clipboard that does not shell out to `wl-clipboard` (`plans/README.md`, `documentation/accessibility-native.md`).
- A window opened on a 150 percent monitor is larger in pixels than the same window at 100 percent, so it hung off a smaller monitor.
- Issues: none.
- Depends on blinkless [#47](https://github.com/chinmay-sawant/blinkless/pull/47) for CJK glyphs in the IME field. It is a separate repository and this PR does not require it; the blinkless pin is unchanged here.

---

## Changes

### Windows UI Automation (`internal/uia`, `internal/window/a11y_uia_windows.go`)

- A pure-Go provider with no cgo for Windows amd64 and arm64. It maps the accessibility snapshot to UIA control types, names, state, bounds, and the Invoke, Value, Toggle, and SelectionItem patterns. Password values are never exposed, and actions go through the same queue as the Linux bridge.
- **Events were never delivered, and now are.** Every raise call returned `S_OK` but no client received anything. Two causes: `get_ProviderOptions` returned `1` (client-side) where a provider hosted in its window must return `2` (server-side), and the window answered `WM_GETOBJECT` only for the root object ID, although UIA sends IDs from zero up when it delivers an event. `OBJID_CLIENT` also gets the MSAA view of the tree, and a focus move sends an `EVENT_OBJECT_FOCUS` WinEvent. Regression tests pin both details.
- **Hit testing.** `ElementProviderFromPoint` takes two doubles, which a Go syscall callback cannot read. One assembly thunk per architecture (`hit_windows_amd64.s`, `hit_windows_arm64.s`) moves their bits into integer registers. `Provider.ElementAt` picks the deepest element, and the later sibling where two overlap.
- `live_events_windows_test.go` hosts the provider in a plain Win32 window and keeps changing focus and names, so a UIA client can check event delivery without Ebiten (`OWNFRAME_UIA_LIVE=1`).

### Windows window fit (`internal/window/window_fit*.go`)

- Once at startup, and once after the window settles on a different monitor, a window is shrunk and moved just enough to lie inside that monitor's work area. It is never enlarged, and one that already fits is left alone. It waits while the window is maximized, minimized, full screen, or dragged.
- The move uses `SWP_ASYNCWINDOWPOS`. A synchronous move from inside `Update` crashed the window.

### Windows IME (`internal/window/ime_*.go`)

- Windows reports the clause cursor at the end of a fresh preedit, so the glyph anchor moved with every key, and the window read that as a caret jump and committed the half-typed composition. A live composition is no longer restarted. Headless tests cover preedit, stale focus, restart, anchor, cancel, and commit.

### Wayland clipboard (`internal/clipboard`)

- A pure-Go Wayland client that binds `ext_data_control_manager_v1` or `zwlr_data_control_manager_v1`, tried before `wl-copy`/`wl-paste` and X11, which remain the fallbacks. Tested against a fake compositor over a unix socket.
- `make lint` found 26 issues in this code and `-race` found a data race in a test (a flag read from another goroutine). The flag is now atomic; eight consecutive `-race` runs of the package are clean.

### CI

- A `windows-arm64` job on the native `windows-11-arm` runner builds, vets, and tests `internal/uia` and `internal/window`, so the arm64 assembly thunk runs on real hardware. It runs on master or with `[ci only=windows-arm64]`.

### Docs

- `documentation/accessibility-native.md` records the validation and the limits. `documentation/platforms.md` and `AGENTS.md` are updated.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | None for pages. The UIA provider does nothing until a client listens (`UiaClientsAreListening`). The fit reads two Win32 calls per update until it settles. |
| **Memory** | Negligible. |
| **Behavior / correctness** | NVDA and Narrator announce controls on Windows. A Windows window that would hang off its monitor now fits. Windows IME composition is no longer committed early. |
| **API / CLI** | None. No public API changed. |
| **Dependencies** | None added. |
| **Binary size / build time** | Windows binaries carry the provider and two small assembly files. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [x] `make test` (152 packages)
- [x] `make lint`: the Go gate, docs, plans, file-size, examples, and import checks pass. The website ESLint step could not run locally because `eslint` is not installed in this worktree; nothing under `frontend/` changed.
- [x] `make graphics-check`
- [x] `go test -race ./internal/clipboard` repeated, and `GOOS=windows` `go vet` for amd64 and arm64
- [ ] `windows-arm64` CI: runs on master or on request. The arm64 thunk is built and vetted here but has not run on arm64 hardware.

### Validated on a Windows 11 x64 machine with a 150 percent and a 100 percent monitor

- **NVDA 2026.3:** announced "Display name, edit, Ada", "Private value, edit, protected" (the password text is not read), "Receive updates, check box, checked", and "Email, radio button, checked" as Tab moved focus, confirmed by ear.
- **Narrator:** read the controls, confirmed by ear.
- **Hit testing:** `AutomationElement.FromPoint` at the center of every reported rectangle returned that element or a descendant, 44 of 44.
- **DPI:** at 150 percent the checkbox rectangle is 19 pixels (the 13 pixel box scaled by 1.5), and screen pixels inside it contain the checked box's accent colour; a rectangle shifted 150 pixels contains none.
- **Window fit:** a window dropped with its bottom 24 pixels past the work area moved to the edge, and a 1470 by 1134 window was shrunk to 1470 by 1032 at the top of the work area, with no crash.
- **Windows IME:** Microsoft Japanese IME composed, showed candidates, and committed.

**Not validated:** JAWS (needs a license), Accessibility Insights, arm64 execution, Windows versions before 11, and the Wayland clipboard on a compositor that offers data-control (WSLg's compositor does not; the fallback path runs there). IME cancel, candidate selection, and the candidate popup following the caret are not checked.

### Commands

```sh
make test
make lint
make graphics-check
GOOS=windows GOARCH=arm64 go vet ./internal/uia ./internal/window
```

---

## Screenshots / sample output

```
no screenshots; NVDA's speech log and the UIA probes above are the evidence
```

---

## Related issues

- None.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled
- [x] Filled body committed under `plans/PR/pr-native-platform-windows-uia-ime-wayland.md`

---

## Follow-ups (out of scope)

- Bump the blinkless pin in `go.mod` and `examples/go.mod` once blinkless #47 is merged and tagged.
- Windows `386` and `arm` have no UIA provider.
- macOS NSAccessibility, Android TalkBack, and macOS IME validation are not started.
- Native Wayland clipboard on GNOME Mutter and WSLg's Weston needs a focused-surface `wl_data_device` path from the window layer.

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 32 | 4419 | 14 |
| `.md` | 12 | 190 | 36 |
| `.s` | 2 | 45 | 0 |
| `.yml` | 1 | 30 | 1 |
| **Total** | **47** | **4684** | **51** |
