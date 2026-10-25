# Phase 02 - keyframe transform model

Goal: `transform` keyframes mean what CSS says they mean. Translate, scale,
rotate, skew, matrix lists, 3D functions, and perspective produce the same
geometry as a browser, on every element shape, with the same origin.

## Where it lives

- `internal/animation/animation.go:19-24` is the whole keyframe payload today:
  `X`, `Y`, `Opacity`, `R`, `G`, `B`, and three presence flags.
- `internal/cssanimation/keyframes.go:20-31` accepts one bare `translate*()` in
  px, a unitless opacity, and a hex colour. A list, a scale, a rotate, or a
  percentage fails the prefix test in `values.go:36-64` and is dropped.
- `internal/animation/compose.go:11-16` samples six scalar fields.
- `internal/cssmotion/paint.go:19-23` writes X, Y, W, H, RGB, Alpha, and paint
  opacity. It never touches the matrix.
- `internal/frame/anim_write.go:9-52` already composes a scale or rotation about
  an op's painted center through `SetXform`. Read it for the composition order,
  but do not import it and do not merge the packages.
- `internal/replay/xform_geom.go:15-26` turns an op matrix into an Ebiten `GeoM`.
  The matrix is already a supported, tested paint field.
- `blinkless@v0.0.3` `internal/layout/transform.go:1274-1288` bakes the composed
  CTM into every descendant op. `layout.Box` exposes X, Y, W, H only.

## What Chrome does

Blink parses the transform list, resolves percentages against the border box,
applies `transform-origin`, and composes a 4x4 matrix. Interpolation walks the
list pairwise when the function lists match, and falls back to matrix
decomposition when they do not (`translate(10px) rotate(5deg)` to
`scale(2)` blends through decomposed translate, scale, rotation, and skew).

## Design

1. Replace the scalar payload with a decomposed 2D transform:

   ```go
   type Transform struct {
       Tx, Ty       float64 // points, already resolved from px and %
       Sx, Sy       float64 // 1 is identity
       Rotate       float64 // radians
       SkewX, SkewY float64 // radians
       Set          bool
   }
   ```

   Keep `X`, `Y` as aliases during migration so `internal/frame` and the preview
   command keep compiling. Delete them only when every caller in the tree reads
   `Transform.Tx/Ty`. The phase is not done while both paths write X/Y.

   The 2D struct is the fast path. A keyframe that contains any 3D function also
   keeps the parsed function list, because a 4x4 matrix cannot be reduced to
   these six fields:

   ```go
   type TransformValue struct {
       Decomposed Transform
       Functions  []Function // present when ThreeD
       ThreeD     bool
   }
   ```

   `animation.Keyframe` stores a `TransformValue`, not just the six floats.

2. Parse the full list in a new `internal/cssanimation/transform.go`. Accept:
   - `translate(x[, y])`, `translateX`, `translateY`, `translate3d`, `translateZ`,
     with px, %, and unitless zero;
   - `scale(x[, y])`, `scaleX`, `scaleY`, `scale3d`, `scaleZ`;
   - `rotate(a)`, `rotateZ(a)`, `rotateX`, `rotateY`, `rotate3d(x,y,z,angle)` with
     deg, rad, grad, turn;
   - `skew(x[, y])`, `skewX`, `skewY`;
   - `matrix(a,b,c,d,e,f)` and the full 4x4 `matrix3d(...)`;
   - `perspective(length)`;
   - `none`.
   A percentage translate resolves against the element's border box at bind
   time. A percentage in `translate()` with no box is a bind error, counted, not
   a silent zero.

   The individual `translate`, `rotate`, and `scale` properties are implemented
   and are separate animation targets. Parse each one independently and compose
   them in CSS order before `transform`: translate, rotate, scale, then the
   transform list. A keyframe that animates `rotate` must not overwrite a static
   `translate` that the stylesheet declares.

