# v0.0.3 - Phase 06: Components, Generated UI, and Module Boundaries

> **Parent:** [README.md](README.md)
> **Status:** In progress (13/18 done: chart skins decoupled to `examples/builder/charts/`, `ui-gen-check` wired into `make lint`; `MOD-10` remaining)
> **Estimated effort:** 3–5 days

---

## Overview

At commit `da4f12cc`, the root library module (`github.com/chinmay-sawant/ownframe`) carries significant non-core breadth before `v0.1.0`:
- `components/` contains **7 separate JavaScript chart-library skins** (`apexcharts`, `chartjs`, `echarts`, `nivo`, `recharts`, `shadcnchart`, `tremor`), each with ~780–902 line `paint.go` files (~5,000+ lines of chart rendering code). Across the entire repository, **zero library packages and only one example (`examples/builder/ui-builder/screen_*.go`) import them**.
- `ui/properties_generated.go` is **10,198 lines** of generated Go code compiled into the core module.
- `examples/` has grown to **84 subdirectories** (2.5 GB on disk), including 8 legacy input/form micro-demos (`bind`, `bind-hooks`, `clipboard`, `controls`, `editing`, `forms`, `input`, `states`) already skipped by `Makefile` `SKIP_OPEN` because `input-lab` superseded them.

Keeping the root library module lean reduces compile times, symbol-table bloat, API maintenance burden, and cognitive load for consumers doing `go get github.com/chinmay-sawant/ownframe`.

## Executive Summary

- Decouple the 7 chart-library skins from the core `ownframe` module (either into their own Go module `components/charts` or into `examples/builder/`, or unify shared chart paint primitives to eliminate massive copy-paste duplication).
- Audit and trim `ui/properties_generated.go` (10,198 LOC) and document its generator cleanly.
- Prune or consolidate superseded micro-examples in `examples/` while preserving test coverage and teaching examples in `examples/maintenance.json`.
- Verify isolated consumer builds via `make module-check` and `make examples-check`.

## Current Source Baseline (`da4f12cc`)

| Area | Baseline Measurement | Consumers / Findings |
|---|---|---|
| `components/apexcharts`..`tremor` (7 chart pkgs) | **7 packages**, ~5,000+ LOC (`paint.go` 779–902 LOC each) | Imported **only** by `examples/builder/ui-builder/screen_*.go` and their own `*_test.go` |
| `components/chadcn` | Primary reusable component set | Imported by `examples/builder/ui-builder` and `examples/builder/loglens`; stays in core or `components/` |
| `ui/properties_generated.go` | **10,198 LOC** (single largest `.go` file in repo) | Generated CSS property builder helpers in `ui/` |
| `examples/` | **84 subdirectories**, 2.5 GB on disk | 8 folders already superseded by `input-lab` (`SKIP_OPEN` in `Makefile`), plus mixed experiments and clones |

---

## Phase 6: Components, Generated UI, and Module Boundaries

### 6.1 Decouple or unify the 7 chart-library skins

- [x] `MOD-01` Audit all 7 chart packages (`components/apexcharts`, `components/chartjs`, `components/echarts`, `components/nivo`, `components/recharts`, `components/shadcnchart`, `components/tremor`). Evidence: **7,331 Go LOC** across 49 files (`4,130` LOC in 5 `paint.go` files); zero SVG (pure HTML/CSS flex mockups); `shadcnchart/paint.go` is **100% identical** to `chartjs/paint.go`, `tremor` 97.1%, `recharts` 94.5%, `apexcharts` 91.7%, with byte-identical 26-line `css/00-base.css`; zero library callers (only `examples/builder/ui-builder/screen_*.go`). Recommendation: **Option A** (move out of root module into `examples/builder/charts/` or nested module).
- [x] `MOD-02` Make and record the module-boundary decision for the 7 chart skins: **Option A selected** — move all 7 chart-library skin packages (`apexcharts`, `chartjs`, `echarts`, `nivo`, `recharts`, `shadcnchart`, `tremor`) out of the root `ownframe` module into `examples/builder/charts/` (which is already inside the separate `examples/` Go module). Rationale: zero library callers, single showcase consumer (`examples/builder/ui-builder`), and 7,331 LOC of HTML/CSS flex mockups.
- [x] `MOD-03` Execute Option A from `MOD-02`, moving all 7 chart packages to `examples/builder/charts/`, updating all 8 `examples/builder/ui-builder/` files, all 7 chart test files, `components/README.md`, and `AGENTS.md`. Evidence: 63 files / 7,331 Go LOC / 259 KB removed from root module; `go test ./components/... ./examples/builder/charts/... ./examples/builder/ui-builder/...`, `make module-check`, and `make examples-check` all exited 0.
- [x] `MOD-04` Audit `components/chadcn` for remaining widget interaction gaps and verify zero `internal/` imports. Evidence: `go list -f '{{.Imports}}' ./components/chadcn` imports only `ownframe`, `ui`, `ui/theme`, and stdlib (0 `internal/` imports); `go test ./components/...` passed across all 8 packages; all 25 `plans/v0.0.4/03-component-interactions.md` items complete.
- [x] `MOD-05` Verify root module zip contents via `make module-check` before and after `MOD-03`. Evidence: `make module-check` dropped from **1,709** archive entries to **1,381** archive entries (including the 63 removed chart files + concurrent page consolidation), and the isolated public consumer compiled with no workspace or replace.

### 6.2 Audit `ui/properties_generated.go` (10,198 LOC)

