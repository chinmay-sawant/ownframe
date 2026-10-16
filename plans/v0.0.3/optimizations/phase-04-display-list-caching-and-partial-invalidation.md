# Phase 04, display list caching and partial invalidation

> **Parent:** `plans/v0.0.3/optimizations/README.md` - optimization family 4,
> "Display list caching and partial invalidation". Status is planning only.
> **Status:** open. No behavior changes yet. Rows close only after the matching
> test or benchmark passes.
> **Estimated effort:** 5 to 7 days, split across cache, page wiring, op indexing,
> and reuse gates.

---

## Overview

ownframe keeps two kinds of state between `Redraw` calls. `internal/render/cache.go`
(`Cache`, lines 17 to 22, and `NewCache` at lines 24 to 44) keeps the parsed HTML
tree and the styled document. The page itself keeps the last `Display` in
`Page.display` (`internal/page/page_draw_stages.go`, `commitDisplay` at lines
64 to 77). What is missing is a cache decision for the layout pass: every
`Redraw` calls `render.DisplayListDocument` (`internal/render/document.go`,
lines 16 to 31) and allocates a fresh operation list even when the rendered
source, viewport, and pointer state did not change. The dirty list already works
well; see `diffDisplay` in `internal/page/page_dirty_diff.go` and the geometry
matching in `internal/page/page_dirty_match.go`.

This phase does two jobs:

1. Cache a complete `Display` at the render layer, with explicit identity keys.
2. Make invalidation map to operations by element, through op-level ownership,
   instead of the box scan in `page_invalidate.go`.

Property trees (phase 01) and incremental layout are not required for this
phase. The engine still runs a full layout on a miss.

## Where it lives

ownframe only. blinkless is untouched, and no pin bump is needed for the core
rows. An op-level element link remains the optional engine seam described in
`plans/v0.0.2/damage-and-invalidation/phase-04-op-element-ownership.md`.

Likely files:

- `internal/render/cache.go` - add display storage and identity fields.
- `internal/render/state.go` - add a comparable identity helper for `State`.
- `internal/render/document.go` - add a cached layout entry point beside
  `DisplayListDocument`.
- `internal/page/page_draw_stages.go` - build and install the op index in
  `commitDisplay`.
- `internal/page/page_invalidate.go` - use the op index before the box scan.
- `internal/page/page_dirty_diff.go` and `internal/page/page_dirty_op.go` -
  reuse equality and diff for partial decisions.
- `internal/page/page_patch_display.go` - reuse when a patch changes nothing.
- `internal/page/page_element.go` - op index storage and lookups.
- Tests under `internal/render` and `internal/page`.

## Proposed shapes

### 1. Display cache in `render.Cache`

The cache keeps a `*layout.Display` with its generating keys:

- source string, through `Cache.Source()` (`internal/render/cache.go`, lines
  46 to 52);
- width and height;
- pointer state signature, from `State` in `internal/render/state.go`;
- image resolver identity, usually the `ImageBytes` function;
- the frame sequence or perf generation when useful.

The entry points stay backward compatible. `DisplayListDocument` continues to
lay out a styled document. A new `DisplayListCached(ctx, cache, state)`
returns the stored display on a key match and rebuilds on any mismatch.

The cache must also expose `InvalidateDisplay()`. A page already drops cached
state when source or theme changes, and that path calls this new method.

### 2. Op-level ownership

The page builds an index during `commitDisplay`:

```go
type opRange struct{ start, end int }
```

The proposed map is `map[string]opRange` or `map[string][]int`, with one entry
per element id that owns operations. The baseline uses `boxes` and `boxByID`,
so the first version derives ranges from `display.Boxes` plus operation bounds,
the same geometric source as `markElement` in `internal/page/page_invalidate.go`.
A future engine element link replaces that scan without changing the callers.

`Invalidate(id)` looks up the range first. If the id is missing, it keeps the
current box scan and calls `markFull` for the missing id. Two ids with
overlapping ranges still coalesce through the dirty list, because `markRect`
already unions small neighbors and caps the list.

### 3. Op-level reuse

Reuse is allocation sharing, not skipped layout. After a fresh display is built,
the page and render layer compare each operation with the retained one through
`sameOp` in `internal/page/page_dirty_op.go`. Matching operations can reuse the
old `DisplayOp` header, its content, and its allocated image bytes. This keeps
the first pointer stable where possible and reduces GC pressure on a no-op
redraw.

The equality boundary already exists, so no new comparison is invented:

- identities such as `Kind`, `Text`, `Font`, numeric fields, transform, group,
  image dimensions, and `ImageBytes()` are compared by `sameOp`;
- `PatchDisplay` and `PatchOp` in `internal/page/page_patch_display.go` already
  mutate and compare operations.

### 4. Reuse gates

Reuse is allowed only when all gates pass. Each gate is a small check with an
explicit fallback.

Gate A, whole display reuse:
- same `Source()`;
- same width and height;
- same state signature;
- same image resolver identity;
- cache present;
- no pending ids in `dirty.pending`;
- no full dirty flag.

