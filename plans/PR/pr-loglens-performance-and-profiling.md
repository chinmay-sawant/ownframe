# fix(loglens): scope hover updates and release retained render state

## Summary

Loglens hover and control updates were doing full-page work, and replaced displays, fallback lists, and replay plans could stay alive through caches. This PR uses the cascade's precise hover scopes, makes the geometric fallback conservative when escaping selectors exist, releases displaced render state, and adds reproducible profiling and capture tooling. It depends on blinkless PR #36 for the engine lifetime fixes and pins the released blinkless `v0.1.1` tag in both modules.

---

## Motivation / context

- Base: `origin/master` at `442d164a`. The branch adds six commits on `fix/loglens-performance-and-profiling`.
- Plans: no plan file is dedicated to this branch. The work follows the Loglens dashboard plan in `plans/v0.0.3/pr/66/pr-chadcn-loglens-dashboard.md` and the patch-path goals in `plans/v0.0.3/performance/phase-07-rebuild-and-patch-paths.md`.
- Issues: none. No open or closed issue tracks this work. Prior context: #66 added the Loglens dashboard.
- The trigger was visible in Loglens: a button hover could take the full-layout path because an unrelated `:has()` or sibling-after-`:hover` rule existed somewhere in the stylesheet. Control toggles also rebuilt the dashboard instead of patching the control that changed.
- The navigation lifetime test fails against blinkless `d805508` and passes against blinkless PR #36, released as `v0.1.1`. Both module pins use that tag, and `go.work` stays limited to this repository's two modules.

---

## Changes

### Hover patch selection (`internal/page`, `internal/page/draw`, `internal/render`)

- Prefer `HoverChangedScopes` from the cascade: it evaluates the full selector per node, including `:has` and sibling combinators, so the scoped patch is exact when those roots fit.
- Keep the geometric fallback page-wide conservative. When the cascade scopes do not fit, any `:has()` or sibling-after-`:hover` selector in the page or theme forces a full layout. The geometric path cannot prove which element a risky selector belongs to, so this prevents a narrower patch from painting differently from a full layout.
- Classify functional pseudos before `:hover` (`:not()`, `:is()`, `:where()`) as global, so the last class inside the argument is never mistaken for the hovered element. Treat `:has(+ ...)` and `:has(~ ...)` as global.
- Make the selector scan skip comments, quoted strings, and attribute selectors. A `+`, `~`, or `}` inside them can no longer end the scan early or be mistaken for a combinator.
- Convert element-relayout panics in `render.RelayoutElement` and `render.RelayoutElements` into errors, so callers take the existing fallback path instead of crashing the host.
- Cover the functional-pseudo and id-less `:has` cases with regression tests, plus the existing sibling, scoped `:has`, hidden-popup, and refused-patch parity tests.

### Display and cache lifetimes (`internal/page`, `internal/replay`, `internal/window`)

- Release the previous display after a normal commit, bitmap fallback, restamp, SVG paint adoption, single-element patch, batched patch, layout adoption, and each intermediate hover patch.
- Release the fallback display that `CommitList` and `CommitBitmap` displace. A bitmap page kept its previous `FallbackDisplay` alive because `p.draw.Display` is nil on that path.
- Reset the motion lookup index when the page cache is invalidated, so a new document cannot resolve against stale entries.
- Evict replay group and chunk plans when the window installs a different display or switches to a bitmap fallback.
- Touch raster-cache entries on a hit so eviction follows least-recently-used order instead of insertion order.
- Cover replaced public displays and fallback displays with release tests, plus a navigation test that warms three cycles, runs twelve more, and requires obsolete displays to be collectible and post-GC growth to stay within 8 MiB.

### Diagnostics and profiling

