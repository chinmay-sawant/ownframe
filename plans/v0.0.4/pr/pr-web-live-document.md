# feat(web): serve the live page and its media in the browser

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `origin/master` (`085ba7ea`)
- Head: `feat/web-live-document`
- Diff against `origin/master`: `5dee69fd`, `dfc55cf8`, and this plan file

---

## Summary

`-web` now serves the application HTML as the browser document, with real inputs, instead of a picture inside a frame. Any picture registered with `SetImage` loads by its source name. `SetAudio` and `SetAudioPlayback` publish the clip, whether it is playing, and the 0 to 1 level, and the browser player follows pause and volume. A click updates the existing tree in place.

---

## Motivation / context

- Plans: `plans/v0.0.4/pr/pr-web-live-document.md`
- The PNG shell could not type, and a later iframe kept paint spans that were not real inputs. Registered images never reached the browser, because `url("logo")` asked for `/logo` and the server had no such route. Playback ignored pause and ran at full volume because the tab owned a separate player from the page.
- Issues: none. This work has no tracker issue.

---

## Changes

### Live document

- `GET /` returns the executed page when it has a source. Text fields, passwords, textareas, and selects are real controls. `?dom=0` and a page with no source keep the PNG shell.
- `POST /change` replaces a whole field value through `BeforeEdit`, bind, and `Change`.
- The boot script morphs the current document against the new one. An id matches that id in the sibling list. Everything else matches by position. New nodes are inserted at that index. The focused field keeps its value and caret.

### Images

- `GET /{name}` returns the bytes any page stored with `SetImage`, including a slash in the name. PNG, JPEG, GIF, WEBP, BMP, icon, SVG, AVIF, and HEIC get an image type. An unknown format is still returned. `.` and `..` are rejected.

### Audio and playback

- `Serve` and `ServeWithOptions` create the same audio context as `Run`, so a page that plays through the host is heard on this machine.
- `Page.SetAudio` stores one encoded clip. The generation moves only when the bytes change. `Page.SetAudioPlayback` stores playing and a volume from 0 to 1.
- `GET /audio` serves that clip. `GET /state` adds `audioGen`, `audioPlaying`, and `audioVolume`. The boot script sets the element volume and calls play or pause from that snapshot.
- The Spotify player publishes its current clip and its percent volume through those methods. A bar at 10% is `0.1` in the browser.
- A nil sound no longer panics when no audio context is running.

### Docs

- `documentation/web.md`, `features.md`, `platforms.md`, and `window.md` describe the document, the image route, and browser playback.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | A state poll reads the clip generation without copying the bytes. Publishing the same clip again does not restart playback. |
| **Memory** | The page keeps one copy of the current clip. `GET /audio` copies it out under the server lock. |
| **Behavior / correctness** | `-web` shows the page HTML, loads registered images, updates clicked regions in place, and follows play, pause, and volume in the browser. |
| **API / CLI** | New `Page` methods: `SetAudio`, `Audio`, `SetAudioPlayback`, `AudioPlayback`. `-web` and `-addr` are unchanged. |
| **Dependencies** | None. |
| **Binary size / build time** | No new module requirements. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | Existing `Run` and `Serve` callers keep working. Apps that want browser audio call `SetAudio` and `SetAudioPlayback`. |

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

`SetAudio` and `SetAudioPlayback` are real page methods, the same shape as `SetImage`. The web server and the app both call them.

## Test plan

- [x] `go test ./internal/web/` for the document, image route, audio route, and `/state`
- [x] `go test -run TestAudio ./internal/page/`
- [x] `go test ./examples/spotify-player/player/`
- [x] `go test ./examples/music/` for a missing audio context
- [ ] `make test`
- [ ] `make lint`
- [ ] `make test-race`
- [ ] `make build`

### Commands

```sh
go test -count=1 ./internal/web/ ./internal/page/ ./examples/music/ ./examples/spotify-player/player/
go run ./examples/login -web
go run ./examples/spotify-player -web
```

---

## Screenshots / sample output

Spotify `-web` on `127.0.0.1:8121` returned all 46 `url(...)` images with `200`. Covers were `image/jpeg`. Icons were `image/png`. The home view showed sidebar art, nav icons, and album covers.

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

- The seek fill and time labels are still painted on the display list. The browser document shows them when the page redraws.
- `landing/index.html` is still untracked and is not in this branch.
- `make test` and `make lint` were not run for this branch.

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
| `.go` | 38 | 4000 | 29 |
| `.md` | 5 | 211 | 34 |
| **Total** | **43** | **4211** | **63** |
