## Summary

Bundle on `chore/fixes` against `master` (234 files, +28459/-1103). The code viewer keeps the earlier drawer, inline card, zoom, multi-select, history, and staged SCM work, and adds block-scoped revert with undo and redo, change-block stepping, gutter and divider revert controls, preview line numbers, and Git hunk-range docs. The window host work from the earlier bundle is unchanged. The website gains separate engine and compare pages, a team section, runnable starter demos with live WASM builds, and a highlighting fix with copy buttons. Build output under `docs/` and `frontend/public/` grows with the new demos. CI gets a site toolchain pin, xvfb for race shards, and two test timing fixes.

## Motivation / context

- Plans: no separate plan; this is a fixes bundle.
- Issues: no issue is linked to this work.
- Baseline: `git diff --shortstat master...HEAD` is 234 files changed, 28459 insertions, 1103 deletions. The earlier record at `797cbee2` covered 128 files (+11385/-524). The commits since add the website series, revert-block work, and CI fixes.

## Changes

### Code viewer: explorer and drawer (unchanged from earlier record)

- Explorer multi-select: Ctrl+click toggles, Shift+click extends the range; Ctrl opens without clearing the selection.
- Sticky open-folder headers above the scroll window; Space toggles the focused folder or opens the focused file.
- Right-click explorer menu: Copy Path, Copy File, Paste; Ctrl+V pastes staged copies, `file://` URI lists, plain text to `paste.txt`, and PNG/JPEG/PDF bytes; `C:\…` paths translate to `/mnt/<drive>/…` on WSL.
- Resizable file/source-control drawer via the `drawer-resize` handle, up to +80 px (~30%), with the widening persisted per workspace.

### Code viewer: editing and inline card (unchanged from earlier record)

- In-place editing keeps line numbers visible in edit mode.
- Line-number clicks and Ctrl+I open the inline card at that line, including unchanged lines (shown under a neutral `Line` kind).
- Multi-line edit selections render as one indented snippet with a single file-plus-range field, instead of one numbered row per line.
- The card survives scrolling by re-anchoring to its line, keeps the user's drag offset across scrolls, drags by header or body in any direction with page clamping, and no longer closes on body clicks (X and Escape still close it).

### Code viewer: source control and diff

- Staged/unstaged sections with per-file and bulk stage/unstage, and discard behind a confirm dialog that blocks other clicks.
- Diff safety limits: oversized files and diffs render a notice row instead of opening or painting; shared line budget across files and diffs.
- Wrapped-row cache per file and column budget reused on scroll; unwrapped rows paint from the visible rune window.
- Block-scoped revert (`b435beae`, `60c576e9`): one static discard arrow per change block, unified `↩` on the first added row and split `→` on the center divider slot; `revert_block.go` resolves `revert-line-<hunk>-<old>-<new>` to the contiguous source block; `revert_hunk.go` applies raw `-U0` patches for worktree and index, preserves tabs and EOL, and isolates adjacent blocks in one displayed hunk. Separate revert undo and redo stacks with Ctrl+Z/Y, cleared on reload and project switch. Open documents refresh in place after revert.
- Change-block stepping (`b435beae`): `hunk-prev/next` crumbs move by contiguous change block through `fileNavChangeStep`, with shared bottom-center highlight and enabled state. Single additions and adjacent add/delete pairs leave both arrows off.
- Gutter and divider paint (`b435beae`, `60c576e9`): dropped flattened `z-index: 2` on `.ln`, `.gutter-mark`, and `.line-revert` so row fills no longer bury text; wrapped continuations clear the marker so ids stay unique; split edit targets rescoped to text cells and the divider arrow shifted +4px.
- Git hunk ranges (`a9487552`): README documents `@@ -old,count +new,count` semantics, change-block stepping, and block revert with Ctrl+Z; `diff_git_ranges_test.go` cross-checks real `git diff` for replacements, insertions, deletions, empty files, missing trailing newline, separated hunks, and `-U0`/`-U3` headers.

### Code viewer: navigation, zoom, preview

- Back/forward file history (100 entries, deduped, missing files pruned) with title-bar buttons and Alt+Left/Right.
- Layout zoom scales chrome, rows, and preview together, with Ctrl+wheel and keyboard chords, per-workspace persistence, and a reset.
- Markdown preview handles hidden elements; history graph renders in explicit pixels; sc-stats right-align; settings focus treatment on buttons.
- Preview line numbers (`b435beae`): `markdown_preview.go` assigns `No/Line`, `pane.go` reserves `previewLineW=44` in column math, `components/editor.html` and `components/preview.css` add `button.pln` and `span.pln`, with zoom rules in `zoom_sections.go`.
- Window-start snap (`b435beae`): `fillCode` and `secondView` snap the window start back to a visual-line start so mid-line offsets do not blank the gutter.

