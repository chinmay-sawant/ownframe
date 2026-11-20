# chore: consolidate internal packages, cut repo footprint, and align docs for the v0.0.3 cleanup

## Summary

- Executes the phase-wise plan in `plans/v0.0.3/chore-improvements/`: 101 of 147 checklist rows are now `[x]` with recorded evidence, 4 are `[~]` (deferred or partial, each with a reason), and 42 stay open.
- Collapses micro-file sprawl without dropping a single test: `internal/page` 317 to 52 files, `internal/window` 406 to 70, `internal/replay` 146 to 29, and the module root 48 to 18 non-test files. All 7 CSS motion packages now live in one `internal/motion` package.
- Adds guardrails so the cleanup cannot regress: a 5 MiB tracked-file gate, a doc-pin drift check, a relative-link check, a generated-`ui` drift check, `testing` and root-`ownframe` import gates for `internal/`, and a manual-vs-CSS motion overlap counter.

This PR changes structure and tooling. It does not add a user-visible feature. One public behavior note: `Page.Window()` desktop controls are now bound to the page that `Run`, `RunWithOptions`, or `NewGame` attaches, and are safe no-ops on detached pages (see Breaking changes).

---

## Motivation / context

- Plans: `plans/v0.0.3/chore-improvements/README.md` (index) and phase files `phase-01` to `phase-07`. Each row carries its evidence beside it.
- Baseline audit at `da4f12cc`: about 200,600 core Go lines, 507 files in `internal/page`, 409 in `internal/window`, 10 separate motion packages, 149 exported root methods, a 3.4 GB `.git`, and docs that cited a stale `blinkless` pin.
- Process: the work landed in waves. Parallel agents owned disjoint directories, and every wave ended with a green `make lint` before it was committed.

---

## Changes

### Phase 01: repository footprint and binary hygiene (8 of 21 rows)

- Removed local leftovers: 14 `*.test` binaries (about 360 MB), `mem.out`, a 0-byte tracked file named `Händen`, stale `.doc-snippets-*` directories, `examples/css-preview/css-preview.test`, and an empty `examples/temp/`.
- `make clean-local` removes build, test, and profile leftovers and leaves `temp/` alone (it prints the size). `.gitignore` also covers `*.prof`, `cpu.out`, and `.pytest_cache/`.
- `scripts/check-repo-size.py` fails on any newly tracked file over 5 MiB. It reads a frozen 28-entry allowlist (`scripts/repo-size-allowlist.txt`) that can only shrink, and it is wired into `make lint` and `make repo-size-check`. Unit tests are in `scripts/test_repo_size.py`.
- Audited `.git` blob history (the largest cumulative paths are the `docs/demos/*.wasm` files and the Android APKs) and recorded it in `phase-01`.
- `CONTRIBUTING.md` documents `--filter=blob:none` and `--depth=1` clones.
- `FOOT-05` and `FOOT-10` are audits only: no Pages workflow exists today, 8 APKs (230.85 MiB) are tracked, and no CI job references them. No binaries were removed.
- `FOOT-08` is `[~]`: the `.webp` captures stay exactly as they are, by owner decision.

### Phase 02: docs, plans, and truth alignment (20 of 21 rows, phase complete)

