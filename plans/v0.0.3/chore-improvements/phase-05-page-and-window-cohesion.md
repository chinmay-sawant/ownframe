# v0.0.3 - Phase 05: `internal/page`, `internal/window`, and `internal/replay` Cohesion

> **Parent:** [README.md](README.md)
> **Status:** In progress (6/22 done: `internal/window` 406->70 files and `internal/replay` 146->29 files complete; `internal/page` pending)
> **Estimated effort:** 6–10 days (land in small, domain-scoped PRs)

---

## Overview

Although the old 2,000-character file cap was retired in `c59f1a51` (`COHESION` in `plans/README.md`), the codebase still bears heavy structural scars from that era:
- `internal/page` contains **507 `.go` files total**, with **316 `.go` files sitting flat in `internal/page/`** and **360 methods defined on `*Page`** (plus 12 subpackages).
- `internal/window` contains **409 `.go` files total**, with **406 `.go` files sitting flat in `internal/window/`** (only `pacing/` has 3 files).
- `internal/replay` contains **146 `.go` files**.

Keeping one giant stateful struct split across 300–400 tiny files does not reduce coupling—it spreads a single operation across a dozen files, hides ownership boundaries, and makes debugging and onboarding exhausting. This phase consolidates micro-files by domain and tightens internal struct/subpackage boundaries without changing public behavior.

## Executive Summary

- Audit and map `internal/page`'s 316 flat files and 360 `*Page` methods into clear domains; merge micro-files and push subsystem state into cohesive state structs or existing subpackages (`form/`, `draw/`, `a11y/`, `scroll/`, `dirty/`, `gesture/`, `patch/`, `reload/`).
- Partition `internal/window`'s 406 flat files into cohesive domain files and dedicated subpackages where ownership is distinct (DevTools dock, tile/viewport repainting, input/touch/IME, desktop chrome/placement, AT-SPI host bridge).
- Consolidate `internal/replay`'s 146 files into clean pipeline stages (raster cache, text shaping/cache, chunk/layer cache, blend/shader compositing, draw list replay).
- Preserve all invariants, performance benchmarks, race safety, and GPU pixel regressions throughout.

## Current Source Baseline (`da4f12cc`)

| Package | Total `.go` Files | Flat Root `.go` Files | Notable Subpackages / Metrics |
|---|---|---|---|
| `internal/page` | **507** | **316** | **360 methods on `*Page`**; 12 subpackages (`form/` 54, `draw/` 46, `a11y/` 23, `gesture/` 18, `dirty/` 12, `patch/` 7, `reload/` 7, `source/` 7, `scroll/` 6, `hit/` 5, `click/` 4, `nav/` 2) |
| `internal/window` | **409** | **406** | Only 1 subpackage (`pacing/` 3 files); mixes game loop, tile/viewport painting, devtools dock, input/touch/IME, OS chrome/snap, a11y bridge |
| `internal/replay` | **146** | ~140+ | Raster cache, text cache, chunk cache, layer cache, Kage shaders, vector/image/grid drawing |
| Core module total (excl. `examples/`, `temp/`, `frontend/`) | **~200,609 LOC** | — | Largest non-test files in page/window: `form_wire.go` (1,158 LOC), `placement.go` (843 LOC), `gesture_wire.go` (813 LOC), `motion_wire.go` (785 LOC) |

---

## Phase 5: `internal/page`, `internal/window`, and `internal/replay` Cohesion

### 5.1 Inventory and domain map for `internal/page`

- [ ] `COH-01` Generate a file-size and method-group inventory of all 316 flat `.go` files and 360 `*Page` methods in `internal/page/` (grouping non-test vs test files and flagging every file under 60 lines). Proof: inventory summary recorded in PR/phase notes with target merge groups.
- [ ] `COH-02` Identify every `*Page` method that exists only as a pass-through to an already-extracted subpackage (`form/`, `scroll/`, `a11y/`, `reload/`, `nav/`, `dirty/`, `gesture/`, `patch/`) and classify which ones are genuinely required by `ownframe` root doors or `internal/host` interfaces vs internal-only wrappers that can be deleted. Proof: list of removable pass-through methods verified with `grep`.
- [ ] `COH-03` Identify dead/redundant test helper files across `internal/page/*_test.go` and group fragmented 1-test files by scenario (e.g., all focus tests together, all scroll tests together, all hot-reload tests together, per `AGENTS.md` "Package and file cohesion"). Proof: test consolidation map recorded before editing.