### Window host (unchanged from earlier record)

- Window placement persistence (`Options.PlacementID`): remembers monitor, position, size, and maximized state under owner-only permissions, with safe fallbacks.
- WSL manual-maximize rework: per-monitor workarea cache, restore sequencing fixes, and a plausibility check before applying a learned rect.
- Titlebar drag smoothing: low-pass cursor filtering, gesture tracking that cancels motion feedback, WSLg cursor-jump rejection, jitter coalescing, and min-visible clamping; relayout storms suppressed mid-drag.
- `DisableKeyZoom` no longer consumes Ctrl+=/-/0, so pages can run their own layout zoom.
- Devtools dock hit-testing: the custom-chrome drag strip no longer steals presses over the dock or resize edge, and the close hit spans the full header.
- Context menus grow an optional shortcut `Hint` label.
- No `internal/window/` changes in `797cbee2..HEAD`. Verified empty.

### Website and starters (`8fc47f91`, `9d9988de`, `0ff9f793`, `098a1481`)

- Home rework: hero with install command and clipboard status, Launch board preview card, showcase grid with live-demo links, `Hello World` and `Split with friends` picker under `#start`.
- New standalone pages: `Architecture.jsx` for `#engine`, `Benchmarks.jsx` for `#compare`, expanded `Team.jsx` with maintainer card and tool list, routing and mobile menu in `main.jsx`.
- Runnable starters under `examples/frontend/`: Launch board root, `hello/` Hello World, `budget/` Split with friends, `workspace/` Workspace Inspector. Displayed source matches compiled source through `?raw` imports in `ExampleCode.jsx` and `starter.js`.
- WASM wiring: `frontend/scripts/build-demo.mjs` builds `frontend`, `workspace`, `hello`, `budget`, `telegram`, and `spotify-player` targets and copies `wasm_exec.js`; `check-site.mjs` adds linked-demo, embedded-demo, starter, mobile, and clipboard checks with `--capture-demo` support.
- Highlight fix (`098a1481`): shared `highlight.js` keeps `{{…}}` template actions in one `hljs-template-tag` span; per-file Copy buttons with live-region status; `highlight.test.js` covers Go, CSS, and HTML cases; `style.css` covers emitted classes and the code bar.

### Test-only, scripts, and config (`f6291bc3`, website series)

- `internal/safego/safego_test.go`: poll up to 5s for the crash report before `TempDir` cleanup. Test-only fix for `TestSurvivesPanic`.
- `scripts/check-doc-snippets.py` and `documentation/snippets.json`: starter snippets now come from `examples/frontend/workspace/main.go` and `examples/frontend/budget/main.go` with template and style assets, replacing `Home.jsx` quick-start entries.
- `examples/spotify-player/main_js.go` and `examples/wispr-flow-dashboard/main_js.go`: `js`-build entrypoints that run the page with `ownframe.Run`.
- `examples/maintenance.json`: adds the landing-page WASM demo entry.
- `.gitignore`: un-ignores built starter and demo WASM under `frontend/public/demos/` and `docs/demos/`.
- Lint and fixture hygiene (`f6291bc3`): named constants for drawer, inline-card, and SCM limits, dead-code removal, `0o600` test perms, `errors.Is`, `ExecCommandContext`, and a `runtime.Caller`-based docs path in `final_shot_test.go`.

### Generated site output

- `docs/` and `frontend/public/` gain built WASM, `wasm_exec.js`, demo pages, preview PNGs, and team photos for the new demos (61 files, +10248/-108 in `797cbee2..HEAD`). Hand-written parts are the demo `index.html` shells, `docs/index.html` title change, and team images.

## Impact

