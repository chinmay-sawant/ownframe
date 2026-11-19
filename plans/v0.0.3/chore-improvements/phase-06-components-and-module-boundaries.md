# v0.0.3 - Phase 06: Components, Generated UI, and Module Boundaries

> **Parent:** [README.md](README.md)
> **Status:** Open (baseline measured at `da4f12cc`)
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

- [ ] `MOD-01` Audit all 7 chart packages (`components/apexcharts`, `components/chartjs`, `components/echarts`, `components/nivo`, `components/recharts`, `components/shadcnchart`, `components/tremor`): measure total LOC, duplicated SVG/axis/legend/paint logic across their `paint.go` files, and confirm zero external or library callers outside `examples/builder/ui-builder`. Proof: audit table of LOC and shared vs unique primitives recorded.
- [ ] `MOD-02` Make and record the module-boundary decision for the 7 chart skins:
  - **Option A (Recommended for lean root module):** Move the 7 chart packages into a separate nested Go module (`components/charts/go.mod` or `examples/builder/charts/`) so `go get github.com/chinmay-sawant/ownframe` does not ship 7 JS chart skin libraries in the core module archive.
  - **Option B (If kept as public reusable charts):** Extract one shared internal chart renderer (`components/chartcore` or `ui/chart`) and thin theme/preset wrappers for the 7 visual styles, cutting duplicated `paint.go` code by >60%.
  Proof: written decision in this phase file before moving code.
- [ ] `MOD-03` Execute the chosen option from `MOD-02`, updating `go.work`, `examples/go.mod`, `examples/builder/ui-builder/`, `components/README.md`, and `scripts/cmd/module-check`. Proof: `make build`, `make test`, and `make module-check` exit 0.
- [ ] `MOD-04` Audit `components/chadcn` for any remaining widget interaction gaps (menus, dialogs, popups, tables) vs `plans/v0.0.4/03-component-interactions.md` and confirm `components/chadcn` imports only public `ownframe` + `ui` packages (zero `internal/` imports). Proof: `go list -f '{{.Imports}}' ./components/chadcn` verified clean.
- [ ] `MOD-05` Verify root module zip contents via `make module-check` before and after `MOD-03` and record archive file count and byte size reduction. Proof: before/after `module-check` archive size recorded.

### 6.2 Audit `ui/properties_generated.go` (10,198 LOC)

- [ ] `MOD-06` Inspect the generator that produces `ui/properties_generated.go` (locate generator script/command, count generated functions/methods, and measure how many are actually used across `components/` and `examples/`). Proof: audit report of total generated helpers vs referenced helpers.
- [ ] `MOD-07` Measure compile-time and binary-size impact of `ui/properties_generated.go` (`go build -toolexec` / `go tool nm -size` on a minimal `ui`-importing binary) and determine whether rarely used CSS properties should use generic `Prop(name, val)` instead of thousands of generated wrapper methods. Proof: measured compile-time and binary-size comparison recorded.
- [ ] `MOD-08` Ensure `go generate ./ui/...` (or documented generator command) is reproducible, deterministic, and checked by lint/CI so `ui/properties_generated.go` cannot drift from its source specification. Proof: re-running generator produces empty `git diff`.

### 6.3 Organize and prune `examples/` (84 subdirectories)

- [ ] `MOD-09` Audit `examples/maintenance.json` against all 84 folders in `examples/`, classifying every folder into: **Teaching** (6 curated core starters), **Showcase / App Demo**, **Conformance / Benchmark Harness**, **Superseded / Duplicate**, or **Stale Scratch** (e.g., `examples/temp/`). Proof: complete inventory table covering all 84 directories.
- [ ] `MOD-10` Migrate any unique regression assertions from the 8 `SKIP_OPEN` legacy form/input folders (`bind`, `bind-hooks`, `clipboard`, `controls`, `editing`, `forms`, `input`, `states`) into `examples/input-lab` or library unit tests (`internal/page/form`), then remove the superseded duplicate example folders and clean `SKIP_OPEN` in `Makefile`. Proof: zero lost test coverage (`go test ./examples/input-lab/... ./internal/page/...`), `Makefile` `SKIP_OPEN` shrunk, and `examples/` directory count reduced.
- [ ] `MOD-11` Remove or clean non-example scratch folders under `examples/` (check `examples/temp/` and untracked `examples/AI/` / `examples/builder/loglens-rust`) and update `examples/readme.md` and `examples/maintenance.json` so every remaining folder has a clear purpose. Proof: `make examples-check` exits 0.
- [ ] `MOD-12` Review audio/media weight in `examples/spotify-player/player/assets/audio/` (~15+ MP3 files, 2–2.6 MB each) and check whether demo clips can be trimmed or kept under a clear asset budget so `examples/` stays manageable. Proof: asset size audit documented.

### 6.4 Module consumer and CI verification

- [ ] `MOD-13` Run `make module-check` to confirm an isolated consumer outside `go.work` resolves and builds cleanly against the packaged library archive. Proof: `make module-check` exits 0.
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

- [ ] `MOD-GATE-01` 7 chart skins either decoupled from the root library module or deduplicated behind a shared paint core (>60% LOC reduction).
- [ ] `MOD-GATE-02` `ui/properties_generated.go` audited, reproducible, and trimmed where dead/bloat is confirmed.
- [ ] `MOD-GATE-03` Superseded legacy micro-examples (`examples/temp`, superseded `SKIP_OPEN` folders) cleaned up with zero lost test assertions.
- [ ] `MOD-GATE-04` `make module-check`, `make examples-check`, `make lint`, and `make test` exit 0.
