# v0.0.4 - Phase 02: Documentation, Plans, and Truth Alignment

> **Parent:** [README.md](README.md)
> **Status:** Complete (`DOCS-10` explicitly deferred `[~]` due to `frontend/` compare routes)
> **Estimated effort:** 2–3 days

---

## Overview

At commit `da4f12cc`, documentation and planning artifacts have outgrown the library's current stage (`v0.0.3`): **69 files in `documentation/`** (including a 225 KB `compatibility-matrix.md`), **277 files in `plans/`**, **28 directories in `skills/`**, and a **31.5 KB `AGENTS.md`**. At this volume, docs drift from code—for example, `AGENTS.md` still states `github.com/chinmay-sawant/blinkless v0.0.4-0.20261025105255-96d7b9d95338` while `go.mod` is on `v0.1.2-0.20261118190733-7abfce32c375`—and newcomers have no clear 3-page path to get started.

In addition, `README.md` and `AGENTS.md` use "1.0 ships native Go apps on desktop (Linux, Windows, macOS) and Android" phrasing even though the library is still `v0.0.3` and lacks Windows UIA, macOS VoiceOver, Android TalkBack, iOS, and native Wayland clipboard support. This phase restores single-source-of-truth accuracy, automates drift checks, and trims documentation bloat.

## Executive Summary