3. Interpolation rule.
   - Same function signature: interpolate component-wise after resolving units,
     in 2D and in 3D. This is the exact CSS path and covers almost every
     authored pair.
   - Different signatures, 2D only: decompose both to 2D matrices, interpolate
     translation, scale, rotation, and skew, then recompose. Use the WebKit
     two-dimensional decomposition.
   - Different signatures, any 3D: decompose the 4x4 matrices per CSS Transforms
     Level 2 (translation, scale, skew, perspective, quaternion) and recompose.
     Do not snap a mismatched 3D list; Chrome blends it.
   Put the 2D and 3D paths in `internal/cssanimation/decompose.go` with
   table-driven tests.

4. Apply through the matrix, not by rewriting X and Y. At bind, capture each
   target op's baked matrix `M_base` and the element's origin. At sample:

   ```text
   M_anim   = T(origin) * compose(t) * T(-origin)
   M_target = M_base * M_anim
   op.SetXform(M_target)
   ```

   `M_base` maps page space to device, so the origin is the element's border-box
   origin in page space, not the op's own center. This is where the current
   scalar path is wrong: two ops in one element translate by the same amount
   today, but a rotation about each op's own center is not a rotation of the
   element.

5. Origin. Read `OriginX`, `OriginY`, and `Box` from `Box.Transform`, exported by
   [priority 0](../priority-0-engine-transform-and-perspective.md). Build the
   animation matrix about that origin. Until the pin lands, a non-default origin
   uses the border-box center and is counted as a fidelity gap; it is not the
   finished state of this phase. Do not guess the origin from the baked matrix; a
   static ancestor transform makes that unrecoverable.

6. Static transform on the animated element. `M_base` already contains the
   element's own static transform when the stylesheet has one, so composing on
   top would double it. Priority 0 exports `Own` and `World`, so replace instead
   of compose:
   - no static transform: the animation matrix is the only one,
   - static transform: `M_target = Parent * anim * World^-1 * M_op`.
   The center fallback and the skip counter exist only while priority 0 is in
   flight. Do not ship a heuristic that looks right on the demo and wrong in a
   nested layout.

7. Text ops take the transform. A rotating card rotates its label. Colour writes
   do not, which phase 03 fixes with the field mask. There is no need to exclude
   text from the matrix path.

8. Damage. A matrix change invalidates the op's transformed bounds, not its
   untransformed box. Reuse `internal/page/dirty/op.go:53-78` or
   `internal/replay/xform_bounds.go:14-36`; do not write a third corner-mapping
   function. Phase 07 wires the result into the damage set.

9. `will-change: transform` is not readable today (`compatibility-matrix.md:840`
   records it as ignored). Do not fake it. Layer promotion in phase 08 keys on
   the property being animated, which is a stronger signal anyway.

10. 3D path. A keyframe with any 3D function, or a `perspective()` function in
    the list, does not write `SetXform`. It serializes the interpolated list to
    the canonical CSS form and calls `RestampTransforms` from priority 0, under
    the phase 04 budget. The 2D matrix path stays for 2D-only lists. A 3D element
    is never promoted to a layer.

### 3D and perspective

`perspective`, `perspective-origin`, and every 3D transform function are
implemented and animatable. They are projective. A 2D matrix cannot represent
them, and a retained layer cannot be re-projected by changing its matrix.

[Priority 0](../priority-0-engine-transform-and-perspective.md) adds
`RestampTransforms`, which re-runs the perspective propagation and the stamp walk
for a subtree with an override. This phase parses and interpolates the keyframe;
phase 04 schedules the restamp under the relayout budget. Do not approximate with
a scale, and do not put 3D or perspective on the layer path.

`transform-style` and `backface-visibility` are discrete: they switch a mode, they
do not interpolate, and they need no frame work beyond the rebind that the mode
change already causes.

The fidelity limit is `preserve-3d`: the engine flattens the chain to a 2D
projection, so intersecting 3D planes and true back-face culling are not
rendered. The animation matches the static projection, which is the correct
target for this track. Do not promise a 3D scene.

## Known limits

- Matching `rotate3d()` lists with different axes interpolate the axis components
  linearly. CSS normalizes the axes and falls back to matrices when they differ.
- A 180 degree rotation through the matrix path has two equally short quaternion
  routes. The midpoint sign follows the WebKit/Blink convention; the endpoints
  are correct either way.
