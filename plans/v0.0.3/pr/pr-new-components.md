# feat(components): add Search, dashboard primitives, MailPulse/Inventory builders and compare harness

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `origin/master` (`f88fa52c`)
- Head: `feat/new-components` (`cea71f84`)
- Diff against `origin/master`: 149 files before this plan file

---

## Summary

This branch adds a `Search` constructor plus dashboard primitives (`AreaChart`, `FilterBar`, `HexGrid`, `StatCard`, `StatusSegments`) to `components/chadcn` and removes background shadows across the set. It adds two builder demos (`builder/mailpulse` 15-tab engagement dashboard, `builder/inventory` Northwind heatmap dashboard) and makes the `builder/ui-builder` catalog fully interactive with live filtering and row windowing. It also adds a naive 1000-row `examples/compare` harness (ownframe vs mygo) plus `examples/mygo-inventory`, and lands generic idle/hover/replay fixes (emoji gate, form-scan skip, viewport idle, windowing, hover+tick batching).

---

## Motivation / context

- Plans: `plans/v0.0.3/components/chadcn/README.md` (first `components/<set>`, `ui` stays experimental).
- Builder demos share one folder under `examples/builder/`; MailPulse and Inventory exercise the new ChadCN primitives end to end.
- Compare harness reframes `examples/compare` as apples-to-apples naive full-rebuild so default-path cost is measured on both toolkits; the earlier windowed/virtualized best-practice round lives in session notes per `examples/compare/README.md`.
- Core fixes unblock the demos at scale: quiet ticks parking on the viewport path, bounded windowing prewarm, hover cursor caching, and form/emoji fast paths.
- Issues: none. This work has no tracker issue.

---

## Changes

### ChadCN component set

- New `Search(options SearchOptions)` in `components/chadcn/search.go` + `css/search.css`: controlled `Open + Query`, host owns state via `OnOpenChange/OnQuery/OnSelect`. Closed renders field only; open renders transparent fixed backdrop (click-away closes) + absolute `role=listbox` panel that never moves siblings. Live case-insensitive filtering over `Label+Value+Description` via `searchMatches`, `FilterSearchItems()` helper for remote lists, `aria-selected` + `✓` on selected row, `No results` / `No results for "q"` empty states. Explicit `box-shadow:none`.
- New dashboard primitives:
  - `AreaChart` (`area_chart.go` + `css/area-chart.css`): div-geometry two-series trend, header with legend dots (`data-series` / shared `chartSeriesSlot`), body columns with per-series dots positioned by `chartHeight(value/max)`, hatched gap band, axis labels, `aria-label`.
  - `FilterBar` (`filter_bar.go` + `css/filter-bar.css`): static legend (`FilterBarLegend{Label,Tone}`) + pill group (`FilterBarOption{Label,Active,Action}`) with `aria-pressed` buttons and `data-action`.
  - `HexGrid` (`hex_grid.go` + `css/hex-grid.css`): honeycomb status map without canvas/SVG/clip-path; 3 divs per cell (`of-hex-top/mid/bottom`), 1-based tone via `chartSeriesSlot`, `--of-hex-columns` clamped 1–24 (default 12).
  - `StatCard` (`stat_card.go` + `css/stat-card.css`): static KPI `{Label,Value,Delta,Tone}`, delta pill only when `Delta != ""`, `of-stat-card-delta--*` tones.
  - `StatusSegments` (`status_segments.go` + `css/status-segments.css`): proportional bars + count legend via `FlexGrow(Value)`, optional `Overdue` header.
- `SidebarSection(label, items...)` (`sidebar.go` + `css/sidebar.css`): groups `SidebarNavItem` rows under `of-sidebar-section-label`, hidden when collapsed.
- `Trace` (`trace.go` + `css/trace.css`): bar layout from `width% + margin-left%` block to flex `gap + bar` with `FlexGrow(start/width/rest)` proportions.
- Shadow removal (`aa69a455`): deletes `box-shadow` from ~19 existing CSS files (alert-dialog, calendar, card, combobox field+popover, command, context-menu, date-picker field+panel, dialog, drawer, dropdown-menu, hover-card, input, menubar, native-select, navigation-menu, popover, select field+popover, sheet, tabs, textarea, toast). New `search.css` is shadowless by design.
- `README.md`: rows for `Search`, `AreaChart`, `HexGrid`, `StatusSegments`, `StatCard`, `FilterBar`; `Sidebar` row lists `SidebarSection`.

### Builder MailPulse (new)

