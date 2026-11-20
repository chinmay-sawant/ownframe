# v0.0.3 - Phase 04: Motion and Animation Package Consolidation

> **Parent:** [README.md](README.md)
> **Status:** In progress (4/19 done: `MOT-01`–`MOT-04` preview relocation and testing-import gate)
> **Estimated effort:** 4–6 days

---

## Overview

At commit `da4f12cc`, animation and transition logic is fragmented across **12 internal packages**:
- Core CSS motion chain (7 packages): `internal/cssmotion` (12 files), `internal/animation` (44 files), `internal/cssanimation` (32 files), `internal/transition` (6 files), `internal/csstransition` (17 files), `internal/csstrigger` (8 files), `internal/motion` (34 files).
- Preview/test helpers living as top-level runtime `internal/` packages (3 packages): `internal/animationpreview` (8 files), `internal/transitionpreview` (8 files), `internal/previewwindow`. Note that `internal/animationpreview` and `internal/transitionpreview` import Go's `testing` package in non-test production files AND reverse-import root package `ownframe`.
- Additional motion packages (2 packages): `internal/viewtransition` (19 files, which also reverse-imports root `ownframe` and is only used by `examples/css-preview` tests) and `internal/frame` (27 files, Go-driven manual animation).

This fragmentation creates artificial cross-package export boundaries, duplicate easing/ownership glue, reverse dependencies on root `ownframe`, and high cognitive overhead when debugging any animation bug.

## Executive Summary

- Move test/CLI preview helpers (`animationpreview`, `transitionpreview`, `previewwindow`) out of the core runtime `internal/` namespace and eliminate `testing` imports from non-test production packages.
- Merge the 7-package CSS motion chain into a clean 2–3 package architecture with unexported internals (`internal/motion` plus at most `internal/motion/css` if a clean parser/binder boundary is warranted).
- Resolve `internal/viewtransition`'s reverse dependency on root `ownframe` and clarify its status (active engine subsystem vs experimental preview helper).
- Define and enforce a clean boundary between automatic CSS `internal/motion` and imperative Go `internal/frame`.

## Current Source Baseline (`da4f12cc`)

| Package | `.go` Files | Key Imports (Measured via `go list`) | Problem |
|---|---|---|---|
| `internal/cssmotion` | 12 | `blinkless/layout` | Tiny shared utility package; forces exported glue types |
| `internal/animation` | 44 | `cssmotion`, `blinkless/layout` | Split from `cssanimation` & `motion` with heavy cross-calls (`compose.go` 1,215 LOC) |
| `internal/cssanimation` | 32 | `animation`, `cssmotion` | Thin CSS binding layer separated from `animation` |
| `internal/transition` | 6 | `cssmotion` | Only 6 files (3 non-test: `easing.go`, `renderer.go`, `step.go`) |
| `internal/csstransition` | 17 | `transition`, `cssmotion` | Binding layer separated from `transition` |
| `internal/csstrigger` | 8 | `animation` | Small trigger resolver used only by `motion`/`animationpreview`/`examples/css-preview` |
| `internal/motion` | 34 | `animation`, `cssanimation`, `cssmotion`, `csstransition`, `transition`, `page/dirty` | Orchestrator importing all 6 sibling packages (`transition.go` 1,292 LOC, `bind.go` 790 LOC) |
| `internal/animationpreview` | 8 | **`testing`**, **`ownframe`**, `previewwindow`, `animation`, `cssanimation`, `csstrigger` | Non-test package importing `testing` and root `ownframe`; only used by `cmd/animation-preview` |
| `internal/transitionpreview` | 8 | **`testing`**, **`ownframe`**, `previewwindow`, `transition`, `csstransition` | Non-test package importing `testing` and root `ownframe`; only used by `cmd/transition-preview` |
| `internal/previewwindow` | ~4 | Window preview harness | Only used by `animationpreview` and `transitionpreview` |
| `internal/viewtransition` | 19 | **`ownframe`** | Reverse-imports root `ownframe`; only called by `examples/css-preview/*_test.go` |
| `internal/frame` | 27 | `blinkless/layout`, `internal/page` | Imperative Go animation path; overlaps retained-op writes with `internal/motion` |