- Fix confirmed pin/version drift across `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, and `documentation/`, and extend `scripts/check-doc-snippets.py` to enforce them automatically.
- Replace ambiguous "1.0 ships..." claims with clear `v0.0.x` current-status statements and an explicit pre-1.0 vs 1.0 roadmap boundary.
- Restructure `documentation/README.md` into a 3-document Quick Start plus categorized guides and generated references; merge or archive redundant notes.
- Trim `AGENTS.md` from 31.5 KB to a crisp, high-signal engineering ruleset (<12 KB) and organize historical `plans/` cleanly under archive indexes.

## Current Source Baseline (`da4f12cc`)

| Artifact | Baseline Measurement | Confirmed Issue |
|---|---|---|
| `AGENTS.md` | **31,525 bytes** | Cites stale `blinkless v0.0.4-0.20261025105255-96d7b9d95338` (line 47) vs `go.mod` `v0.1.2-0.20261118190733-7abfce32c375`; embeds long per-example narratives that drift |
| `documentation/` | **69 files** (840 KB) | Flat folder mixing onboarding guides, 8 separate `compare-*.md` files, historical review notes, and a 225 KB `compatibility-matrix.md` |
| `plans/` | **277 files** | Mixes historical v0.0.1/v0.0.2/PR records, superseded `prod-ready-task`, completed v0.0.4 tracks, and active work |
| `README.md` | Line 13, 119 | Says "1.0 ships native Go apps..." while module is `v0.0.2` / `v0.0.3` cycle and lacks Windows/macOS/Android a11y and native Wayland clipboard |

---

## Phase 2: Documentation, Plans, and Truth Alignment

### 2.1 Fix confirmed documentation drift and automate pin checks

- [x] `DOCS-01` Update `AGENTS.md` line 47 to remove the hardcoded stale `blinkless v0.0.4-0.20261025105255-96d7b9d95338` version string; point to `go.mod` as the single source of truth for exact `blinkless`, `ebiten/v2`, and Go version pins. Evidence: updated `AGENTS.md` line 47; `grep -n "v0\.0\.4-0\.2026" AGENTS.md` returns empty.
- [x] `DOCS-02` Audit `README.md`, `CONTRIBUTING.md`, `documentation/*.md`, and `examples/readme.md` for hardcoded Go (`1.26.*`), `blinkless`, `ebiten`, and `golangci-lint` version strings that disagree with `go.mod`, `examples/go.mod`, and `Makefile`. Evidence: fixed stale `1.26.10` -> `1.26.11` in `CONTRIBUTING.md` line 7 and stale `blinkless` pin in `AGENTS.md` line 47; verified `v2.15.0` matches `Makefile` `GOLANGCI_LINT_VERSION`.
- [x] `DOCS-03` Extend `scripts/check-doc-snippets.py` (via `scripts/doc_claims.py`) to scan `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, and `documentation/android-runtime.md` for hardcoded `blinkless` / `go 1.xx.x` pins and fail if any contradict `go.mod`. Evidence: added `check_doc_pins` in `scripts/doc_claims.py` and `test_stale_doc_pins_fail` in `scripts/test_doc_claims.py`; `make tools-check` and `make docs-check` exit 0.
- [x] `DOCS-04` Audit `AGENTS.md` against current codebase reality for `internal/page`, `internal/window`, `examples/`, and `components/` (verifying every path and method name cited still exists). Evidence: audited all path-like tokens in rewritten `AGENTS.md` against the checkout; zero dead file references.

### 2.2 Align "1.0" vs `v0.0.x` support claims

- [x] `DOCS-05` Revise `README.md` (lines 13 and 119) to replace "1.0 ships..." with honest current-cycle phrasing (`v0.0.x` preview cycle; no `v1.0` tag yet), explicitly linking unshipped pre-1.0 gaps. Evidence: updated `README.md` lines 13 and 119; `make docs-check` exit 0.
- [x] `DOCS-06` Update `documentation/platforms.md` and `documentation/features.md` with a concise "Current (`v0.0.x`) vs Planned (`v1.0`)" matrix covering desktop windowing, Android, IME, clipboard, and accessibility. Evidence: added `Current (v0.0.x) vs planned (v1.0)` table in `documentation/platforms.md` and aligned intro in `documentation/features.md` to match `plans/README.md`.
- [x] `DOCS-07` Update `AGENTS.md` wording ("1.0 native hosts are...") to match the revised `documentation/platforms.md` terminology so agent prompts and public docs never contradict each other. Evidence: updated `AGENTS.md` line 7 to `Target native hosts (currently in the v0.0.x preview cycle; no v1.0 tag yet)`.
- [x] `DOCS-08` Verify `documentation/accessibility-native.md` and `documentation/clipboard.md` clearly state platform prerequisites upfront and are linked from `README.md`. Evidence: updated `documentation/accessibility-native.md` intro to `Currently (v0.0.x; no v1.0 tag yet)`, confirmed `documentation/clipboard.md` states `wl-copy`/`wl-paste` on line 3, and linked `accessibility-native.md` directly from `README.md`.

### 2.3 Rationalize `documentation/` (69 files) and create a 3-doc onboarding path

- [x] `DOCS-09` Redesign `documentation/README.md` into four explicit tiers (Start Here 3 docs, Core Guides, Comparisons & Reference, Historical & Point-in-Time Notes) with zero orphan files. Evidence: updated `documentation/README.md`; verified all 66 sibling `.md` files linked with 0 orphans and 0 broken relative links under `make docs-check`.
- [~] `DOCS-10` Consolidate the 8 thin `compare-*.md` files into one `documentation/comparisons.md`. Deferred: `frontend/src/compareDisplay.js`, `frontend/src/compareDisplay.test.js`, `frontend/src/Benchmarks.jsx`, `frontend/scripts/check-site.mjs`, and `documentation/snippets.json` route and test each `#docs/compare-*` slug individually. Kept files intact and grouped under Section 3 of `documentation/README.md`; reconsider when refactoring `frontend/` comparison routes.
- [x] `DOCS-11` Review point-in-time historical reports in `documentation/` (`lint-findings.md`, `loglens-profiling.md`, `mermaid-code-review.md`, `migrate-experimental.md`, `rebrand.md`, `maintenance-review.md`) and mark their status/scope at the top. Evidence: added top-of-file `> **Status:**` banners to all 5 unbadged files (`lint-findings.md` already had its historical banner) and grouped all 6 under Section 4 of `documentation/README.md`.
- [x] `DOCS-12` Review `documentation/compatibility-matrix.md` (225 KB) provenance and codegen usage. Evidence: verified top-of-file header cites upstream Blinkless `7abfce32c375` and `testdata/css/catalog/properties.json`; confirmed downstream generators `scripts/gen-ui-properties/main.go` and `examples/blinkless-all/props_gen.go` consume it directly; documented in `documentation/README.md`.
- [x] `DOCS-13` Check `frontend/` website doc loaders and tests to ensure doc changes do not break website tests. Evidence: `npm --prefix frontend test` passed (13/13 tests green).

### 2.4 Trim `AGENTS.md` and organize historical `plans/` (277 files)

- [x] `DOCS-14` Refactor `AGENTS.md` from 31.5 KB down to <=12 KB while preserving all engineering invariants, build/test/lint rules, public API constraints, and pointers to canonical docs. Evidence: trimmed `AGENTS.md` from 31,525 bytes to **9,053 bytes** (71% reduction; <=12,288-byte target) with zero dead path references.
- [x] `DOCS-15` Audit `plans/` (277 files) and distinguish active ledgers from archived plan trees in `plans/README.md`. Evidence: added `## Plan trees (active vs archived)` table to `plans/README.md`; `make plans-check` passed (24 rows with distinct states).
- [x] `DOCS-16` Audit `skills/` (28 directories) and inventory active ownframe workflows vs vendored external templates. Evidence: created `skills/README.md` categorizing all active and vendored `ponytail*` skill packs.
- [x] `DOCS-17` Verify `scripts/check-plans.py` passes and add a broken-relative-link checker for `README.md`, `AGENTS.md`, `documentation/README.md`, and `plans/README.md`. Evidence: added `check_relative_links` in `scripts/doc_claims.py` and `test_broken_relative_links_fail` in `scripts/test_doc_claims.py`; `make docs-check`, `make plans-check`, and `make tools-check` (54 tests) exit 0.

---

## Dependencies

- `DOCS-10` and `DOCS-11` must update `frontend/` link references before merging so `make site-check` stays green.
- `DOCS-14` (trimming `AGENTS.md`) must preserve the exact public API freeze rules referenced by `alias_gate_test.go` and `api_internal_gate_test.go`.

## Risks

- **Broken external or website links:** Moving files in `documentation/` could break links from `frontend/` or GitHub issues/PRs. Verify all references with `grep -rn "documentation/"` before moving any file.
- **Over-trimming `AGENTS.md`:** Removing critical non-obvious invariants (like `GOWORK=off` for gomobile, `clipboard.UseMemory` in tests, or not running `go build ./...`) would hurt agent and contributor velocity. Keep all operational rules; move only narrative example walkthroughs.

## Closure Gates

- [x] `DOCS-GATE-01` Zero stale Go / `blinkless` / `ebiten` version pins across `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, and `documentation/`. Evidence: `make docs-check` (with new `check_doc_pins`) passed (39 snippets compiled, 0 pin mismatches).
- [x] `DOCS-GATE-02` `AGENTS.md` trimmed to <=12 KB with example narratives referenced via per-example docs. Evidence: `wc -c AGENTS.md` = 9,053 bytes (<=12,288).
- [x] `DOCS-GATE-03` `documentation/README.md` presents a clear 3-doc Quick Start and zero orphan files. Evidence: all 66 sibling `.md` files linked; 0 orphans.
- [x] `DOCS-GATE-04` `make docs-check`, `make plans-check`, `make tools-check`, and `npm --prefix frontend test` exit 0. Evidence: 39 snippets compiled, 24 ledger rows valid, 54 Python tests + Go scripts tests OK, 13/13 frontend unit tests passed.
