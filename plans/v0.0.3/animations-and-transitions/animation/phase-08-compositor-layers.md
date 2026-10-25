# Phase 08 - compositor layers

Goal: an animated element with a transform, opacity, filter, or clip is rendered
once into a retained texture, then moved and composited per frame. The page under
it is repainted only where the old and new positions cover, and the element's
descendants are not re-rasterized.

This is the Chrome behaviour the current architecture is missing, and it is the
reason this phase exists. It is not a second thread. Ebiten draws on the game
goroutine; the win comes from not repainting content, not from parallelism.

## First milestone landed

The replay side is in:

- `layer.go` captures an element's owned ops in paint order and renders the base
  pose into a texture. The content signature clears `XformSet`, so the animation
  matrix never re-bakes the texture.
- `layer_cache.go` keys layers by element id, box index, and generation, with
  16-layer and 32 Mi-pixel caps, LRU eviction, and Begin/Demote/Close lifecycle.
- `layer_paint.go` composites at the first paint-order position and skips every
  owned op. The window merges old and new pose damage.
- `group_composite.go` shares the matrix and alpha composite path.
- `internal/window/replay_partial.go` and `replay_tile_paint.go` consume a
  `LayerSource`, but no page implements it yet.
- A rotated-card benchmark pays 40 op draws once, then one background op and one
  layer blit per frame.

Still open:

- [ ] Add `LayerSpecs(display)` to the page or its adapter and expose active
      entries, property masks, and group alpha from `internal/motion`.
- [ ] Do not bake the sampled alpha into the ops when a layer applies it.
- [ ] Group-contained promotion. A page with any blend, isolation, or mask group
      currently demotes every layer.
- [ ] Nested and overlapping promoted elements; they are demoted today.
- [ ] Viewport and full-replay paths. Only the tile path consumes layers.
- [ ] Layer counters in `Probe().Motion()`.
- [ ] Filter, clip-path, and mask animations on a layer.
- [ ] Add the gated GPU layer checks to `make graphics-check`.
- [ ] 0/50/100 percent goldens and a 43-card benchmark.

## Where it lives

- `internal/replay/group_layer.go:10-37` already renders a subtree into an
  `*ebiten.Image` and composites it. The image is allocated and deallocated
  inside one call.
- `internal/replay/group_composite.go:15-47` composites with a translation only.
  The blend path copies the destination and runs a Kage shader.
- `internal/replay/group_plan.go:18-28` holds group identity and bounds.
  `group_first.go:29-56` computes the bounds once per display pointer.
- `internal/replay/chunk_cache.go:16-55` retains a run texture with a content
  signature and re-bakes in place. This is the invalidation model to copy.
- `internal/replay/chunk_sig_op.go:11-47` and `chunk_sig_extra.go:10-52` show
  which fields must enter a signature: geometry, colour, alpha, opacity, matrix,
  text, image bytes, grid segments, group id and mode.
- `internal/viewtransition/ownership.go:5-76` captures one element's ops in paint
  order through `Box.OperationIDs` and `OpOwners`. Reuse it; do not write a second
  subtree walker.
- `internal/window/replay_tile_paint.go:33-49` repaints a tile with
  `DrawVisibleUnfixedZ`. `internal/replay/pinned_z.go:45-60` culls by tile.
- `internal/render/replayable_group.go:8-44` enforces depth 4, separable blend
  modes, and no fixed or z-indexed members inside a masked or non-normal group.
- `documentation/optimizations.md:432-439` names the gap: "What is missing is a
  layer the window can blit without repainting the page under it."

## What Chrome does

Paint produces a layer tree. The compositor owns a texture per layer. A transform
or opacity animation changes the layer's matrix or alpha, not its pixels. Damage
is the union of the old and new layer bounds; the area below the old position is
repainted from the layers beneath.

## Design

### 1. Layer record

```go
type Layer struct {
    ElementID string
    Gen       uint64
    Box       image.Rectangle // page space, content scale
    Tex       *ebiten.Image
    Sig       uint64
    Ops       []int           // owned op indices, paint order
    Alpha     float64
    Matrix    mat2D           // animation matrix in page space
    LastUse   uint64
    Valid     bool
}
```

