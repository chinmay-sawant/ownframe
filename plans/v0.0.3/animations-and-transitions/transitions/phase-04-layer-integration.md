# Phase 04 - layer integration

Goal: a transform, opacity, filter, clip, or display transition uses the retained
layer from animation phase 08. It moves or fades a texture instead of re-baking
the subtree, and the two tracks share one cache and one budget.

## Where it lives

- `internal/replay/layer.go` (animation phase 08) holds the layer record, the
  cache, and the composite matrix and alpha path.
- `internal/replay/group_layer.go:10-37` is the existing render-to-texture code
  the layer builds on.
- `internal/replay/group_composite.go:15-47` is the translate-only composite the
  layer replaces.
- `internal/motion` owns both registries, so a promotion decision can see an
  element with an animation and a transition at the same time.

## Design

1. One cache, one budget. Do not create a transition-specific layer store. A
   layer is keyed by element id and generation and is shared by both tracks. An
   element that has both an animation and a transition owns one layer, not two.

2. Promotion at capture. When a capture produces a transform, opacity, filter, or
   clip transition, ask the layer cache for that element. If the layer does not
   exist and the element passes the promotion predicate, render it once in the
   base pose. If it exists, reuse it. The transition then only changes the matrix
   or alpha per frame.

3. Group alpha for opacity. A transition with opacity on a subtree that overlaps
   its own children must composite with `ColorScale` alpha, not per-op alpha. The
   layer gives the correct result; per-op fades double-darken the overlap. This
   closes the fidelity gap recorded in animation phase 03.

4. Filter transitions. The layer texture is retained; the filter is applied at
   composite time. A filter parameter change re-runs the shader, not the subtree
   raster. Interpolate filter functions that have the same shape; a different
   function list is discrete.

5. Clip transitions. Rasterize the clip shape once per frame into the mask
   texture and apply it with `BlendDestinationIn`, the same path
   `group_mask.go` uses. The layer content stays cached.

6. Display cross-fade. A `display` transition that removes an element needs the
   outgoing layer after the new display has dropped the ops. Capture the outgoing
   element into a layer at commit, bind a fade-out on it, and release the layer
   at 100%. An entry transition fades the incoming layer in. Two layers exist at
   once only for the cross-fade, and both count against the budget.

7. Lifetime. A layer promoted for a transition is released when the transition
   ends, cancels, or is replaced by a non-layer property. Release is explicit and
   does not wait for LRU eviction; the cache is a budget, not a garbage collector.

8. Damage. The transition damages `old ∪ new` layer bounds, exactly like an
   animation. A display cross-fade damages the union of both layers. The tile
   repaint skips the promoted element's ops and composites the layer once.

9. Conflict with animation. If the same element has a promoted animation layer,
   the transition uses that layer and its matrix and alpha are combined. Define
   the order: transition matrix first, animation matrix second, alpha multiplied.
   That matches the cascade, where the animation wins on the same field, but the
   two fields are independent transforms, so composition is the only sensible
   answer. Record the combined case in a test.

10. Demotion. A transition on a large subtree can be cheaper per op than as a
    layer, particularly a colour transition with no transform. The promotion
    predicate from animation phase 08 already excludes colour-only changes; keep
    that rule and do not promote every transition.

## Checklist

### A. Shared cache

- [ ] Use the phase 08 cache and budget; no second store.
- [ ] Key by element id and generation.
- [ ] One layer for an element with both an animation and a transition.
- [ ] Release on end and cancel.
- [ ] `Close` releases transition layers too.

### B. Composite

- [ ] Transform transition through the layer matrix.
- [ ] Opacity transition through group alpha.
- [ ] Overlapping children fade once, not twice.
- [ ] Filter transition re-runs only the shader.
- [ ] Clip transition rasterizes only the mask.
- [ ] Combined animation and transition matrix order is tested.

### C. Display cross-fade

- [ ] Capture the outgoing element before the new display drops it.
- [ ] Fade out for the full duration.
- [ ] Fade in the incoming layer at 0%.
- [ ] Release both layers at the end.
- [ ] Damage covers both rects.

### D. Damage and paint

- [ ] Old ∪ new bounds.
- [ ] Skip the promoted ops in tile repaint.
- [ ] Composite inside a group when the group encloses the element.
- [ ] No stale pixels after release.
- [ ] Golden frames at 0%, 50%, and 100%.

### E. Budget

- [ ] Cross-fade counts two layers.
- [ ] A full budget demotes with a reason.
- [ ] A transition never evicts a live animation layer it does not own.

Gate: a transform transition and an opacity transition each keep their subtree
texture across the run, a display exit keeps its pixels, and the animation layer
plus transition layer case composites in the documented order.

## Risks

- A display cross-fade is the one place where two layers for one element are
  correct. Do not generalise that to a layer tree.
- A transition that promotes and then demotes mid-run will flicker. Decide
  promotion at capture and keep it until the end, unless the budget forces a
  demotion; if it does, snap the remaining transition rather than blend paths.
- Group alpha and per-op alpha cannot be mixed for one element in one frame. The
  layer is authoritative while it exists.
