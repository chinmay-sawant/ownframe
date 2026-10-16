# Phase 01 - Property trees

Recorded 2026-10-16 against the working tree on `master`. Planning only. No
behavior change ships with this file. Rows stay open or deferred until code,
tests, and a measured result close them.

Parent ledger: [README.md](README.md). Charter row: optimization family 1,
"Transform, clip, and effect state that updates without repainting children."
See README.md section "Phase-wise checklist (master status)" for the master row.

## Why this is deferred

Property trees are the optimization that avoids rebuilding and repainting
children when transform, clip, or effect state changes. They do that by storing
those three kinds of state in small tree-shaped nodes and pointing display
operations and groups at node IDs instead of baking the values. Today ownframe
bakes the values into each op and group, and Chrome-style node indirection is
not present in either the engine or replay.

The baked path is already correct for the supported transforms and group composites. Upstream (Chromium)
keeps property trees cheap by invalidating nodes and repainting only when needed.
ownframe's equivalent invalidation surface is tile damage, which is documented
under plans/v0.0.2/damage-and-invalidation/phase-01-damage-list.md. Adding
node indirection now would only move work around unless the engine emits
stable nodes and a real case updates transform, clip, or effect without
changing child ops.

So this phase is deferred under the README's `[~]` convention: baseline
already works, and the node linkage needed to make property trees useful is
not available yet. The baseline rows below are recorded as deferred with their
next gate.

## Current state - transform mapping

Source: `internal/replay/xform_geom.go:1-156`.

- `xformGeom` reads `op.Transform()` directly and converts the returned
  `Matrix2D` into an `ebiten.GeoM`. Translation is scaled by `pxPerPt`, so
  the matrix is treated as unitless in local space
  (`internal/replay/xform_geom.go:14-29`).
- `shiftLocal` moves a geom origin in local space instead of dst space, which
  is required for the shift after `xformRaster` returns cached raster bounds
  (`internal/replay/xform_geom.go:31-36`).
- `xformLocalBox` computes an op's untransformed paint box for fill rect,
  stroke rect, and line ops. This is the geometry that must remain stable if
  only the transform node changes (`internal/replay/xform_geom.go:38-66`).
- Text and grid runs take separate branches through `drawXform`
  (`internal/replay/xform_draw.go:10-24`), and `drawXformShape` replays a
  cached raster through the baked matrix
  (`internal/replay/xform_draw.go:26-46`).
- There is no separate transform tree. Every paint reads the baked matrix
  from the op pointer and has no node ID, parent pointer, or invalidation
  hint.
- Cache keys for transformed rasters are based on the current op pointer and
  raster. A property node would need the node matrix added, or would need the
  resolver to invalidate the cache when the node matrix changes. Neither is
  present.

## Current state - group build and compositing

Source: `internal/replay/group_build.go`, `group_layer.go`, `group_shader.go`.

- `buildPlan` sorts ops into a tree by `op.Group()` pointers and detects
  group begin/end markers. A group whose mask is non nil gets a node with
  a `maskBox` and `hasMask`, so masking is attached to the group instead of
  to an effect node (`internal/replay/group_build.go:8-33`).
- `attachNode` creates one `groupNode` per `*layout.DisplayGroup` and attaches
  it under the parent group pointer. There is no node deduplication beyond the
  display pointer, and there is no effect node indirection
  (`internal/replay/group_build.go:35-53`).
- `paintGroup` allocates one `ebiten.NewImage` layer sized to the group window,
  paints each member, applies the group mask using `BlendDestinationIn`, and
  composites once with the group mode (`internal/replay/group_layer.go:11-34`).
- Mode and mask are read from `n.group` directly. There is no opacity field,
  blend mode field, or mask on an effect node. A change in any group compositing
  property therefore implies a new `*layout.DisplayGroup`, which implies a new
  display list, because that is how the engine links group state.
- `groupShader` compiles the group blend program once and returns nil on a
  compile failure. Callers fall back to source-over. This is the only effect
  state fallback today (`internal/replay/group_shader.go:18-33`).
- `groupNode` carries `group`, `items`, `box`, `maskBox`, `maskTex`, and
  sampling flags, but no node IDs for transform, clip, or effect
  (`internal/replay/group_plan.go:10-40`).
- `planFor` caches the group tree by `*layout.Display` pointer with two
  entries. A replaced display always builds a fresh group tree, even if the
  diff is small (`internal/replay/group_plan.go:38-72`).

## Baseline rows

These rows are checked-in, correct baseline behavior and stay closed. They
are marked as deferred work rather than completions because the optimization
does not use them as property nodes yet.

- [~] `op.Transform()` reaches replay and drives `xformGeom`. Baseline works.
  Next gate: a transform-only display change that does not change op identity.
- [~] `shiftLocal` keeps cached transformed rasters aligned. Baseline works.
  Next gate: property node matrix added to raster cache keys.