---

## Phase 4: Motion and Animation Package Consolidation

### 4.1 Remove test/preview helpers from top-level runtime `internal/`

- [x] `MOT-01` Audit `cmd/animation-preview/main.go` and `cmd/transition-preview/main.go` along with all callers of `internal/animationpreview`, `internal/transitionpreview`, and `internal/previewwindow`. Evidence: `grep -rn "animationpreview\|transitionpreview\|previewwindow"` confirmed zero callers outside `cmd/animation-preview`, `cmd/transition-preview`, and those three packages.
- [x] `MOT-02` Remove `testing` imports from non-test `.go` files in `animationpreview` and `transitionpreview` (split `preview_test_helpers.go` into `preview_helpers_test.go` so `testing` flags are never linked into compiled binaries). Evidence: renamed both files to `preview_helpers_test.go`; `go list -f '{{.ImportPath}} {{.Imports}}' ./internal/... ./cmd/... | grep -E ' testing(\b|\])'` returned empty for all runtime and `cmd` packages (`internal/perfcheck` is the sole standalone test-assertion helper package).
- [x] `MOT-03` Relocate `internal/animationpreview`, `internal/transitionpreview`, and `internal/previewwindow` under `cmd/internal/` (`cmd/internal/animationpreview`, `cmd/internal/transitionpreview`, `cmd/internal/previewwindow`), removing 3 non-runtime packages from top-level `internal/`. Evidence: moved directories, updated `cmd/animation-preview/main.go`, `cmd/transition-preview/main.go`, and relative `testdata/` paths; `go test -race ./cmd/...` passed (`animationpreview` 1.18s, `transitionpreview` 1.93s).
- [x] `MOT-04` Add a test gate ensuring no non-test file under `internal/` or `cmd/` imports `testing` (allowlisting only the standalone test-assertion package `internal/perfcheck/perfcheck.go`). Evidence: added `TestInternalNoTestingImport` in `api_internal_gate_test.go`; `go test -race .` passed.

### 4.2 Collapse the 7-package CSS motion chain into a cohesive subsystem

- [ ] `MOT-05` Map every exported symbol across `cssmotion`, `animation`, `cssanimation`, `transition`, `csstransition`, `csstrigger`, and `motion` to identify which symbols are actually consumed by `internal/page` and `internal/window` (external boundary of the motion subsystem). Proof: dependency boundary document listing exact types/functions `internal/page/motion_wire.go` needs.
- [ ] `MOT-06` Merge `internal/transition` (3 non-test files) and `internal/csstransition` into `internal/motion` (or `internal/motion/transition`), unexporting internal binding and stepping helpers and deduplicating easing logic against `cssmotion/easing*.go`. Proof: `internal/transition` and `internal/csstransition` removed; all transition behavior tests (`behavior_*.go`, `renderer_test.go`, `bind_*_test.go`) pass in the consolidated package.
- [ ] `MOT-07` Merge `internal/cssmotion` and `internal/csstrigger` into `internal/motion` (sharing one easing implementation, one ownership index, and one paint-capture helper). Proof: `internal/cssmotion` and `internal/csstrigger` removed; zero duplicate easing functions across `internal/motion/...`.
- [ ] `MOT-08` Merge `internal/animation` and `internal/cssanimation` into `internal/motion` (or keep at most one internal subpackage `internal/motion/timeline` if separation keeps files and types genuinely cleaner). Unexport all helper types not required by `internal/page`. Proof: `internal/animation` and `internal/cssanimation` removed; total motion-related `internal/` packages reduced from 10 to <=2 (excluding `internal/frame`).
- [ ] `MOT-09` Review the two largest merged files (`internal/motion/transition.go` 1,292 LOC and `compose.go` 1,215 LOC) and organize by clear cohesive responsibilities (e.g., `bind.go`, `step.go`, `compose.go`, `layers.go`, `transition.go`, `easing.go`, `damage.go`) without re-creating 15-line micro-files. Proof: clean file layout in `internal/motion/` where every file owns one clear stage of the motion pipeline.
- [ ] `MOT-10` Update `internal/page/motion_wire.go` and `scripts/example-internal-allowlist.txt` (if any `examples/css-preview` entries were not already moved in Phase 03) to match the consolidated package path. Proof: `make build`, `make internal-import-check`, and `go test -race ./internal/motion/... ./internal/page/...` exit 0.

