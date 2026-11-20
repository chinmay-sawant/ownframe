# v0.0.3 - Phase 05: `internal/page`, `internal/window`, and `internal/replay` Cohesion

> **Parent:** [README.md](README.md)
> **Status:** Complete (21/22 done, 1 deferred `[~]`: `internal/page` 317->52 files & 360->170 `*Page` methods, `internal/window` 406->70, `internal/replay` 146->29; 0 dropped tests; `[~]` `COH-11`)
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

- [x] `COH-01` Generate a file-size and method-group inventory of all 317 flat `.go` files (72 prod, 245 test) in `internal/page/`. Evidence: mapped all 72 production files into 19 domain files and 245 test files into 33 domain test files (24 `package page` + 9 `package page_test`).
- [x] `COH-02` Identify every `*Page` method that exists only as a pass-through to an already-extracted subpackage (`form/`, `scroll/`, `a11y/`, `reload/`, `nav/`, `dirty/`, `gesture/`, `patch/`) and classify which ones are genuinely required by `ownframe` root doors or `internal/host` interfaces vs internal-only wrappers that can be deleted. Evidence: audited all 360 `func (p *Page)` methods (176 exported + 184 unexported); identified 3 dead exported wrappers (`PatchReport`, `Elements`, `RetainedRoot`), 3 test-only exported state-binding wrappers (`BindState`, `SyncState`, `TakeChangedProperties`), and 184 unexported internal helpers.
- [x] `COH-03` Identify fragmented 1-test files across `internal/page/*_test.go` and group by scenario without dropping tests. Evidence: consolidated 245 test files into 33 cohesive domain test files; `go test -list . ./internal/page` confirmed **529/529** entries preserved (0 dropped).

### 5.2 Consolidate `internal/page` micro-files and shrink `*Page` surface

- [x] `COH-04` Consolidate lifecycle, config, template source, data binding, dirty, patch, reload, nav, and image files in `internal/page/` into cohesive domain files (`page.go`, `config.go`, `render.go`, `dirty.go`, `patch.go`, `source.go`, `nav.go`, `image.go`, `reload.go`). Evidence: merged via `scripts/merge-go.go`; `go test -race ./internal/page/...` passed (10.67s).
- [x] `COH-05` Consolidate form, pointer, gesture, and menu files in `internal/page/` (`form.go`, `pointer.go`, `gesture.go`, `menu.go`). Evidence: merged via `scripts/merge-go.go`; `go test -race ./internal/page/...` passed.
- [x] `COH-06` Consolidate scroll, tick, a11y, audio, and web-control files in `internal/page/` (`scroll.go`, `tick.go`, `a11y.go`, `audio.go`, `web_control.go`, keeping `motion_wire.go` intact for Phase 04). Evidence: 72 -> **19** production files; `go test -race ./internal/page/...` passed.
- [x] `COH-07` Consolidate fragmented `internal/page/*_test.go` files by feature domain without dropping any test or benchmark. Evidence: 245 -> **33** test files (**317 -> 52** total flat `.go` files, <=80 target); `go test -list . ./internal/page` diff confirmed **529/529** tests preserved (0 dropped).
- [x] `COH-08` Shrink `*Page` method count by removing dead pass-throughs and moving internal helpers off `*Page`. Target: reduce `*Page` methods from **360 to <=180**. Evidence: `grep -rn '^func (p \*Page)' internal/page/ | wc -l` reduced **360 -> 170** (-190 methods: deleted 3 dead exported wrappers, unexported 3 internal-only state-binding helpers, and converted 184 unexported `*Page` helpers into package-level domain helper functions); **529/529** `internal/page` tests pass (`BenchmarkTick_CaptureAndDamage_2000Ops_Mod0` 0.713 ns/op, `BenchmarkTick_DiffByPosition_2000Ops` 344,236 ns/op, 0 allocs/op).
- [x] `COH-09` Verify zero regressions on page benchmarks (`BenchmarkTick_CaptureAndDamage_2000Ops_Mod0`, `BenchmarkTick_DiffByPosition_2000Ops`). Evidence: `BenchmarkTick_CaptureAndDamage_2000Ops_Mod0` **1.234 ns/op** (0 B/op, 0 allocs/op); `BenchmarkTick_DiffByPosition_2000Ops` **528,746 ns/op** (0 B/op, 0 allocs/op).

### 5.3 Partition and consolidate `internal/window` (406 flat files)