- [~] `buildPlan` groups ops into a valid blend and isolation tree. Baseline
  works. Next gate: per-group effect node with stable identity.
- [~] `paintGroup` renders layer, mask, and blend in one composite. Baseline
  works. Next gate: effect node opacity and clip handling with the same pixel
  output.
- [~] `groupShader` returns one compiled shader with a source-over fallback.
  Baseline works. Next gate: property effect mode lookup instead of group mode.
- [~] `planFor` caches a group tree by display pointer. Baseline works. Next
  gate: node-aware partial invalidation, not only pointer replacement.

## Current state - first related tests

- `internal/replay/xform_geom_test.go` covers the transform geometry helpers.
- `internal/replay/group_plan_test.go` covers group tree and plan behavior.
- `internal/replay/group_shader_test.go` covers shader compile and fallback.
- `internal/replay/chunk_group_test.go` covers grouped replay behavior.
- `gpu_group_clip_test.go`, `gpu_group_mask_test.go`, `gpu_group_nested_test.go`,
  `gpu_group_pixels_test.go`, `gpu_group_test.go`, and `gpu_group_photo_pixels_test.go`
  cover the group visual parity suite.
- Repaint surfaces: `repaint_equality_test.go`, `repaint_equality_helpers_test.go`,
  and `repaint_picture_compare_test.go`.
- `visible.go` already handles viewport visibility (`internal/replay/visible.go:11-56`),
  but that is culling, not a clip property tree.
- `make golden` runs the public `TestDisplay` in `./layout`.

## Proposed model

The model has three node kinds and one root per frame.

- TransformNode carries a parent node ID and one `Matrix2D` matrix. It
  represents one CSS transform stack that may be shared by many ops.
- ClipNode carries a parent node ID and one CSS clip (content box, border box,
  circle, ellipse, or path). It represents one CSS clip and overflow boundary.
- EffectNode carries a parent node ID, opacity, blend mode, isolation flag,
  CSS mask pointer, and an optional animated or shared visual effect state.

Rows are owned by the engine where the values are computed, but the pointers
are replay-side in `Display` and `DisplayGroup`.

### Proposed engine-side contract

Owner: github.com/chinmay-sawant/blinkless. These rows are planned here but
must land in blinkless first, then be pinned in `go.mod` and `examples/go.mod`.

1. Add a node registry to `layout.Result` or to `layout.Display`, keyed by
   integer IDs with a root sentinel.
2. Add resolved node IDs to `DisplayOp` and `DisplayGroup`, or add pointers
   to the same structs. Default zero means "no property node; use the baked
   fields", preserving today's contract.
3. Reuse node IDs across renders when the CSS input for that node is
   unchanged. This is the invariant that lets transform changes update state
   without repainting children.
4. Keep `Transform()`, `BlendModeName`, and group masks working as fallback.
   The fallback is never removed in ownframe; only the property path is added.
5. Add engine tests showing that one transform change changes one node and
   leaves every child op pointer stable.

### Proposed ownframe contract

Owner: ownframe.

1. `paintDisplay` and `paintNode` read transform and effect state through a
   resolver that accepts node IDs. The resolver returns the baked values when
   the id is zero, so baseline remains unchanged.
2. `xformGeom` takes either an op with a transform node id or the existing op.
   Resolved matrices are converted through the same `ebiten.GeoM` conversion.
3. `buildPlan` constructs group nodes and records the group's effect node id.
   `paintGroup` resolves mode and mask from that id before allocating the layer.
4. Clip nodes are the only addition that changes draw calls. Start with a
   conservative no-op: store the id without changing clipping. Move clipping
   only after visual parity and timing prove the need.
5. Cache and invalidation:
   - `planFor` includes the root node ID in its cache key.
   - Raster caches include the transform node ID or its matrix.
   - Tile damage remains the visual output. If a state change can be reused
     with no pixel changes, the node path simply wins against the raster path by
     skipping the raster work and recording a skip count.
6. Parity proof:
   - Every property-tree paint result equals the baked paint result for the
     golden and repaint suite before any optimized state-update path ships.

## Phases

| Phase | Owner | Job |
|---|---|---|
| 0 | ownframe | Record current baseline, bundle timings, and the cases where transform or group state changes without child-op changes |
| 1 | blinkless | Design property node types and ID allocation in the engine and bump the pin after review |
| 2 | blinkless | Emit transform nodes and group effect nodes on `Display` and `DisplayGroup` with fallback behavior |
| 3 | ownframe | Add the node resolver and read baked fallback when IDs are zero |
| 4 | ownframe | Resolve `op.Transform()` through a transform node in `xformGeom` and `xformRaster` caching |
| 5 | ownframe | Resolve group mode, mask, and isolation through an effect node in `paintGroup` |
| 6 | ownframe | Store clip node IDs without changing clipping, then prove parity |
| 7 | ownframe | Add node-aware partial work and skip counts without removing tile damage |
| 8 | ownframe | Add tests and measurement comparison |
| 9 | docs | Update owner docs and phase ledger |

