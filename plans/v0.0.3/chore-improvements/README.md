# v0.0.3 - Chore Improvements Plan

> **Parent:** [plans/v0.0.3/README.md](../README.md) (and [maintenance ledger](../../README.md))
> **Status:** In progress (baseline `da4f12cc`; **91 `[x]`**, **3 `[~]`**, **53 `[ ]`** of 147 rows — Phase 02 complete; `page` 317->52, `window` 406->70, `replay` 146->29, root 48->18 files / 149->144 methods, 7 chart skins decoupled, `viewtransition` & `frame` decoupled, `ui-gen-check` + `release.yml` size reporting live)
> **Estimated effort:** 3–5 weeks across 7 incremental, independently mergeable phases

---

## Overview

This directory is the canonical phase-wise plan to address structural debt, repository footprint, API leaks, package fragmentation, documentation drift, binary weight, and pre-1.0 readiness gaps audited on `master` at commit `da4f12cc` (2026-11-20).

Creating these phase files records the baseline and work breakdown; it does not complete any checklist row. Each linked phase file owns its atomic task rows. This index holds no duplicate task checklist.

## Executive Summary

Measured repository baseline at `da4f12cc`:

| Area | Baseline Measurement | Target Direction | Owning Phase |
|---|---|---|---|
| **Repository & Git Footprint** | **3.4 GB `.git`**, 668 MB `docs/` (9 committed 52–66 MB `.wasm`/`.wasm.gz` files), 207 MB of 7 committed Android APKs, 7.0 GB local `temp/`, ~360 MB stray `*.test` binaries | Decouple built `.wasm` and `.apk` binaries from main git history; add file-size guardrails; clean workspace | [Phase 01](phase-01-repo-footprint-and-binary-hygiene.md) |
| **Docs, Plans & 1.0 Claims** | **69 files in `documentation/`**, 277 files in `plans/`, 31.5 KB `AGENTS.md` (with stale `blinkless v0.0.4-...` pin vs `go.mod` `v0.1.2-...`), premature "1.0" phrasing | Sync all pins with automated checks; establish 3-doc onboarding path; archive old plans; align 1.0 scope claims | [Phase 02](phase-02-docs-plans-and-truth-alignment.md) |
| **Public API & Ergonomics** | **63 root `.go` files**, 149 exported methods across `Page` and 7 doors; global `window.*` calls on instance `page.Window()`; 4-path update model; 16 allowlisted example internal imports | Separate global host calls from page instances; consolidate root micro-files; extract list/chrome helpers; drain allowlist | [Phase 03](phase-03-api-boundaries-and-developer-ergonomics.md) |
| **Motion Package Sprawl** | **10 motion/transition `internal/` packages** + `internal/frame` + `internal/previewwindow`; non-test preview packages importing `testing` and root `ownframe` | Move test/preview helpers out of runtime `internal/`; merge core CSS motion chain to 2–3 packages; clarify `frame` vs `motion` seam | [Phase 04](phase-04-motion-and-animation-consolidation.md) |
| **Page & Window Cohesion** | **~200,609 core Go LOC**; `internal/page` has 507 `.go` files (316 flat in root) and 360 `*Page` methods; `internal/window` has 409 `.go` files (406 flat in root); `internal/replay` has 146 files | Consolidate legacy micro-files into cohesive domain files; narrow `*Page` coordinator surface; partition `internal/window` by ownership | [Phase 05](phase-05-page-and-window-cohesion.md) |
| **Components & Module Scope** | **7 JS chart-library skins** (`apexcharts`..`tremor`) inside root module used only by `examples/builder/ui-builder`; 10,198-line `ui/properties_generated.go`; 84 folders / 2.5 GB in `examples/` | Decouple chart skins from core library module or share paint core; trim generated UI code; prune superseded micro-examples | [Phase 06](phase-06-components-and-module-boundaries.md) |
| **Dependencies, Binary Size & 1.0 Gaps** | **`go 1.26.11`**, pseudo-versions of `blinkless` and `ebiten/v2` alpha, 43–46 MB compiled test binaries, no native Windows/macOS/Android a11y adapters or native Wayland clipboard | Move to tagged upstream pins; profile and trim binary bloat; define explicit 1.0 platform & a11y exit criteria | [Phase 07](phase-07-dependencies-binary-size-and-platform-readiness.md) |

## Phase Index

Phases are ordered by risk and leverage: stop ongoing git/workspace bloat and fix documentation truth first, then clarify public API seams before consolidating internal packages, then refine module boundaries and release/platform readiness.

| Order | Phase | Focus | Primary Gate | Canonical Checklist |
|---|---|---|---|---|
| 1 | **Phase 01** | Repo footprint & binary hygiene | No tracked `.wasm`/`.apk` build outputs in `master`; size check in CI | [phase-01-repo-footprint-and-binary-hygiene.md](phase-01-repo-footprint-and-binary-hygiene.md) |
| 2 | **Phase 02** | Docs, plans & truth alignment | `make docs-check plans-check` green; zero pin drift; honest 1.0 scope | [phase-02-docs-plans-and-truth-alignment.md](phase-02-docs-plans-and-truth-alignment.md) |
| 3 | **Phase 03** | API boundaries & developer ergonomics | No singleton `window.*` calls on `Page.Window()`; `alias_gate_test.go` + `make internal-import-check` green | [phase-03-api-boundaries-and-developer-ergonomics.md](phase-03-api-boundaries-and-developer-ergonomics.md) |
| 4 | **Phase 04** | Motion & animation consolidation | <=3 core motion packages; zero `testing` imports in non-test `internal/` | [phase-04-motion-and-animation-consolidation.md](phase-04-motion-and-animation-consolidation.md) |
| 5 | **Phase 05** | `internal/page`, `window` & `replay` cohesion | >60% reduction in flat micro-files; `make test-race` + `make graphics-check` + `make soak-check` green | [phase-05-page-and-window-cohesion.md](phase-05-page-and-window-cohesion.md) |
| 6 | **Phase 06** | Components & module boundaries | Lean root module; `make module-check examples-check` green | [phase-06-components-and-module-boundaries.md](phase-06-components-and-module-boundaries.md) |
| 7 | **Phase 07** | Dependencies, binary size & platform readiness | Tagged upstream deps; measured binary budget; explicit 1.0 exit criteria | [phase-07-dependencies-binary-size-and-platform-readiness.md](phase-07-dependencies-binary-size-and-platform-readiness.md) |

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
- **Phase 02 before Phase 03/06:** Agree on honest pre-1.0 vs 1.0 positioning and documentation layout before reorganizing public APIs or moving `components/` modules.
- **Phase 03 before Phase 05:** Clarify the public `ownframe.Page` / `Window` / `Edit` / `Form` contracts before shrinking `internal/page` and `internal/window` internals so refactors align with the final public seams.
- **Phase 04 before Phase 05:** Collapsing the 10 motion packages simplifies `internal/page/motion_wire.go` and `internal/window` motion hooks prior to page/window consolidation.
- **Phase 06 independent after Phase 03:** Component module splits and `examples/` cleanup can run in parallel with Phase 05 once public helpers for virtual lists/pinned chrome are settled.
- **Phase 07 depends on upstream tags:** Moving `blinkless` and `ebiten/v2` off pseudo-versions requires tagged releases in those repositories; mark blocked items explicitly rather than guessing.
