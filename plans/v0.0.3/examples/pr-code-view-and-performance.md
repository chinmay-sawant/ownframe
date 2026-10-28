# PR: Code Review example and the hot-path perf pass

Base: `master`. Head: `example/code-view-and-performance`.
Range: `2d6df76e..2b76dade` before this record, 20 commits. `examples/code-view` is new on this branch.

## Summary

This branch adds the Code Review example and the six-phase hot-path pass in `plans/v0.0.3/performance`. Code Review is one HTML page for a live workspace: explorer, editor, source control, history, search, Markdown preview, and six color themes. The perf pass cuts work on idle ticks, DOM diffs, syntax highlighting, and soft-wrap resize. The tables below are the bench figures recorded in that plan.

## Motivation / context

- Plans: `plans/v0.0.3/examples/code-view/` and `plans/v0.0.3/performance/`
- Issues: see **Related issues**
- Code Review is the desktop viewer for this cycle. The perf plan is the follow-up that removes the allocations those screens were hitting.

## Changes

### Code Review

`go run ./examples/code-view` opens the current directory in a normal OS window through `ownframe.Run`. An empty root loads the 17-file sample. `-web` serves the same page on `127.0.0.1:8145` and does not start the file watcher. There is no script tag, no canvas, and no import of `ownframe/internal`.

The default frame is 1500×900. The minimum is 720×540. The activity bar switches Explorer and Source Control and opens quick open, Details, Settings, and the devtools dock. Devtools start on. A saved session does not hide the dock unless the launch passes `-dev=false`. `-perf` defaults to on. The host scrollbar is `ScrollbarHidden`. Each editor pane paints its own thumb. Ctrl+wheel and the zoom chords do not scale the page.

The editor has two groups with separate vertical offsets, in-place edit, soft wrap, horizontal scroll when wrap is off, gutter marks, and an in-page tab and edit menu. Source Control reads `git diff HEAD` plus untracked files. History shows up to 20 commits at first and grows on a worker. The branch picker loads another branch's log without changing the checkout. `Ctrl+P`, `Ctrl+F`, and `Ctrl+Shift+F` cover quick open, in-file find, and workspace search. Markdown preview renders fixed 20 px rows, including tables and text flowcharts and sequence diagrams. Settings stores one of six themes on the workspace session: Midnight, Dracula Night, Light, Solarized Light, Nord, and High Contrast.

The session file lives under the user config directory, mode `0700` for the directory and `0600` for the file. Paths stored in it are workspace-relative.

### Performance

`b6d87627` added the hot-path benchmarks and reads a projective transform through `Op.ProjectiveTransformValue`. That accessor change has no timing. `54e4e4da` is phases 01 through 06: one reused draw-options struct, retained layer scratch slices, an idle-tick bypass when nothing mutated, a pooled diff arena, a line token cache in Code Review, a retained wrap-plan buffer, and reused tile blit options.

The figures below are the ones written into `plans/v0.0.3/performance/` after `go test -benchmem -benchtime=200ms -p 1` on an i7-13700HX. There is no saved bench log in the repo. One sample, not a median of three. The **After** column is the recorded result.

| Hotspot | Before | **After** |
| --- | --- | --- |
| Idle full tick, 2000 unchanged ops | 522 µs | **26.59 ns, 0 allocs** |
| DOM position diff, 2000 ops | 954 µs, 313 KB, 2,011 allocs | **399 µs, 88 B, 0 allocs** |
| Go highlight, 1000 unchanged lines, warm cache | 671 µs, 365 KB, 9,395 allocs | **6.74 µs, 24.5 KB, 1 alloc** |
| Python highlight, same shape | 1.21 ms, 776 KB, 14,832 allocs | **7.04 µs, 24.5 KB, 1 alloc** |
| Soft-wrap resize, 1000 lines | 656 µs, 1.82 MB, 6,350 allocs | **21.1 µs, 66 B, 0 allocs** |