- [x] `MOD-06` Inspect `scripts/gen-ui-properties/` and `ui/properties_generated.go` (10,198 LOC / 371.8 KB). Evidence: generates **688** `(*Element)` methods, **688** `*Value` interfaces, **122** keyword consts, **1,589** unexported marker methods from `documentation/compatibility-matrix.md`; only **60/688 (8.7%)** of `(*Element)` methods are used across `components/` + `examples/` (70/688 = 10.2% repo-wide — **89.8% unused**).
- [x] `MOD-07` Measure compile-time and binary-size impact of `ui/properties_generated.go`. Evidence: minimal `ui`-importing binary links **1,433 `ui` symbols (206.8 KB)** and compiles `ui` in ~1.65s; `ui` also imports `ownframe` + `reflect` (22.85 MB minimal binary). Recommendation: keep typed helpers for ~100 core layout/visual properties and route rarely-used CSS properties through generic `Prop(name, val)`.
- [x] `MOD-08` Ensure `go generate ./ui/...` is reproducible, deterministic, and checked by lint/CI. Evidence: added `-out` flag to `scripts/gen-ui-properties/main.go`, created non-mutating `scripts/check-ui-generated.py` + `scripts/test_ui_generated.py`, added `make ui-gen-check`, and wired into `scripts/run-lint.py`; `make ui-gen-check` (371,862 bytes match) and `make tools-check` (57 tests OK) exit 0.

### 6.3 Organize and prune `examples/` (84 subdirectories)

- [x] `MOD-09` Audit `examples/maintenance.json` against all directories in `examples/` and `Makefile` `SKIP_OPEN`. Evidence: 75 top-level dirs (7 `maintained`, 64 `experimental`, 1 tracked non-Go `android-host`, 3 gitignored local dirs `AI/`, `airmouse/`, `gallery/`). Confirmed **0 of 10 `SKIP_OPEN` folders are pure duplicates of `examples/input-lab`** (8 tests): every folder carries unique test assertions (e.g., `bind` checkbox/select, `forms` exact snapshot, `bind-hooks` veto error, `clipboard` `LastFocusID`/selection, `controls` `SetPicker`/multiline wrap, `states` replay check) that must be migrated before deletion in `MOD-10`.
- [ ] `MOD-10` Migrate any unique regression assertions from the 8 `SKIP_OPEN` legacy form/input folders (`bind`, `bind-hooks`, `clipboard`, `controls`, `editing`, `forms`, `input`, `states`) into `examples/input-lab` or library unit tests (`internal/page/form`), then remove the superseded duplicate example folders and clean `SKIP_OPEN` in `Makefile`. Proof: zero lost test coverage (`go test ./examples/input-lab/... ./internal/page/...`), `Makefile` `SKIP_OPEN` shrunk, and `examples/` directory count reduced.
- [x] `MOD-11` Remove or clean non-example scratch folders under `examples/` (`examples/temp/`) and verify `examples/maintenance.json`. Evidence: removed empty untracked `examples/temp/`; `make examples-check` passed (`7` maintained + `64` experimental/support roots).
- [x] `MOD-12` Review audio/media weight in `examples/spotify-player/player/assets/audio/`. Evidence: **52.82 MiB (55.39 MB)** across 73 tracked files (30 MP3s = 50.79 MiB, 40 cover JPGs = 1.96 MiB, metadata = 59 KB), embedded via `//go:embed` in `audio_catalog.go` (drives the 66 MB `spotify-player/ownframe.wasm.gz`). Confirmed completely excluded from root `ownframe` library archive by nested `examples/go.mod` and `scripts/cmd/module-check`.

### 6.4 Module consumer and CI verification

- [x] `MOD-13` Run `make module-check` to confirm an isolated consumer outside `go.work` resolves and builds cleanly against the packaged library archive. Evidence: `make module-check` exited 0 (`1381 archive entries; isolated consumer compiled with no workspace or replace`).
- [ ] `MOD-14` Run `make test`, `make lint`, and `make examples-check` across both modules (`root` and `examples/`). Proof: all three commands exit 0.

---

## Dependencies

- `MOD-03` must update `scripts/cmd/module-check` if a new nested `go.mod` is introduced under `components/`.
- `MOD-10` must migrate all assertions from superseded examples *before* deleting any folder so test coverage never drops.
- Coordinates with **Phase 02 (`DOCS-14`)**: each kept major example owns its own `README.md` instead of living inside `AGENTS.md`.

## Risks

- **Breaking `examples/builder/ui-builder`:** `ui-builder` imports all 7 chart packages directly (`screen_apexcharts.go` through `screen_tremor.go`). Keep `ui-builder` compiling and passing tests in the same PR as `MOD-03`.
- **Breaking public `ui` callers:** Trimming `ui/properties_generated.go` could remove helpers used by `components/chadcn` or `examples/`; verify against full `make build` (`go vet ./... ./examples/...`) and `make docs-check` before removing any generated method.

## Closure Gates

- [x] `MOD-GATE-01` 7 chart skins decoupled from the root library module into `examples/builder/charts/` (63 files / 7,331 Go LOC removed from root module). Evidence: `make module-check` and `go test ./examples/builder/charts/... ./examples/builder/ui-builder/...` exit 0.
- [ ] `MOD-GATE-02` `ui/properties_generated.go` audited, reproducible, and trimmed where dead/bloat is confirmed.
- [ ] `MOD-GATE-03` Superseded legacy micro-examples (`examples/temp`, superseded `SKIP_OPEN` folders) cleaned up with zero lost test assertions.
- [ ] `MOD-GATE-04` `make module-check`, `make examples-check`, `make lint`, and `make test` exit 0.
