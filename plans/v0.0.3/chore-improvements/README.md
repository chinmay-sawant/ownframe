# v0.0.3 - Chore Improvements Plan

> **Parent:** [plans/v0.0.3/README.md](../README.md) (and [maintenance ledger](../../README.md))
> **Status:** In progress — **Phases 02, 03, 04, 05 & 06 Complete** (**121 `[x]`** · **5 `[~]`** · **21 `[ ]`** of 147 rows = **85.7% resolved**)
> **Estimated effort:** Only external/CI artifact decoupling (Phase 01: 9 rows) and upstream tags/host validation (Phase 07: 12 rows) remain

---

## Overview

This directory is the canonical phase-wise plan to address structural debt, repository footprint, API leaks, package fragmentation, documentation drift, binary weight, and pre-1.0 readiness gaps audited on `master` at baseline commit `da4f12cc` (2026-11-20), with Waves 1–4 merged into `master` via [PR #83](../../v0.0.4/pr/pr-chore-improvements-v003.md) (`63464cd0`) and Wave 5 completing all in-repo code and refactor phases (Phases 03, 04, 05, and 06).

Each linked phase file owns its atomic task rows and evidence. This index holds the before/after scoreboard, phase progress breakdown, and the roadmap for the remaining 21 external/CI open rows without duplicating active checklists.

## Executive Summary (Baseline `da4f12cc` vs Current Wave 5)

| Area | Baseline (`da4f12cc`) | Current after Wave 5 | Remaining Direction | Owning Phase |
|---|---|---|---|---|
| **Repository & Git Footprint** | **3.4 GB `.git`**, 668 MB `docs/` (10 tracked `.wasm`/`.wasm.gz` = 581.8 MB), 8 tracked APKs (230.85 MiB), 7.0 GB `temp/`, ~360 MB stray `*.test` binaries | Stray `*.test`/`mem.out`/`Händen`/`examples/temp` removed; **`make repo-size-check` (5 MiB gate)** live in `make lint`; `.git` blob & CI/APK audits complete; `.webp` kept (`FOOT-08` `[~]`) and `.git` tag immutability decision recorded (`FOOT-16` `[~]`) | Move `.wasm` builds to `pages.yml` (`FOOT-06`–`09`) and 8 APKs to GitHub Releases (`FOOT-11`–`13`) | [Phase 01](phase-01-repo-footprint-and-binary-hygiene.md) |
| **Docs, Plans & 1.0 Claims** | **69 files in `documentation/`**, 277 in `plans/`, 31.5 KB `AGENTS.md` (stale `blinkless` pin), premature "1.0" phrasing | **Phase Complete:** `AGENTS.md` trimmed **31.5 KB -> 9.0 KB**; 4-tier `documentation/README.md` (0 orphans); automated pin & relative-link checks in `make docs-check`; honest `v0.0.x` vs `v1.0` matrix | Complete (`DOCS-10` `[~]` keeps `compare-*.md` separate for `frontend/` routes) | [Phase 02](phase-02-docs-plans-and-truth-alignment.md) |
| **Public API & Ergonomics** | **48 non-test root `.go` files**, 149 exported methods; global `window.*` calls on `Page.Window()`; 16 allowlisted example internal imports | **Phase Complete:** Root non-test `.go` files **48 -> 18**; exported methods **149 -> 144**; `Page.Window()` bound to `windowHost` (0 `internal/window` imports in `page_window.go`); `ui/virtual` + `updates.md` + `ManualOverlap` diagnostic added; allowlist drained **16 -> 1** (`desktop-cat`) | Complete (`API-09` `[~]` freezes the 5 experimental `.Testing()` harnesses) | [Phase 03](phase-03-api-boundaries-and-developer-ergonomics.md) |
| **Motion Package Sprawl** | **10 motion/transition `internal/` packages** + `internal/frame` + `internal/previewwindow`; non-test preview packages importing `testing` and root `ownframe` | **Phase Complete:** Collapsed 7 CSS motion packages into 1 `internal/motion` and consolidated **153 -> 34 `.go` files** (18 prod + 16 test, 365/365 tests preserved); moved 3 preview packages to `cmd/internal/`; decoupled `internal/viewtransition` from root `ownframe`; unified `frame.OpBox` with `dirty.OpBounds` | Complete (19/19 `[x]`) | [Phase 04](phase-04-motion-and-animation-consolidation.md) |
| **Page, Window & Replay Cohesion** | `internal/page` **317** flat files (**360 `*Page` methods**); `internal/window` **406** flat files; `internal/replay` **146** files (**869** flat files total) | **Phase Complete:** Flat `.go` files reduced **869 -> 151 (-718 micro-files)**: `internal/page` **317 -> 52**, `internal/window` **406 -> 70**, `internal/replay` **146 -> 29** with **1,106/1,106 tests preserved**; `*Page` methods trimmed **360 -> 170 (-190)** | Complete (`COH-11` `[~]` keeps `devtools*.go`/`replay*.go` in `internal/window`) | [Phase 05](phase-05-page-and-window-cohesion.md) |
| **Components & Module Scope** | **7 JS chart-library skins** (7,331 Go LOC) in root module; 10,198-line `ui/properties_generated.go`; 75 `examples/` folders | **Phase Complete:** Moved all 7 chart skins to `examples/builder/charts/` (root archive **1,709 -> 1,264 entries**); trimmed `ui/properties_generated.go` **10,198 -> 2,053 LOC (-79.9%)** & `ui.a` **12.37 -> 4.24 MB (-65.7%)**; migrated & removed 6 superseded `SKIP_OPEN` legacy input demos (69 files) | Complete (18/18 `[x]`) | [Phase 06](phase-06-components-and-module-boundaries.md) |
| **Dependencies, Binary Size & 1.0 Gaps** | **`go 1.26.11`**, pseudo-versions of `blinkless` and `ebiten/v2` alpha, 45 MB unstripped / 33 MB stripped binaries, silent Wayland clipboard failure | Audited `go mod graph`, `make vuln` (0 code vulns), symbol profile & `//go:embed` footprint (`ownframe/internal` = 42.6 KB vs `blinkless` fonts = 3.09 MB); added `ErrWaylandToolsMissing` + `clipboard.Diagnose()`; documented 2-module release tag sequence & CI binary-size logging | Tag upstream `blinkless`/`ebiten` (`DEP-02`–`03`), track upstream `blinkless` split (`DEP-09`), validate Windows/macOS hosts (`DEP-12`, `DEP-14`), final suite (`DEP-17`–`20`) | [Phase 07](phase-07-dependencies-binary-size-and-platform-readiness.md) |

## Phase Progress Breakdown (`121 [x]` · `5 [~]` · `21 [ ]` of 147 rows)

| Order | Phase | `[x]` Done | `[~]` Deferred/Partial | `[ ]` Open | Total | Status | Canonical Checklist |
|---|---|---:|---:|---:|---:|---|---|
| 1 | **Phase 01:** Repo footprint & binary hygiene | **10** | 2 | 9 | 21 | In progress (57%) | [phase-01-repo-footprint-and-binary-hygiene.md](phase-01-repo-footprint-and-binary-hygiene.md) |
| 2 | **Phase 02:** Docs, plans & truth alignment | **20** | 1 | **0** | 21 | **Complete (100%)** | [phase-02-docs-plans-and-truth-alignment.md](phase-02-docs-plans-and-truth-alignment.md) |
| 3 | **Phase 03:** API boundaries & developer ergonomics | **21** | 1 | **0** | 22 | **Complete (100%)** | [phase-03-api-boundaries-and-developer-ergonomics.md](phase-03-api-boundaries-and-developer-ergonomics.md) |
| 4 | **Phase 04:** Motion & animation consolidation | **19** | 0 | **0** | 19 | **Complete (100%)** | [phase-04-motion-and-animation-consolidation.md](phase-04-motion-and-animation-consolidation.md) |
| 5 | **Phase 05:** `internal/page`, `window` & `replay` cohesion | **21** | 1 | **0** | 22 | **Complete (100%)** | [phase-05-page-and-window-cohesion.md](phase-05-page-and-window-cohesion.md) |
| 6 | **Phase 06:** Components & module boundaries | **18** | 0 | **0** | 18 | **Complete (100%)** | [phase-06-components-and-module-boundaries.md](phase-06-components-and-module-boundaries.md) |
| 7 | **Phase 07:** Dependencies, binary size & platform readiness | **12** | 0 | 12 | 24 | In progress (50%) | [phase-07-dependencies-binary-size-and-platform-readiness.md](phase-07-dependencies-binary-size-and-platform-readiness.md) |
| — | **Total** | **121** | **5** | **21** | **147** | **85.7% resolved** | — |

## Summary of Remaining Open Work (21 `[ ]` rows + 5 `[~]` rows)

### 1. All In-Repo Code & Refactor Phases Are Complete (Phases 02, 03, 04, 05, 06 = 100%)
- **Phase 02 (21/21 resolved):** `AGENTS.md` trimmed 31.5 KB -> 9.0 KB; 4-tier `documentation/README.md` with 0 orphan guides; honest `v0.0.x` vs `v1.0` support claims.
- **Phase 03 (22/22 resolved):** Root non-test `.go` files consolidated 48 -> 18; `Page.Window()` bound via `windowHost`; `ui/virtual` and `documentation/updates.md` shipped; `scripts/example-internal-allowlist.txt` drained **16 -> 1** (`desktop-cat`).
- **Phase 04 (19/19 resolved):** 7 CSS motion packages collapsed into `internal/motion` and consolidated **153 -> 34 `.go` files** (18 prod + 16 test, 365/365 tests preserved); 3 preview packages moved to `cmd/internal/`; `internal/viewtransition` decoupled from root `ownframe`.
- **Phase 05 (22/22 resolved):** `internal/page`, `internal/window`, and `internal/replay` consolidated **869 -> 151 flat `.go` files** (-718 micro-files, 1,106/1,106 tests preserved) and `*Page` methods trimmed **360 -> 170** (`<= 180`).
- **Phase 06 (18/18 resolved):** 7 chart skins moved to `examples/builder/charts/` (root archive **1,709 -> 1,264 entries**); `ui/properties_generated.go` trimmed **10,198 -> 2,053 LOC (-79.9%)**; 6 superseded `SKIP_OPEN` legacy input demos migrated into `examples/input-lab` and deleted.

### 2. CI / Release Artifact Decoupling (Phase 01 — 9 open rows)
- **`FOOT-06`, `FOOT-07`, `FOOT-09`:** Add `.github/workflows/pages.yml` (building all 9 `.wasm` demos in CI, including adding `wispr-flow` to `frontend/scripts/build-demo.mjs`), untrack the 10 `.wasm`/`.wasm.gz` binaries (~581.8 MB), and shrink `scripts/repo-size-allowlist.txt`.
- **`FOOT-11`–`FOOT-13`:** Publish the 8 reference Android APKs (230.85 MiB) to GitHub Releases/CI artifacts, untrack `*.apk`, and update `showcase.md` / `frontend/src/showcase.js` download links.
- **`FOOT-GATE-01`, `FOOT-GATE-02`, `FOOT-GATE-04`:** Close Phase 01 gates once `.wasm` and `.apk` binaries are untracked.

### 3. Upstream Pins, Hardware Validation & Program Closure (Phase 07 — 12 open rows)
- **Upstream tags & `blinkless` package split (`DEP-02`, `DEP-03`, `DEP-09`, `DEP-GATE-01`):** Tag `blinkless` and `ebiten/v2` stable releases, and file/track upstream `blinkless` issue to avoid pulling PDF/TeX/font packages into every GUI binary.
- **Native host & APK validation (`DEP-12`, `DEP-14`, `DEP-16`):** Real-host Windows `NATIVE-IME` validation, macOS & Windows desktop host validation records, and fresh `scripts/android-release.sh` build check.
- **Final closure battery (`DEP-17`–`DEP-20`, `DEP-GATE-04`):** Full `make test`, `make test-race`, `make lint`, `make graphics-check`, final before/after scoreboard, and reconciliation into [plans/README.md](../../README.md).

### 4. Deferred / Partial Rows (`[~]` — 5 rows)
- **`FOOT-08`** ([Phase 01](phase-01-repo-footprint-and-binary-hygiene.md)): Keep existing `.webp` demo captures intact at current quality per owner instruction.
- **`FOOT-16`** ([Phase 01](phase-01-repo-footprint-and-binary-hygiene.md)): Do not rewrite `.git` history across published Go module tag commits (`v0.0.1`, `v0.0.2`) to preserve `sum.golang.org` immutability; rely on `make repo-size-check` and `--filter=blob:none` shallow clones.
- **`DOCS-10`** ([Phase 02](phase-02-docs-plans-and-truth-alignment.md)): Keep the 8 `compare-*.md` files separate because `frontend/` routes and tests each `#docs/compare-*` slug individually.
- **`API-09`** ([Phase 03](phase-03-api-boundaries-and-developer-ergonomics.md)): AST gate `TestNoUnauthorizedPageTestingCalls` is live; 5 pre-existing experimental example harnesses remain frozen in `allowExamples`.
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