| Hotspot | Before | **After** |
| --- | --- | --- |
| Animation layers, 1000 ops | 32.16 µs, 45.3 KB, 12 allocs | **15.31 µs, 3.3 KB, 6 allocs** |
| Raster blit | 662.5 ns, 588 B, 4 allocs | **567.7 ns, 580 B** |
| Letter spacing | 51.40 µs direct, or 79.31 µs cold | **41.08 µs, 21.0 KB, 215 allocs** |
| Full tile paint, 800×600 | 20.24 µs, 14.2 KB, 100 allocs | **21.59 µs, 14.2 KB** |

Idle tick, DOM diff, syntax, soft wrap, layers, blit, and letter spacing got faster and allocate less. Full tile paint did not get faster. The plan reuses its buffers and the recorded time went from 20.24 µs to 21.59 µs.

Two labels in the plan overshoot the benches. The 19,634× line is 522 µs to 26.59 ns on the idle full tick, which returns as soon as `tickMutated` is false. It is not 374 µs to 0.756 ns. The syntax rows hit a process-global cache of unchanged lines, so they are not a measured keystroke. A real edit still rescans from the dirty line. All six phase checklists are marked completed in that directory.

### Library, window, and web

`MenuItem` is a new root alias for `page.MenuItem`, named in `plans/v0.0.3/prod-ready-task/phase-01-public-api.md`. `Page.ContextMenuFor` and `Page.MenuAction` run a custom right-click menu. `Handlers.Wheel` sees the desktop wheel first. `Config.DisableWheelZoom` and `Config.DisableKeyZoom` default off, so other examples still zoom. `Serve` now always exposes `GET /state`, `POST /key`, and `POST /wheel`, and ticks a page at about 60 Hz. `ServeOptions.Listener` uses a caller-supplied listener. `Click`, `Title`, `Boxes`, `WantFrameAt`, `WantFrameIn`, `ScrollTo`, and `ScrollBy` moved out of the testing files. Their signatures stayed.

Linux desktop `Run` can dial AT-SPI and report `Live() == true` after the name is claimed. Non-Linux `Live` stays false. This PR does not validate that bus with Orca.

`253705c1` also pools patch ops, keeps a spare viewport texture across a one-frame scroll pause, moves MP3 and WAV decode off the frame thread, and makes `secret.Secret` JSON and text marshaling emit `[REDACTED]`.

### Dependencies

The library `go.mod` is unchanged. The examples module adds `github.com/fsnotify/fsnotify v1.10.1` for the watcher.

## Impact

| Area | **Impact** |
| --- | --- |
| **Performance** | Idle tick, DOM diff, syntax cache, and soft-wrap resize allocate less and return sooner. See the tables above. Tile paint time did not drop. |
| **Memory** | Diff arena, wrap plan, layer scratch, and draw options are reused. Syntax cache holds tokenized lines for the process. The viewport keeps one spare window-sized texture. |
| **Behavior / correctness** | Code Review is a new example. Serve ticks and captures wheel and keys. Linux desktop startup may publish an AT-SPI tree. Textarea paste keeps newlines. Ctrl+Shift+V no longer pastes. |
| **API / CLI** | New menu methods, two zoom flags, and `ServeOptions.Listener`. Existing `Page` method signatures that moved files did not change. |
| **Dependencies** | Examples module only: fsnotify. |
| **Binary size / build time** | Not measured. |

## Breaking changes / migration

| Item | Migration |
| --- | --- |
| `secret.Secret` JSON and text | `Marshal` and `MarshalText` now write `[REDACTED]`. Read the value in process. Do not marshal it. |
| Ctrl+Shift+V | No longer pastes. Shift+Insert still does. |
| Textarea paste | Newlines stay. A custom `Handlers.Paste` receives them. |
| `Serve` / `-web` | Every page ticks at about 60 Hz. `GET /state` is always registered. The shell captures wheel and keys. The browser wheel does not scroll a tall picture unless the page handles `Wheel`. |
| Linux AT-SPI | Desktop `Run` dials the bus and may publish the tree. Non-Linux stays dark. |
| Unkeyed `Config` or `Handlers` composite literals | New fields would break them. Keyed literals are fine. None were found in this repo. |

