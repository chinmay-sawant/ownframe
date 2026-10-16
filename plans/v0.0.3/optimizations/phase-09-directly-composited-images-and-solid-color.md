# Phase 09, directly composited images and solid color

Goal: promote simple content in `internal/replay` to direct compositor passes, so
solid-color fills and plain decoded images stop being merged into the tile
rasters at paint time. Ownframe stays a `layout.DisplayList` engine and keeps
the tile grid from phase 02. The change is in the replay layer and in what the
window keeps as a retained surface, not in engine layout.

## Where it lives

Baseline pieces that already ship, cited from the tree on 2026-10-16:

- `internal/replay/op.go`: `drawOp` is the single dispatch (`op.go:14-46`). A
  transformed rect is rerouted to `drawXform`; otherwise `DisplayOpFillRect`
  calls `fillRect`, and `DisplayOpImage` calls `drawImage`.
- `internal/replay/fill.go`: `fillRect` and `snapped`. A plain fill with no
  radius uses `vector.FillRect` after snapping all four edges to whole pixels.
  A rounded fill checks the raster cache first and only paints when the cache
  misses.
- `internal/replay/raster_fill.go` and `internal/replay/raster_cache.go`:
  repeated rounded fills bake once and blit on later damage.
- `internal/replay/image.go`: `drawImage` decodes through `decodedImage`, then
  draws the whole image with `ebiten.DrawImage`, linear filter, and an
  alpha color scale when the op opacity is below 1.
- `internal/replay/image_cache.go` lines 1-70: `imageCache`, `decodedImage`,
  64-entry cap, read-lock hit, lock-free decode, write-lock insert, second
  hit check, LRU-by-first-insertion eviction. It disposes the oldest entry
  when the cap is reached.
- `internal/replay/image_key.go`: `imageBytesKey` folds the payload bytes and
  the `IsJPEG` flag into an FNV-1a key; `imageGeom` applies the op rectangle,
  page offset, `pxPerPt` scale, and the op's CSS transform.
- `internal/masking/image_decode.go`: `DecodeImage` reads `op.ImageBytes`,
  runs `image.Decode`, rejects empty bounds, and returns alpha. `SkipImageSize`
  rejects negative or zero geometry.
- `internal/masking/radii.go`: `FillRadii` resolves circular corners and rejects
  elliptical radii, which is what keeps plain `FillRect` joins safe for a
  direct pass.
- `internal/replay/chunk_core.go` and `internal/replay/chunk_paint.go`: element
  runs can paint as one cached raster (`paintRun`) or fall back per op
  (`drawRunDirect`). A mixed translation, a fractional translation, or a
  translucent member forces the per-op path.
- `internal/replay/group_paint.go` and `internal/replay/group_composite.go`:
  non-normal and masked groups buffer into a layer and composite through a Kage
  shader. Normal source-over groups paint straight through.
- `internal/replay/color.go`: `rgba`, `withAlpha`, `channel`, `clamp` map an op
  color to `color.RGBA` and fold opacity the same way the engine painters do.
- `internal/window/replay_partial.go`: `drawReplayPartial` allocates plans per
  rect and calls `applyRepaint`; the tile grid keeps untouched tiles and blits
  only the damaged ones.
- `internal/window/replay_tile_blit.go`: `blitFrame` copies damaged tiles only
  on a normal frame.

## What Chrome does

Chromium's compositor identifies cheap content that the GPU can draw without a
rasterized tile: solid colors, images that can stay as textures, and text runs
backed by a glyph atlas. Those are kept as direct draws or dedicated layers,
while only the genuinely complex content pays for tile raster. The two
consequences ownframe cares about are lower tile write cost on scroll and less
re-paint of content that never changed.

ownframe pieces that already match that idea: opacity, alpha, rounded fills that
bake through the raster cache, and images that decode once and are retained as
`*ebiten.Image`. What it lacks is a policy that says which ops may bypass the
tile raster entirely, a promotion record the window can read later, and the
escape hatches for groups, clips, transformed op boxes, and partial opacity.

## Current status as of 2026-10-16

- [x] Fills and decoded images paint directly into the tile destination. There
  is no intermediate page bitmap on the replay path.
- [x] Image payloads decode once and are cached by encoded bytes, with a
  64-entry cap. Transformed keeps its own geometry path.
- [x] Repeated rounded fills bake through the raster cache instead of repainting.
- [x] Normal-mode groups paint straight through; only non-normal modes buffer.
- [ ] No promotion policy exists. `drawOp` always paints into whatever `dst`
  it receives, which means the tile and a direct-to-screen pass look identical.
- [ ] No per-op "direct" or "promoted" flag is carried in the replay tree, so a
  window cannot skip raster when the whole op list is eligible.
- [ ] Solid-color fills with zero radius, no clip, and normal blend still go
  through `vector.FillRect` on every paint, even when the rect is large and
  repeatable.
