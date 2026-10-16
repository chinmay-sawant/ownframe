# Phase 06, occlusion culling


> **Parent:** `plans/v0.0.3/optimizations/README.md` - occlusion culling is optimization family 6.
> **Status:** planning only. Viewport visibility culling ships; element-level occlusion does not.
> **Estimated effort:** roughly four to six days, split across the backdrop model, tile integration, ordering guards, and parity tests.

---

## Overview

ownframe paints a `layout.DisplayList` through `internal/replay` and the
256 px tile grid in `internal/window`. The only culling today is viewport
visibility: `DrawVisible` and `DrawVisiblePinned` in `internal/replay/visible.go`
(lines 11 to 56) skip ops whose bounds do not touch the destination rect.
That helps on tall pages, but it does nothing for content that is on screen and
hidden behind a later opaque element. A card with a solid background, a fixed
header, a modal backdrop, or a dense list of overlapping rows still paints
everything underneath every frame, including into tiles whose pixels are never
visible. On the departure-board dashboard the worst frames paint hundreds of ops
for rows that a header or a tile background already covers.

This phase adds element-level occlusion culling. The input is unchanged:
`layout.DisplayList` (`layout/displaylist.go:118-138`) and its `Order` slice
(`layout/displaylist.go:124-125`). The output is unchanged: pixels. The change
is a conservative paint-time decision that skips an op when every pixel in its
painted bounds is already provided by a later, fully opaque backdrop in the same
coordinate frame. Damage tracking, invalidation, and tile allocation keep their
current meaning. No engine or layout change is required, and no engine pin bump
is needed.

Scope is ownframe only. `github.com/chinmay-sawant/blinkless` is not edited.

A `[x]` row would be baseline code checked in on the date named. A `[ ]` row
is open work. `[~]` is reserved for a deliberately deferred item, with a
reason and a gate; this phase has none yet, and deprecated items move to a
later ledger instead of being silently dropped.

## Why this is needed

The repaint cost has two parts even after tiling: which ops paint into a tile,
and how many of those ops are visible. Tiling (phase 02) already limits the
pixels touched by a damage rect. It does not limit the ops that paint into the
touched pixels. A 256 px tile under an opaque card still walks every earlier
op whose bounds intersect that tile, shapes its text, decodes images, and
draws them, only for the card to overwrite those pixels. The cost scales with
document size rather than with the visible scene.

Viewport culling (`internal/replay/visible.go:35-56`) cannot catch this because
both the backdrop and the content are on screen. Damage culling (phase 05)
cannot catch it because a change to a hidden op still produces dirty rects and
tile paints. Occlusion is the missing filter.

The practical targets are:

- Dense opaque surfaces: cards, panels, buttons, headers, footers, fixed
  overlays.
- Back-to-front painter's order already encodes front/back. The front element
  acts as an occluder; the engine does not need to annotate anything.
- Tiles under a full-tile backdrop can skip every content op in that tile.

## Where it lives

Baseline pieces that already ship and that this phase builds on:

- `internal/replay/visible.go` - `DrawVisible`, `DrawVisiblePinned`, `visibleRect`, and
  `visitVisible`. Viewport and pin handling must stay intact.
- `internal/replay/paint_bounds.go` and `internal/replay/op_bounds.go` - `opBounds`
  converts an op to a whole-pixel box and reports false for unbounded ink
  (transforms, rotation, unknown bounds). This is the conservative boundary.
- `internal/replay/chunk_box.go` - `chunkBoxFor` already rejects fixed,
  z-indexed, translucent, unbounded, or too-large ops. The same rejection rules
  define the first occlusion accept list.
- `internal/replay/chunk_run.go` - `runOpaque` already checks `op.Opacity() < 1`
  for a whole run before raster caching. It is the template for backdrop
  eligibility.
- `internal/masking/paint_color.go` - `PaintColor` clamps `op.Opacity()` to 1.0.
  Opacity eligibility uses the clamped value.
- `internal/replay/group_build.go` - blend and isolation groups are built before
  painting. Occlusion must treat a group as one atomic paint unit.
- `internal/window/replay_tile*.go` and `internal/window/replay_partial.go` - the
  tile grid (`tileSize = 256` at `internal/window/replay_tile.go:14`) and
  per-tile paint. Tile integration hooks in here.
- `internal/replay/paint_bounds_box_test.go`, `internal/replay/visible_test.go`,
  `internal/page/page_picture_compare_test.go` and sibling picture tests - the
  parity surface this phase must keep green.

