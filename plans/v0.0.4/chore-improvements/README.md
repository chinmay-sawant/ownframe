# v0.0.4 - Chore Improvements Plan

> **Parent:** [plans/v0.0.4/README.md](../README.md) (and [maintenance ledger](../../README.md))
> **Status:** In progress (129 complete, 9 partial, 9 open)

---

## Overview

This directory is the canonical phase-wise plan to address structural debt, repository footprint, API leaks, package fragmentation, documentation drift, binary weight, and pre-1.0 readiness gaps audited on `master` at baseline commit `da4f12cc` (2026-11-20), with Waves 1–4 merged into `master` via [PR #83](../pr/pr-chore-improvements-v003.md) (`63464cd0`) and Wave 5 completing all in-repo code and refactor phases (Phases 03, 04, 05, and 06).

Each phase file owns its task rows and evidence. This scoreboard compares
`da4f12cc` with the proposed local tree based on `b27c49c2`. File sizes
are Git blob bytes at baseline and proposed tracked file bytes at review.

| Measure | Baseline | Proposed |
| --- | ---: | ---: |
| Tracked `docs/` | **589,435,263 B** | **61,131,046 B** |
| Tracked `examples/` | **359,129,491 B** | **117,225,014 B** |
| Tracked WASM and gzip payloads | **10** | **0** |
| Tracked Android APKs | **8** | **0** |
| Root production Go files | **48** | **18** |
| Exported root Go methods | **149** | **145** |
| Flat Go files in `internal/page` | **316** | **52** |
| Flat Go files in `internal/window` | **406** | **70** |
| Flat Go files in `internal/replay` | **146** | **29** |
| Internal animation, motion, and transition package directories | **9** | **2** |
| `AGENTS.md` | **31,525 B** | **9,246 B** |
| Allowlisted example internal imports | **16** | **1** |
| Blinkless | **`v0.1.2` pseudo-version** | **`v0.1.2`** |
| Ebitengine | **`v2.11.0` alpha pseudo-version** | **`v2.10.6`** |
| Stripped desktop Hello binary | **about 33 MB** | **33,429,032 B** |

The local Git directory remains about **3.4 GiB**. Existing release tags and
historical blobs are retained. Ten media entries remain in the >5 MiB
allowlist after the 18 WASM and APK entries are removed.

The matched upstream experiment measured a stripped Hello build at
**33,416,808 B** by default and **26,952,424 B** with `blinkless_minimal`.
The production modules use the default Blinkless release. The
[artifact and host record](../../validation/2026-11-20-artifact-and-host-checks.md)
contains the commands and the open checks.

## Phase progress

| Phase | Complete | Partial | Open |
| --- | ---: | ---: | ---: |
| [Repository footprint](phase-01-repo-footprint-and-binary-hygiene.md) | 14 | 4 | 3 |
| [Docs and plan alignment](phase-02-docs-plans-and-truth-alignment.md) | 20 | 1 | 0 |
| [API boundaries](phase-03-api-boundaries-and-developer-ergonomics.md) | 22 | 0 | 0 |
| [Motion and animation](phase-04-motion-and-animation-consolidation.md) | 19 | 0 | 0 |
| [Page, window, and replay](phase-05-page-and-window-cohesion.md) | 21 | 1 | 0 |
| [Components and modules](phase-06-components-and-module-boundaries.md) | 18 | 0 | 0 |
| [Dependencies and platform readiness](phase-07-dependencies-binary-size-and-platform-readiness.md) | 15 | 3 | 6 |
| **Total** | **129** | **9** | **9** |


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