Gate B, op reuse after a rebuild:
- old and new operation counts are equal;
- for each position, `sameOp` is true, or only changed ops fall in the
  invalidated ranges;
- no resize, no source change, and no state signature change;
- every invalidated id has a non-empty op range.

Gate C, range reuse for `Invalidate`:
- the id has a non-empty range;
- the range is fully inside the current operation slice;
- the previous and current budgets do not overlap with fixed or pinned ops that
  can move independently.

Gate D, fallback to a full rebuild:
- viewport change;
- source change;
- state signature change with no known element map;
- an invalid id or missing range;
- more than the configured invalidated id count, proposed as 8 ids;
- merged dirty area above one third of the frame, matching `diffDisplay`.

When any gate fails, the implementation takes the normal path and records the
reason in perf counters or a debug note. It does not silently keep stale ops.

## Checklist

### Phase 04.1, display cache

- [ ] Add `display *layout.Display` and its keys to `Cache` in
      `internal/render/cache.go`.
- [ ] Add a comparable `State.Identity()` method in `internal/render/state.go`.
- [ ] Implement `Display(ctx, width, height, state) (*Display, error)` and
      `InvalidateDisplay()` in `internal/render`.
- [ ] Keep `DisplayListDocument` working for callers that do not have a cache.
- [ ] Reuse the display when all Gate A conditions hold, proven by a pointer
      equality test.
- [ ] Rebuild on a source change, size change, state change, or image resolver
      change.
- [ ] Return `fetch.ErrNilContext` for a nil context, as `DisplayListDocument`
      already does.

### Phase 04.2, page wiring

- [ ] Create or reuse one `render.Cache` per page instead of the global styled
      path where the page already uses it.
- [ ] Route `drawStaged` through the cached display entry point.
- [ ] Keep `Redraw` counters and `Stats` correct after a cache hit, with zero
      layout time only when a full hit occurs.
- [ ] Store the generation of the cached display and close it on a cache hit.
- [ ] Keep `TakeDirty`, `TakeDirtyRects`, and the window path unchanged.

### Phase 04.3, op index

- [ ] Add op range storage to `Page`, with a generation tag.
- [ ] Build the index in `commitDisplay`, using boxes and op bounds for the
      first version.
- [ ] Update `Invalidate` to use the index before the box scan.
- [ ] Keep the box scan as a fallback for pages without tagged ops.
- [ ] Test nested elements, overlapping boxes, and ids introduced at a later
      generation.
- [ ] Add a benchmark for `Invalidate` before and after the index.

### Phase 04.4, op-level reuse

- [ ] Compare old and new displays through `sameOp` in a helper with tests.
- [ ] Reuse operation payloads when Gate B passes.
- [ ] Measure allocation and CPU cost with `Page` perf counters.
- [ ] Keep image payloads byte-equal before reuse.
- [ ] Fall back to the fresh operation on any mismatch.

### Phase 04.5, reuse gates and observability

- [ ] Implement Gates A through D with explicit fallbacks.
- [ ] Record the gate that rejected a reuse in a debug or perf field.
- [ ] Avoid counters that allocate per frame, following the tiling rule in
      `phase-02-tiling-and-sparse-rasterization.md`.
- [ ] Add devtools exposure only after the perf numbers settle.

### Phase 04.6, tests and benchmarks

- [ ] `internal/render` cache reuse test for the whole display.
- [ ] `internal/page` tests for op index range accuracy and fallback.
- [ ] `internal/page` tests for op reuse preserving `sameOp` equality.
- [ ] `internal/page` tests for Gate A through Gate D fallbacks.
- [ ] Budget test using `page_perf_test.go` style counters.
- [ ] Exercise at least two example pages with no changed output.

### Phase 04.7, docs and closure

- [ ] Update `documentation/repaint.md` with display cache and op identity
      rules.
- [ ] Update any plan or note that claims caching beyond the styled document.
- [ ] Run `make golden` after display or dirty changes.
- [ ] Run `make test` once at the end.

## Exit criteria

- A `Redraw` with identical inputs returns the same cached `Display` pointer.
- Source, viewport, or state changes rebuild and invalidate the cache.
- `Invalidate(id)` uses op ranges when available and falls back cleanly.
- Op reuse keeps `diffDisplay` and the existing dirty tests green.
- Gate failures fall back without staling the retained list or dirty regions.
- `make golden` and `make test` pass.

## Risks

- The op index needs the engine phase for true element links. The geometric
  first version may need a scan, which this phase must measure before it can
  claim a gain.
- Display cache keys include a function value. Function equality is valid in
  Go, but callers must pass the stable resolver. Document that contract.
- A cache hit can hide a stale display if an `Invalidate` call is missed. Keep
  `Invalidate` easy, and treat any cache failure as a rebuild rather than an
  error.
- Reusing payloads can hide a mutation that an operation makes at replay time.
  Only reuse immutable payloads, and never reuse a mutable buffer.
- Full invalidation counts can still exceed the dirty cap; reuse does not
  remove the cap or the full-frame fallback.