- `AGENTS.md` shrank from 31,525 to about 9,050 bytes. It kept every build, lint, test, API, and clipboard rule and dropped per-example narratives. The stale `blinkless v0.0.4-...` pin is gone and the file now points at `go.mod`.
- `scripts/doc_claims.py` now fails on stale `blinkless` or Go version pins in `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, and `documentation/android-runtime.md`. It also checks relative links in `README.md`, `AGENTS.md`, `documentation/README.md`, and `plans/README.md`. Both have regression tests.
- Replaced "1.0 ships ..." with honest `v0.0.x` wording in `README.md`, `AGENTS.md`, `documentation/platforms.md`, `documentation/features.md`, and `documentation/accessibility-native.md`. `platforms.md` gained a current-versus-planned table.
- `documentation/README.md` is now four tiers with a three-document start path, and links all 66 sibling docs with no orphans. Historical docs carry status banners.
- `plans/README.md` gained an active-versus-archived table, and `skills/README.md` is a new inventory.
- `DOCS-10` is `[~]`: the eight `compare-*.md` files stay separate because `frontend/src/compareDisplay.js`, `frontend/scripts/check-site.mjs`, and `documentation/snippets.json` route each slug on its own.

### Phase 03: API boundaries and ergonomics (16 of 22 rows)

- `Page.Window()` desktop controls (`CursorPosition`, `Minimize`, `ToggleMaximize`, `Maximized`, `Close`) go through an unexported `windowHost` that `RunWithOptions` and `NewGame` bind. `page_window.go` no longer imports `internal/window`. `TestPageWindowHostControlsDetachedAndBound` covers nil, detached, and bound pages.
- Root non-test files went from 48 to 18 using `scripts/merge-go.go`. `alias_allow_test.go` has the new file names and the same 31 aliases.
- Removed 5 unused exported methods (`Page.Audio`, `Page.AudioPlayback`, `Probe.Elements`, `Scroll.ViewportPinZ`, `Testing.TakeChangedProperties`). Exported receiver methods went from 149 to 144. `Testing.DevTools` stays because three example tests call it.
- New `ui/virtual` package (`Window` computes the visible row range and both spacer heights) with edge-case tests.
- New `documentation/updates.md` (which update API to call), plus a pinned-chrome section in `documentation/scrolling.md` and a `Config.RetainState` note in `documentation/css-motion.md`.
- `go doc .` fixes: the `Page` doc comment was detached by the new interface and now reads correctly, and a duplicate package comment in `elements.go` is gone.
- `MotionStats.ManualOverlap` and `ManualOverlapID` count ticks that write an op an active CSS animation or transition owns (`motion.Runtime.CheckManualOverlap`, test in `internal/page/motion_overlap_test.go`).
- `TestNoUnauthorizedPageTestingCalls` blocks new non-test `.Testing()` callers. `API-09` is `[~]` because 5 older example harnesses stay frozen in its allowlist.
- `scripts/example-internal-allowlist.txt` went from 16 to 10 pairs.

### Phase 04: motion consolidation (16 of 19 rows)

- `internal/transition`, `csstransition`, `cssmotion`, `csstrigger`, `animation`, and `cssanimation` merged into `internal/motion`. Collisions were resolved by renaming: `BindAnimation`, `BindTransition`, `BindTransitionTo`, `TriggerEvent`, `DispatchTrigger`, and unexported `resolveTriggers`. The duplicate easing wrapper was dropped, and `clamp01`, `lerp`, `percentDivisor`, and `decomposeMatrix2D` each have one copy.
- `internal/animationpreview`, `transitionpreview`, and `previewwindow` moved to `cmd/internal/`. Their `testing`-importing helpers became `_test.go` files.
- `internal/viewtransition` no longer imports root `ownframe` (it uses `blinkless/layout`; the root types are aliases, so callers are unchanged).
- `internal/frame` and `internal/motion` now share one bounds routine (`dirty.OpBounds`, `dirty.LocalPaintBox`). Four duplicate helpers were deleted and a parity test was added. This also fixes an inconsistency: stroke-width growth now applies on the identity-transform path too.
- Gates: `TestInternalNoTestingImport` and `TestInternalNoRootOwnframeImport`.

### Phase 05: page, window, and replay cohesion (16 of 22 rows)

- File counts: `internal/page` 317 to 52, `internal/window` 406 to 70, `internal/replay` 146 to 29. Test inventories are identical before and after: 529, 441, and 136 entries (1,106 in total, none dropped, checked with `go test -list .` diffs).
- Build-tagged files in `internal/window` stay separate and match their tags. `go vet` passed for Linux, Windows, macOS, wasm, and `-tags ownframe_debug`.
- Page, tile, and replay benchmarks were re-run (for example `BenchmarkTick_CaptureAndDamage_2000Ops_Mod0` at 0 allocs). The raw numbers are in the phase file.
- `COH-15`: audited every package-level `var` in `internal/window`. None were accidental per-loop state. `update.go` `stages` became a fixed-length array.
- `COH-11` is `[~]`: `devtools*.go` and `replay*.go` stay in `internal/window` because splitting them would need a wide forwarding adapter back into `*shell`.

### Phase 06: components and module boundaries (13 of 18 rows)

- Moved the 7 chart-skin packages from `components/` to `examples/builder/charts/` (63 files, 7,331 Go lines). `components/chadcn` stays. `make module-check` archive entries went from 1,709 to 1,381.
- `scripts/check-ui-generated.py` regenerates `ui/properties_generated.go` into a temp file and compares bytes without touching the tree. `scripts/gen-ui-properties` gained an `-out` flag. It runs in `make lint` and as `make ui-gen-check`.
- Audits recorded in `phase-06`: 89.8% of the 688 generated `Element` methods are never called in the repo, `ui` pulls in `ownframe`, and the spotify audio assets are 52.82 MiB across 73 tracked files and are excluded from the root module archive.

### Phase 07: dependencies, binary size, and platform readiness (10 of 24 rows)

- `go mod graph`: every heavy indirect (`fpdf`, `tdewolff/canvas`, `goldmark`, `latin-modern`, `star-tex`, `knuth`, and the old BurntSushi packages) enters through `blinkless`. `make vuln` reports no vulnerability affecting this code.
- Measured on Linux x86_64: `hello` and `login` are 45 MB unstripped and 33 MB stripped, and the `login` wasm is 53 MB (14 MB gzipped).
- `internal/clipboard`: new `ErrWaylandToolsMissing`, `Status`, and side-effect-free `Diagnose()` on every build-tag partition. `Read` still satisfies `errors.Is(err, ErrUnavailable)`. Tests stay on `UseMemory(true)`.
- `release.yml` logs packaged desktop and wasm binary sizes. `documentation/accessibility-native.md` gained an explicit v1.0 accessibility policy.
- `DEP-01` is an audit: no dependency needs Go newer than 1.26.0. The `go` directive was not changed.
- `make module-check` and `make release-check` pass.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | None intended. Page, window, and replay benchmarks were re-run after the merges with 0 B/op and 0 allocs on the tick paths. `frame` bounds now match `motion` bounds on the identity-transform stroke path. |
| **Memory** | None intended. |
| **Behavior / correctness** | Detached pages no longer reach global window state through `Page.Window()`. Stroke damage bounds in `internal/frame` follow `dirty.OpBounds`. |
| **API / CLI** | 5 unused exported methods removed (149 to 144). `go doc .` reads correctly. `ui/virtual` is new. `cmd/*-preview` packages moved under `cmd/internal/`. |
| **Dependencies** | No `go.mod` change. 7 chart packages left the root module archive. |
| **Binary size / build time** | Unchanged. Profiled and documented only; release builds now log sizes. |
| **Repo size** | Working tree only. `.git` history is untouched and no tracked `.wasm`, `.apk`, or `.webp` was removed. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| `Page.Window().Minimize`, `ToggleMaximize`, `Maximized`, `Close`, and `CursorPosition` act only on a page attached by `Run`, `RunWithOptions`, or `NewGame` | Nothing for normal apps. A page that was never run now no-ops (`Maximized` is false, `CursorPosition` is `0, 0`). |
| Removed `Page.Audio`, `Page.AudioPlayback`, `Probe.Elements`, `Scroll.ViewportPinZ`, `Testing.TakeChangedProperties` | No caller exists in `examples/`, `components/`, `ui/`, `cmd/`, or `documentation/`. The library is `v0.0.x`. |
| Chart packages moved from `ownframe/components/<pkg>` to `ownframe/examples/builder/charts/<pkg>` (apexcharts, chartjs, echarts, nivo, recharts, shadcnchart, tremor) | Update the import path. No library code imported them. |
| `internal/` packages merged or moved | Internal only. `scripts/example-internal-allowlist.txt` now lists 10 pairs. |

---

## Library freeze (until Page wrap)

- [x] No new `type X =` at the module root except documented re-exports (`TestRootAliasAllowlist` passes; same 31 aliases)
- [x] No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows
- [x] No new example `ownframe/internal/` imports (`make internal-import-check` is clean at 10 pairs, down from 16)
- [x] No synonym names
- [x] New exported methods return named types in package `ownframe` or a public subpackage (the new `Window` host methods return plain values; `ui/virtual.Range` is a public type)
- [x] Library code panics only for invariants; new failure paths return `error`
- [x] `context.Context` is first where I/O or cancellation applies; no hidden `time.After`
- [x] No new examples that resemble Telegram, Teams, Spotify, or Wispr

## Test plan

Checked on this branch, with the last full `make lint` at `84b033da`:

- [x] `make lint` passes all 10 stages: tools-check, linter version regression, golangci-lint v2.15.0 on both modules, docs snippets, plan states, repo-size gate, examples inventory, example internal imports, ui-generated drift, and website ESLint
- [x] `go test -race` on `internal/page/...`, `internal/motion/...`, `internal/frame`, `internal/viewtransition`, `internal/window/...`, `internal/replay/...`, `internal/clipboard/...`, `ui/...`, `components/...`, `cmd/...`, and the root package
- [x] `go test ./examples/motion-conformance/... ./examples/css-preview/... ./examples/builder/charts/... ./examples/builder/ui-builder/...`
- [x] `make docs-check`, `make plans-check`, `make repo-size-check`, `make internal-import-check`, `make examples-check`, `make ui-gen-check`, `make module-check`, `make release-check`
- [x] `make build BUILD_P=4` (last run at `a0123506`; not re-run after the later merges)
- [x] Cross-OS `go vet ./internal/window/...` (Windows, macOS, wasm, Linux, `ownframe_debug`)
- [ ] Full `make test` across both modules: not re-run on the final head
- [ ] `make test-race` across the whole library module: only the packages listed above were run
- [ ] `make graphics-check`: not run after the window, replay, and page merges. Run it before merging.
- [ ] `make soak-check`: not run

### Commands

```sh
make lint
make test
make test-race
make graphics-check
make module-check release-check
```

---

## Screenshots / sample output

```
PASS maintenance tooling
PASS linter version regression
PASS Go library and examples (v2.15.0)
PASS documentation snippets
PASS plan states
PASS tracked file size guardrail
PASS example maintenance inventory
PASS example internal imports
PASS ui generated properties drift
PASS website ESLint

module-check: 1382 archive entries; isolated consumer compiled with no workspace or replace
release-check: 2 archives verified; no release published
```

---

## Related issues

- Relates to the v0.0.3 cleanup plan in `plans/v0.0.3/chore-improvements/`. No tracked issue number exists for this work; edit this line if one is created.

---

## PR metadata checklist (author)

- [ ] Self-assigned (`--assignee @me`)
- [ ] Labels applied (`enhancement`, `documentation`)
- [x] Related issues filled (no ticket exists yet; see above)
- [x] Filled body committed under `plans/v0.0.4/pr/pr-chore-improvements-v003.md`

---

## Follow-ups (out of scope)

- Phase 01: build the `.wasm` demos in a Pages workflow and move the 8 tracked APKs to releases (`FOOT-06` to `FOOT-13`), then decide on any history rewrite (`FOOT-16`). This needs repository settings changes.
- Phase 03: drain the remaining 10 `examples/css-preview` allowlist pairs (`API-16`, `API-17`), and refactor the 5 `.Testing()` example harnesses off the allowlist (`API-09`).
- Phase 04: split the two largest merged motion files by pipeline stage (`MOT-09`) and run `make graphics-check` plus motion benchmarks (`MOT-16`).
- Phase 05: shrink `*Page` from 360 methods to 180 (`COH-02`, `COH-08`).
- Phase 06: migrate the unique assertions out of the 10 `SKIP_OPEN` examples before removing them (`MOD-10`), then trim `ui/properties_generated.go`.
- Phase 07: move `blinkless` and `ebiten` off pseudo-versions (needs upstream tags), validate Windows IME and real macOS and Windows hosts, and set a stripped binary-size budget.
- Docs: add the two-step `vX.Y.Z` then `examples/vX.Y.Z` release sequence to `documentation/release-checklist.md`.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API changes documented (removed methods, chart import paths, `Page.Window()` binding)
- [ ] No new root aliases, synonym methods, or example `internal/` imports
- [ ] Merged files keep their tests: compare the before and after counts above
- [ ] Build-tagged `internal/window` and `internal/clipboard` files still match their tags
- [ ] `.golangci.yml` path entries follow the renamed files
- [ ] PR has assignee and labels
- [ ] Related issues use the correct keywords
- [ ] No secrets or generated artifacts committed
- [ ] Diff-stat-by-extension table pasted at the bottom
- [ ] `make graphics-check` passes locally before merge

---

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh master` at `84b033da`. Most of the `.go` churn is file renames and merges, so insertions and deletions are close to equal.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.css` | 7 | 0 | 0 |
| `.go` | 1188 | 54044 | 58061 |
| `.md` | 35 | 963 | 175 |
| `.py` | 7 | 266 | 6 |
| `.txt` | 2 | 37 | 12 |
| `.yml` | 2 | 32 | 4 |
| No extension | 3 | 18 | 2 |
| **Total** | **1244** | **55360** | **58260** |
