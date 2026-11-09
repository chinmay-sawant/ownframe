# Phase 07: Full Rebuild, Single Patch, and Batched Patch Paths

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)
> **Status:** Proposed
> **Target Subsystems:** page rebuild pipeline (`ui` builder, `internal/page`, `internal/render`), element patch path (`internal/page/patch*`, blinkless `RelayoutElement`), app composition patterns
> **Primary Goal:** Halve the cost of the three data-change paths (full rebuild, click-through rebuild, element patch) on large documents, and record which path owns which change size.

---

## 1. Problem Statement & Baseline Benchmark

A data change on a large document (~9k boxes, ~9k display ops, ~600 KB source) pays the whole document every time through the default rebuild path. Measured on i7-13700HX, headless, `go test -bench`:

*   **Full build, changed data:** **304.1 ms/op, 3.3 fps, 45.9 MB/op, 316k allocs/op**.
*   **Click + full rebuild:** **304.1 ms/op, 3.3 fps** (same cost for a 2-element change as for a 1000-element change).
*   **Click + single-element patches** (2 class edits + 1 text edit, no reparse, no recascade): **55.7 ms/op, 18 fps, 42.5 MB/op, 34k allocs/op**.
*   **1000 per-element patches in one batch:** **19650 ms/op, 0.05 fps, 12.9 GB/op, 13M allocs/op** (anti-pattern, measured to forbid it).
*   **Global change via stylesheet swap + redraw:** **264.8 ms/op, 3.8 fps** (one cascade, layout, paint; no builder render).

Sampled stage split inside a full rebuild (Perf on): template 0 ms, style 7.9 ms, layout 256.7 ms, paint 0 ms headless. Layout and display-list build is ~90% of the cost. Builder HTML render is ~9 ms (~4%). Faster parsing or faster HTML generation cannot halve anything.

Two structural facts drive the plan:

1.  Each `PatchElement` re-resolves styles for the whole document (~17 ms here), so N patches cost N re-resolves. Small touches win (3 patches beat a 304 ms rebuild 5.5x); loops over large surfaces lose (1000 patches take 19.6 s).
2.  A global restyle needs exactly one cascade + layout + paint whether it arrives as a rebuild or a stylesheet swap. Per-element patching must never be the vehicle for full-surface changes.

---

## 2. Target Metrics & Estimated Win

| Path | Baseline | Target | Estimated win | Owning step |
| :--- | :---: | :---: | :--- | :--- |
| Full build, changed data | 304.1 ms/op, 3.3 fps | **<=152 ms/op, >=6.6 fps** | **>=2x via windowing; engine layout work upstream** | Step 3, Step 5 |
| Click + rebuild | 304.1 ms/op, 3.3 fps | **<=152 ms/op** | **5.5x already banked by routing selection through the patch path (55.7 ms measured)** | Step 1 |
| Click + patch (few edits) | 55.7 ms/op, 18 fps | **<=28 ms/op, >=36 fps** | **~2x via one shared style re-resolve for N edits** | Step 2 |
| Global restyle (whole surface) | 264.8 ms/op via swap | **<=152 ms/op** | **Reaches target only with Step 3 or Step 5; per-element loops stay forbidden** | Step 3, Step 4, Step 5 |
| Idle re-render | 0.014 ms/op, 0 allocs | no change | Already free; unchanged by this phase | none |

---

## 3. Implementation Summary

### Step 1: Route small touches through the patch path by default

1.  Selection, toggles, and other few-element updates commit `PatchElement` edits (class/attribute/text) instead of a full view rebuild.
2.  Every patch keeps a full-refresh fallback on refusal, so the page never goes stale.
3.  Keep both benchmarks: the rebuild number stays the regression baseline, the patch number is the interactive target.

### Step 2: Batched multi-element patch (one re-resolve for N edits)

1.  Add a batch patch entry point: apply N tree edits, run one style re-resolve, re-lay out each touched element, splice once, draw once.
2.  Refused elements fall back individually to refresh; the rest of the batch still commits.
3.  Expected effect: a 3-edit interaction drops from ~3 re-resolves (~56 ms) to ~1 (~25 ms).

### Step 3: Window large homogeneous content

1.  For collections past the windowing threshold, lay out only the visible window plus spacers (existing spacer pattern), not the whole collection.
2.  Selection and edits address elements by id against memory truth; off-window targets scroll into view or update without boxes.
3.  Benchmarks measure the visible window. Full-build cost scales with viewport, not collection size (expected ~10x on thousand-row shapes, which clears the 2x target with margin).

### Step 4: Forbid per-element loops for full-surface changes

1.  Global changes (theme, palette, full-surface restyle, and later chart/material swatches) travel as one stylesheet swap or one rebuild, never as N patches.
2.  Record the 1000-patch measurement as the regression tripwire for this rule.

### Step 5: Upstream engine layout cost (parallel track)

1.  Profile placement + display-list emission on wide flat documents (~9k boxes, 257 ms layout) in the layout repository.
2.  Consume the improvement through the dependency pin. This is the only lever that halves full builds without windowing or fewer boxes per row.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-07-01` Route few-element updates through `PatchElement` with refresh fallback; click-path benchmark reads <=152 ms/op.
- [x] `OPT-07-02` Add batched multi-element patch (one style re-resolve, per-element relayouts, single draw); few-edit benchmark reads <=28 ms/op.
- [x] `OPT-07-03` Window large homogeneous collections (visible window + spacers); full-build benchmark scales with viewport and reads <=152 ms/op.
- [ ] `OPT-07-04` Keep a per-element-loop regression benchmark proving full-surface patching stays forbidden (must remain orders of magnitude slower than swap/rebuild).
- [x] `OPT-07-05` Stylesheet-swap path for global restyles with one cascade + layout + paint; benchmark at or under rebuild cost.
- [ ] `OPT-07-06` Upstream layout-profile issue filed with the 257 ms / 9k-box baseline; pin bump consumed when the engine halves it.
- [ ] `OPT-07-07` Run the data-change benchmark set plus `make test`, `go vet`, and `gofmt` clean.
