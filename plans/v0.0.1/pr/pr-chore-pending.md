## Summary

Complete the v0.0.1 workstream: finish the form and replay fidelity work, and add the pieces around it: desktop file dialogs, live themes, key events, pointer states, scrollbars, a page-background underlay, replay scaling, frame helpers for `SetTick` animation, and a suite of 22 new example folders with headless tests. The window still replays the layout as vector operations or draws the engine bitmap with a `bitmap fallback` badge; no PDF is written anywhere.

## Motivation / context

- Plans: [`plans/v0.0.1/compare.md`](plans/v0.0.1/compare.md), the Electron gap list that defines v0.0.1. [`PHASES.md`](../../PHASES.md) records the integration state and the known engine and replay limits.
- Branch `chore/pending`: 37 commits on top of `master` (`ac2ddc4`); the `feature/v0.0.1` stack and its follow-ups are merged in. This PR carries the remaining integration: 545 files, +20,764 / -380.
- No issue is linked; the repo has no issues, same as the earlier v0.0.1 PR.

## Changes

### Rendering: replay coverage and state

- `internal/render`: the replay gate now accepts CSS outlines, letter-spaced text, transformed images, and masked or elliptical strokes (`StrokeMask` within `0x0F`). New `State{Focus, Hover, Active, Theme, Images}` with `DisplayListState` and `PaintState` entry points; `RadiiXY` resolves elliptical corner radii.
- `internal/replay`: new `stroke_path.go` traces per-corner elliptical arcs as 8-segment polylines and masked strokes side by side; image geometry applies the op rect, CSS transform, scroll, and `pxPerPt`; letter-spaced text is measured and drawn per rune.
- `internal/frame` (new): `Fill`, `Fills`, `Text`, and `BoxUnits` find fill and text operations in the retained `layout.Display`, so a `SetTick` callback can animate without another `Redraw`. Returned pointers invalidate after the next `Redraw`.
- An unsupported operation still keeps the whole page on the bitmap path with the `bitmap fallback` badge, so one frame never mixes vector and bitmap.

### Window and input

- New `scrollbar.go` and `scrollbar_drag.go`: thumbs appear when content overflows; a drag consumes the pointer so content clicks behind the thumb are skipped.
- New `background.go`: the window fills the screen with the page's `html`/`body` fill colour (or the bitmap corner), default white, including under scaled replay.
- New `replay_scale.go`: a stretched replay renders through an offscreen buffer and scales to the window; wheel scrolling is disabled while stretched.
- Pointer: `pointer.go` sends `Hover` every frame and `Press` plus `Release` on tracked edges; `pointer_touch.go` turns fresh touch IDs into taps without double-firing a mouse click.
- Keys: `key_events.go` forwards `KeyDown` and `KeyUp` with lowercase names from `keyname.go`; `key_watch.go` drops OS auto-repeats; `chord_watch.go` replaces the fired-chord logic and `chord_swallow.go` stops a held chord key from typing its letter.

### Page, forms, and file dialogs

- `internal/filepick` (new): a desktop dialog per OS: Linux `zenity`/`qarma`/`matedialog`/`kdialog`; Windows `comdlg32!GetOpenFileNameW`; macOS `osascript`; WSL through `powershell.exe`/`pwsh.exe` with `wslpath` or UNC conversion. `Run` installs it, while wasm, mobile, and `Serve` keep the typed-name fallback. `OWNFRAME_FILEPICK_DEBUG=1` prints fallback reasons.
- `internal/page`: a file input opens the picker and fires `BeforeEdit` and then `Change`; `BeforeEdit` can veto an edit; `Cut` flows through binding and is undoable; long values wrap with the caret kept on the line; the button CSS workaround is gone now that the engine styles buttons.
- New page API: `Config.Theme` and `SetTheme`, `SetTick` and `Tick`, `SetImage`, `Hover` / `Press` / `Release`, `KeyDown` / `KeyUp`. `Handlers` gains `BeforeEdit`, `KeyDown`, and `KeyUp`.
- `host.Screen` gains `Hover`, `Press`, `Release`, `KeyDown`, and `KeyUp`; the new optional `host.Ticker` calls `Tick` once per frame.
- `Run` and `BindMobile` create the Ebiten audio context (`audio_context.go`); `Serve` deliberately has none.

