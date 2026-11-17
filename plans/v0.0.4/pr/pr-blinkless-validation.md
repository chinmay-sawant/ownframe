# feat(examples): add the blinkless validation windows and a way to test them by driving the window

## Summary

- Adds the OwnFrame side of the blinkless v0.1.2 work: three example apps
  (`blinkless-all`, `blinkless-scroll`, a preview that opens blinkless
  fixtures) that show the new CSS properties, plus the window and render
  changes they need (shared headers, `-webkit-text-stroke` paint, a silent
  preview with no sound device).
- Fixes two bugs found by testing `blinkless-all` in the real window, adds live
  state to its `-web` mode, and documents how to test an app by driving the
  window with the mouse and keyboard (`scripts/gui-drive.sh`).
- Clears all `make lint` findings so `make lint` and `make test` exit 0.

**Depends on blinkless pull request #38.** This branch uses engine features
that the pinned `blinkless v0.1.1` does not have (for example
`DisplayOp.TextStroke`). It builds through the committed `go.work`, which uses
the local `../blinkless` checkout. With `GOWORK=off` it fails to compile against
`v0.1.1`. Merge and tag blinkless first, then bump the pin here. Until then
this is a draft.

---

## Motivation / context

- Plans: the blinkless `plans/v0.1.2/` checklists name `examples/blinkless-all`
  as the place a person checks the new styles by hand.
- Testing every row of `blinkless-all` against its description found two
  disagreements, both in the host demos and not in the engine (see Changes).

---

## Changes

### Examples and window

- `examples/blinkless-all`: one window for the 61 properties that were not yet
  implemented, in six tabs (navigation and UI, draft borders and shapes, motion
  path, fragmentation, reading order, SVG leftovers). Live sessions react to the
  pointer, keys and clock on the navigation and motion tabs.
- `examples/blinkless-scroll`: the host side of scroll timelines, with
  timeline lists and scrollers.
- Preview: opens a blinkless fixture, opens the ten-style page with no
  arguments, adds an `env()` swatch, and opens silently when there is no sound
  device.
- Window and render: shared customizable headers across desktop applications,
  and the `-webkit-text-stroke` outline paint.
- Window screenshot on demand (`-snapshot`) and a calmer motion tab.

### Bugs found by testing in the window

- `shape-image-threshold`: the page never registered `url(mask)`, so the 0.5 box
  read no alpha and started its text 80px in while the readout said 40px.
  `NewPage` now registers the mask.
- `copy-into`: the demo printed the running header but never drew the heading
  its description says still paints. It does now.
- The scrolled action sheet painted over the pinned header. Its re-layout trick
  wrote a growing `z-index` onto the sheet box, and the pinned block has
  `z-index:10`. The trigger is now the custom property `--sheet-offset`.
- Regression tests assert on the painted page, not the readout
  (`page_paint_test.go`, `nav_sheet_test.go`).

### Web mode

- `-web` now attaches the live sessions (`web_live.go`): hover, click, key and
  typed text from the web routes drive the same sessions the window uses. A
  click carries the box under the point, not the point, so a slider click lands
  mid-track; dragging stays window-only. `web_live_test.go` covers the path.

### Testing and docs

- `scripts/gui-drive.sh` drives the real window with `xdotool` (`setup` fetches
  it without root) and reads it back through `-snapshot`.
- `documentation/testing.md` has a new section, "Testing an app by driving it",
  with the gotchas found on WSLg (keys need `--window`, the wheel hits whatever
  is under the pointer, the click origin is measured). `AGENTS.md` carries a
  short form.

### Lint

- Clears the `golangci-lint` findings in `internal/window`, the two
  `text_stroke.go` files and the examples: named consts, a clamped `int32`
  conversion, `Deallocate()` in place of the deprecated `Dispose()`, contexts
  passed to PNG capture and `exec`, and `//nolint` with a written reason where
  a path is operator-supplied.
- Behavior change to review: demo and `-png` output files are now written with
  mode `0o600` and output directories with `0o750`.
- `scripts/test_graphics_gate.py` lists `TestWindowSharedHeaderPixels`, which
  `check-graphics.sh` already ran.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Not measured in this PR. |
| **Memory** | Not measured in this PR. |
| **Behavior / correctness** | Two example demos now match their descriptions; the window code change is limited to lint-driven edits. |
| **API / CLI** | No public API change. Example flags: `-web` now has live state; output file modes are tighter. |
| **Dependencies** | None added. Needs the next blinkless release (see Summary). |
| **Binary size / build time** | Not measured. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| Demo and `-png` output files are `0o600`, directories `0o750` | Loosen with `chmod` if another user must read them. |

---

## Test plan

- [x] `make test` exits 0 (162 packages `ok`), run against the local blinkless
  checkout on `feature/v0.1.2-pending-css`
- [x] `make lint` exits 0
- [x] `go test ./examples/blinkless-all` exits 0, including the new painted-page,
  sheet and web-input tests
- [x] `make docs-check` exits 0
- [x] Manual: every one of the 45 `blinkless-all` examples checked in the real
  window (click, keys, drag, type, press-and-hold, toggle, pause), driven with
  `scripts/gui-drive.sh` on WSL2 with WSLg
- [ ] `make graphics-check` was **not run**
- [ ] `GOWORK=off` build fails until the blinkless pin is bumped (see Summary)

### Commands

```sh
make test
make lint
go test ./examples/blinkless-all
```

---

## Screenshots / sample output

Full-page window pictures of each tab were read during the manual pass. They
are not committed.

---

## Related issues

- Depends on https://github.com/chinmay-sawant/blinkless/pull/38.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`)
- [x] Related issues filled
- [x] Filled body committed under `plans/v0.0.4/pr/pr-blinkless-validation.md`

---

## Follow-ups (out of scope)

- After blinkless #38 merges and is tagged, bump `go.mod` and `examples/go.mod`
  to that tag and take this PR out of draft.
- Run `make graphics-check`; the testing guide still says twelve tests where the
  script runs fourteen.
- Selections from one row of `blinkless-all` stay highlighted when another row
  is clicked.

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.go` | 98 | 11630 | 24 |
| `.html` | 1 | 8 | 0 |
| `.json` | 2 | 15 | 2 |
| `.md` | 7 | 469 | 4 |
| `.py` | 1 | 1 | 0 |
| `.sh` | 2 | 150 | 0 |
| `.work` | 1 | 1 | 0 |
| **Total** | **112** | **12274** | **30** |

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