| Area | Impact |
|------|--------|
| **Behavior** | Drawer, card, zoom, multi-select, history, staged SCM, block revert with undo, change-block stepping, preview line numbers, and starter demos are new or changed. Window placement, drag feel, and zoom-chord routing change host behavior. |
| **Performance** | Wrapped-row cache and visible-window painting cut per-scroll work; placement writes are debounced off the update path. |
| **API / CLI** | Adds `Options.PlacementID` and the menu-item `Hint` field. Adds `go run ./examples/frontend[/hello][/budget]` starters and `frontend/scripts/build-demo.mjs` targets. No other CLI changes. |
| **Dependencies** | None. |
| **Binary size / build time** | No meaningful change to shipped binaries. Checked-in WASM and demo assets grow the checkout. |

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
- [x] New revert and nav suites: `revert_hunk_test.go`, `revert_block_regression_test.go`, `file_nav_change_test.go`, `gutter_paint_test.go`, `split_divider_revert_test.go`, `diff_git_ranges_test.go`
- [x] Frontend checks: `highlight.test.js`, `examples/frontend` and `budget` `main_test.go`
- [x] CI fixes verified locally: `go vet -p 1 ./examples/code-view/ ./examples/music/` clean, `gofmt` clean
- [x] `go test -race -count=1 -p 1 ./examples/music/` green (async-decode poll fix)
- [x] `go test -count=1 -p 1 ./examples/code-view/ -run TestFinalShots` skips without the env gate, skips without a display when gated, passes under `xvfb-run` with `OWNFRAME_CODE_VIEW_SCREENSHOT=1`
- [x] `bash -n scripts/test-examples-race.sh` clean, `ci.yml` parses as YAML
- [ ] Full `make test` was not run locally (final-shot image tests need one `RunGame` per process and a display; CI covers it).
- [ ] `make lint` was not run locally; CI covers it.
- [x] Full `make site-check` green locally (npm test, generated-output check, `check:site` with local Chrome).

### Commands

```sh
go vet -p 1 ./examples/code-view/ ./examples/music/
go test -race -count=1 -p 1 ./examples/music/
go test -count=1 -p 1 ./examples/code-view/ -run TestFinalShots
OWNFRAME_CODE_VIEW_SCREENSHOT=1 xvfb-run -a go test -count=1 -p 1 ./examples/code-view/ -run TestFinalShots
go test ./examples/code-view -run 'TestInline|TestFeatureInline|TestActionControls|TestSelection' -count=1 -p 1
go test ./examples/code-view -count=1 -p 1 -run 'Test[^FD]|TestD[^r]|TestF[^i]'
make docs-check plans-check examples-check
```

### CI failures fixed in this update (run `41011673307`)

- `site`: added pinned `setup-go 1.26.10` to the `site` job; demo wasms build stamp-free, path-free, and pinned to the `go.work` toolchain so bytes are identical everywhere; the generated-output check compares only files git would commit so the ignored 108 MB unpacked Spotify wasm cannot fail it; the wheel scroll test now targets the sticky header with an iframe guard and polls for the scroll instead of assuming a fixed point and sleep; rebuilt `docs/` and `frontend/public/` output is committed.
- `test` and `examples-race (1)`: folded the four `final_shot_test.go` cases into one gated `TestFinalShots` with a single `RunGame` per process (second `RunGame` panics with `MarkGuest is called more than once`), gated on `OWNFRAME_CODE_VIEW_SCREENSHOT=1` with a display skip.
- `examples-race (1)/(2)`: `examples-race` job now installs `xvfb xauth libgl1-mesa-dri` and runs shards under `xvfb-run -a`, matching the `test` job.
- `examples-race (2)`: `TestEngineAsyncDecode` polls `Update()` up to 2s instead of assuming one `Update` lands the background-decoded voice.

## Screenshots / sample output

None; no visual test harness output is checked in. `frontend/public/demo/preview-dark.png`, `preview-light.png`, `demos/code-view.png`, and `demos/spotify-still.png` are the checked-in preview stills for the site.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [x] `bug` and `enhancement` labels applied when opened.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/PR/pr-chore-fixes.md`.

## Follow-ups (out of scope)

- Full `make test` and `make lint` verification happens in CI.
- Website build output and team-page assets under `docs/` and `frontend/public/demos/` may deserve a split into a site-only change. This update keeps them in one bundle so the PR body matches the branch.

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
| --- | ---: | ---: | ---: |
| `.css` | 17 | 2531 | 252 |
| `.go` | 107 | 14180 | 484 |
| `.gz` | 2 | Binary | Binary |
| `.html` | 27 | 1304 | 38 |
| `.jpg` | 2 | Binary | Binary |
| `.js` | 22 | 9324 | 111 |
| `.json` | 3 | 18 | 16 |
| `.jsx` | 7 | 480 | 159 |
| `.md` | 8 | 268 | 19 |
| `.mjs` | 2 | 277 | 13 |
| `.png` | 14 | Binary | Binary |
| `.py` | 3 | 59 | 10 |
| `.wasm` | 12 | Binary | Binary |
| `.webp` | 6 | Binary | Binary |
| `.yml` | 1 | 6 | 1 |
| No extension | 1 | 12 | 0 |
| **Total** | **234** | **28459** | **1103** |