## Checklist

### Phase 00 - Baseline and evidence

- [ ] Run the departure-board paced dump twice and store the Ops and Frame tab
      timings under the same window size and tile count:
      `go run ./examples/departure-board -dump -pace 16ms`.
- [ ] List all cases in `examples/departure-board` where transform, clip,
      opacity, blend, or group mode changes independently of child ops.
- [ ] Confirm that no current caller updates transform or group state without a
      full engine render. If such a caller exists, name it and route it through
      the engine property node path instead of ownframe.
- [ ] Record the golden suite command used for all parity work:
      `make golden`.
- [ ] Record the group parity test list used for all visual proof:
      `go test ./internal/replay -run '^TestGPUGroup' -count=1`.

### Phase 01 - Engine linkage (blinkless side)

- [ ] Define node types with stable IDs, parent IDs, and root sentinel.
- [ ] Add resolved transform node IDs to ops with a zero fallback.
- [ ] Add resolved effect node IDs to groups with a zero fallback.
- [ ] Add a simple transform-only change test that does not alter child op identity.
- [ ] Merge the blinkless change, then bump the pin in `go.mod` and `examples/go.mod`
      in ownframe.
- [~] This phase cannot close from this repo alone. Engine pull request and pin
      bump are required.

### Phase 02 - Node resolver in ownframe

- [ ] Add a helper that maps the display root node to the registered tree and
      falls back to zero identifiers.
- [ ] Add unit tests for zero-ID fallback so baseline remains byte-for-byte.
- [ ] Add a `planFor` cache key that includes the root node ID.

### Phase 03 - Transform node resolution

- [ ] Resolve the transform matrix in `xformGeom` when a transform node ID is
      present.
- [ ] Keep non-transform ops using the baked matrix path.
- [ ] Add transform node IDs to the `xformRaster` cache key or invalidate
      entries after node matrix change.
- [ ] Add targeted tests in `internal/replay/xform_geom_test.go`.

### Phase 04 - Effect node resolution

- [ ] Resolve mode, opacity, isolation, and mask from effect node IDs in
      `paintGroup`.
- [ ] Keep group field fallback when the effect node ID is zero.
- [ ] Preserve the blend shader fallback path:
      `groupShader()` or nil (`internal/replay/group_shader.go:18-33`).
- [ ] Add group parity tests for opacity, blend, mask, and isolation variants.

### Phase 05 - Clip node baseline

- [ ] Store clip node IDs without changing clip behavior.
- [ ] Document the reason clipping stays conservative: axis-facing ownframe
      only clips through tile and viewport bounds today; CSS clip-shapes are
      baked at the raster level in the engine.
- [ ] Add explicit test that a clip node does nothing at replay yet.

### Phase 06 - Measurement and optimization

- [ ] Add a skip counter for paints that reuse an unchanged property tree.
- [ ] Record paired before and after paced dump numbers using the same command.
- [ ] Only claim the optimization if at least one repaint path skips work
      while preserving `TakeDirty` and tile damage behavior.
- [ ] Otherwise close the phase as deferred with the measured reason.

### Phase 07 - Docs and packaging

- [ ] Update this ledger with the measured outcome.
- [ ] Update README.md optimization rows from deferred to shipped only if a
      runtime row closes.
- [ ] Update `PHASES.md` indexes when the files ship.

## Tests and exit criteria

Required before any deferred row flips to shipped:

- `go build ./...`
- `go test ./internal/replay`
- `make golden`
- paired paced dumps recorded from
  `go run ./examples/departure-board -dump -pace 16ms`
- group parity suite:
  `go test ./internal/replay -run '^TestGPUGroup' -count=1`
- repaint suite:
  `go test . -run 'Repaint' -count=1` if applicable in the root package,
  otherwise the replay repaint tests

Exit criteria, either path is valid:

1. Deferred path:
   - Numbers show no state-only case that reuses unchanged children.
   - The baseline rows remain deferred and documented.
   - `make golden` passes.
2. Shipped path:
   - Engine linkage and pin bump are in place.
   - Property-tree paints equal baked paints on the golden and GPU group suite.
   - At least one state-only update skips child repaint or raster work without
     changing tile damage rules.
   - Docs record the before and after numbers.
   - `make test` and `make lint` pass.

## Dependencies

- blinkless property linkage before ownframe can claim node-aware state.
- Tile damage and display list invalidation ledgers for visual proof.
- Group raster and shader parity suite for any effect node change.

## Status note

As of 2026-10-16, this optimization is deferred. Baseline transform and
group compositing work through baked values, and no node indirection exists.
Runtime rows stay open until engine linkage and a paced measurement name a gap.
EOF