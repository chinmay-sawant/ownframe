# Historical record (not live)

This file is a snapshot of the v0.0.2 phase diary. It is not the live 1.0 checklist. Current status is in `plans/README.md`.

# Phase status

Recorded 2026-10-04.

## v0.0.2 (on feature/v0.0.2)

| Area | What shipped |
|---|---|
| Pipeline split | `internal/render` caches the parsed tree and styled document; `SetSize` and state changes call `css.Relayout`, so a drag or a hover reparses nothing. Page counters feed `host.Inspector`. |
| Dynamic resize | `MaxWidth`/`MaxHeight` are window bounds; a drag commits one relayout per settled size (100 ms motion throttle); hover and active re-resolve after a relayout; the scroll clamp runs on relayout and navigation; `examples/resize`. |
| Incremental repaint | `Page.Invalidate`, the display-list diff, `TakeDirty`, `replay.DrawRect`, and a persistent buffer sized to the page content; byte-equality tests hold the page after an interaction to the same PNG as a full `Redraw`, and the op filter and buffer plan have unit tests. The bitmap fallback keeps a full repaint. |
| Hot reload | `Config.File`/`ThemeFile`, the 250 ms poll, pending retry on a broken file, state survival, `Serve` reload plus shell refresh, `examples/reload`. |
| DevTools | `Page.Stats`, `host.Inspector`, the F12 or Ctrl+Shift+I right dock with box picking, the Elements JSON tab, the Frame counters tab, and the Ops list with operation outlines, `GET /debug/state`, `examples/devtools`. The dock never enters `Page.PNG`, the display list, or the box list. |
| Drag and drop | `host.Dropper`, the window drop pass, `Handlers.Drop`, `examples/drop`. `-web` has no window loop and receives no drops; the wasm canvas does. |
| Packaging | `scripts/package.sh` with the `-n` dry run, archive layouts and `SHA256SUMS`, `documentation/packaging.md`. |
| Input and interaction | Caret and selection range, click-to-offset through `internal/textrun`, Tab focus traversal, drag/double/triple selection, the context menu, cursor shapes, `ScrollTo`/`ScrollBy`, touch drag and pinch, F11. |
| Window options | `RunWithOptions` and `WindowOptions` (`Transparent`, `Borderless`, `Floating`, `FixedSize`, `MousePassthrough`, `Interactive`, `Draggable`, `BottomRight`, `Margin`), and `examples/desktop-cat`. |
| Engine | `css.Relayout` on gowkhtmltopdf `chore/changes-for-go-gpui` (merged from `feature/v002-relayout`), with a sheet cache for repeated collection of one tree. |

## v0.0.2 pending

1. Open the PR for `feature/v0.0.2`; the body is ready at `plans/v0.0.2/pr/pr-v0.0.2.md`.
2. After the merge, recreate the local `v0.0.2` tag on the merge commit and push it with the release notes.

## v0.0.2 release state (recorded 2026-10-05)

- gowkhtmltopdf `chore/changes-for-go-gpui` is pushed at `1a39183`; `go.mod` pins `v0.2.7-0.20261004151708-1a3918301a68`, and the local `replace` is gone (local, uncommitted).
- `feature/v0.0.2` is pushed at `2d758fd`, 107 commits ahead of `master`.
- `plans/v0.0.2/release-notes.md` holds the tag details. The local `v0.0.2` tag points at `2d758fd` and predates the local pin and notes.

## v0.0.2 limits and engine gaps

- IME: Ebiten v2.10.4 ships the experimental `exp/textinput` package, so the work is window wiring plus the page-side composing run, not an upstream ask. The route and the design are recorded in `plans/v0.0.2/input-interaction.md`.
- `@media (height)`, `(min-height)`, and `(max-height)` never match: the engine accepts only width and inline-size names (`internal/css/container.go` in gowkhtmltopdf). `TestMediaHeightFollowsTheFrame` is skipped with this reason.
- The bitmap fallback repaints the whole canvas; a partial bitmap path needs a new engine call and is out of scope for v0.0.2.
- Canvas, accessibility, video, WebGL, and auto-update stay after v0.0.2 with the reasons in `plans/v0.0.2/electron-gaps.md`.
- Up/Down in a textarea and selection beyond the visible scroll of a field stay out; the value model is a single line.
- A page that fetches during layout or runs a clock can change pixels the diff does not know about. `Invalidate()` with no argument is the escape hatch.
- The inspector shows geometry and attributes, not computed styles. The engine styles a document internally and publishes no per-element reader; a computed-style call for one element id is the follow-up.

## v0.0.1 status (previous)

Recorded 2026-10-03.

- gowkhtmltopdf: `chore/changes-for-go-gpui` at `54a29b6` locally (remote at `7f8164f`).
- ownframe: `chore/pending` at `614d83a` locally, from `master` `ac2ddc4`; `master` untouched.

## Done (on the local branches)

| Area | What shipped |
|---|---|
| A. Template-faithful form controls | Author attributes kept; `data-ownframe-field/focus/selected/placeholder`; caret; placeholder; default stylesheet after `<head>`; login on real inputs. |
| Buttons | Engine UA face; submit-like inputs rewritten to buttons; ownframe workaround removed. |
| B. Replay coverage | Rounded strokes, images, elliptical and masked strokes, transformed images, letter-spaced text, CSS outlines. |
| C. Binding layer | `data-bind` two-way binding; `Handlers.BeforeEdit` and `Handlers.Change`. |
| A leftovers | `FormSelected`; `Cut` writes through binding, fires Change, and is undoable in the login example. |
| D. Engine gaps | Named colors; `oklch`/`oklab`/`color-mix`/`light-dark`; `min`/`max`/`clamp`, general `calc`, `dvh`/`svh`/`lvh`; `@supports`/`@layer`/`@property`; `data:` fonts and null-transform woff2; `conic-gradient`; basic `clip-path`; input value/placeholder painting; control UA faces. |
| Focus and pointer state | `:focus`/`:focus-visible`/`:hover`/`:active`/`:checked`; `css.Options.Focus/Hover/Active`; `Page.Hover/Press/Release`; window pointer tracking; `host.Screen` additions. |

## Integration (closed 2026-10-05)

1. Engine branch pushed; `chore/changes-for-go-gpui` is on origin at `1a39183`.
2. `chore/pending` merged; `master` moved to `5a428f4` (PR #3 and PR #4).
3. `go.mod` pins `v0.2.7-0.20261004151708-1a3918301a68`; the local `replace` is gone.

## Pending — engine limits

- WOFF2: transformed `glyf/loca` (Google Fonts) unsupported; TTF/OTF and null-transform WOFF2 work.
- `:disabled`, `:target` never match; `:focus-within` not implemented.
- `::placeholder`, `::selection` pseudo-elements never match; ownframe covers them with `data-ownframe-placeholder` and `data-ownframe-selected`.
- `:hover`/`:active`/`:focus` match the exact id only: no ancestor hover, no state for elements without ids.
- `@layer`: per-stylesheet order; `!important` does not reverse layers.
- `clip-path`: raster-only (backgrounds/images); `@property` syntax parsed but not enforced; bare inputs have no UA width.

## Pending — replay fallbacks

- Blend/isolation groups, rotated text, fake oblique, font features, autospace.
- Elliptical fills, unknown stroke masks, non-image transforms, images without a decodable payload.

## Optional cleanups

- Drop the span rewrite once the engine exposes value mutation for editing.
- `chore/miscellaneous` is superseded by `chore/pending`.