Likely new files:

- `internal/replay/occlusion.go` - backdrop collection, mask, and skip marking.
- `internal/replay/occlusion_test.go` - unit and comparison tests.
- Touchpoints in `internal/replay/paint_bounds.go` if bounds helpers need
  coordinate-frame variants.
- Touchpoints in `internal/window/replay_tile_paint.go` for per-tile use.

Engine side stays untouched.

## What occlusion means here

Occlusion is a property of paint order, not of layout. An element is already
painted when a later element in `display.Order` fully covers it. If that later
element is fully opaque and uses normal source-over blending, the earlier
element's pixels cannot affect the final image. Skipping them is correct by
construction.

Two scopes matter:

1. Backdrop skipping. A single opaque op covers one or more earlier ops and
   those ops are dropped from the walk.
2. Tile-level occlusion. A backdrop may cover only part of a tile, or an
   earlier op may be visible in one tile and hidden in another. The skip
   decision must be evaluated per `(op, tile)` pair, not only globally.

Both share one model: the occluder set accumulated from the front of
`display.Order`.

## Opaque backdrop skipping

### Eligible backdrop

An op is an occluder only when all of these hold:

- `op.Opacity()` is exactly 1 after `Clamp01` (see `internal/masking/paint_color.go`).
  A value of 0.999 is not 1.0; alpha blending rounds per channel, so a near-
  opaque op cannot promise byte parity.
- It has no non-normal blend mode. `BlendModeName()` must be empty or
  `"source-over"` (checked where `group_composite.go` resolves modes).
- It is not inside a blend or isolation group, or it is the group's own chrome
  drawn after the group. A group composites as a unit; partial coverage inside
  it cannot be skipped without recomputing the group.
- `opBounds` returns a non-empty, non-error box. Unbounded ink (rotation,
  transformed geometry, text without bounds) is never an occluder. This reuses
  `internal/replay/op_bounds.go:20-60`.
- Its transform is identity. `op.XformSet` with a non-identity transform makes
  the painted box a rotated or skewed polygon; the axis-aligned backdrop model
  does not apply. Baked `op.Transform()` in `internal/replay/xform_geom.go`
  still works for geometry, but occlusion uses the simpler axis-aligned path.
- It is not `Fixed` or viewport-pinned unless the skipped op shares that same
  fixed or pinned layer. Fixed and pinned ops live in viewport space and can
  occlude content, but content must never occlude them.
- Its kind qualifies. Initial policy: `DisplayOpFillRect` and `DisplayOpStrokeRect`
  with full stroke (StrokeMask == 0) qualify. `DisplayOpImage` qualifies only when
  the payload is known to be fully opaque; that knowledge does not exist yet and
  lands in phase 10 with decode metadata. `DisplayOpText`, `DisplayOpBullet`,
  `DisplayOpLine`, `DisplayOpGridRun` do not qualify, because their pixels do
  not cover every point inside the bounds box. This is deliberately conservative.

Affected earlier ops:

- Any op whose `opBounds` box lies entirely inside the occluder box.
- Any op whose bounds intersect the occluder but are not fully covered is not
  skipped. Partial coverage produces visible slivers that must paint.
- An op that is already outside the viewport, or outside the tile, stays
  handled by the existing paths and is not counted here.

### Backdrop algorithm

Paint order is ascending in `display.Order` (`layout/displaylist.go:120-121`:
higher layers are later). An occluder can only skip ops that come before it.
Therefore the pass runs twice:

1. Reverse pass, from the end of `display.Order` to the start. Maintain a list
   of occluder boxes in page space. For each op, if its box is fully inside at
   least one occluder box from a later op, mark it skippable. Then, if the op
   itself qualifies as a backdrop, add its box to the list.
2. Forward paint pass uses the skip flags. The normal `paintChunked` and
   `DrawVisible` paths check the flag before shaping, decoding, or drawing.

The reverse order is what makes "later means front" safe without changing the
forward drawing order. Blending order, group compositing, and chunk raster
parity are all preserved because skipped ops are simply absent from the walk;
the remaining ops paint in the same order they did before.

Overlapping occluders union their boxes. A back op fully covered by the union
of two front op boxes can be skipped even when no single op covers it. The
union is computed per coordinate frame (see ordering assumptions). For the first
implementation, require a single occluder per skipped op; union support is an
enhancement row with its own parity test, because union coverage must account
for sub-pixel seams.