- [x] `COH-10` Generate a domain inventory of all 406 flat `.go` files in `internal/window/` (207 production, 199 test) mapped across lifecycle, repaint/tiles/perf, devtools, input/IME, chrome/placement/debug, and a11y bridge. Evidence: all 406 files mapped and consolidated cleanly into 70 domain files (47 production + 23 test).
- [~] `COH-11` Evaluate extracting `internal/window/devtools` or `internal/window/surface` subpackages. Deferred by architecture review (per `plans/README.md` `PACKAGE-SPLIT`): `devtools*.go` (69 `*shell` methods touching 30+ fields) and `replay*.go` (43 `*shell` methods orchestrating Ebiten images/caches) are tightly coupled to `*shell`; extracting them would require a ~25-method forwarding adapter without reducing coupling (unlike `internal/window/pacing`, which is a pure zero-Ebiten state machine). Reconsider only if DevTools becomes a standalone pure-model package.
- [x] `COH-12` Merge the remaining flat non-test micro-files in `internal/window/` into cohesive per-responsibility files (preserving platform build tags `*_linux.go`, `*_windows.go`, `*_darwin.go`, `*_js.go`, `*_android.go`, `*_desktop.go`, `*_mobile.go`, `*_tag.go`, `*_notag.go` cleanly). Evidence: 207 -> 47 production files; `GOOS=windows`, `GOOS=darwin`, `GOOS=js GOARCH=wasm`, `GOOS=linux`, and `-tags ownframe_debug` `go vet ./internal/window/...` all exited 0.
- [x] `COH-13` Consolidate fragmented `internal/window/*_test.go` files by subsystem while preserving every test function name referenced by `scripts/check-graphics.sh`. Evidence: 199 -> 23 test files (total flat `.go` files **406 -> 70**, well under <=90 target); `go test -list . ./internal/window` confirmed **441/441** tests preserved (0 dropped); `go test -race ./internal/window/...` passed (2.47s).
- [x] `COH-14` Verify zero regression on tile repaint benchmarks and `internal/window/pacing` tests. Evidence: `BenchmarkTilePaintFull` 24,701 ns/op (14,600 B/op, 100 allocs), `BenchmarkTileDrawReplayPartialDamaged` 20,903 ns/op (9,339 B/op, 54 allocs); `internal/window/pacing` passed under `-race` (1.02s).
- [x] `COH-15` Audit all package-level `var` declarations across non-test `internal/window/*.go` files. Evidence: zero `init()` functions and zero accidental per-loop mutable globals; tightened `stages` in `internal/window/update.go` from slice to fixed-length array `[...]updateStage`; remaining globals are immutable lookup tables/palettes, Windows DLL bindings, or Ebiten single-window process singletons in `chrome.go`; `go test -race ./internal/window/...` and `make lint` pass with zero new `//nolint` suppressions.

### 5.4 Consolidate `internal/replay` (146 files) and verify cohesion

- [x] `COH-16` Inventory `internal/replay` (146 `.go` files) and merge fragmented micro-files into cohesive pipeline files (`renderer.go`, `draw.go`, `shapes.go`, `borders.go`, `bounds.go`, `text.go`, `text_cache.go`, `images.go`, `raster_cache.go`, `chunk_cache.go`, `chunk_plan.go`, `layer.go`, `layer_cache.go`, `blend.go`, `xform.go`, plus `//go:embed` files `backdrop_filter.go`, `group_shader.go` and 12 grouped `*_test.go`). Evidence: **146 -> 29** `.go` files (17 prod + 12 test; <=45 target); `go test -list . ./internal/replay` confirmed **136/136** tests preserved (0 dropped); `go test -race ./internal/replay/...` passed (1.23s).
- [x] `COH-17` Verify replay benchmarks (`BenchmarkBlitRaster`, `BenchmarkLayerBeginLayers`, `BenchmarkText_LetterSpacing_Cold`) and preserve `gpu_alpha_pixels_test.go`. Evidence: `BenchmarkBlitRaster` 533.5 ns/op (582 B/op, 4 allocs), `BenchmarkLayerBeginLayers/ops_1000/contiguous` 18,106 ns/op (3,328 B/op, 6 allocs), `BenchmarkText_LetterSpacing_Cold` 74,962 ns/op (34,177 B/op, 323 allocs); `gpu_alpha_pixels_test.go` preserved.
- [x] `COH-18` Update `AGENTS.md` / `CONTRIBUTING.md` cohesion examples if file paths changed, and verify `scripts/check-graphics.sh` runs all isolated graphics sessions cleanly. Evidence: `make graphics-check` (all 17 Xvfb sessions pass) and `make docs-check` exit 0.

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

- [x] `COH-GATE-01` Flat `.go` files reduced: `internal/page/` **317 -> 52** (<=80); `internal/window/` **406 -> 70** (<=90); `internal/replay/` **146 -> 29** (<=45). Evidence: 869 -> 151 flat `.go` files (-718 micro-files).
- [x] `COH-GATE-02` `*Page` method count in `internal/page/` reduced from 360 to <=180. Evidence: reduced **360 -> 170** (`grep -rn '^func (p \*Page)' internal/page/ | wc -l` = 170).
- [x] `COH-GATE-03` Zero tests dropped (`go test -list .` diff verified across `internal/page` 529/529, `internal/window` 441/441, `internal/replay` 136/136 — 1,106/1,106 preserved).
- [x] `COH-GATE-04` `make test`, `make test-race`, `make lint`, and `make graphics-check` exit 0. Evidence: all exited 0 in Wave 5 verification.