### 4.3 Resolve `internal/viewtransition` and clarify `internal/frame` seam

- [ ] `MOT-11` Audit `internal/viewtransition` (19 files): since it reverse-imports root `ownframe` and is only referenced by `examples/css-preview/catalog_view_transition_*_test.go`, decouple it from root `ownframe` (operate on `layout.Display` / narrow interfaces instead) OR move it into `internal/motion/viewtransition` once cycle-free. Proof: `go list -f '{{.Imports}}' ./internal/motion/...` contains zero imports of root `github.com/chinmay-sawant/ownframe`.
- [ ] `MOT-12` Check whether any non-test `internal/` package still imports root `github.com/chinmay-sawant/ownframe` (preventing circular-feeling layering). Proof: `go list -f '{{.ImportPath}} {{.Imports}}' ./internal/... | grep 'github.com/chinmay-sawant/ownframe '` audited and empty for runtime packages.
- [ ] `MOT-13` Audit `internal/frame` (27 files): consolidate overlapping retained-op helpers with `internal/motion` (sharing op-patching/damage helpers via `internal/page/dirty` or a tiny shared helper) so both declarative CSS motion and manual Go `SetTick` motion compute damage identically. Proof: shared damage helper tested across both paths.
- [ ] `MOT-14` Update `documentation/css-motion.md`, `documentation/frames.md`, and `AGENTS.md` to reflect the consolidated package architecture. Proof: `make docs-check` exits 0.

### 4.4 Regression and conformance verification

- [ ] `MOT-15` Run full motion race tests and conformance suites: `go test -race -p 1 ./internal/motion/... ./internal/frame/... ./internal/page/...` and `go test ./examples/motion-conformance/... ./examples/css-preview/...`. Proof: exit 0 with test counts recorded.
- [ ] `MOT-16` Run `make graphics-check` and `BenchmarkLayerBeginLayers` / motion benchmarks to confirm zero performance regression from package consolidation. Proof: benchmark output and `make graphics-check` exit 0 recorded.

---

## Dependencies

- `MOT-01`–`MOT-04` (moving preview helpers) have zero library dependencies and can land as the first standalone PR.
- `MOT-06`–`MOT-08` should land sequentially (one merge PR per step or one clean coordinated PR) to keep diffs reviewable and `git blame` readable via `git mv`.
- Coordinates with `API-16` in Phase 03 if `examples/css-preview` tests still import `internal/cssmotion`/`internal/csstransition`.

## Risks

- **Symbol collisions during merges:** `animation`, `transition`, and `cssmotion` each define local `easing.go` / `bind.go` helpers. Audit duplicate function names before merging and keep one canonical easing implementation.
- **Hidden test-helper dependencies:** `examples/css-preview` and `cmd/*-preview` depend on several currently exported internals. Update or relocate those callers in the same commit so `make build` never breaks.

## Closure Gates

- [ ] `MOT-GATE-01` Motion-related packages under `internal/` reduced from 11 (`animation*`, `css*`, `motion`, `transition*`, `viewtransition`, `previewwindow`) to <=2 runtime packages (plus `internal/frame`).
- [ ] `MOT-GATE-02` Zero non-test files under `internal/` import `testing` or reverse-import root `ownframe`.
- [ ] `MOT-GATE-03` `make test`, `make test-race`, `make lint`, `make internal-import-check`, and `make graphics-check` exit 0.
