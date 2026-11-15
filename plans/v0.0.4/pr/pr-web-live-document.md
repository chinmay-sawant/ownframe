# feat(web): serve the live page and its media in the browser

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `feat/web-live-document` (12 commits, head `9953557e`)
- Release cycle: v0.0.4

---

## Summary

`-web` serves the application HTML as the browser document, with real inputs, instead of a picture inside a frame. The live document now follows the server: ticks patch it in place, wheel and caret input reach the page, registered images and audio play in the tab, and several previews that did not respond (Forms, Dino, Flappy Bird, Loglens, Spotify, Blackhole) work again.

---

## Motivation / context

- Plans: `plans/v0.0.4/pr/pr-web-live-document.md`
- The PNG shell could not type, and registered images and audio never reached the browser. After the live document landed, a browser sweep of the example previews found input, scroll, and animation failures. The causes sat in the shared web layer and in a few examples.
- Issues: none. This work has no tracker issue.

---

## Changes

### Live document (`internal/web`)

- `GET /` returns the executed page with real text fields, passwords, textareas, selects, checkboxes, and radios. `?dom=0` and a page with no source keep the PNG shell.
- `POST /change` replaces a field value or moves its caret. The browser sends its own caret after a click, so a click inside the text puts the caret there instead of at the end.
- Snapshots arrive over server-sent events. `domGeneration` is separate from the frame generation, so an in-place tick does not refetch the whole document.
- Tick patches update leaf text, inline geometry, and fill widths by element id, so a game scene and a seek bar move without a full morph. Hidden text can be cleared again.
- Checkbox and radio state is written back from the server, so a selection no longer snaps back.
- The browser scrolls the document itself when it can. The wheel handler only forwards to the server when the document cannot scroll, which fixes a page that did not scroll.
- A click resolved by element id or action ignores the server's frame scroll offset. Before, a click that focused a control left an offset that made later clicks miss, which broke Loglens calendar next and previous.

### Images and audio

- `GET /{name}` returns bytes stored with `SetImage`. `GET /audio` serves the clip from `SetAudio`, and `/state` reports `audioGen`, `audioPlaying`, and `audioVolume`.
- `Serve` creates the audio context `Run` uses.

### Examples

- **Spotify:** clicking Your Library or a playlist no longer restarts the track. The player samples the audio clock before it commits a navigation click.
- **Dino:** the HI score follows the active run and the scene moves in the live document.
- **Flappy Bird:** a wasm canvas preview, the hidden game-over label, and cloud and ground motion in `-web`.
- **Loglens:** calendar navigation and page scroll in `-web`, with a regression test for month changes.
- **Blackhole:** the scene frame swap goes through `SetData` again. A `PatchElement` on the inline `<img>` was refused by the layout engine on every push, so new frames never reached the screen and each refusal dropped the layout cache. The HUD values are template data, so a re-render keeps them. The scene push stays at 100 ms because a faster pace gave blank GPU readbacks.
- **Telegram and Teams:** a `-web` thread view with fixed chrome, and long-word wrapping with composer scrollbars.

### Docs

- `documentation/web.md`, `features.md`, `platforms.md`, and `window.md` describe the document, the image route, and browser playback.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Ticks send small id-anchored patches instead of refetching the document. Blackhole no longer reparses the page on every scene push. |
| **Memory** | The page keeps one copy of the current clip. |
| **Behavior / correctness** | `-web` previews accept typing, caret placement, radios, wheel scroll, calendar clicks, and animated scenes. Spotify navigation keeps playback position. |
| **API / CLI** | `Page.SetAudio`, `Audio`, `SetAudioPlayback`, `AudioPlayback`. `-web` and `-addr` are unchanged. |
| **Dependencies** | None. |
| **Binary size / build time** | No new module requirements. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing `Run` and `Serve` callers keep working. |

---

## Library freeze (until Page wrap)

Until `plans/v0.0.3/prod-ready-task/` wraps `Page`, reject the PR if it does any of the following.

- [x] No new `type X =` at the module root except documented re-exports
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`make internal-import-check`)
- [x] No synonym names (`PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR
- [x] Library code panics only for invariants; new user-facing failure paths return `error`
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's ctx
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr

`SetAudio` and `SetAudioPlayback` are page methods of the same shape as `SetImage`.

## Test plan

- [x] `go test ./internal/web ./internal/page`
- [x] `go test ./examples/blackhole ./examples/dino/... ./examples/spotify-player/... ./examples/builder/loglens`
- [x] Browser checks in a muted headless Chrome: Loglens wheel scroll and calendar next/previous, Blackhole HUD buttons, Dino and Flappy Bird motion, Forms radios, Spotify playlist click
- [x] Blackhole desktop window launched for 15 s: zero refused patches, scene frames reach the display
- [ ] `make test`
- [ ] `make lint`
- [ ] `make test-race`
- [ ] `make build`

### Commands

```sh
go test -count=1 ./internal/web ./internal/page ./examples/blackhole ./examples/builder/loglens
go run ./examples/builder/loglens -web
go run ./examples/blackhole
```

---

## Screenshots / sample output

Loglens On-call calendar in the browser: `November 2026` → `December 2026` → `November 2026`.

---

## Related issues

- None. This work has no tracker issue.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues section filled. There is no tracker issue for this work.
- [x] Filled body committed under `plans/v0.0.4/pr/pr-web-live-document.md`

---

## Follow-ups (out of scope)

- Blackhole still reads each frame back from the GPU, encodes a JPEG, and swaps the image. Blank readbacks still happen, and a smooth version needs the shader output drawn straight to the window.
- Blackhole `-web` shows a still frame, so drag and spin work only in the desktop and wasm builds.
- Fetch `-web` was not verified because its remote image service timed out.
- `make test`, `make lint`, `make test-race`, and `make build` were not run for this branch.

---

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

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 3 | 59 | 5 |
| `.go` | 116 | 7294 | 219 |
| `.html` | 7 | 54 | 9 |
| `.md` | 7 | 272 | 36 |
| **Total** | **133** | **7679** | **269** |