Group handling: a blend group opens at `IsGroupBegin` and closes at the matching
end. If a whole group's composite bounds are fully covered by a later backdrop,
the whole group (including its markers) is skipped. If only part is covered, no
op inside the group is skipped, because skipping inside a group changes the
group's alpha composite. `group_build.go` exposes the boundaries used for this.

### Fixed, pinned, and z-index layers

- `op.Fixed` ops ignore scroll translation (`drawOp` sets `dx, dy = 0, 0`). Build
  a separate occluder list for the fixed layer. Fixed backdrops occlude fixed
  ops and content; content backdrops occlude content only.
- Pinned ops (`pinZ > 0 && ZIndex >= pinZ`) paint at `pinDx, pinDy`. Build the
  occluder list per pin layer. A backdrop from the scrolled layer never occludes
  a pinned op, because the two are painted with different translations.
- Z-indexed ops use a higher compositing layer. When `ZIndexSet` is true, treat
  the op in its own z band: an occluder only skips ops with `ZIndex <=` its own.
  This prevents a high-z content op from culling ops that paint later in the
  same list but above it.

### Tile-level occlusion

Tiles are 256 CSS pixels (`internal/window/replay_tile.go:14`). Paint into a
tile uses the same `display.Order` and the same occluder relation, but the
relevant bounds are intersected with the tile rectangle in content space before
the inside test.

Per tile:

1. Convert the occluder boxes into tile-local page space by clipping them to
   the tile rectangle. An occluder that does not intersect the tile contributes
   nothing.
2. Build the skip set for ops whose bounds are fully inside the union of the
   clippings.
3. Paint the tile with skipped ops omitted.

Global skip flags from the non-tile pass are a subset, because a global skip
requires full-page coverage. Tile skips can fire for ops that are only fully
covered inside that tile. Both sets are applied: an op is skipped in a tile if
either its global flag is set or its per-tile flag is set.

A tile can become empty after skipping. In that case `paintTile` returns without
touching the retained pixels, but only when at least one backdrop intersecting
the tile was painted during this frame. If every dynamic op in the tile was
skipped because the backdrop is already in the retained screen, the tile is
still marked clean and skipped. This is the path that saves work on
steady-state frames. A flag, `tileOccluded` in the partial state, records it for
the damage probe (phase 05).

Partial coverage inside a tile uses the tile rectangle, not the backdrop's full
box. A backdrop whose bounds cover the whole tile skips the tile's content;
one that covers most but not all of a tile leaves the content at the edges.

Occlusion never changes tile geometry, tile allocation, or `ensureTiles`
(`internal/window/replay_tile_life.go`). It is a paint filter.

### Ordering assumptions

These are the assumptions that make the skip safe. Each one is a hard gate;
violating any of them disables occlusion for the affected fraction rather than
producing wrong pixels.

1. `display.Order` is the paint order. The blinkless doc comment states it
   honors z-index, the outline paint layer, and the chrome-below-content rule
   (`layout/displaylist.go:120-121`). This phase does not reorder anything.
2. A later entry is in front for the same coordinate frame. If the engine
   changes `PaintOrder` semantics (blend groups, outlines), the reverse pass
   must be updated in the same change.
3. Ops with `Kind == DisplayOpNoop` and `DisplayOpUnknown` boundaries carry no
   paint (`layout/displaylist.go:26-37`). They do not act as occluders and are
   not checked for coverage. Group markers are skipped or grouped as a unit.
4. `opBounds` is the authoritative conservative box. If it returns false, the op
   is neither an occluder nor skippable. This matches `chunkBoxFor` in
   `internal/replay/chunk_box.go:14-36`.
5. Outlines paint on a separate layer. An opaque fill behind an element does
   not occlude the element's outline if the outline is a later entry. The
   outline stays if it is visible from the front. This is why the reverse pass
   uses the actual order instead of a depth sort.
6. Fixed and pinned ops are in different coordinate frames and never occlude
   across frames. See the fixed/pinned rules above.
7. Chunk raster caching (`internal/replay/chunk_cache.go`) stores pre-rendered
   runs. A chunk whose ops are all skipped can be removed from the plan for
   that frame, but the cached raster stays valid. Occlusion does not invalidate
   the cache; it only bypasses it. This keeps chunk parity and `runOpaque`
   behavior intact.
8. Damage is independent. Occlusion changes how a dirty rect is painted, not
   which rects are dirty. `TakeDirtyRects` (`internal/page/page_dirty_take.go`)
   is unchanged.

### Coordinate frame for boxes

