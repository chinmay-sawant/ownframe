## Summary

Bundle of code-view and desktop-window fixes on `chore/fixes`, from the merge base with `master` (`65b43475`). The code viewer gains a resizable file drawer, a snippet-style inline card that survives scrolling and drags in any direction, layout zoom, explorer multi-select with clipboard paste, sticky folders, back/forward file history, staged/unstaged source-control sections, and diff safety limits. The window host gains placement persistence, a WSL maximize rework, titlebar-drag smoothing, devtools dock hit-test fixes, and `DisableKeyZoom` no longer swallowing its chords.

## Motivation / context

- Plans: no separate plan; this is a fixes bundle.
- Issues: no issue is linked to this work.
- Baseline: `git diff --stat master...chore/fixes` is 128 files, +11385/−524.

## Changes

### Code viewer: explorer and drawer

- Explorer multi-select: Ctrl+click toggles, Shift+click extends the range; Ctrl opens without clearing the selection.
- Sticky open-folder headers above the scroll window; Space toggles the focused folder or opens the focused file.
- Right-click explorer menu: Copy Path, Copy File, Paste; Ctrl+V pastes staged copies, `file://` URI lists, plain text to `paste.txt`, and PNG/JPEG/PDF bytes; `C:\…` paths translate to `/mnt/<drive>/…` on WSL.
- Resizable file/source-control drawer via the `drawer-resize` handle, up to +80 px (~30%), with the widening persisted per workspace.

### Code viewer: editing and inline card

- In-place editing keeps line numbers visible in edit mode.
- Line-number clicks and Ctrl+I open the inline card at that line, including unchanged lines (shown under a neutral `Line` kind).
- Multi-line edit selections render as one indented snippet with a single file-plus-range field, instead of one numbered row per line.
- The card survives scrolling by re-anchoring to its line, keeps the user's drag offset across scrolls, drags by header or body in any direction with page clamping, and no longer closes on body clicks (X and Escape still close it).

### Code viewer: source control and diff

- Staged/unstaged sections with per-file and bulk stage/unstage, and discard behind a confirm dialog that blocks other clicks.
- Diff safety limits: oversized files and diffs render a notice row instead of opening or painting; shared line budget across files and diffs.
- Wrapped-row cache per file and column budget reused on scroll; unwrapped rows paint from the visible rune window.

### Code viewer: navigation, zoom, preview

- Back/forward file history (100 entries, deduped, missing files pruned) with title-bar buttons and Alt+Left/Right.
- Layout zoom scales chrome, rows, and preview together, with Ctrl+wheel and keyboard chords, per-workspace persistence, and a reset.
- Markdown preview handles hidden elements; history graph renders in explicit pixels; sc-stats right-align; settings focus treatment on buttons.

### Window host

- Window placement persistence (`Options.PlacementID`): remembers monitor, position, size, and maximized state under owner-only permissions, with safe fallbacks.
- WSL manual-maximize rework: per-monitor workarea cache, restore sequencing fixes, and a plausibility check before applying a learned rect.
- Titlebar drag smoothing: low-pass cursor filtering, gesture tracking that cancels motion feedback, WSLg cursor-jump rejection, jitter coalescing, and min-visible clamping; relayout storms suppressed mid-drag.
- `DisableKeyZoom` no longer consumes Ctrl+=/-/0, so pages can run their own layout zoom.
- Devtools dock hit-testing: the custom-chrome drag strip no longer steals presses over the dock or resize edge, and the close hit spans the full header.
- Context menus grow an optional shortcut `Hint` label.

## Impact

| Area | Impact |
|------|--------|
| **Behavior** | Drawer, card, zoom, multi-select, history, and staged SCM are new or changed viewer behaviors; window placement, drag feel, and zoom-chord routing change host behavior. |
| **Performance** | Wrapped-row cache and visible-window painting cut per-scroll work; placement writes are debounced off the update path. |
| **API / CLI** | Adds `Options.PlacementID` and the menu-item `Hint` field. No CLI changes. |
| **Dependencies** | None. |
| **Binary size / build time** | No meaningful change. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `DisableKeyZoom` passes Ctrl+=/-/0 to the page instead of swallowing them | Pages that relied on the swallow need their own handling; the code viewer already implements layout zoom. |
| Scroll and programmatic scroll no longer dismiss the inline card | Close with X or Escape, or open another file or mode. |

## Test plan

- [x] `go vet -p 1 ./examples/code-view`
- [x] Targeted code-view tests: inline, feature-inline, action-controls, selection suites
- [x] Full `examples/code-view` suite minus final-shot image tests (36s, green)
- [x] Drawer tests (`TestDrawer*`, green)
- [x] `gofmt` clean on hand-written files
- [x] `make docs-check plans-check examples-check`
- [ ] Full `make test` was not run locally (final-shot image tests are sensitive to parallel Ebiten runs; CI covers it).
- [ ] `make lint` was not run locally; CI covers it.

### Commands

```sh
go vet -p 1 ./examples/code-view
go test ./examples/code-view -run 'TestInline|TestFeatureInline|TestActionControls|TestSelection' -count=1 -p 1
go test ./examples/code-view -count=1 -p 1 -run 'Test[^FD]|TestD[^r]|TestF[^i]'
make docs-check plans-check examples-check
```

## Screenshots / sample output

None; no visual test harness output is checked in.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] `bug` and `enhancement` labels applied when opened.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/PR/pr-chore-fixes.md`.

## Follow-ups (out of scope)

- Full `make test` and `make lint` verification happens in CI.
- Website build output and team-page assets under `docs/` may deserve a split into a site-only change.

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
| `.go` | **86** | **10756** | **417** |
| `.css` | **12** | **300** | **21** |
| `.html` | **8** | **164** | **27** |
| `.md` | **6** | **50** | **13** |
| `.js` / `.jsx` / `.mjs` | **5** | **112** | **43** |
| images (`.png` / `.webp`, binary) | **10** | **0** | **0** |
| **Total** | **128** | **11385** | **524** |
