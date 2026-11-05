# feat(examples): add ChadCN set and LogLens dashboard

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../../skills/PR/PR_TEMPLATE.md).

- Base: `origin/master` (`2d2cb57e`)
- Head: `chore/new-examples`
- Diff against `origin/master`: 382 files before this plan file

---

## Summary

This branch adds `components/chadcn`, a reusable set of about 70 components built on `ui` with host-owned open and selected state, plus light and dark palettes. It adds the LogLens observability dashboard and the ui-builder gallery under `examples/builder`, and it wires `-perf` and tag-gated `-debug-addr` across the examples. It also switches idle windows to event-driven pacing, makes the web preview scroll accurate, and bumps Ebiten to the November v2.11 alpha.

---

## Motivation / context

- Planned under `plans/v0.0.3/components/chadcn/README.md`. The set is the first under `components/<set>` and `ui` stays experimental.
- Builder demos moved under `examples/builder` so the gallery and the full dashboard share one folder.
- Idle work came from LogLens tick plus hover patching keeping windows awake. The debug server gives a loopback-only way to sample that.
- Issues: none. This work has no tracker issue.

---

## Changes

### ChadCN component set

- New `components/chadcn/` with pure constructors that return `*ui.Element`. State stays in the host: `Open bool` for dialog, sheet, popover and 12 more, `Open []string` for accordion, `Selected` maps for tables and toggles, `Selected time.Time` for calendar and date picker.
- `Stylesheet` embeds `css/*.css` with `00-base.css` first, then the rest sorted. `theme.go` sets `--of-*` variables for light and dark. Usage is `theme.Dark + chadcn.Stylesheet`.
- Moved `ui/components` controls and tests into `components/chadcn`. `ui/README.md` now points at the new set.
- `ui/builder.go` keeps generated control IDs stable across conditional siblings and allows `option` and `optgroup` tags. `ui/mount_test.go` pins the ID scheme.

### Builder demos

- `examples/builder/loglens` is a four-tab ops dashboard with seeded telemetry, a 1s tick, ten rows per page, and Escape closing every overlay. It exercises buttons, tables, tabs, command palette, dialog, sheet, drawer, charts, traces, and chat in one shell.
- `examples/builder/ui-builder` is the isolated catalog, one file per tab, with a notice line per panel.
- Both take `-web`, `-addr`, `-dev`, `-perf`, and `-debug-addr`. LogLens adds `-headless` on 8132 and drags from the top bar with `CustomChrome`. Gallery sits on 8131 with `-log-enabled` and `-log`.

### Window idle pacing and perf

- New `host.FrameScheduler` beside `FrameWanter`. Screens that report notifications park with no timer. Polling screens keep the 125 ms keep-alive tick.
- `Update` skips `shell.Update` and `Draw` on spurious event-wait wakes. Pixel cache trim waits for 5 s of clean idle. DevTools open no longer forces dirty.
- `NoIdlePace` keeps its meaning. Uncapped TPS now forces `SyncWithFPS`.
- Perf sampler uses a 2 s idle gap against 500 ms active, and adds heap sys, GC count, alloc per second, and repaints per second. `presentFPS` backfills one sample.

### Debug server and devtools

- `WindowOptions.DebugAddr` serves `GET /debug/state` and `GET /debug/pprof/*` only with `-tags ownframe_debug`. The stub build binds nothing. The live server allows loopback only, checks Host, caps headers, and strips CR and LF from lines.
- DevTools dock caches to a texture keyed on geometry, tab, scroll, selection, cursor, and pause state. The Frame tab refreshes metrics at 1 Hz with pause and resume.
- Shared op-kind names let one sampler read web and native snapshots.

### Web preview

- New `internal/web/viewport.go` tracks server-side `scrollY`, clamps it, and notifies `host.ScrollObserver`. Clicks offset by `scrollY`. `/frame.png` renders the viewport slice with a PNG cache keyed on generation and size.
- Shell polls at 50 ms, sends throttled `POST /hover`, and clears hover on mouse leave.

### Page tick and hover

- `Page` implements `host.FrameScheduler` with a coalesced buffered channel. Dirty, scroll, patch, devtools, and perf paths notify the host. Tick work sets a guard so it does not wake itself.
- `NextFrameAt` merges the explicit deadline with the caret blink deadline. Selection and file inputs suppress the caret deadline.
- Hover patch exits early when `Hover` is unchanged, prefers cascade-reported scopes with a geometric fit check, caches roots per render key minus Hover, and caps the cache at 256. `PatchElement` clears it.

### Examples, Ebiten, and tooling

- Ebiten `v2.10.4` to `v2.11.0-alpha.0.20261103125056-58976b602532` in both modules, plus blinkless `1b88147`, `x/image v0.47.0`, `oto/v3 v3.5.1`, and purego update.
- Eleven examples gain `-debug-addr` and help text that notes sampling feeds the debug address. READMEs updated for animations, jmap-mail, ui-builder, code-view, live-log-viewer, spreadsheet, tetris, and trade.
- `make debug` runs LogLens with `-dev -perf -debug-addr 127.0.0.1:8129`. New `scripts/profile-idle.py` snapshots `/debug/state` with CPU and RSS. `perf-sample.py` adds heap, GC, peak RSS, and alloc rate columns.

