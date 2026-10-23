# Phase 1: Codebase Cohesion & De-fragmentation

> **Parent:** [README.md](README.md)  
> **Area:** Package Structure, Code Quality & File Cohesion  
> **Status:** Open `[ ]`  

---

## Overview

The historical 2,000-character-per-file constraint led to severe hyper-fragmentation:
- **3,800+ tracked Go files** across the repository, averaging only ~50 lines per file.
- Single types and responsibilities were sliced arbitrarily into dozens of helper files (e.g., `_door.go`, `_api.go`, `_split.go`, `*_more_test.go`).
- `internal/page` expanded to ~381 files, `internal/window` to ~356 files, `internal/replay` to ~134 files, and `examples/` to ~2,325 files (with a simple starter like `examples/login` spanning 29 files).

While `AGENTS.md` and `plans/README.md` retired the 2,000-character cap under the `COHESION` initiative, the legacy files remain. Navigating, reviewing, and safely evolving the code requires systematically consolidating related methods into cohesive, domain-driven files.

---

## Strategy & Principles

1. **Group by Responsibility:** Methods on a single struct belong in the same file or a cohesive, domain-named file (e.g. `page_form.go`, `page_draw.go`, `page_scroll.go`), not sliced across artificial single-method files.
2. **Eliminate Synonym/Indirection Wrappers:** Remove single-line wrapper files that exist only to forward calls without adding semantic value.
3. **Preserve Public API & Git History:** Merges must be done using AST-aware tooling (`scripts/merge-go.go`), preserving comments, docstrings, and license notices, verified by `make test`, `make test-race`, and `make lint`.
4. **Teach Clear Patterns in Examples:** Starter examples like `examples/login` should be concise, idiomatic reference code (1–3 files), not 29 micro-files.

---

## Checklist

### 1.1 Tooling & Verification Automation

- [ ] `COHESION-01` Validate and extend `scripts/merge-go.go` to support automated AST-level merging with sorted unique imports, preserved package-level comments, and conflict detection.  
  *Acceptance Criteria:* `scripts/merge-go_test.go` passes; tool cleanly merges a set of 5 test files into a single valid Go file without syntax or import errors.

- [ ] `COHESION-02` Establish an inventory script or test that tracks total file count and average lines-per-file per package, verifying that consolidation PRs monotonically decrease fragmentation without dropping tests.  
  *Acceptance Criteria:* Script reports package file counts and alerts on regression or orphan files.

### 1.2 Subsystem Consolidation: `internal/page`

- [ ] `COHESION-03` Consolidate `internal/page` form and edit subsystems into cohesive domain files (`form_state.go`, `edit_input.go`). Retire `page_form_door.go`, `page_edit_door.go`, and their single-method companions.  
  *Acceptance Criteria:* `go test -p 1 ./internal/page` passes; file count in `internal/page` reduced by at least 40%.

- [ ] `COHESION-04` Consolidate `internal/page` layout, drawing, and scroll subsystems (`draw_pipeline.go`, `scroll_window.go`, `dirty_tracker.go`). Remove redundant delegate files.  
  *Acceptance Criteria:* `make test` and `make graphics-check` pass; zero regressions in repaint diff or tile damage calculations.

### 1.3 Subsystem Consolidation: `internal/window`

- [ ] `COHESION-05` Consolidate `internal/window` around the subsystems declared in `internal/window/SUBSYSTEMS.md`:
  - `pacing/`: Pacing, idle mode detection, and deadline scheduling.
  - `devtools/`: Overlay rendering, tab switching, and element/display inspector.
  - `events/`: Input event translation, mouse/touch routing, drop forwarding.
  *Acceptance Criteria:* File count in `internal/window` reduced by at least 50%; all window tests and `TestWindowIdleDeadline` pass.

### 1.4 Subsystem Consolidation: `internal/replay` and Root Package

- [ ] `COHESION-06` Merge fragmented display operations, shader passes, and glyph caches in `internal/replay` into grouped domain files (`ops_shape.go`, `ops_text.go`, `ops_image.go`, `group_layer.go`).  
  *Acceptance Criteria:* `go test -p 1 ./internal/replay` and GPU alpha pixel tests pass.

- [ ] `COHESION-07` Merge root `page_*.go` door files into the canonical `page_wrap.go` and `api.go` files, preserving the exact public method set on `ownframe.Page`.  
  *Acceptance Criteria:* `api_test.go`, `api_internal_gate_test.go`, and `make module-check` pass with no exported API changes.

### 1.5 Examples Consolidation

- [ ] `COHESION-08` Consolidate over-fragmented examples:
  - `examples/login`: Reduce 29 micro-files into `main.go`, `view.go`, and `handlers.go`.
  - `examples/ops-console`: Group into `main.go`, `console/console.go`, and `console/model.go`.
  *Acceptance Criteria:* `make build` and `make examples-check` pass; examples are easily readable as starter templates.