### Examples

- New audio stack: `examples/music` (Openverse resolver, SHA-1 disk cache, MP3/WAV decode, Ebiten playback) and the `examples/spotify-player` now bar (seek, volume, and EQ animated from the audio position through `SetTick` and `internal/frame`).
- New apps: `examples/spotify-player`, `examples/dino` (runner game with `Handlers.KeyDown`/`KeyUp` and display mutation), and `examples/wispr-flow-dashboard`.
- New feature proofs: `controls`, `editing`, `clipboard`, `states`, `bind-hooks`, `theme`, `fetch`, `ipc`, `history`, `crash`, `platform`, and the smaller `layout`, `png`, `replay`, `scrolling`, `shapes`, `web` demos.
- `examples/login` updated for the disabled button, `BeforeEdit` undo, and Cut; `examples/readme.md` indexes every folder and its `-web` port.

### Docs and build

- New `documentation/theming.md`, `frames.md`, `keys.md`, and `features-examples.md`; updates to `forms.md`, `fetch.md`, `ipc.md`, `screen.md`, `binding.md`, `features.md`, and `README.md`.
- `Makefile` (new): `make test TEST_P=...` and `make open [example] [ARGS=-web]`.
- `PHASES.md` status refresh; `.gitignore` covers the example binaries and `temp/`.
- `go.mod` / `go.sum`: three indirect modules for audio (`ebitengine/oto/v3`, `hajimehoshi/go-mp3`, `jfreymuth/pulse`); direct dependencies unchanged; the local `replace ../gowkhtmltopdf` stays until the engine branch is pushed.

## Impact

| Area | Effect |
|---|---|
| **Performance** | Replayable pages skip whole-page rasterization per redraw; scrollbar, background, and scaling work is per frame; letter-spaced text costs one measure and draw per rune. |
| **Memory** | A replayable page holds no RGBA page buffer; the window disposes textures it replaces. `internal/frame` returns pointers into the retained display, valid only until the next `Redraw`. |
| **Behavior / correctness** | Fewer pages fall back to the bitmap badge. Key handling now distinguishes down/up and drops repeats; a held chord key no longer types. A desktop file input opens a real dialog. |
| **API / CLI** | New `Page` methods, `Config.Theme`, and `Handlers` fields; `host.Screen` additions. The examples add no CLI beyond `-web` and `-addr`. |
| **Dependencies** | Three new indirect modules from the audio examples; no direct dependency changes. |
| **Binary size / build time** | Not measured in this PR; desktop and `GOOS=js GOARCH=wasm` builds compile. |

## Breaking changes / migration

| Item | Migration |
|---|---|
| `host.Screen` gains `Hover`, `Press`, `Release`, `KeyDown`, `KeyUp` | Add the methods to a custom screen, or use `*ownframe.Page`, which implements them. |
| Unkeyed `Config` and `Handlers` literals break from the new fields | Use keyed literals. |
| `New` can return a CSS parse error for an invalid `Config.Theme` | Validate the theme source; empty or blank means no theme. `SetTheme` returns the same error. |
| A desktop file input now opens a dialog under `Run` | Keep the typed-name path for hosts with no picker; tests install a fake with `page.InstallPicker`. |
| `go.mod` replaces `gowkhtmltopdf` with `../gowkhtmltopdf` | Build with the sibling checkout; drop the replace and pin a published commit once the engine branch merges. |

## Test plan