- `examples/builder/mailpulse`: light `#f6f6f5` email-engagement dashboard, 15 canonical tabs (`overview, audience, campaigns, content, engagement, revenue, subscribers, automations, segments, integrations, analytics, reports, learn, help, settings`) via `tabs.go:contentBody()`, router in `data_nav.go` normalizing `nav:`/`nav-`/`tab:`.
- Shell: custom chrome with draggable topbar, minimize/maximize/close, DevTools toggle with dot, fixed sidebar (`left 0`, `z-index 50`, 248px margin), fixed topbar (`left 248px`, `z-index 40`, 71px content offset), working search with dropdown + Enter-to-go, notification bell with dismissible seeded items, 5s auto-expiring toast, profile card.
- Engagement cards: KPI tiles, send-time heatmap with hover readout (no select), funnel, top links, per-tab dummy pages with range-window helpers backing range-aware KPI labels.
- Motion: `escort.go` two tick escorts — `escortEnter` mounts growth chart at `scaleY(0)`, waits one presented zero frame, flips to real values over 3s slow transition then strips slow class after 3.2s; `escortMotion` keeps loop hot 1.5s after range switch; growth bars use `scaleY` transform path, `mp-enter` 0.6s rise, dot share bars with stable IDs (`mp-eng-dot-N`, `mp-src-dot-N`).
- Tests: `tabs_test.go` 15 tests (all-tabs render, nav/select, notif dismiss, hover CSS, export/range every tab, heat/growth labels, subscriber search, heat-hover-only, range updates, hands-free range switch, entrance-from-zero, stable dot IDs, entrance-replay-on-revisit).

### Builder Inventory (new)

- `examples/builder/inventory`: Northwind Goods inventory-health dashboard, fixed 1440×960 `LockView` canvas (min 900×600), static header + inner body-viewport wheel scroll, top-60px drag strip with overlay drag guard.
- Content: 5-KPI strip, stock-health table with status filter pills, bin-utilization heat grid (red→green cubic capacity, A–J aisles), inbound PO timeline, value-by-category bar, search palette (backdrop dim, finite scroll window, hint, `OnSelect` picked SKU), New-PO dialog, alerts popover pinned to 64px slot (no topbar shove), account menu, export toast.
- Fixed chrome: alerts popover slot + fixed topbar + shell offset keep overlay surfaces hit-testing above body via pinned layer (`pinned_layer_test.go:TestPinnedChromeLayer`); 28px body side padding matching loglens; heat cells keep hover readout without title tooltips.
- Tests: 25 tests across `interactions_test.go` (14, incl. search backdrop, picked line, panel-on-screen, finite cap, scroll window, popover-still, click-away), `bins_scale_test.go` (5), `fit_viewport_test.go` (2), `scroll_wheel_test.go` (3), `pinned_layer_test.go` (1).

### Ui-builder catalog (modified)

- `examples/builder/ui-builder/screen_catalog.go` (+~700 lines, 78 cards): every card gets `data-action` tap target reporting to notice text; Command/Combobox/Search filter live via `handleChange` with focus preservation.
- New `catalog_actions.go`: `handleCatalog()` for `catalog:` actions (filter pills, nav rows, check/switch/choice/toggle/hover/tooltip/context/sidebar/toast).
- New `catalog_window.go`: row-window slicing on measured average pitch + 2-row overscan + top/bottom spacers; pinned `catalogGeo*` from full renders; `catalogWindowPlan`.
- `app.go`: ~40 new `catalog*` state fields, `RetainState:true`, `Change` handler (catalog-command/combobox/search + widgets-search with `FilterSearchItems`), `refreshLogged` clears pointer focus, `refreshBlur`; `scroll.go` center-scroll keeps full height with spacers; `screen_widgets.go` Search card; `main.go` `MaxFPS=60`.
- Tests: extended `catalog_test.go` (every entry renders, sidebar open, pill/nav notice, wheel defers rebuild), new `catalog_window_test.go` (range/spacer/entry-mapping/windowed-slice/plumbing), new `search_test.go` (widgets search open/filter/close, catalog live filter).

### Compare harness + mygo-inventory

- `examples/compare/`: same 1000-row styled table (header, dark `Num/Title/Description` header, alternating white/gray rows, blue `#dbeafe` selection + header readout) in both toolkits, naive form on both sides: render all rows, rebuild whole view per click.
  - `ownframe/` (in-module): `ui.Mount` + `View.Update`, `listRows=1000`, `ID=compare-row-N` + `data-action=row:N`, no `SetWindowing`/`SetScrollWindow`/tick bridge/retained state/spacers; `main.go` `-bench/-bench-file` measures `redraw-same-data` (200× `Redraw`), `rebuild-same-data` (100× `refresh`), `click-select-rebuild` (30× `pagetest.ClickAction row:500`) with `MemStats` deltas.
  - `mygo/` (isolated module, MyGo `v0.7.4` needs Go `>=1.27.3` vs repo `1.26.10` + `GOTOOLCHAIN=local`, `GOWORK=off`, out of `make build/test/lint`): `Scroll` + 1000 `Row`s, `-bench` with `build-full-view` + `click-select-reframe`.
  - Tests renamed `TestListShowsWindow` → `TestListShowsRows`, assert `Item 1000` / `compare-row-500`, drop window/spacer assertions.