- [ ] Images are redrawn into the tile via `ebiten.DrawImage` every time the
  tile repaints, even though the same cached source exists unchanged in
  `decodedImage`.
- [ ] No metrics tell the developer how many ops could have been promoted on a
  frame, only how many blits happened.
- [x] No step in this phase changes blinkless. The engine keeps returning a plain
  display list.

## Design

1. Keep the current paint path as the reference. `drawOp` and `fillRect` do not
   change their output for promoted work; promotion only changes where the op is
   painted and whether the tile keeps a copy.
2. Define a conservative predicate for a directly composable image:
   - `op.Kind == layout.DisplayOpImage`,
   - `decodedImage(op) != nil`,
   - `!masking.SkipImageSize(op)`,
   - the op belongs to a normal-mode group with no mask,
   - `op.XformSet` may be true only when the baked matrix is identity on axes
     it does not need to re-sample,
   - no enclosing clip that the destination tile does not already carry.
3. Define a consistent predicate for a solid color fill:
   - `op.Kind == layout.DisplayOpFillRect`,
   - `masking.FillRadii(op)` succeeds,
   - all four resolved radii equal zero,
   - `masking.PaintRGBA(op)` alpha equals the clipped op alpha with no alpha
     math or individual corner radius override,
   - normal blend mode, no group mask.
   These two predicates are the only classes promoted in this phase. Text keeps
   its current bake path and solid color text is out of scope.
4. Introduce a `directPass` kind on the replay side, distinct from tile paint.
   A pass holds a small list of promoted candidates, an image destination, and
   the page offset. It paints into a scratch buffer or directly into the final
   surface, depending on group state.
5. Keep tile paint authoritative for overlap and z order. Promotion is allowed
   only for ops that do not rely on tile order to blend; normal source-over
   into a transparent destination needs its source opaque or alpha 1.0. Partial
   opacity still uses tile paint.
6. Add a promotion record the window can inspect: counts of promoted images and
   promoted solid fills on a frame, plus the rect of each. This record is
   reset at the start of `drawReplayPartial`, matching `partial.paints`.
7. Solid-color reduction first. Before any copy, a run of consecutive
   zero-radius fills sharing a color and an X position can be merged into one
   tall rect when the spans are contiguous and no other op paints between them.
   A merged fill runs once through `vector.FillRect` instead of N calls.
8. Image layer second. A decoded image that repeats in size and payload on a
   frame can be kept on a local `*ebiten.Image` whose bounds match the op box,
   so the later tile repaint blits the retained image rather than re-requesting
   the decode path. The image cache already covers decode; this covers the
   final draw count per tile.
9. Keep the bad-payload escape hatch. A nil decode, a skip size, or a key
   collision still falls back to tile paint, in the same way it already skips
   draw. Promotion must never turn a bad image into a crash.
10. Update damage feedback so a promoted op still reports its box as damaged.
    A promoted fill or image paints into the final surface directly; the tile
    the geometry lives over is marked damaged all the same, so nothing goes
    untracked and a later partial tile blit does not overwrite it.

## Checklist

### Phase A, baseline capture

- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` and record the
  frame ops count, blit count, paint count, and image op count in the notes,
  before any change.
- [ ] Add a harmless promotion counter that is always zero in this build, so
  the numbers include the new column. This is measurement only, not behavior.
- [ ] Identify one representative fill-heavy screen and one image-heavy screen
  from the example set, and measure both.
- [ ] Run `make golden` so the drawing list snapshot is clean before the work
  starts.

### Phase B, predicate and escape hatches

- [ ] Add `imagePromotable(op, group) bool` in `internal/replay/image.go`
  beside `drawImage`, with the predicate from design item 2.
- [ ] Add `solidFillPromotable(op, group) bool` in `internal/replay/fill.go`
  beside `fillRect`, with the predicate from design item 3.
- [ ] Include group kind and mask state in both predicates through the same
  inputs `paintItems` already has.
- [ ] Keep non-normal groups, masked groups, and clips falling back to tile
  paint, and document the exact exclusion in a godoc line on each predicate.
- [ ] Add unit tests asserting each escape hatch forces tile paint, including a
  translucent alpha, a transformed image, a rounded fill, and an elliptical
  radius.

### Phase C, solid-color reduction

- [ ] Add `mergeConsecutiveSolidFills` returning a new op slice or index list
  that collapses contiguous zero-radius fills of one RGBA value into one fill
  op covering their combined vertical or horizontal span.
- [ ] Run the merge only on normal-mode, non-masked group bodies and only
  when both source fills have zero radius and identical non-transparent color.
- [ ] Keep the original op order and z-index behavior. A merged fill must paint
  at the same point in the order list, covering the exact same union pixels.
- [ ] Keep the merge lossless. If the merged region has a hole or a different
  op in between, the merge stops and both fills stay separate.
- [ ] Report merged op count in the debug counters.
- [ ] Parity tests: every fill-heavy golden must still match pixel for pixel.
  Partial merges only join pixels that were already solid and contiguous.

### Phase D, image layer

- [ ] Introduce a small retained holder for a promoted image: payload key, op
  bounds, and the `*ebiten.Image` from `decodedImage`, with a generation
  stamp.
- [ ] Retain at most a handful of promoted images per frame and evict when the
  frame ends. Do not grow a second long-lived cache beside `imageCache`.
- [ ] Paint a promoted image by blitting the retained holder into the final
  destination at its own size, keeping the linear filter. Do not skip the
  bounds check.
- [ ] Keep the existing `imageGeom` transform path, so an identity op paints
  the whole decoded area directly.
- [ ] Fall back to tile paint on a missing holder, an empty bounds result, or
  a masked group. The fall back must not allocate on a still frame.
- [ ] Add a test that counts `ebiten.DrawImage` calls for an image-only page
  and asserts the promoted path uses fewer draws than the tile path.

### Phase E, window integration

- [ ] Add a `promoted` slice to the partial state in `internal/window`, matching
  `partial.paints` and `partial.blits`. Clear it at the start of
  `drawReplayPartial`, beside the existing resets.
- [ ] Add a window seam `paintDirectPass(dst, pass) bool` that returns false
  and paints nothing in the first implementation. `drawReplayPartial` calls
  it after tile paint and before `blitFrame`, so the call site and accounting
  exist before behavior changes.
- [ ] When promotion is turned on, paint promoted fills and images through that
  seam into the final surface instead of into the tile grid, then mark the
  same tile rects damaged.
- [ ] Add `promotedImageCount` and `promotedFillCount` to the perf counters
  only when the dock is open, so a still generation allocates nothing.
- [ ] Keep `TakeDirty()` and `TakeDirtyRects()` unchanged. A promoted op reports
  its damaged rect the same way as before.
- [ ] Guard the whole pass behind a flag so early runs can disable it without
  reverting code. Default off until Phase F parity passes.

### Phase F, metrics, tests, and parity

- [ ] Unit tests for `imagePromotable` and `solidFillPromotable`, covering
  every escape hatch from section B and at least one positive case each.
- [ ] Merge parity test that compares before and after pixel output on a
  fill-heavy fixture.
- [ ] Image parity test that compares tile path and promoted path output
  byte-for-byte on an image-heavy fixture, including one with a linear
  transform.
- [ ] Cross-check with image diff tooling when available, but keep the
  authoritative gate as `make golden`.
- [ ] Run `make test` on `internal/replay` and `internal/window` before any
  closing rows.
- [ ] Add a bench, `BenchmarkDirectPassPromotion`, that measures the paint
  count and blit count on the departure board at 16 ms pace.

### Phase G, docs and tracking

- [ ] Document the promotion predicates in `internal/replay` package comments
  at the relevant files, so a later maintainer sees the invariants.
- [ ] Update `documentation/repaint.md` with the distinction between tile paint
  and direct pass, and which ops can move.
- [ ] Update `documentation/performance.md` with the promoted count column and
  what zero promotion means.
- [ ] Update `PHASES.md` with commit hashes and before and after numbers.
- [ ] Change the row status in this file's parent README from `Partial` to
  `In progress` only when Phase B lands; from `In progress` to `Shipped`
  only when phases C, D, and E each have a passing gate.

## Exit

- Two predicates exist with tests covering their escape hatches.
- Every fixture that exercises a solid-color or image-heavy page still passes
  `make golden`.
- A developer can read, from a run, how many fills and images were promoted and
  how many stayed in tile paint.
- No promotion path allocates on a still generation when devtools are closed.
- No change to `layout.DisplayList` or to blinkless. The engine stays a pure
  function returning ops.

## Risks

- Promoting into the final surface can reorder paint relative to tile draws.
  The escape hatch that checks blend mode and group state catches the common
  case, but a missed case silently blends the wrong way. Coverage on real
  fixtures is the only proof, which is why Phase F keeps parity.
- Merging fills blindly across ops can hide a later operation that painted on
  top of the gap. The merge stops at the first non-continuing op, documented
  above, but the geometric test must be exact.
- Holding decoded images in a second short-lived layer can double memory for a
  large page. The early bound of a handful per frame limits that cost and
  ties eviction to frame end.
- A promoted op that is forgotten in damage tracking creates ghost pixels on a
  partial frame. Phase E reuses the same damaged rect list, so promotion and
  damage always agree.

## Non-goals

- No change to `internal/chunk` baking policy. Chunk runs remain the feature of
  choice for mixed content that does not qualify as a solid fill or an image.
- No text promotion in this phase. Text keeps `text_raster` and the bake path.
- No Ebiten version change and no backend-specific swap path. That work is
  phases 05 and 03, not this one.
- No async decode. Decoding stays synchronous on first use per
  `internal/replay/image_cache.go`.
