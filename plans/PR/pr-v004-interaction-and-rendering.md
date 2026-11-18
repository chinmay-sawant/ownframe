# feat(window): add Linux accessibility actions, shared component interactions, viewport sticky, and faster resize

## Summary

- Finishes the four v0.0.4 tracks: accurate docs and a measured resize speedup, Linux accessibility actions (a screen reader can focus and press controls), shared tab, slider and dialog behavior in `components/chadcn`, and viewport `position: sticky` with height media queries.
- Pins the engine to blinkless `master` at `f68817e582e2`, which now contains the sticky and height-query work from blinkless pull request 39.
- Merges `master` into the branch. The merge needed a snippet-manifest fix and the engine pin change below, and no Go source edits.

---

## Motivation / context

- Plans: `plans/v0.0.4/README.md` and the four track files `01` to `04` next to it. Each row is closed with its command, exit status and evidence path.
- Sticky and height queries live in the engine, so track 04 needed a blinkless release first. That is blinkless pull request 39, merged as `f68817e582e2`.
- Issues: see **Related issues**.

---

## Changes

### 01 Documentation and performance

- Corrects the flex-absolute support claim in the docs.
- Adds a list workload benchmark (`examples/perf-benchmarks`, `scripts/perf_baseline.py`) and a before and after record in `plans/v0.0.4/performance-evidence.md`.
- `internal/page/page_size.go` keeps template data clean when only the window size changes and no scroll callback is active. Layout still reruns, so media queries and geometry update. A test fails without the change (one extra template run).
- Same host, perf off, medians: resize 25.7 ms to 23.4 ms (about 9% lower), bytes per op down about 11.7%, allocations down about 41.9%. Text update, unchanged redraw and windowed scroll stayed within the 15% bound. These are page-work numbers. They say nothing about native frame time.

### 02 Linux accessibility actions

- `internal/atspi/action_queue.go` queues actions that an assistive client asks for and waits for the real result. It handles cancel, shutdown and a full queue. `serve_action.go` and `live_action_linux.go` connect it to the bus, and `internal/window/a11y_atspi.go` wakes the idle window so the action runs and paints.
- `scripts/check-atspi-actions.py` checks live D-Bus delivery. Results are in `plans/v0.0.4/02-linux-accessibility-actions.md`.
- Only Linux is covered. Windows, macOS and Android are not claimed.

### 03 Component interactions

- `components/chadcn/interactions.go` adds `chadcn.Interactions` (`NewInteractions`, `Mount`, `Update`, `Handle`, `Close`). An app installs tab keyboard handling, slider keys and pointer drag, and dialog focus and Escape once, instead of writing them again.
- `examples/builder/ui-builder` and `examples/builder/loglens` use it. The slider drag code in `ui-builder` shrinks by about 100 lines.
- `TestManagedInteractionsNative` drives these on the real update loop.

### 04 Viewport sticky and height media queries

- `internal/page/page_sticky.go` and `internal/window/sticky.go` ask the engine for the sticky view at the current scroll and paint and hit-test from it. `SetScrollViewport` is a host-interface method called by the window. It is not on the public `ownframe.Page`.
- `TestStickyViewportPixelsAndHits` covers activation, release at the containing block, scroll return, overlap, edits while scrolled, and zoom. `scripts/check-sticky-reference.mjs` and `scripts/compare-sticky-captures.py` compare eight frames with Chrome.

### Merge with master and engine pin

- `documentation/snippets.json` drops the `chadcn.md#3` entry, because the branch removed that snippet from `documentation/chadcn.md`.
- `go.mod`, `go.sum`, `examples/go.mod` and `examples/go.sum` pin blinkless `v0.1.2-0.20261118113835-f68817e582e2`.
- `plans/v0.0.4/04-css-conformance-and-engine-adoption.md` says the pin moved and which checks ran at the new pin. Older evidence files still name `27722d9cedd2`, and are kept as recorded.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Resize is faster as measured above. Everything else measured within 15% of before. |
| **Memory** | Resize allocates about 11.7% fewer bytes. |
| **Behavior / correctness** | Sticky headers now stay pinned in the window and are hit-testable where drawn. Height media queries follow the window size. Accessibility clients can focus and press controls on Linux. |
| **API / CLI** | Adds `chadcn.Interactions`. No new `ownframe` root symbol, alias or `Page` method. |
| **Dependencies** | Blinkless pin moves forward. No new module. |
| **Binary size / build time** | Not measured. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root. The diff shows none.
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows. The new exported one, `SetScrollViewport`, is called by `internal/window` through `internal/host`.
- [x] No new example `ownframe/internal/` imports (`make internal-import-check` passes inside `make lint`)
- [ ] No synonym names. Not audited line by line.
- [x] New exported methods return named types in package `ownframe` or a public subpackage. `chadcn.Interactions` methods return `error` and `*ui.View`.
- [ ] Library code panics only for invariants. Not audited line by line.
- [x] `context.Context` is first on the new functions that do I/O or can be cancelled (`Mount`, `Update`, `Close`, `NewActions`, `Request`).
- [x] No new examples that resemble Telegram, Teams, Spotify or Wispr. The only new example directory is `examples/perf-benchmarks`.

## Test plan

- [x] `make test`
- [x] `make lint` (Go, docs, plans, internal imports, website ESLint)
- [x] `make build`
- [x] `make docs-check plans-check examples-check`
- [x] `make graphics-check`, run at blinkless `2c9d915cc234`. The final pin has the same engine code plus one plan document, and `graphics-check` was not re-run at it.
- [ ] `make test-race` was not run.

Also driven by hand in a real window: all six tabs of `examples/blinkless-all`. Pictures matched descriptions and readouts, arrow-key focus moved as read out, and the caret field typed and deleted. Linux and WSLg only.

### Commands

```sh
make test
make lint
make build
make docs-check plans-check examples-check
make graphics-check
```

---

## Screenshots / sample output

No new screenshots. Sticky frames and manual test captures are under `plans/v0.0.4/evidence/`.

---

## Related issues

- None.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`enhancement`, `documentation`)
- [x] Related issues filled (none)
- [x] Filled body committed under `plans/PR/pr-v004-interaction-and-rendering.md`

---

## Follow-ups (out of scope)

- Windows UIA, macOS VoiceOver and Android TalkBack actions.
- Native Wayland and hardware-GPU checks. The graphics gate runs on Xvfb software rendering.
- `make test-race` for this branch.
- Re-run `make graphics-check` on the final pin when convenient.

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 1 | 1 | 1 |
| `.go` | 71 | 2677 | 300 |
| `.html` | 1 | 1 | 0 |
| `.json` | 13 | 1111 | 9 |
| `.md` | 18 | 1072 | 39 |
| `.mjs` | 1 | 100 | 0 |
| `.mod` | 2 | 2 | 2 |
| `.png` | 23 | Binary | Binary |
| `.py` | 5 | 244 | 4 |
| `.sh` | 1 | 5 | 0 |
| `.sum` | 2 | 4 | 8 |
| `.txt` | 24 | 1593 | 0 |
| **Total** | **162** | **6810** | **363** |