`opBounds` returns whole-pixel CSS boxes (`internal/replay/op_bounds.go:62-70`,
dividing by `display.PointsPerPixel`). All occlusion boxes are in those CSS
pixels. Translations from `dx, dy` happen at paint time and are not part of the
mask; tile integration applies them consistently with `DrawVisible`. This keeps
the mask independent of scroll and zoom.

## Risks

- Wrong skip corrupts output invisibly. The only detection is a pixel compare.
  Every change must pass the picture tests in `internal/page` before closing a
  row. A future contributors list should treat occlusion as high-risk code.
- Images with alpha. Treating every `DisplayOpImage` as opaque would skip
  content hidden behind transparent or pierced images and produce missing
  pixels. Images are therefore non-occluders in phase 1, and the decode-metadata
  route from phase 10 is the only path to make them occluders.
- Blend and group compositing. A backdrop inside a group does not cover content
  outside the group, and content inside a partially transparent group cannot be
  culled. The guard is conservative: any group begin or end marker disables
  skipping for every op between it and the next non-group backdrop tested.
- Sub-pixel seams. Axis-aligned boxes are rounded outward with `boundsPad`
  (`internal/replay/op_bounds.go:17-19`). An occluder and occludee that share
  an edge may still leave a one-pixel sliver. Rounding both to the same
  outward rule and requiring containment with the pad avoids over-skipping.
  Parity tests cover the boundary.
- Transform and rotation. Any op with `op.RotateDeg != 0` or a non-identity
  transform has no box (`opBounds` returns false). It can never be skipped or
  be a backdrop. This is intentional and matches existing chunk handling.
- Performance of the mask. A naive O(n^2) reverse pass over large documents
  could cost more than it saves. Keep an interval index per row band and
  start with O(n log n) backdrop insertion. Benchmark before claiming a win.
- Stale frames. A dirty frame that reuses retained tiles must not paint skipped
  ops that the tile already contains from the previous frame without the
  backdrop. Occlusion only removes ops from the current paint list; it does not
  remove retained pixels. Since the backdrop is painted or retained in the same
  tile, pixels stay consistent. The damage path already forces a full tile
  repaint on stale frames (`internal/window/replay_partial.go`).
- Debugging. Add an opt-in devtools overlay that draws occluder boxes and
  marks skipped ops. Keep it behind a flag, consistent with phase 05, so
  production builds allocate nothing per frame.
- Thread safety. Build the mask during paint on the draw goroutine. Do not
  share it across frames.

## Check-list

### Phase A, baseline capture

- [ ] Run the departure board and desktop-cat examples with a document seed and
  record op counts per frame in `examples/departure-board/PERFORMANCE.md`.
  Separate frames by whether the viewport moved.
- [ ] Record a `visibleOps` and `paintedOps` split in the perf ring so the
  baseline shows how many ops are already skipped by viewport culling.
- [ ] Add `opBounds` coverage tests for every kind currently in
  `internal/replay/op_bounds_more.go` and `op_bounds_text.go`, so the occlusion
  boundaries are documented before the mask is written.
- [ ] Confirm `make test` passes before window or replay changes.

### Phase B, backdrop model

- [ ] Add `internal/replay/occlusion.go` with `occlusionMask`:
  `build(display, clip image.Rectangle)`, `covered(box image.Rectangle) bool`,
  and `add(box image.Rectangle)`.
- [ ] Implement `buildOccluders(display, clip)` that walks `display.Order` in
  reverse and produces skip booleans indexed by op position in `display.Ops`.
- [ ] Centralize backdrop eligibility in `isBackdrop(op, display) bool`
  (`Op()`s opacity, kind, transform, fixed, z, group).
  Reuse the rules already written in `chunkBoxFor`.
- [ ] Keep single-occluder containment for the first pass: `covered` returns
  true only when one occluder box fully contains the candidate box with the
  outward pad.
- [ ] Unit tests: containment, pad boundary, non-backdrop kinds, opacity
  0.999 versus 1.0, transformed op exclusion.

### Phase C, skip integration into the paint path

- [ ] Add a `Skip []bool` field to a small `paintOcclusion` handle, parallel to
  `display`, so the draw path carries it without changing `layout.Display`.
- [ ] Guard the mask in `paintChunked`: before the item loop, build the mask
  for the full display and check `Skip[index]` for direct ops.
- [ ] Guard `paintGroup` so a fully covered group skips as a unit, and a
  partially covered group paints all members.
- [ ] Guard `paintRun`: a run whose members are all skipped does not paint; a
  run with some skipped members paints only those members.