### Frontend and site

- Homepage starter picker gains Builder basics backed by `examples/frontend/builder`. `build:demo` emits `demos/builder/demo.wasm`.
- New `frontend/public/demos/builder/index.html` loader. Generated `docs/demos/builder/` output ships with the branch.
- `check-site.mjs` pins three picker labels and four builder tabs, and requires the builder WASM to load.

### Docs and plans

- New `plans/v0.0.3/components/chadcn/README.md` records phases 00 through 06 as done and leaves interactive open-state gallery coverage plus `make graphics-check` open.
- `AGENTS.md` and `README.md` point at the new set and the gallery. `.gitignore` un-ignores the builder docs WASM and ignores local AI reference shots and the `loglens-rust` checkout.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Static event-driven windows park with no timer wakes. Hover patches only the changed scopes. Active ticking pages keep their prior cost. |
| **Memory** | Pixel caches stay for 5 s after parking, so wake is fast and resident stays higher briefly. Perf sampler adds heap and GC counters. |
| **Behavior / correctness** | `-perf` works in every example. Debug profiling is off unless the tag and a loopback address are set. Builder demos moved to `examples/builder`. |
| **API / CLI** | Adds `components/chadcn` constructors. No `ownframe.Page` method added or removed. CLI adds `-debug-addr` to eleven examples. |
| **Dependencies** | Ebiten November alpha, blinkless `1b88147`, image, oto, purego bumps in both modules. |
| **Binary size / build time** | Six WASM artifacts rebuilt. Library delta is idle, page, web, and window files plus tests. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `ui/components/` deleted | Import `components/chadcn` instead. Scrollbar and control tests moved with it. |
| `go.mod` bumps | Sync both modules. Blinkless pin must stay at `1b88147` with this Ebiten. |
| `perf-sample.py` CSV columns | Parsers must accept `heap_sys`, `total_alloc`, `num_gc`, `peak_rss_bytes`, and `allocation_bytes_per_second`. |

---

## Library freeze (until Page wrap)

- No new `type X =` alias at the module root. `grep ^+type` on the diff is empty.
- No new `func (p *Page)` that exists to grow `ownframe.Page`. New `FrameEvents` and `FramePolling` feed the window scheduler through `internal/host`.
- No new example `ownframe/internal/` imports. `make internal-import-check` is clean with 16 allowlisted pairs.
- No synonym names. `DrawerOptions` follows the existing options convention.
- No new exported method returns `host.Stats` or `*state.Store`.
- New failure paths return error with `context.Context` first. The only new `panic` is embedded stylesheet load, which cannot fail at runtime.
- No new Telegram, Teams, Spotify, or Wispr clone. Spotify and Wispr diffs are flag plumbing only.

---

## Test plan

- [x] `make site-check` passed on `3f88c3d8` (npm test, generated-site match, `check:site` PASS with 140 layout checks)
- [x] `make internal-import-check` clean (16 allowlisted pairs)
- [ ] `make test` was not run as its own target in this pass
- [ ] `make lint` / `go vet` was not run as its own target in this pass
- [ ] `make test-race` was not run (shared mutable state changed in idle and page paths, so it should run before merge)
- [ ] `make build` was not run as its own target in this pass

### Commands

```sh
make site-check
make internal-import-check
bash scripts/pr-diff-stat.sh master
```

---

## Screenshots / sample output

Second `make site-check` after the rebuild:

```text
site-check: generated output matches the source
PASS: 140 layout checks, themes, mobile menu, anchors, showcase, keyboard focus, and images.
```

First run rebuilt hashed JS and WASM output, which this branch commits as `chore(site)`.

---

## Related issues

- None. No issue is closed or linked by this PR.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [x] Related issues filled: there is no tracker issue
- [x] Filled body committed under `plans/v0.0.3/pr/66/pr-chadcn-loglens-dashboard.md`

---

## Follow-ups (out of scope)

- Run `make test`, `make lint`, `make build`, and `make test-race` before merge. Idle and page paths share mutable state.
- Measure parked CPU and wake latency on a static page with the debug sampler.
- Update CSV consumers for the new `perf-sample.py` columns.
- Watch duplication between `loglens` and `ui-builder` constructors as the set grows.

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
| `.css` | 71 | 5124 | 0 |
| `.go` | 249 | 18797 | 620 |
| `.gz` | 1 | Binary | Binary |
| `.html` | 4 | 229 | 1 |
| `.js` | 4 | 1051 | 63 |
| `.json` | 2 | 28 | 2 |
| `.jsx` | 1 | 5 | 4 |
| `.md` | 30 | 1298 | 274 |
| `.mjs` | 2 | 7 | 2 |
| `.mod` | 2 | 11 | 11 |
| `.py` | 4 | 89 | 5 |
| `.sh` | 1 | 4 | 1 |
| `.sum` | 3 | 24 | 1 |
| `.wasm` | 6 | Binary | Binary |
| No extension | 2 | 12 | 0 |
| **Total** | **382** | **26679** | **984** |