One cache in `Renderer`, beside the chunk store, with an explicit pixel budget
and a count cap. Eviction is by `LastUse`, not insertion order, because the layers
that matter are the ones still animating.

### 2. Capture and render

At the first frame an element is promoted, capture its owned ops in paint order
through the existing `viewtransition` helper. Render them into a texture sized to
the union of their painted boxes, not the element's border box: content that
overflows is part of the layer.

Render in the base pose, with the matrix the display already carries. Do not bake
the animation matrix into the texture. The animation matrix is applied at
composite time, so a moving layer never re-bakes.

### 3. Composite

Replace `compositeLayer`'s translation-only path with:

```go
op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
op.GeoM = layerMatrixToGeoM(l.Matrix, l.Box, pageOffset)
op.ColorScale.ScaleAlpha(float32(l.Alpha))
dst.DrawImage(l.Tex, op)
```

For a filter animation, draw the texture through the existing filter shader
instead of the plain image draw. For `clip-path`, apply a mask texture with
`BlendDestinationIn` before compositing, the same way `group_mask.go` does. For
`mask-position` and `mask-size`, transform the mask texture instead of
re-rasterizing the layer; `mask-border-slice`, `mask-border-width`, and
`mask-border-outset` rebuild the mask texture only, not the layer. A
`mask-image` gradient can interpolate when the stop shape matches; otherwise it
is discrete and switches at the midpoint. `mix-blend-mode` and `isolation` are
discrete mode switches, never interpolated.

### 4. Paint integration

During a tile repaint, when the plan reaches the first op owned by a live layer:
- skip every op in `layer.Ops` for this tile,
- composite the layer through its matrix and alpha, clipped to the tile,
- continue with the ops after the subtree.

The skip must be by op index set, not by element id inside the hot loop. Build a
`map[int]*Layer` once per frame, or a sorted index range when the ops are
contiguous. Prefer the range; `OperationIDs` from one element are contiguous in
practice, and the layer records the range.

Do not paint a layer into a group buffer and again into a tile. A promoted
element inside a blend or isolation group is composited inside that group's
buffer. A promoted element that contains the group is composited in the tile.

### 5. Damage and the hole

A moving layer damages `old ∪ new` transformed bounds. The repaint of that union:
1. fills the background,
2. replays every op that intersects the rect except the layer's own ops,
3. composites the layer once at its current matrix.

This is correct for both rects in one pass: the old position is repainted from
the page content beneath, and the new position receives the layer. Do not try to
"erase" the old layer; the tile is rebuilt from ops.

A layer over a fixed or pinned element needs the same band rules the group code
already has. If the layer's rect crosses a band, demote it.

### 6. Promotion policy

Promote when all of these hold:
- the element has an active animation whose property set includes transform,
  opacity, filter, or clip-path,
- the transform is 2D only. `Box.Transform.HasTransform3D`, a 3D function in the
  keyframe, or a `perspective` animation demotes the element: a projective map
  cannot be a layer matrix,
- it owns at least two ops, or one op that is text or an image,
- it does not animate a layout property in the same frame (phase 04 owns that),
- its captured pixel area is under the layer cap,
- it is not inside a non-separable blend group or a masked group it would
  duplicate,
- its ops do not mix fixed, pinned, or z-indexed content with normal content.

Demote when the animation ends, when the signature changes every frame in a way
that forces a re-bake (a colour animation on a large subtree is cheaper per op),
or when the budget is full. Count every promotion and demotion with a reason.

`will-change` cannot be read (`compatibility-matrix.md:840`), so the animated
property is the hint. Do not add a public promotion API in this phase.

### 7. Invalidation

Re-render the layer when any of these change:
- generation,
- content signature over the owned ops (reuse the chunk signature fields),
- element size or content box,
- theme or font,
- an ancestor layer re-renders and contains this layer.