- A singular 2D endpoint such as `scale(0)` still uses the 2D decomposition,
  which drops the zero scale.
- `perspective(none)` inside a `transform` list is not parsed. Only the
  `perspective` property can express no perspective.

## Checklist

### A. Model

- [ ] Add `TransformValue` to `internal/animation`: the 2D decomposition plus a
      parsed function list when any 3D function is present.
- [ ] Keep X and Y readable for one release, then delete them and update
      `internal/frame` call sites in the same change.
- [ ] Teach `compose.go` to interpolate the transform and to leave W, H, Alpha,
      and the other fields alone.
- [ ] Add a table test for each function, each unit, and a 3-item list.

### B. Parser

- [ ] `parseTransformList` with a tokenizer that rejects unknown functions
      instead of silently skipping the value.
- [ ] Percentage resolution against the element box at bind time.
- [ ] `matrix()` and the full 4x4 `matrix3d()`.
- [ ] `rotateX`, `rotateY`, `rotateZ`, `rotate3d`, `translate3d`, `translateZ`,
      `scale3d`, `scaleZ`, and `perspective()`.
- [ ] Matching 3D lists interpolate component-wise.
- [ ] Mismatched 3D lists decompose through the CSS 4x4 algorithm, not a snap.
- [ ] Individual `translate`, `rotate`, and `scale` properties, composed before
      `transform` in the fixed CSS order.
- [ ] `perspective` and `perspective-origin` through `RestampTransforms`, under
      the phase 04 budget.
- [ ] `transform-style` and `backface-visibility` treated as discrete.
- [ ] Reject a malformed list with one counted error per element, not per frame.
- [ ] Tests copied from the CSS transform spec examples.

### C. Apply

- [ ] Capture `M_base` and the element origin at bind.
- [ ] Compose about the origin and call `SetXform` once per target per frame.
- [ ] Serialize a 3D list to canonical CSS and call `RestampTransforms` under the
      phase 04 budget.
- [ ] Never promote a 3D element to a layer.
- [ ] Skip the write when the composed matrix equals the last written matrix.
- [ ] Keep the text path on the same matrix, with no special case.
- [ ] Verify against a nested transform: parent translated, child rotated.

### D. Engine gap

- [ ] Land [priority 0](../priority-0-engine-transform-and-perspective.md) in
      blinkless: the matrix type, `TransformInfo` on `Box`, and
      `RestampTransforms`.
- [ ] Pin the new blinkless revision and update `go.mod` and `examples/go.mod`.
- [ ] Remove the center-origin fallback and the skip counter once the export
      exists.
- [ ] Read `Box.Transform.Parent`, `Own`, and `World` instead of composing from
      op matrices.
- [ ] Record the fallback in `compatibility-matrix.md` only while it is active.

### E. Example parity

- [ ] Add rotate, scale, skew, and matrix cards to the animations example.
- [ ] Add rotateX, rotateY, translateZ, and matrix3d cards.
- [ ] Add a perspective and a perspective-origin card.
- [ ] Add a card with `transform-origin: top left`.
- [ ] Compare each card with the same CSS in a browser at 25%, 50%, and 75%.

Gate: `make golden` for a rotate card with text and an image, plus a manual
browser comparison at three timestamps for 2D, 3D, and perspective. The 2D matrix
path must not re-lay out the page; the 3D path must match a fresh layout.

## Risks

- Matrix composition order is easy to get backwards. Write the invariant in a
  comment: `M_target = Parent * anim * World^-1 * M_op`, and `M_anim` is built
  around the origin.
- 3D matrix decomposition is the hardest code in this phase. Port the CSS
  Transforms algorithm and test against Chrome, not against intuition.
- Serializing a transform list per frame can allocate. Use a reusable buffer and
  measure; a 3D animation is already on the slow path, but zero allocations is
  still the target.
- A transformed op that is already in a blend group re-bakes its chunk because
  the signature includes the matrix. That is expected here and is what phase 08
  removes for promoted elements.
- Scale below 1 multiplies into glyph raster quality. Do not re-rasterize per
  frame; the layer path is the answer, except for 3D, which never promotes.