- Add cumulative `FullLayouts`, `HoverPatches`, `HoverFallbacks`, and `HoverPatchErrors` counters to the inspector snapshot. `HoverPatchErrors` includes retries.
- Add `HoverPatchFailures`, grouped by stage, stable reason, and target ID. Stages are `selector`, `geometry`, and `selection`; reasons include `target-missing`, `empty-display`, and `escaping-hover-selector`. The map is capped at 256 keys including `<other>`, and a regression test fills it past the cap.
- Add `scroll_x` and `scroll_y` to the window debug state.
- Extend the runtime sampler with `heap_inuse`, `heap_idle`, `heap_released`, `sys`, `stack_sys`, and `gc_pause_total_ns`.
- Add `scripts/windows-loglens-dump.ps1`. It launches a fresh Loglens process per repetition, prompts the operator for idle, hover, scroll, resize, navigation, and idle-after phases, verifies the content viewport, records a SHA-256 and launch arguments, samples process and GPU memory counters every 500 ms, and saves CPU, allocation, and heap profiles around each phase. It does not inject mouse or keyboard input.
- Add `scripts/loglens-report.py`, which derives rates from counter differences and the actual sample clocks, reports medians across repetitions, and rejects failed captures, counter resets, non-finite values, mixed binaries or settings, and phases that show no evidence of the requested work. A present but null `HoverPatchFailures` list is treated as empty, and GPU counters keep both the counter path and the adapter instance.
- Add `scripts/loglens-phase-allocs.sh` for separate hover and resize allocation diffs from the before/after profiles.
- Add `documentation/loglens-profiling.md` and link it from `documentation/performance.md`. The guide records the measurement limits: profiler overhead, rolling p99 versus input latency, sampled redraw allocation versus per-present allocation, and Go-managed memory versus Windows working set and private bytes.
- Add report tests for counter deltas, independent runtime clocks, resets and non-finite values, zero-present idle periods, missing action evidence, failed captures, null hover-failure snapshots, hover failure deltas, GPU counter summaries, and per-adapter GPU counters.

### Loglens interactions and reusable controls

- Batch tooltip and hover-card edits, and patch toggle, switch, checkbox, and slider values instead of rebuilding the dashboard.
- Give avatars an optional stable ID and derived image and fallback IDs, so the on-call identity updates in place.
- Add outside-month calendar selection as an explicit `AllowOutsideSelection` option, pass it through the date picker, and enable adjacent-day selection in Loglens.
- Track on-call coverage by date and add a responder reassignment dialog.
- Lay out edge-region rows independently so hovering a second row cannot shift the first.
- Give chart hover tracks their full interactive height, and adjust dashboard overlays, calendar styling, and control state tests.
- Keep the closed tooltip label stable. A hidden tooltip that aged every second could be included in a patch batch with no box, causing the batch to be refused and the page to reparse. The tick still updates the label while the tooltip is open.
- Patch the threshold description when the nudge buttons move the slider.
- Pin the calendar tests to November 2026 so they do not depend on the current month.
- Add Loglens tests for scoped button hovers, sequential calendar clicks, responder swapping, region cell width, and fine-grained state updates that must not reparse or recascade.

### Dependency and CI

