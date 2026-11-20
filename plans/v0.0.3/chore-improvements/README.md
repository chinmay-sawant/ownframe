# v0.0.3 - Chore Improvements Plan

> **Parent:** [plans/v0.0.3/README.md](../README.md) (and [maintenance ledger](../../README.md))
> **Status:** In progress — **PR #83 merged to `master` at `63464cd0`** (**101 `[x]`** · **4 `[~]`** · **42 `[ ]`** of 147 rows = **71.4% resolved**)
> **Estimated effort:** Remaining 42 rows split across Wave 5 (code/refactor) and external/release gates

---

## Overview

This directory is the canonical phase-wise plan to address structural debt, repository footprint, API leaks, package fragmentation, documentation drift, binary weight, and pre-1.0 readiness gaps audited on `master` at baseline commit `da4f12cc` (2026-11-20), with Waves 1–4 merged into `master` via [PR #83](../../v0.0.4/pr/pr-chore-improvements-v003.md) (`63464cd0`).

Each linked phase file owns its atomic task rows and evidence. This index holds the before/after scoreboard, phase progress breakdown, and the roadmap for the remaining 42 open rows without duplicating active checklists.

## Executive Summary (Baseline `da4f12cc` vs Current `63464cd0`)

| Area | Baseline (`da4f12cc`) | Current on `master` (`63464cd0`) | Remaining Direction | Owning Phase |
|---|---|---|---|---|
| **Repository & Git Footprint** | **3.4 GB `.git`**, 668 MB `docs/` (10 tracked `.wasm`/`.wasm.gz` = 581.8 MB), 8 tracked APKs (230.85 MiB), 7.0 GB `temp/`, ~360 MB stray `*.test` binaries | Stray `*.test`/`mem.out`/`Händen`/`examples/temp` removed; **`make repo-size-check` (5 MiB gate)** live in `make lint`; `.git` blob & CI/APK audits complete; `.webp` files kept (`FOOT-08` `[~]`) | Move `.wasm` builds to `pages.yml` (`FOOT-06`–`09`) and 8 APKs to GitHub Releases (`FOOT-11`–`13`); record `.git` history decision (`FOOT-16`) | [Phase 01](phase-01-repo-footprint-and-binary-hygiene.md) |
| **Docs, Plans & 1.0 Claims** | **69 files in `documentation/`**, 277 in `plans/`, 31.5 KB `AGENTS.md` (stale `blinkless` pin), premature "1.0" phrasing | **Phase Complete:** `AGENTS.md` trimmed **31.5 KB -> 9.0 KB**; 4-tier `documentation/README.md` (0 orphans); automated pin & relative-link checks in `make docs-check`; honest `v0.0.x` vs `v1.0` matrix | Complete (`DOCS-10` `[~]` keeps `compare-*.md` separate for `frontend/` routes) | [Phase 02](phase-02-docs-plans-and-truth-alignment.md) |
| **Public API & Ergonomics** | **48 non-test root `.go` files**, 149 exported methods; global `window.*` calls on `Page.Window()`; 16 allowlisted example internal imports | Root non-test `.go` files **48 -> 18**; exported methods **149 -> 144**; `Page.Window()` bound to `windowHost` (0 `internal/window` imports in `page_window.go`); `ui/virtual` + `updates.md` + `ManualOverlap` diagnostic added; allowlist **16 -> 10** | Drain remaining 10 `examples/css-preview` internal imports (`API-16`–`18`) and close `API-GATE-03`/`04` | [Phase 03](phase-03-api-boundaries-and-developer-ergonomics.md) |
| **Motion Package Sprawl** | **10 motion/transition `internal/` packages** + `internal/frame` + `internal/previewwindow`; non-test preview packages importing `testing` and root `ownframe` | **Collapsed 7 CSS motion packages into 1 `internal/motion`**; moved 3 preview packages to `cmd/internal/`; decoupled `internal/viewtransition` from root `ownframe`; unified `frame.OpBox` with `dirty.OpBounds`; added `TestInternalNoTestingImport` & `TestInternalNoRootOwnframeImport` | Organize `internal/motion/` files (`MOT-09`) and run `make graphics-check` + motion benchmarks (`MOT-16`, `MOT-GATE-03`) | [Phase 04](phase-04-motion-and-animation-consolidation.md) |
| **Page, Window & Replay Cohesion** | `internal/page` **317** flat files (360 `*Page` methods); `internal/window` **406** flat files; `internal/replay` **146** files (**869** flat files total) | Flat `.go` files reduced **869 -> 151 (-718 micro-files)**: `internal/page` **317 -> 52**, `internal/window` **406 -> 70**, `internal/replay` **146 -> 29** with **1,106/1,106 tests preserved (0 dropped)** | Shrink `*Page` pass-through method count (`COH-02`, `COH-08`, `COH-GATE-02`) and run `make graphics-check` + `make soak-check` (`COH-18`, `COH-GATE-04`) | [Phase 05](phase-05-page-and-window-cohesion.md) |
| **Components & Module Scope** | **7 JS chart-library skins** (7,331 Go LOC) in root module; 10,198-line `ui/properties_generated.go`; 75 `examples/` folders | Moved all 7 chart skins (63 files / 7,331 LOC) to `examples/builder/charts/` (root archive **1,709 -> 1,382 entries**); wired **`make ui-gen-check`** into `make lint`; audited `ui/` (89.8% methods unused) & `SKIP_OPEN` examples | Migrate unique assertions from `SKIP_OPEN` legacy input demos (`MOD-10`), trim `ui/properties_generated.go` (`MOD-GATE-02`), close Phase 06 gates | [Phase 06](phase-06-components-and-module-boundaries.md) |
| **Dependencies, Binary Size & 1.0 Gaps** | **`go 1.26.11`**, pseudo-versions of `blinkless` and `ebiten/v2` alpha, 45 MB unstripped / 33 MB stripped binaries, silent Wayland clipboard failure | Audited `go mod graph` (all heavy indirects enter via `blinkless`), `make vuln` (0 code vulns), & symbol profile; added `ErrWaylandToolsMissing` + `clipboard.Diagnose()`; added CI binary-size logging in `release.yml` & v1.0 a11y policy | Tag upstream `blinkless`/`ebiten` (`DEP-02`–`03`), gate optional assets (`DEP-08`–`09`), validate Windows/macOS hosts (`DEP-12`, `DEP-14`), final suite (`DEP-17`–`20`) | [Phase 07](phase-07-dependencies-binary-size-and-platform-readiness.md) |

## Phase Progress Breakdown (`101 [x]` · `4 [~]` · `42 [ ]` of 147 rows)

| Order | Phase | `[x]` Done | `[~]` Deferred/Partial | `[ ]` Open | Total | Status | Canonical Checklist |
|---|---|---:|---:|---:|---:|---|---|
| 1 | **Phase 01:** Repo footprint & binary hygiene | **10** | 1 | 10 | 21 | In progress (52%) | [phase-01-repo-footprint-and-binary-hygiene.md](phase-01-repo-footprint-and-binary-hygiene.md) |
| 2 | **Phase 02:** Docs, plans & truth alignment | **20** | 1 | **0** | 21 | **Complete (100%)** | [phase-02-docs-plans-and-truth-alignment.md](phase-02-docs-plans-and-truth-alignment.md) |
| 3 | **Phase 03:** API boundaries & developer ergonomics | **16** | 1 | 5 | 22 | In progress (77%) | [phase-03-api-boundaries-and-developer-ergonomics.md](phase-03-api-boundaries-and-developer-ergonomics.md) |
| 4 | **Phase 04:** Motion & animation consolidation | **16** | 0 | 3 | 19 | In progress (84%) | [phase-04-motion-and-animation-consolidation.md](phase-04-motion-and-animation-consolidation.md) |
| 5 | **Phase 05:** `internal/page`, `window` & `replay` cohesion | **16** | 1 | 5 | 22 | In progress (77%) | [phase-05-page-and-window-cohesion.md](phase-05-page-and-window-cohesion.md) |
| 6 | **Phase 06:** Components & module boundaries | **13** | 0 | 5 | 18 | In progress (72%) | [phase-06-components-and-module-boundaries.md](phase-06-components-and-module-boundaries.md) |
| 7 | **Phase 07:** Dependencies, binary size & platform readiness | **10** | 0 | 14 | 24 | In progress (42%) | [phase-07-dependencies-binary-size-and-platform-readiness.md](phase-07-dependencies-binary-size-and-platform-readiness.md) |
| — | **Total** | **101** | **4** | **42** | **147** | **71.4% resolved** | — |

## Summary of Remaining Open Work (42 `[ ]` rows + 4 `[~]` rows)

### 1. In-Repo Code & Refactor Work (Ready for Wave 5 Subagents — 18 rows)
- **Phase 03 (5 open rows):** `API-16`–`API-18`, `API-GATE-03`, `API-GATE-04` — relocate `examples/css-preview` internal motion tests into `internal/motion` / `internal/viewtransition` tests so `scripts/example-internal-allowlist.txt` drops from 10 pairs to `<= 1` pair (`desktop-cat`), plus optionally refactoring the 5 `[~]` `API-09` `.Testing()` example harnesses.
- **Phase 04 (3 open rows):** `MOT-09`, `MOT-16`, `MOT-GATE-03` — consolidate/organize the 153 merged files in `internal/motion/` into clean pipeline-stage files and run `make graphics-check` + motion benchmarks.
- **Phase 05 (5 open rows):** `COH-02`, `COH-08`, `COH-18`, `COH-GATE-02`, `COH-GATE-04` — audit and trim internal pass-through methods on `*Page` in `internal/page/` (from 360 toward `<= 180`), and run `make graphics-check` + `make soak-check`.
- **Phase 06 (5 open rows):** `MOD-10`, `MOD-14`, `MOD-GATE-02`–`MOD-GATE-04` — migrate unique test assertions out of the superseded `SKIP_OPEN` legacy input micro-examples (`bind-hooks`, `clipboard`, `controls`, `editing`, `input`, `states`) into `examples/input-lab` or `internal/page/form`, remove the redundant folders, and trim unused generated methods in `ui/properties_generated.go`.

### 2. CI / Release Artifact Decoupling (Phase 01 — 10 open rows)
- **`FOOT-06`, `FOOT-07`, `FOOT-09`:** Add `.github/workflows/pages.yml` (building all 9 `.wasm` demos in CI, including adding `wispr-flow` to `frontend/scripts/build-demo.mjs`), untrack the 10 `.wasm`/`.wasm.gz` binaries (~581.8 MB), and shrink `scripts/repo-size-allowlist.txt`.
- **`FOOT-11`–`FOOT-13`:** Publish the 8 reference Android APKs (230.85 MiB) to GitHub Releases/CI artifacts, untrack `*.apk`, and update `showcase.md` / `frontend/src/showcase.js` download links.
- **`FOOT-16`, `FOOT-GATE-01`, `FOOT-GATE-02`, `FOOT-GATE-04`:** Record final `.git` history decision (preserving published Go module tag commit immutability) and close Phase 01 gates.

### 3. Upstream Pins, Hardware Validation & Program Closure (Phase 07 — 14 open rows)
- **Upstream tags & binary weight (`DEP-02`, `DEP-03`, `DEP-08`, `DEP-09`, `DEP-GATE-01`, `DEP-GATE-02`):** Tag `blinkless` and `ebiten/v2` stable releases, decouple optional assets / file upstream `blinkless` issue to avoid pulling PDF/TeX/font packages into every GUI binary.
- **Native host & APK validation (`DEP-12`, `DEP-14`, `DEP-16`):** Real-host Windows `NATIVE-IME` validation, macOS & Windows desktop host validation records, and fresh `scripts/android-release.sh` build check.
- **Final closure battery (`DEP-17`–`DEP-20`, `DEP-GATE-04`):** Full `make test`, `make test-race`, `make lint`, `make graphics-check`, final before/after scoreboard, and reconciliation into [plans/README.md](../../README.md).

### 4. Deferred / Partial Rows (`[~]` — 4 rows)
- **`FOOT-08`** ([Phase 01](phase-01-repo-footprint-and-binary-hygiene.md)): Keep existing `.webp` demo captures intact at current quality per owner instruction.
- **`DOCS-10`** ([Phase 02](phase-02-docs-plans-and-truth-alignment.md)): Keep the 8 `compare-*.md` files separate because `frontend/` routes and tests each `#docs/compare-*` slug individually.
- **`API-09`** ([Phase 03](phase-03-api-boundaries-and-developer-ergonomics.md)): AST gate `TestNoUnauthorizedPageTestingCalls` is live; 5 pre-existing experimental example harnesses remain in `allowExamples`.
- **`COH-11`** ([Phase 05](phase-05-page-and-window-cohesion.md)): Keep `devtools*.go` and `replay*.go` in `internal/window` rather than introducing wide forwarding adapters back into `*shell`.

---

## Phase 1: Execution and Evidence Rules

### 1.1 Checklist conventions

- Use consistent row states across all phase files:
  - `[ ]` not started or not proven;
  - `[x]` implemented and validated with current evidence recorded beside the row;
  - `[~]` intentionally deferred or partial, with reason, owner boundary, and next gate.
- Never mark a row `[x]` on plan prose or a bare compile alone. Record the exact commit hash, command, exit status, and measured result (before/after counts, sizes, or test output).
- Keep non-documentation changes gated by `make lint` and `make test`. Changes touching paint, replay, motion, or windowing also require `make graphics-check` and `make test-race`.
- Documentation-only plan updates run `make docs-check plans-check` without claiming code/runtime verification.
- Preserve public API invariants in `AGENTS.md`: `ownframe.Page` stays wrapped with an unexported `*page.Page`, no unapproved `type X =` root aliases, and no new `examples/` imports of `ownframe/internal/`.

## Dependencies

- **Phase 01 before history/site changes:** Stop committing new `.wasm` and `.apk` artifacts before evaluating any git history rewrite.
- **Phase 02 (complete):** Honest pre-1.0 vs 1.0 positioning and documentation tiers are in place.
- **Phase 03 & Phase 04 before remaining Phase 05 (`COH-02`/`COH-08`):** Public `ownframe.Page` doors and `internal/motion` consolidation are landed; `*Page` internal method trimming in `internal/page/` can now proceed cleanly.
- **Phase 07 depends on upstream tags and OS hosts:** Moving `blinkless` and `ebiten/v2` off pseudo-versions requires tagged releases in those repositories, and Windows/macOS validation requires those hosts; mark blocked items explicitly rather than guessing.