- [ ] Guard `DrawVisible` and `DrawVisiblePinned` in
  `internal/replay/visible.go` so the flag is honored through the visitor.
- [ ] Add counters `occludedOps` and `occlusionBuildMicros` to the perf ring.
  Keep them off a still generation unless the devtools dock is open, matching
  phase 03 allocation rules.
- [ ] Verify `make golden` and every `internal/page` picture test before and
  after.

### Phase D, tile-level occlusion

- [ ] In `paintTile` (or the tile paint entry in
  `internal/window/replay_tile_paint.go`), build a tile-clipped mask using the
  tile rectangle and the backdrop list from phase B.
- [ ] Apply global skip flags first, then tile skip flags. Log the reason per op
  behind the probe flag, never in a release build.
- [ ] Empty-tile handling: if every op in a tile is skipped and at least one
  backdrop intersecting the tile was painted this frame, leave retained pixels
  untouched and mark the tile clean.
- [ ] Thread dock clip (`dockClipX`) and chrome rects into the backdrop
  rectangle list, so a dock resize still damages both regions.
- [ ] Add a tile occlusion counter to the frame counters tab in the devtools
  dock, reusing the existing tab rather than adding a new page.

### Phase E, image and group metadata (depends on Phase 10)

- [ ] Define a decode-side `FullyOpaque bool` on the image cache key or value,
  populated only after `masking.DecodeImage` proves alpha is all 255. Phase 10
  (async image decoding) changes the producer; do not modify blinkless.
- [ ] Allow `DisplayOpImage` with `FullyOpaque` true to be an occluder, behind
  an unlock flag so existing picture tests can opt in one example at a time.
- [ ] Add a real-world case: desktop-cat overlay has large opaque backgrounds and
  should become cheaper, not faster, on the first opt-in.
- [ ] Group composite bounds for an animated opacity group are never occluders.

### Phase F, measurements, metrics, and tests

- [ ] Unit tests for `occlusionMask`: containment, outward pad, union of two
  backdrop boxes (enhancement), mixed transforms, group exclusion.
- [ ] Comparison tests in `internal/page`: run each picture comparison with
  occlusion enabled and assert byte-identical PNGs. Use the existing helpers;
  do not add a second comparison framework.
- [ ] Add `BenchmarkOcclusionCull` in `internal/replay` with three inputs:
  a staircase of overlapping opaque rects, a grid of cards with transparent
  images, and a page with no occlusion at all as the overhead control.
- [ ] Add a fuzz test that draws random small displays twice, once direct and
  once with occlusion, and compares the resulting image. This is the safety net
  for regressions in the skip logic.
- [ ] Record op counts before and after for the departure board and store them
  in `examples/departure-board/PERFORMANCE.md`. A net regression of more than
  five percent on frames with no occlusion disables the feature by default.
- [ ] Run `go vet -p 1 ./... ./examples/...` and `make test` from the repo root.

### Phase G, docs and tracking

- [ ] Update `documentation/repaint.md` with a short section on occlusion:
  definition, qualifying backdrop kinds, and the per-tile pass.
- [ ] Update `documentation/performance.md` with the op-count table and the
  benchmark row.
- [ ] Update `PHASES.md` with commit hashes and measurement numbers.
- [ ] Change the row in `plans/v0.0.3/optimizations/README.md` from `Partial`
  to `In progress` only when Phase B lands, then to `Done` after Phase F
  tests pass.

## Exit criteria

- Occlusion is conservative and off by default until Phase F proves no parity
  change. When turned on, every example keeps byte-identical output on the
  picture tests and `make golden`.
- Frames with opaque surfaces show lower `paintedOps` without a rendering
  change. Frames with no occlusion show no more than a five percent overhead
  from mask construction.
- Empty tiles under occlusion are left untouched, and tiles remain 256 CSS
  pixels.
- No new files cross the 2000-line soft limit, and `make lint` is green.
- The devtools overlay, when enabled, draws occluder boxes and marks skipped
  ops; when disabled it allocates nothing per frame on a still generation.

## Non-goals

- No change to `layout.DisplayList`, blinkless, or the published engine pin.
- No change to damage detection or invalidation semantics. Occlusion is a
  paint filter, not a state change.
- No image occlusion until phase 10 provides decode-side opacity proof.
- No support for non-axis-aligned backdrop regions, non-normal blend modes,
  or per-group partial skipping.
- No second tile size or dynamic grid. Tile geometry is fixed by phase 02.
- No WebAssembly-specific divergence; browser and desktop share the replay path.