- `examples/mygo-inventory/` (isolated module, same Go-version reason): native MyGo port of the Northwind dashboard (topbar, period/category filters, `+New PO` modal, 5 KPIs, stock table, bin grid A–J, inbound POs, category value, 8×10 product heatmap via `ui/chart` with 250ms transition, `heat.go` click-select).
- `examples/maintenance.json`: registers `compare` and `mygo-inventory` as experimental.

### Core library fixes

- Emoji ASCII fast path (`internal/emoji/emoji_ascii.go`, `emoji_replace.go`): `hasNonASCII` → `maybeHasEmoji`, checks only lead pairs `E2+{9A,9C,9D,AD}` and `F0 9F`; `TestReplaceGateSkipsNonEmoji` covers `😂` color path and zero-alloc skip for `café/·/→/—`.
- Form scan/wire skip (`internal/page/form/scan.go`, `form_wire.go`): extracted `HasControls(source)`; `syncForm` skips button rewrite/scan/merge/binds when cached live set empty + no controls, still handles focus clear/apply/clamp, `lastSyncBody`/`formDirty`, `emojiPass`; `TestHasControls` + `form_sync_skip_test.go`.
- Gesture hover (`internal/page/gesture_wire.go`): `Hover` uses innermost-first `hoverIDAt` stopping at first id-bearing box; `CursorShape` short-circuits when gesture already non-default.
- Window hover cursor (`internal/window/hover_cursor.go`): `hoverState` caches `cx,cy/cgen/cid/cok` + `pushed`; skips `applyCursor` box+link scan when point+generation+hover id unchanged.
- Idle probe (`internal/window/idle_probe.go`): viewport path requires `frameMarker` when ticking; clean iff `blitOnly && !frameOptIn` (ticking) or `blitOnly` (non-ticking); quiet tick parks, `FrameDirty`/dirty-rect stays dirty; `125b128b` drops compare tick workaround to permanent tick. New `idle_viewport_tick_test.go`.
- Replay windowing (`replay_oversize.go`, `replay_partial.go`, `replay_viewport_prewarm.go`): new `windowing()` opt-in via `host.ScrollObserver.Windowing()`; partial routes windowing to viewport path even when tile budget fits; prewarm sweeps `viewport+overscan` instead of full spacer/12MP; new `replay_windowing_test.go` (viewport-sized buffer, opt-out keeps tiles, full spacer scroll range, bounded warmup follows scroll).
- Update batching (`internal/window/update.go`, `hover_cursor.go` in `93316249`): `updateStage.batch` (pointer/touch/hold/drop/tick=true); wraps batched span in `host.RedrawBatcher Begin/EndRedrawBatch`, single flush before `syncImage` fuses hover+tick into one layout per frame.
- Touch-ups: `internal/atspi/live.go` 1-line, `page_set_data_cache_test.go` 1-line, `replay_budget_partial_test.go` expectation update, `replay_viewport_prewarm_test.go` call-site update, `replay_oversize.go` +12.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Quiet viewport ticks park (no wake); hover+tick fused to one layout per frame; emoji/form fast paths skip heavy passes on text/model pages; windowing prewarm bounded to viewport+overscan. No measured regressions claimed here. |
| **Memory** | Windowing keeps viewport-sized buffer instead of full-page tiles when opted in; prewarm no longer sweeps 12MP neighbourhood. Pixel-cache/heap behavior otherwise unchanged. |
| **Behavior / correctness** | Catalog/search/command/combobox filter live; inventory/mailpulse overlays hit-test above body; trace bars use flex proportions; ~19 component CSS files lose `box-shadow` (visual flattening); compare measures naive full-rebuild on both sides. |
| **API / CLI** | Adds `chadcn.Search`, `AreaChart`, `FilterBar`, `HexGrid`, `StatCard`, `StatusSegments`, `SidebarSection`. No `ownframe.Page` method added or removed. CLI: ui-builder `MaxFPS=60`; mailpulse `-headless/-addr 127.0.0.1:8148`; inventory `-addr 127.0.0.1:8135`; compare `-bench/-bench-file`. |
| **Dependencies** | `mygo`/`mygo-inventory` isolated modules on MyGo `v0.7.4` (out of `make` gates until repo Go moves). No library `go.mod` pin change in this diff. |
| **Binary size / build time** | Three new runnable examples (mailpulse, inventory, compare/ownframe) plus isolated mygo modules. No WASM/docs rebuild in this diff. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None (no Go API removal) | Fully backwards-compatible at the Go API level. |
| Component CSS loses `box-shadow` (~19 files) | Visual flattening only; re-add local shadow if a product needs elevation. |
| `Trace` bar geometry moved to flex proportions | Snapshot/golden expectations on bar boxes need refresh; semantics unchanged. |
| `examples/maintenance.json` gains `compare`, `mygo-inventory` experimental roots | Tooling that enumerates experimental roots sees two more entries. |