## Library freeze (until Page wrap)

Until `plans/v0.0.3/prod-ready-task/` wraps `Page`, reject the PR if it does any of the following.

- [x] No new `type X =` at the module root except documented re-exports. The one new alias is `MenuItem`, listed in `phase-01-public-api.md` and `alias_allow_test.go`.
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows. The new methods are used by the window or the menu path.
- [x] No new example `ownframe/internal/` imports. `scripts/example-internal-allowlist.txt` is unchanged.
- [x] No synonym names. Zoom policies stay separate. `ContextMenuFor` does not rename `ContextMenu`.
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR. New returns are `[]MenuItem`, `error`, or `bool`.
- [x] Library code panics only for invariants. The new `panic` is in a web tick test.
- [ ] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline. Public menu and wheel handlers take `ctx` first. `newLiveBus` in `internal/atspi/live_linux.go` dials with `net.DialTimeout` and no `context.Context`.

## Test plan

- [x] `make test` passed on `2b76dade` (`go test -p 1 ./... ./examples/...`). This record is Markdown only.
- [ ] `make lint` / `go vet`. Not run for this pull request.
- [ ] `make test-race` (library module; when shared mutable state changes). Not run. The tick flag and the syntax cache are shared mutable state, so this is still worth a race run.
- [ ] `make build` (when binary output is part of the change). Not run. `go vet` of the whole tree was not repeated for this record.
- [ ] `make run` wall time vs baseline. That target belongs to the template's other repo. Not run here.
- [ ] `make reference-metrics` / gopdfsuit hard metrics if detector surface changed. Not applicable.

### Commands

```sh
make test
```

Code Review tests live in `examples/code-view` and cover the shell, syntax, diff, history, edit, session, themes, editor thumbs, and the public menu bridge. Screenshot tests write under `/tmp` only when `OWNFRAME_CODE_VIEW_SCREENSHOT=1`.

## Screenshots / sample output

No new window capture is attached. The bench tables in **Performance** are the sample output for this PR.

## Related issues

- Relates to #48. Linux desktop `Run` can dial AT-SPI and claim the name. Orca and Accerciser were not used to check the tree.
- No open issue is fully completed here. #44 and #45 stay open. This PR does not add TalkBack, and it does not compare Code Review with Chromium.

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`, `documentation`)
- [x] Related issues filled with real ticket IDs
- [x] Filled body committed under `plans/v0.0.3/examples/pr-code-view-and-performance.md`

## Follow-ups (out of scope)

- `examples/code-view/README.md` and `plans/v0.0.3/examples/code-view/` still say the minimum is 840×700. The code and `header_layout_test.go` use 720×540.
- The example README still says the two groups share one page scroll and that there is no horizontal scroll. `split_scroll.go` and `hscroll.go` do both.
- `SetTick` still says Serve does not tick. `Serve` does tick.
- `internal/atspi/doc.go` still says `NewLiveBus` never reports `Live`.
- `ServeOptions.Listener` is not checked against `InsecureAllowRemoteBind` unless `addr` is empty or ends in `:0`.
- Phase 05 through 08 still list fuzzy search, go to line, a phone layout, and a high-contrast audit as open. A High Contrast theme exists. That is not the audit.
- Re-run the benches and keep the log if these tables need to be cited outside the plan.

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

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh master` on `2b76dade`, then this file added as one `.md`.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 16 | 3193 | 0 |
| `.diff` | 1 | 833 | 0 |
| `.go` | 264 | 43025 | 476 |
| `.html` | 13 | 866 | 0 |
| `.json` | 2 | 20 | 0 |
| `.md` | 45 | 2406 | 196 |
| `.mod` | 1 | 4 | 0 |
| `.py` | 2 | 309 | 0 |
| `.sum` | 1 | 4 | 10 |
| `.yaml` | 1 | 30 | 0 |
| No extension | 1 | 15 | 0 |
| **Total** | **347** | **50705** | **682** |