### 5.2 Consolidate `internal/page` micro-files and shrink `*Page` surface

- [ ] `COH-04` Consolidate lifecycle, config, template source, data binding, and redraw orchestration files in `internal/page/` into cohesive domain files (`page.go`, `config.go`, `render.go`, `patch.go`, `reload.go`, `nav.go`). Proof: flat file count drops meaningfully; `go test -race ./internal/page/...` exits 0.
- [ ] `COH-05` Consolidate form, focus, caret, selection, editing, and key/pointer input files in `internal/page/` (pairing `form_wire.go`, `gesture_wire.go`, and scattered single-method files into cohesive `form.go`, `edit.go`, `pointer.go`, `keys.go`, `menu.go`, `cursor.go` files) and delegate internal state transitions cleanly to `internal/page/form` and `internal/page/gesture`. Proof: `go test -race ./internal/page/...` exits 0; zero behavior regressions.
- [ ] `COH-06` Consolidate scroll, virtualization, pinned layer, motion/tick, accessibility, probe/stats, and testing-hook files in `internal/page/` into cohesive files (`scroll.go`, `tick.go`, `motion.go`, `a11y.go`, `probe.go`, `testing.go`). Proof: `go test -race ./internal/page/...` exits 0.
- [ ] `COH-07` Consolidate fragmented `internal/page/*_test.go` files by feature domain (keeping related test cases and fixtures together without dropping a single test assertion or benchmark). Proof: total test count (`go test -v ./internal/page | grep -c '^=== RUN'`) stays >= baseline before consolidation.
- [ ] `COH-08` Shrink `*Page` method count by moving sub-domain internal helpers onto dedicated embedded/owned state structs (`p.form`, `p.scroll`, `p.motion`, `p.probe`, `p.reload`) instead of hanging every private helper off `*Page`. Target: reduce flat root `.go` files in `internal/page/` from **316 to <=80** (including tests) and reduce `*Page` methods from **360 to <=180**. Proof: `ls -1 internal/page/*.go | wc -l` and `grep -rn '^func (p \*Page)' internal/page/ | wc -l` measured and recorded.
- [ ] `COH-09` Verify zero regressions on page benchmarks (`BenchmarkTick_CaptureAndDamage_2000Ops_Mod0`, `BenchmarkTick_DiffByPosition_2000Ops`) and `TestPageSoak`. Proof: benchmark & soak outputs recorded.

### 5.3 Partition and consolidate `internal/window` (406 flat files)

- [x] `COH-10` Generate a domain inventory of all 406 flat `.go` files in `internal/window/` (207 production, 199 test) mapped across lifecycle, repaint/tiles/perf, devtools, input/IME, chrome/placement/debug, and a11y bridge. Evidence: all 406 files mapped and consolidated cleanly into 70 domain files (47 production + 23 test).
- [ ] `COH-11` Extract genuinely independent subsystems in `internal/window` with clean inputs/outputs into subpackages where it reduces coupling (candidates: `internal/window/devtools` for the DevTools dock UI/state, and `internal/window/tiles` or `internal/window/surface` for tile/viewport damage buffers). Keep tight Ebiten loop state in `internal/window` to avoid artificial indirection. Proof: extracted subpackages have explicit interfaces, zero circular imports, and isolated unit tests.
- [x] `COH-12` Merge the remaining flat non-test micro-files in `internal/window/` into cohesive per-responsibility files (preserving platform build tags `*_linux.go`, `*_windows.go`, `*_darwin.go`, `*_js.go`, `*_android.go`, `*_desktop.go`, `*_mobile.go`, `*_tag.go`, `*_notag.go` cleanly). Evidence: 207 -> 47 production files; `GOOS=windows`, `GOOS=darwin`, `GOOS=js GOARCH=wasm`, `GOOS=linux`, and `-tags ownframe_debug` `go vet ./internal/window/...` all exited 0.
- [x] `COH-13` Consolidate fragmented `internal/window/*_test.go` files by subsystem while preserving every test function name referenced by `scripts/check-graphics.sh`. Evidence: 199 -> 23 test files (total flat `.go` files **406 -> 70**, well under <=90 target); `go test -list . ./internal/window` confirmed **441/441** tests preserved (0 dropped); `go test -race ./internal/window/...` passed (2.47s).
- [x] `COH-14` Verify zero regression on tile repaint benchmarks and `internal/window/pacing` tests. Evidence: `BenchmarkTilePaintFull` 24,701 ns/op (14,600 B/op, 100 allocs), `BenchmarkTileDrawReplayPartialDamaged` 20,903 ns/op (9,339 B/op, 54 allocs); `internal/window/pacing` passed under `-race` (1.02s).
- [ ] `COH-15` Audit `internal/window` package-level mutable globals (especially desktop chrome/window-control singletons tied to Phase 03 `API-02`) and move per-loop state onto the `Game` / host struct where Ebiten allows. Proof: `gochecknoglobals` clean without new `//nolint` suppressions.