- Advance `go.mod` and `examples/go.mod` to `github.com/chinmay-sawant/blinkless v0.1.1`, the release containing PR #36. Both `go.sum` files move with it.
- Remove `../blinkless` from `go.work`. The committed workspace lists only `.` and `./examples`, so a clean checkout and CI resolve the same engine from the module proxy. The workspace removal is in the first commit, so the performance job's `HEAD^` baseline also has a valid workspace; the final module pin is the released tag.
- Make `scripts/check-graphics.sh` tolerate non-JSON lines in the `go test -json` stream. `xvfb-run` redirects the command's stderr into stdout, so a fresh runner's `go: downloading ...` line used to fail the gate before the first test ran. Non-JSON lines now go to stderr, and the named test still must emit a pass event.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Hover-only changes use the cascade's changed scopes instead of a full layout when those scopes fit. The geometric fallback is more conservative when escaping selectors exist, which can cost a full layout on pages that use them. Batched tooltip and control edits avoid the previous full dashboard rebuild. The measured local Linux resize workload's median allocation over three runs fell from 71,715,266 to 17,974,044 B/op, about 74.9%, with the blinkless PR #36 engine. This is the complete benchmark workload, including its initial no-op iteration, not a universal interactive frame cost or Windows FPS claim. |
| **Memory** | Superseded displays, fallback lists, replay plans, and raster entries are released or evicted instead of being retained by commit history and caches. The hover failure map is bounded at 256 keys. Navigation tests require obsolete displays to be collectible and post-GC growth to stay within 8 MiB. No hardcoded Go memory limit or idle `debug.FreeOSMemory` call is added. Native Windows and graphics memory were not measured. |
| **Behavior / correctness** | Hover output stays equal to full layout for sibling selectors, scoped `:has()`, functional pseudos, id-less `:has` ancestors, hidden popups, and refused patches. Global `:has()` rules still force full layout. Relayout panics surface as errors and take the fallback path. Loglens gains date-based on-call coverage, a responder dialog, adjacent-month calendar selection, stable avatar IDs, independent region rows, a stable closed-tooltip label, and a patched threshold description. |
| **API / CLI** | Additive component options: `chadcn.AvatarOptions.ID`, `chadcn.CalendarOptions.AllowOutsideSelection`, and `chadcn.DatePickerOptions.AllowOutsideSelection`. No new page API or Go CLI flags. The new capture script takes its own parameters (`-Repeats`, `-SecondsPerPhase`, `-TopologyLabel`, `-DevTools`), and the report script takes `--out`. `host.Stats` gains fields for the devtools inspector, and `/debug/state` gains `scroll_x` and `scroll_y`. |
| **Dependencies** | Blinkless moves to `v0.1.1` in both modules. No new third-party Go modules. The Python report uses the standard library only. The branch no longer requires a sibling blinkless checkout. |
| **Binary size / build time** | Not measured. Apart from the Blinkless pin, the change adds Go code and tests but no new third-party module. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| Public library API | None. The new component options are additive and existing call sites keep their behavior. |
| Hover patch selection | A page with a scoped `:has()` or sibling-after-`:hover` rule can still take a scoped patch when the cascade scopes fit. When they do not fit, any escaping selector in the sheets now forces a full layout. No caller change. |
| Diagnostics schema | `host.Stats` and `/debug/state` gain fields. JSON decoders that ignore unknown keys need no change; consumers with a closed schema must add the fields. |
| Dependency revision | Both modules require blinkless `v0.1.1`, the release containing PR #36. No migration for normal builds; advance the pin together with the engine. |
| Windows capture workflow | The capture script prompts for manual input and does not send synthetic mouse or keyboard events. An automated pipeline that expected input injection must drive the app itself. |

---

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports.
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows.
- [x] No new example `ownframe/internal/` imports. `make lint` runs `make internal-import-check` and passes.
- [x] No synonym names.
- [x] New exported methods return named types in package `ownframe` or a public subpackage. This PR adds no exported methods; the new options are fields on existing public types.
- [x] Library code panics only for invariants. New relayout failure paths return errors, and the hover patch consumes them through the fallback path.
- [x] `context.Context` is the first parameter when the function already does I/O, cancel, or deadline work. No new exported I/O API is added.
- [x] No new examples whose primary purpose is resembling Telegram, Teams, Spotify, or Wispr.

---

## Test plan

- [x] `make test TEST_P=2` (library and examples)
- [x] `make lint` (both Go modules and frontend ESLint; this also runs the documentation, plan, and internal-import checks)
- [x] `make test-race` (library module)
- [x] `make build`
- [x] `make graphics-check`
- [x] `make docs-check plans-check examples-check`
- [x] `python3 -m unittest discover -s scripts -p test_loglens_report.py` (9 tests)
- [x] Functional-pseudo and id-less `:has` hover regression tests
- [x] Replaced fallback display release test
- [x] Bounded hover failure map test
- [x] Loglens race test `TestFineGrainedStateUpdatesUseElementPatching` with `-count=3`
- [x] Windows PowerShell parser validation for `scripts/windows-loglens-dump.ps1`
- [x] Targeted Windows debug build of Loglens
- [ ] Live Windows capture with `scripts/windows-loglens-dump.ps1` (blocked by antivirus; see the limits below)
- [ ] `make run` wall time versus baseline (this repository has no `make run` target, and no wall-time result is claimed)
- [ ] `make reference-metrics` (this repository has no such target)