- [x] `make test TEST_P=4`
- [x] `go vet ./...`
- [x] `GOOS=js GOARCH=wasm go build ./...`
- [ ] Desktop smoke via `make open` (scrollbars, background underlay, fallback badge, file dialog); not run in the PR-preparation session.
- [ ] No CI is configured in this repo.

### Commands

```sh
make test TEST_P=4
go vet ./...
GOOS=js GOARCH=wasm go build ./...
```

## Screenshots / sample output

```
$ make test TEST_P=4
go test -p 4 ./...
ok   github.com/chinmay-sawant/ownframe/examples/spotify-player/player
ok   github.com/chinmay-sawant/ownframe/examples/dino/dino
ok   github.com/chinmay-sawant/ownframe/internal/filepick
ok   github.com/chinmay-sawant/ownframe/internal/frame
ok   github.com/chinmay-sawant/ownframe/internal/page
ok   github.com/chinmay-sawant/ownframe/internal/render
ok   github.com/chinmay-sawant/ownframe/internal/replay
ok   github.com/chinmay-sawant/ownframe/internal/window
... all packages ok; no failures
```

Desktop captures were not taken in the PR-preparation session.

## Related issues

None. No tickets exist in this repo; `plans/v0.0.1/compare.md` is the reference.

## PR metadata checklist (author)

- [x] Body copy written at `plans/v0.0.1/pr/pr-chore-pending.md` (kept local, not committed).
- [ ] Self-assigned with `--assignee "@me"` (pending `gh pr create`).
- [ ] Labels applied: `enhancement`, `documentation`.
- [ ] Ticket IDs linked: not applicable, no issues exist.

## Follow-ups (out of scope)

- Push the gowkhtmltopdf branch, pin the engine commit in `go.mod`, and drop the `replace` directive.
- Refresh stale `PHASES.md` hashes: the ownframe tip is `f62dedc` (the file says `614d83a`); the engine branch is `2111b36` locally (the file says `54a29b6`, remote `7f8164f`).
- `documentation/features.md` says the features "did not add modules"; the three audio modules contradict that.
- `documentation/README.md` does not index `binding.md` or `features-examples.md`.
- Replay gaps: elliptical fills, blend/isolation groups, rotation, fake oblique, font features, autospace, and images without a decodable payload.
- Example polish: spotify queue/devices/back-forward have no handlers; the wispr `#download` button does nothing; fetch defaults to an external URL and shows error text offline; dino `-web` is static because `Serve` does not tick.
- The planned doom example was dropped before merge; `examples/doom` is absent.

## Reviewer checklist

- [ ] Behavior matches summary and test plan; replay gates and the renderer draw the same operation set.
- [ ] No unrelated changes in the diff; no secrets or generated artifacts committed.
- [ ] Public API changes are documented (`SetTheme`, `SetTick`, `SetImage`, pointer and key methods, `Handlers` fields, `host.Screen` and `host.Ticker`).
- [ ] `host.Screen` break is acceptable, or a compatibility path is added.
- [ ] The local `replace` is understood as temporary.
- [ ] Diff-stat-by-extension table pasted at the bottom.

## Diff stat by extension

Generated from `git diff master...HEAD --numstat`, grouped by final extension as the PR template describes (the template's `scripts/pr-diff-stat.sh` lives in the gowkhtmltopdf repo, not this one). No renames or binary files; the file with no extension is `Makefile`.

| Extension | Files | Insertions | Deletions |
|-----------|-------|------------|-----------|
| `.go` | 388 | 17594 | 341 |
| `.css` | 22 | 1208 | 0 |
| `.html` | 39 | 1166 | 1 |
| `.md` | 15 | 404 | 38 |
| `.svg` | 77 | 321 | 0 |
| `(none)` | 1 | 55 | 0 |
| `.sum` | 1 | 8 | 0 |
| `.gitignore` | 1 | 5 | 0 |
| `.mod` | 1 | 3 | 0 |
| **Total** | 545 | 20764 | 380 |