A colour or text change inside an animated subtree changes the signature, so the
layer re-bakes once and then moves cheaply. That is the intended trade.

### 8. Nested layers

Promote the outermost animated element and keep descendants in its texture. If a
descendant also animates with a different timing, either promote it separately
and make the parent's texture a place where the child is composited at parent
render time (which re-bakes the parent every frame), or demote the child and let
phase 02 patch its matrix inside the parent texture. Choose demotion when the
parent is promoted; record the conflict. Do not build a recursive layer tree in
this phase.

### 9. Caps

Start with 16 layers and 32 Mi pixels per page, both configurable constants. A
layer larger than the cap is demoted, not clipped. Evict the least recently used
when the cap is hit, and make the eviction visible in `Probe()` so a page can be
tuned.

## Checklist

### A. Layer cache

- [ ] `Layer` record and cache in `Renderer`, beside `chunkStore`.
- [ ] Key by element id and generation.
- [ ] Content signature over owned ops.
- [ ] Pixel and count caps with LRU eviction.
- [ ] `Close` releases every texture.
- [ ] Unit test insert, hit, invalidate, evict, and close.

### B. Capture

- [ ] Reuse `viewtransition`'s ordered subset helper.
- [ ] Render the base pose into a texture sized to the painted union.
- [ ] Record the owned op range for the paint skip.
- [ ] Test an element with text, an image, and a rounded fill.

### C. Composite

- [ ] Matrix path in `compositeLayer`, not only translation.
- [ ] Group alpha through `ColorScale`.
- [ ] Filter shader path for filter animations.
- [ ] Clip mask path for clip-path animations.
- [ ] Mask transform for `mask-position` and `mask-size`.
- [ ] Mask rebuild for `mask-border-*`, without re-baking the layer.
- [ ] Discrete switch for `mix-blend-mode`, `isolation`, and a mismatched
      `mask-image` gradient.
- [ ] Test opacity 0.5 with two overlapping children: the overlap darkens once,
      not twice.

### D. Paint

- [ ] Skip owned ops by range during tile repaint.
- [ ] Composite the layer clipped to the tile.
- [ ] Paint promoted content inside a group buffer when the group encloses it.
- [ ] Do not composite a layer into a tile and into a group in the same frame.

### E. Damage

- [ ] Old ∪ new transformed bounds.
- [ ] Repaint skips the layer ops and composites once.
- [ ] Scroll under a layer: the layer moves with the content, not with the tile.
- [ ] Fixed chrome over a layer paints after it.

### F. Policy and metrics

- [ ] Promotion predicate with the six conditions.
- [ ] Demotion for `HasTransform3D` and any 3D or perspective animation.
- [ ] Demotion with a reason counter.
- [ ] Pixel and count caps.
- [ ] Report promoted, demoted, rebaked, and bytes in `Probe()`.
- [ ] Benchmarks: rotate, scale, opacity, filter, and a 43-card page.

### G. Tests

- [ ] Rotate a card with text and an image: descendants do not re-bake.
- [ ] Scale 0.5 and 2.0: no per-frame text re-shaping.
- [ ] Opacity on overlapping children matches a browser.
- [ ] Demotion when the budget is full still renders correctly.
- [ ] Golden frames for each case at 0%, 50%, and 100%.
- [ ] A layer torn down on finish leaves no stale pixels.

Gate: `make graphics-check` plus a paced dump showing fewer `DrawImage` calls and
fewer chunk re-bakes on the rotate card than the phase 02 path, with identical
golden frames.

## Risks

- A layer that is damaged incorrectly leaves a ghost. The old ∪ new union and the
  skip-and-replay rule are the whole correctness argument; test them before
  optimising.
- A layer changes text antialiasing because it is sampled through a matrix.
  Accept it for animated content and keep static text on the direct path.
- The signature must include the matrix only if the matrix is baked. In this
  design it is not, so exclude it; including it would re-bake every frame and
  defeat the phase.
- Group caps and blend modes are load-bearing for correctness. A promoted element
  inside a non-separable blend group must not bypass the group shader.