### Commands

```sh
make test TEST_P=2
make lint
make test-race
make build
make graphics-check
make docs-check plans-check examples-check
python3 -m unittest discover -s scripts -p test_loglens_report.py

go test -count=1 ./internal/page -run 'TestHoverPatchFunctionalPseudoScopeFallsBack|TestHoverPatchIDlessHasScopeFallsBack|TestFallbackDisplayReleaseOnCommit'
go test -count=1 ./internal/page/draw -run 'TestHoverPatchRisky'
cd examples && go test -race -count=3 ./builder/loglens -run '^TestFineGrainedStateUpdatesUseElementPatching$'
```

---

## Screenshots / sample output

Local Linux resize workload, median allocation over three runs with blinkless `v0.1.1`:

```text
before  71,715,266 B/op
after   17,974,044 B/op
change  -53,741,222 B/op (-74.9%)
```

This is the complete benchmark workload, including its initial no-op iteration. It is not a universal interactive frame cost or a Windows FPS claim. Navigation checks also prove obsolete displays collect after repeated screen changes.

Windows capture limits reported during implementation:

```text
The new Windows capture script passed parsing, but live execution was blocked
by antivirus. Different window sizes and workloads prevent a causal
single-monitor versus dual-monitor memory conclusion. Native Windows and
graphics memory remain a separate profiling priority. No hardcoded Go memory
limit or idle debug.FreeOSMemory call is added.
```

---

## Related issues

- None. No tracked issue covers this work.
- Prior work: #66 added the Loglens dashboard.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`bug`, `documentation`, `benchmark`)
- [x] Related issues checked: none matches this branch
- [x] Filled body committed under `plans/PR/pr-loglens-performance-and-profiling.md`

---

## Follow-ups (out of scope)

- Keep both module requirements on released blinkless versions and do not reintroduce the local sibling workspace.
- Run the live Windows capture after the antivirus block is resolved, including GPU memory counters. Compare single-monitor and dual-monitor runs only with matched window size, viewport, and workload.
- Use the `HoverPatchFailures` counters to find remaining full-layout hovers in Loglens and decide whether they are expected fallbacks or new patch gaps.
- Add native Windows graphics memory attribution when a tool is available. The current GPU counters report process and adapter totals, not individual D3D objects or driver allocations.
- The `html/attribute.go` id and dirty registries in blinkless still key on the document and remain a separate cache-lifetime item.
- Fix the pre-existing `make site-check` stale generated output on `master` in a direct-to-master change. Feature branches use `--build-only` for this reason.
- `make catalog-check` in blinkless is red on `master` with `unknown handler function: wrap-inside -> wrapBreakInsideAvoid`; it is pre-existing and outside this diff.

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
- [ ] Both modules resolve the same released blinkless version
- [ ] The conservative hover fallback is acceptable for pages with escaping selectors

---

## Diff stat by extension

Branch `fix/loglens-performance-and-profiling`, base `master` (`442d164a`), 10 commits.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 1 | 3 | 0 |
| `.go` | 40 | 2066 | 171 |
| `.md` | 3 | 362 | 0 |
| `.mod` | 2 | 2 | 2 |
| `.ps1` | 1 | 171 | 0 |
| `.py` | 2 | 335 | 0 |
| `.sh` | 2 | 43 | 1 |
| `.sum` | 2 | 4 | 4 |
| **Total** | **53** | **2986** | **178** |