### 5.4 Consolidate `internal/replay` (146 files) and verify cohesion

- [x] `COH-16` Inventory `internal/replay` (146 `.go` files) and merge fragmented micro-files into cohesive pipeline files (`renderer.go`, `draw.go`, `shapes.go`, `borders.go`, `bounds.go`, `text.go`, `text_cache.go`, `images.go`, `raster_cache.go`, `chunk_cache.go`, `chunk_plan.go`, `layer.go`, `layer_cache.go`, `blend.go`, `xform.go`, plus `//go:embed` files `backdrop_filter.go`, `group_shader.go` and 12 grouped `*_test.go`). Evidence: **146 -> 29** `.go` files (17 prod + 12 test; <=45 target); `go test -list . ./internal/replay` confirmed **136/136** tests preserved (0 dropped); `go test -race ./internal/replay/...` passed (1.23s).
- [x] `COH-17` Verify replay benchmarks (`BenchmarkBlitRaster`, `BenchmarkLayerBeginLayers`, `BenchmarkText_LetterSpacing_Cold`) and preserve `gpu_alpha_pixels_test.go`. Evidence: `BenchmarkBlitRaster` 533.5 ns/op (582 B/op, 4 allocs), `BenchmarkLayerBeginLayers/ops_1000/contiguous` 18,106 ns/op (3,328 B/op, 6 allocs), `BenchmarkText_LetterSpacing_Cold` 74,962 ns/op (34,177 B/op, 323 allocs); `gpu_alpha_pixels_test.go` preserved.
- [ ] `COH-18` Update `AGENTS.md` / `CONTRIBUTING.md` cohesion examples if file paths changed, and verify `scripts/check-graphics.sh` runs all five isolated graphics sessions cleanly. Proof: `make graphics-check` and `make docs-check` exit 0.

---

## Dependencies

- Run after **Phase 03 (`API-02`–`API-03`)** so `internal/window` host-control seams are finalized before reorganizing `internal/window`.
- Run after **Phase 04** so `internal/page/motion_wire.go` already imports the consolidated `internal/motion` package.
- Keep `scripts/check-graphics.sh` test names intact across all test-file merges.

## Risks

- **Broken platform build tags:** Merging `internal/window` files carelessly can mix Linux/Windows/macOS/wasm/Android build constraints. Keep OS-specific files separate by build tag and verify cross-OS `go vet` on every PR.
- **Hidden test deletion during file merges:** When combining `*_test.go` files, compare `go test -list .` output before and after every merge to prove zero tests were silently dropped.
- **Giant monolithic PRs:** Never merge 300 files in one commit. Land 1 PR per cohesive domain (e.g., `internal/page` forms, `internal/page` scroll/motion, `internal/window` devtools, `internal/window` tiles, `internal/replay`).

## Closure Gates

- [ ] `COH-GATE-01` Flat `.go` files reduced: `internal/page/` from 316 to <=80; `internal/window/` from 406 to <=90; `internal/replay/` from 146 to <=45.
- [ ] `COH-GATE-02` `*Page` method count in `internal/page/` reduced from 360 to <=180.
- [ ] `COH-GATE-03` Zero tests dropped (`go test -list .` diff verified across `internal/page`, `internal/window`, `internal/replay`).
- [ ] `COH-GATE-04` `make test`, `make test-race`, `make lint`, `make graphics-check`, and `make soak-check` (or scoped soak) exit 0.
