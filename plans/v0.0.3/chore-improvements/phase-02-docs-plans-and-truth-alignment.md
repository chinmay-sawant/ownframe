# v0.0.3 - Phase 02: Documentation, Plans, and Truth Alignment

> **Parent:** [README.md](README.md)
> **Status:** Open (baseline measured at `da4f12cc`)
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
- [ ] `DOCS-04` Audit `AGENTS.md` against current codebase reality for `internal/page`, `internal/window`, `examples/`, and `components/` (e.g., verify every path and method name cited in `AGENTS.md` still exists). Proof: script or grep audit with zero dead file/symbol references in `AGENTS.md`.

### 2.2 Align "1.0" vs `v0.0.x` support claims

- [x] `DOCS-05` Revise `README.md` (lines 13 and 119) to replace "1.0 ships..." with honest current-cycle phrasing (`v0.0.x` preview cycle; no `v1.0` tag yet), explicitly linking unshipped pre-1.0 gaps. Evidence: updated `README.md` lines 13 and 119; `make docs-check` exit 0.
- [x] `DOCS-06` Update `documentation/platforms.md` and `documentation/features.md` with a concise "Current (`v0.0.x`) vs Planned (`v1.0`)" matrix covering desktop windowing, Android, IME, clipboard, and accessibility. Evidence: added `Current (v0.0.x) vs planned (v1.0)` table in `documentation/platforms.md` and aligned intro in `documentation/features.md` to match `plans/README.md`.
- [x] `DOCS-07` Update `AGENTS.md` wording ("1.0 native hosts are...") to match the revised `documentation/platforms.md` terminology so agent prompts and public docs never contradict each other. Evidence: updated `AGENTS.md` line 7 to `Target native hosts (currently in the v0.0.x preview cycle; no v1.0 tag yet)`.
- [x] `DOCS-08` Verify `documentation/accessibility-native.md` and `documentation/clipboard.md` clearly state platform prerequisites upfront and are linked from `README.md`. Evidence: updated `documentation/accessibility-native.md` intro to `Currently (v0.0.x; no v1.0 tag yet)`, confirmed `documentation/clipboard.md` states `wl-copy`/`wl-paste` on line 3, and linked `accessibility-native.md` directly from `README.md`.

### 2.3 Rationalize `documentation/` (69 files) and create a 3-doc onboarding path

- [ ] `DOCS-09` Redesign `documentation/README.md` into four explicit tiers:
  1. **Start Here (3 docs):** Quick Start (`screen.md` or new `quickstart.md`), Interaction & Forms (`binding.md`), Layout & Platforms (`platforms.md`).
  2. **Core Guides:** Windowing, theming, motion, frames, scrolling, testing, packaging, security.
  3. **Comparisons & Reference:** Consolidated comparisons + generated `compatibility-matrix.md`.
  4. **Internal / Historical Notes:** Profiling captures, lint findings, rebrand, migration notes.
  Proof: `documentation/README.md` lists every file in `documentation/` in one of the four tiers with zero orphan docs.
- [ ] `DOCS-10` Consolidate the 8 thin `compare-*.md` files (`compare-ebitengine.md`, `compare-electron.md`, `compare-fyne.md`, `compare-gio.md`, `compare-rust-gpui.md`, `compare-syntax.md`, `compare-wails.md`, `compare-webview.md`, each 1–2.8 KB) into one structured `documentation/comparisons.md` (keeping redirect stubs or updating all internal/website links). Proof: total `documentation/*.md` count drops by >=7 and zero broken markdown links (`scripts/check-doc-snippets.py` / link check green).
- [ ] `DOCS-11` Review point-in-time historical reports in `documentation/` (`lint-findings.md`, `loglens-profiling.md`, `mermaid-code-review.md`, `migrate-experimental.md`, `rebrand.md`, `maintenance-review.md`) and either move them under `documentation/archive/` or clearly mark their date and scope at the top. Proof: every non-evergreen doc carries a top-of-file status banner or lives in `documentation/archive/`.
- [ ] `DOCS-12` Review `documentation/compatibility-matrix.md` (225 KB): verify how it is generated, confirm it is not hand-edited, and add a top-of-file generator header + summary table so humans don't have to scroll 225 KB to understand CSS coverage. Proof: generator command documented and reproducible.
- [ ] `DOCS-13` Check `frontend/` website doc loaders to ensure any moved or merged `documentation/` files do not break Vite builds or broken-link tests. Proof: `make site-check` exits 0.

### 2.4 Trim `AGENTS.md` and organize historical `plans/` (277 files)

- [ ] `DOCS-14` Refactor `AGENTS.md` from 31.5 KB down to <=12 KB by removing long prose descriptions of individual examples (`examples/code-view`, `examples/telegram`, `examples/spotify-player`, `examples/desktop-cat`, `examples/wispr-flow-dashboard` — move those into each example's own `README.md`) while retaining all engineering invariants, build/test/lint rules, public API constraints, and pointers to canonical docs. Proof: `wc -c AGENTS.md` <= 12,288 bytes and all build/lint/API rules preserved.
- [ ] `DOCS-15` Audit `plans/` (277 files): group completed historical folders (`v0.0.1`, `v0.0.2`, `PR`, `v0.0.3/pr/`, superseded `v0.0.3/prod-ready-task`) under clear index summaries in `plans/README.md` / `plans/v0.0.3/README.md` so contributors can immediately distinguish active ledgers from closed history. Proof: `plans/README.md` clearly lists active vs archived plan trees and `make plans-check` passes.
- [ ] `DOCS-16` Audit `skills/` (28 directories): document which skills are active ownframe workflows vs legacy/external templates (e.g., `ponytail*` / CodeHound references in `skills/phase-wise-checklist/SKILLS.md`), and remove or archive unused ones. Proof: `skills/README.md` inventory created or unused templates removed.
- [ ] `DOCS-17` Verify `scripts/check-plans.py` still passes and add a broken-relative-link checker for `README.md`, `AGENTS.md`, `documentation/`, and `plans/README.md`. Proof: `make docs-check plans-check` exits 0.

---

## Dependencies

- `DOCS-10` and `DOCS-11` must update `frontend/` link references before merging so `make site-check` stays green.
- `DOCS-14` (trimming `AGENTS.md`) must preserve the exact public API freeze rules referenced by `alias_gate_test.go` and `api_internal_gate_test.go`.

## Risks

- **Broken external or website links:** Moving files in `documentation/` could break links from `frontend/` or GitHub issues/PRs. Verify all references with `grep -rn "documentation/"` before moving any file.
- **Over-trimming `AGENTS.md`:** Removing critical non-obvious invariants (like `GOWORK=off` for gomobile, `clipboard.UseMemory` in tests, or not running `go build ./...`) would hurt agent and contributor velocity. Keep all operational rules; move only narrative example walkthroughs.

## Closure Gates

- [x] `DOCS-GATE-01` Zero stale Go / `blinkless` / `ebiten` version pins across `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, and `documentation/`. Evidence: `make docs-check` (with new `check_doc_pins`) passed (39 snippets compiled, 0 pin mismatches).
- [ ] `DOCS-GATE-02` `AGENTS.md` trimmed to <=12 KB with example narratives moved into per-example READMEs.
- [ ] `DOCS-GATE-03` `documentation/README.md` presents a clear 3-doc Quick Start and zero orphan files.
- [ ] `DOCS-GATE-04` `make docs-check`, `make plans-check`, `make site-check`, and `make tools-check` exit 0.