---

## Library freeze (until Page wrap)

Until `plans/v0.0.3/prod-ready-task/` wraps `Page`, reject the PR if it does any of the following.

- [x] No new `type X =` at the module root except documented re-exports (`grep ^+type` on the diff is empty)
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows (only unexported `hoverIDAt`; no exported `func (p *Page) [A-Z]` added)
- [x] No new example `ownframe/internal/` imports (`make internal-import-check` pair set unchanged; `git diff` shows no new `ownframe/internal/` import in `examples/`)
- [x] No synonym names (`PreviousX` for `X`)
- [x] New exported methods return named types in package `ownframe` or a public subpackage, never `host.Stats` or `*state.Store` unless those types are re-exported in the same PR (new constructors return `*ui.Element`)
- [x] Library code panics only for invariants; new user-facing failure paths return `error`
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline; no hidden `time.After` that ignores the caller's ctx
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr (mailpulse/inventory/compare are engagement/inventory/table demos)

---

## Test plan

- [x] `go test ./components/chadcn/...` passes (root module)
- [x] `go test ./builder/inventory/... ./builder/mailpulse/... ./builder/ui-builder/...` passes (examples module)
- [x] `go test ./compare/ownframe/...` passes (examples module)
- [x] `go test ./internal/emoji/... ./internal/page/form/...` passes (root module)
- [x] `go test ./internal/window/ -run "TestPartialBudgetViewportIdles|TestWindowing|TestViewport"` passes (root module)
- [x] `go vet -p 1 ./components/chadcn/... ./internal/emoji/... ./internal/window/...` passes
- [ ] `make test` (full both-module run) was not run as its own target in this pass
- [ ] `make lint` / full `go vet -p 1 ./... ./examples/...` was not run as its own target in this pass
- [ ] `make test-race` was not run (shared mutable state changed in idle/window paths, should run before merge)
- [ ] `make build` was not run as its own target in this pass

### Commands

```sh
go test ./components/chadcn/...
go test ./internal/emoji/... ./internal/page/form/...
go test ./internal/window/ -run "TestPartialBudgetViewportIdles|TestWindowing|TestViewport" -count=1
cd examples && go test ./builder/inventory/... ./builder/mailpulse/... ./builder/ui-builder/... ./compare/ownframe/...
bash scripts/pr-diff-stat.sh master
```

---

## Screenshots / sample output

```text
ok  github.com/chinmay-sawant/ownframe/components/chadcn  0.339s
ok  github.com/chinmay-sawant/ownframe/examples/builder/inventory  2.780s
ok  github.com/chinmay-sawant/ownframe/examples/builder/mailpulse  1.794s
ok  github.com/chinmay-sawant/ownframe/examples/builder/ui-builder  5.799s
ok  github.com/chinmay-sawant/ownframe/examples/compare/ownframe  (cached)
ok  github.com/chinmay-sawant/ownframe/internal/emoji  0.004s
ok  github.com/chinmay-sawant/ownframe/internal/page/form  0.021s
ok  github.com/chinmay-sawant/ownframe/internal/window  0.021s
```

---

## Related issues

- None. No issue is closed or linked by this PR.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [ ] Related issues filled: there is no tracker issue
- [x] Filled body committed under `plans/v0.0.3/pr/pr-new-components.md`

---

## Follow-ups (out of scope)

- Run `make test`, `make lint`, `make build`, and `make test-race` before merge. Idle, window, and replay paths share mutable state.
- Refresh MailPulse `README.md` screenshot section (still describes single engagement card, not 15 tabs).
- Decide whether `mygo/` and `mygo-inventory/` rejoin `make` gates once repo Go moves past `1.27.3`.
- Watch duplication between `builder/inventory` and `mygo-inventory` fixtures as both evolve.
- Measure parked CPU/wake latency on a static page with the new viewport idle rule.

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
| `.css` | 29 | 614 | 41 |
| `.go` | 109 | 13117 | 260 |
| `.json` | 1 | 2 | 0 |
| `.md` | 6 | 296 | 1 |
| `.mod` | 2 | 26 | 0 |
| `.sum` | 2 | 28 | 0 |
| **Total** | **149** | **14083** | **302** |
