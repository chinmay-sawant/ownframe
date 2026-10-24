# Priority 0 - engine transform, 3D, and perspective support for animation

> **Status:** planning, blocks animation phase 02
> **Track:** [animations and transitions](README.md)
> **Upstream:** `github.com/chinmay-sawant/blinkless`
> **Why priority 0:** without these exports, an animated `transform` is wrong on
> any element with a non-default `transform-origin`, a `transform-box` other than
> the default, or a static transform in the same rule. An animated `perspective`
> or a 3D transform function (`rotateY`, `translateZ`, `matrix3d`) is impossible
> at all, because a 2D layer matrix cannot carry a projective map.

## The problem

Blinkless implements `transform`, `transform-origin`, `transform-box`,
`perspective`, and `perspective-origin`. It resolves them, bakes the element's own
matrix about the resolved origin, multiplies it by the ancestor CTM, and stamps
the result onto every descendant op
(`internal/layout/transform.go:1439-1460`, `stampBoxTransforms` at `:1274`).

`layout.Box` exposes X, Y, W, H, animation records, transition records, and
operation ids. It does not expose any transform value
(`layout/layout.go:10-33`). The information exists during layout and is thrown
away before the host sees the tree.

Consequences for the animation runtime:

1. `transform-origin: top left` cannot be applied. The host only has the element
   box, so it defaults to the center and the rotation is visibly wrong.
2. `transform-box: content-box` and `view-box` cannot be replicated. The origin
   reference box differs from the border box.
3. An element with a static `transform: rotate(10deg)` and an animated
   `transform` would be composed on top of the baked static one, so the element
   rotates 20 degrees on the first frame.
4. A nested element inherits its ancestor's stamped matrix. The host cannot tell
   which part came from the parent and which part belongs to the element.
5. A changing `perspective` changes the projection of every descendant. The host
   only receives flattened 2D matrices, one per op, and cannot derive the new
   projection from them.
6. A 3D transform function is flattened to 2D at stamp time. The host can
   interpolate `rotateY` or `translateZ` values, but it cannot project the result
   onto the page without the engine's camera math.

Two facts make the export urgent:

- `internal/layout.Matrix2D` is internal. Hosts cannot name it, so
  `internal/frame/anim_matrix.go:8-13` defines its own `mat6` and copies fields
  through a zero `DisplayOp` just to obtain a value of the type
  (`anim_write.go:13-16`).
- `identityMatrix`, `BakeOrigin`, `Mul`, and `Apply` already exist internally
  (`transform.go:73-136`). Exporting them is a wrapper, not new math.

## 1. A nameable matrix type

Proposal for blinkless `layout`:

```go
// layout/matrix.go
type Matrix2D = ilayout.Matrix2D

func IdentityMatrix() Matrix2D
func BakeOrigin(m Matrix2D, ox, oy float64) Matrix2D
```

`Matrix2D` already has `A, B, C, D, E, F`, `IsIdentity`, `Mul`, and `Apply`. The
alias makes the type nameable outside the engine. `BakeOrigin` is the
origin-composition helper the host needs to build a 2D animation matrix.

Do not export `ProjectiveTransform` or the 3D chain. The host supplies a transform
value, the engine projects it.

## 2. Per-element transform metadata

```go
// layout/layout.go
type TransformInfo struct {
    // OriginX and OriginY are in layout points, the same unit as Matrix2D's
    // E and F. Convert to CSS pixels with Display.PointsPerPixel when needed.
    OriginX, OriginY float64
    Box              string // border-box | content-box | fill-box | stroke-box | view-box

    // Parent maps the parent coordinate space to device space. Own is this
    // element's own baked transform. World is Parent * Own, the matrix the
    // engine stamps on the element's ops.
    Parent Matrix2D
    Own    Matrix2D
    World  Matrix2D

    HasMatrix      bool
    HasTransform3D bool
    Perspective    float64 // layout points, 0 when none
    PerspectiveOriginX, PerspectiveOriginY float64
    HasPerspective bool
    Style          string // flat | preserve-3d
    Backface       string // visible | hidden
}
```

`Box` gains one field:

```go
Transform TransformInfo
```

Why all four matrices:

- `Own` is what an animation replaces.
- `Parent` is needed because `Own` and the animation matrix live in different
  spaces. The host cannot apply a page-space delta to an op whose matrix already
  contains the ancestor CTM.
- `World` is a convenience for the common case and a cross-check in tests.
- `HasTransform3D` tells the host to use the restamp path and to skip layer
  promotion. It is not a skip flag anymore.

Composition for a 2D-only element, once these fields exist:

```text
base       = Parent * Own              // element's accumulated matrix
M_new_op   = Parent * anim * base^-1 * M_op
```

`base^-1 * M_op` removes the element and its ancestors, leaving the descendant's
local matrix; `Parent * anim` rebuilds the chain with the animated transform.
This is correct for every op in the element's subtree, including text and images.

A 3D element does not use this formula. It goes through `RestampTransforms`.

### Populate at the same point as the stamp

`walkPlaced` builds `PlacedElement` (`internal/layout/placed.go:40-70`). The
transform values are resolved in `boxTransformAccum`
(`internal/layout/transform.go:1439-1460`) and the origin in
`resolveTransformOriginWithBox` (`:165`). The two functions must not diverge.

Options:

1. Compute the values once in `walkPlaced` from `boxNode.style` and the box.
2. Record them on the `box` node during the stamp walk, then copy in `walkPlaced`.
3. Return them from `boxTransformAccum` and pass them down.

Prefer 1 if the origin and the own matrix can be recomputed cheaply and
identically. Prefer 2 if the preserve-3d chain makes a second computation unsafe.
The acceptance test is equality with the stamped op matrix, not which option was
chosen.

Note the unit rule: `Box.X/Y/W/H` are CSS pixels (`ptToPx`), `Matrix2D.E/F` and
`OriginX/Y` are layout points. Document the exception on the field, and add a
test that catches a unit mix-up.

## 3. 3D transform and perspective re-projection

This is the second half of priority 0, and it cannot be solved on the host.

### Why a 2D matrix cannot carry it

`perspective` belongs to an element but applies to its children
(`internal/layout/perspective_pass.go:1-45`). The pre-pass propagates the nearest
ancestor's distance and origin into every 3D-bearing descendant. At stamp time,
`perspectiveFlattenArgs` (`perspective_pass.go:53-63`) picks the camera, and
`boxTransformAccum` flattens the 3D matrix to a 2D one
(`transform.go:1445-1452`).

A projective map is not affine. Ebiten's `GeoM`, the op matrix, and the retained
layer all hold a 2D affine transform. If `perspective` changes, every descendant's
flattened matrix changes, often non-linearly across the subtree. The same is true
for any 3D function: `rotateY(45deg)` has no 2D affine equivalent, and its
foreshortening depends on the camera that the engine applied.

The engine already has the correct math. The missing piece is a way to run it
again for a subtree, with an overridden transform or perspective, without
re-running the cascade or layout.

### Proposed API

```go
// layout/reproject.go
type TransformOverride struct {
    ElementID string

    // Transform replaces the element's own transform for this pass.
    // A raw CSS transform list. HasTransform=false leaves it unchanged;
    // HasTransform=true with "" means none.
    Transform    string
    HasTransform bool

    // Origin optionally replaces transform-origin, in layout points.
    OriginX   float64
    OriginY   float64
    HasOrigin bool

    // Perspective optionally replaces the perspective property for this
    // element's descendants, in layout points. Distance 0 means none.
    Perspective          float64
    HasPerspective       bool
    PerspectiveOriginX   float64
    PerspectiveOriginY   float64
    HasPerspectiveOrigin bool
}

func RestampTransforms(ctx context.Context, doc *css.Document, prev *Display,
    overrides []TransformOverride, options Options) (*ElementPatch, error)
```

The transform value is a CSS list, not a matrix, so the engine parses it with the
same `parseTransformList` the cascade uses. That keeps 2D functions, 3D
functions, `matrix()`, `matrix3d()`, and a `perspective()` function inside the
list on one code path, and it avoids freezing an internal representation.

Semantics:

- Find each element by id in the retained box tree.
- Treat `Transform`, `Origin`, and `Perspective` as computed values for this pass
  only. Do not write them into the tree or the style cache.
- If both a transform and a perspective are overridden, apply the transform to
  the element, then the perspective to its descendants, in the same order the
  cascade would.
- Re-run the perspective propagation for that element's subtree, then the stamp
  walk for the same subtree. Do not re-cascade and do not re-place. Seed the
  subtree walk with the nearest ancestor's resolved perspective, so overriding an
  element to "none" restores the outer camera instead of inventing an
  orthographic one.
- Preserve the 3D chain and preservation rules the static path uses. A
  `preserve-3d` subtree must flatten the same way it does statically.
- Return an `ElementPatch` with `Old` and `New` damage, like `RelayoutElement`, so
  the host adopts it through the existing splice path.
- Require `RetainState`, like `RelayoutElement`. Without the retained box tree the
  pass cannot run.
- With no overrides, return an empty patch and leave every matrix untouched.

Why an override list rather than one element: a page can animate several 3D cards
and perspective cards in one frame. One call with a list avoids N stamp walks.

### What the host sends

For 2D-only transforms, the host keeps the current fast path and never calls this
function. For 3D transforms and perspective, the host serializes the interpolated
transform list to CSS and calls `RestampTransforms`:

```text
rotateY(37.5deg) translateZ(20px)
perspective(500px) rotateX(12deg)
matrix3d(1,0,0,0, 0,1,0,0, 0,0,1,0, 0,0,0,1)
```

Use a reusable buffer and `strconv.FormatFloat` with enough precision to match the
engine's parser. Measure the format cost; it is on the slow path, but it must not
allocate per frame.

If the string cost ever matters, a later API can accept a parsed value list. Do
not add that now.

### What it deliberately does not do

- It does not build a 3D scene graph. The engine flattens a `preserve-3d` chain to
  the same 2D projection it uses statically. Intersecting 3D planes and true
  back-face culling are out of scope; `backface-visibility` remains a discrete
  engine behavior.
- It does not expose `ProjectiveTransform` or the 3D chain.
- It does not promote a 3D element to a compositor layer. A 3D projection cannot
  be a layer matrix.
- It does not change the static path. A page that never calls it produces the
  same display list as today.

### Cost

The pass is O(boxes + ops) in the subtree, with no text shaping and no cascade.
A leaf 3D transform is one box and its ops; a perspective change is the whole
subtree. Both go through the same per-frame budget as the phase 04 relayout
queue. Measure a 20-card 3D and perspective page before raising the budget.

## What not to export

- A full 3D matrix or the preserve-3d chain. The host does not need it, and
  exposing it would freeze an internal representation.
- The raw `transform` parser.
- A `SetTransform` method on the box. The host writes 2D through `op.SetXform`
  and 3D through `RestampTransforms`.
- Per-op transform ancestry. `World` plus the op matrix is enough for 2D.

## Phases

### P0.1 - API shape and review

- [ ] Add `layout/matrix.go` with the `Matrix2D` alias, `IdentityMatrix`, and
      `BakeOrigin`.
- [ ] Add `TransformInfo` to `layout/layout.go`.
- [ ] Add `TransformOverride` and `RestampTransforms` to `layout`.
- [ ] Document every field's unit: CSS pixels for the box, layout points for the
      matrix, origin, and perspective distance.
- [ ] Get upstream review on the field names and the restamp signature before
      implementing.
- [ ] Decide whether `OriginX/Y` ships in points only, or with px convenience
      fields. Do not ship both without a reason.

### P0.2 - Populate transform metadata

- [ ] Resolve the origin once, with the same `transform-box` rules as the stamp
      walk.
- [ ] Compute `Parent`, `Own`, and `World` for every element box.
- [ ] Set `HasMatrix` only when the element has a transform.
- [ ] Set `HasTransform3D` when the element participates in a 3D composition.
- [ ] Export `Perspective`, its origin, `Style`, and `Backface`.
- [ ] Populate Parent and origin for every element box, including one with no
      static transform. `HasMatrix` reports a static transform only.
- [ ] Do not change op emission. The display list must be byte-identical when
      the fields are ignored.

### P0.3 - 3D and perspective re-projection

- [ ] Refactor `propagatePerspective` to accept a root box and an override map,
      or add a sibling function that does not mutate the cached style.
- [ ] Refactor the stamp entry to run on a subtree root.
- [ ] Implement `RestampTransforms` with the override list, retained state, and
      an `ElementPatch` result.
- [ ] Parse the `Transform` override with the cascade's transform parser, so 2D
      and 3D functions stay on one path.
- [ ] Support `rotateX`, `rotateY`, `rotateZ`, `rotate3d`, `translate3d`,
      `translateZ`, `scale3d`, `scaleZ`, `matrix3d`, and a `perspective()`
      function in the list.
- [ ] Preserve nested overrides: outermost first, inner wins.
- [ ] Preserve the static 3D chain and `preserve-3d` flattening.
- [ ] Return an empty patch when no override changes a value.
- [ ] Add a test that `RestampTransforms` with a transform T produces exactly the
      matrices a fresh layout with that declaration produces.

### P0.4 - Blinkless tests

- [ ] `transform-origin: 0 0` and `50% 50%` match Chrome 143 at four corners.
- [ ] `transform-box: content-box`, `fill-box`, and `view-box` resolve the origin
      against the same box the stamp uses.
- [ ] A static transform plus a nested element: `World` equals the stamped op
      matrix.
- [ ] A parent transform plus a child transform: `Parent` and `Own` separate
      cleanly, and `World` reconstructs the op matrix.
- [ ] `perspective` and `perspective-origin` values match the resolved style.
- [ ] `rotateY`, `rotateX`, `translateZ`, `scaleZ`, `rotate3d`, and `matrix3d`
      each match a fresh layout at 25%, 50%, and 75%.
- [ ] `perspective(500px) rotateY(45deg)` matches a fresh layout.
- [ ] A parent `perspective` plus a child `rotateY` matches a fresh layout.
- [ ] A `preserve-3d` chain flattens the same way statically and through an
      override.
- [ ] Animating perspective 500 to 1000 matches a fresh layout at 25%, 50%, and
      75%.
- [ ] `perspective-origin: top left` matches Chrome.
- [ ] Nested perspective: the inner override wins for its subtree.
- [ ] No-transform elements report zero values and `HasMatrix == false`.
- [ ] With no override, `RestampTransforms` leaves every matrix untouched.

### P0.5 - Release and pin

- [ ] Merge upstream, tag the next blinkless version.
- [ ] Bump `go.mod` and `examples/go.mod` to that tag. No replace.
- [ ] Run `make build` on both modules.
- [ ] Record the new pin in `AGENTS.md`.

### P0.6 - Consume in OwnFrame

- [ ] Replace the center-origin fallback in animation phase 02.
- [ ] Delete the `mat6` duplicate in `internal/frame` only if the exported type
      makes it unnecessary. Do not merge the package.
- [ ] Use `Parent * anim * World^-1` for a 2D target matrix, not a per-op guess.
- [ ] Add a test: an element with a static rotation and an animated rotation does
      not double.
- [ ] Add a test: `transform-origin: top left` matches a browser.
- [ ] Route 3D transform keyframes and perspective keyframes through
      `RestampTransforms` under the phase 04 budget.
- [ ] Never promote a 3D element to a layer in phase 08.
- [ ] Add a test: rotateY and perspective animations match a browser at three
      timestamps.
- [ ] Remove the blocked rows from `coverage.md` when the pin lands.

## Acceptance

- `Box.Transform.World` equals the matrix stamped on the element's first op, for
  every box on a page with nested transforms.
- A host can replace the element's 2D transform with an animated one and get the
  browser result, with no access to blinkless internals.
- A host can animate a 3D transform or `perspective` and get the same op matrices
  a fresh layout would produce at that value.
- A page with no transforms and no perspective produces a byte-identical display
  list before and after the change.
- The public API is additive. Existing blinkless callers compile unchanged.

## Risks

- Unit mismatch is the most likely bug. Points in the matrix, pixels in the box,
  and a test that fails on a mix-up.
- Preserve-3d composition is not a 2D chain. Do not export a `Parent` that
  pretends otherwise. Reuse the existing chain logic; do not reimplement it.
- Recomputing the origin in `walkPlaced` can drift from the stamp. Add the
  equality test before choosing the implementation.
- The matrix alias exposes an internal type. Keep its methods documented and its
  layout stable; a breaking change there breaks hosts.
- The re-projection pass must not mutate the cached style. An override that leaks
  into the next static frame turns an animation into a permanent change.
- String serialization per frame can allocate. Give the host a buffer and measure
  it; a 3D transform animation is already on the slow path, but zero allocations
  is still the target.

## Known limits after the first implementation

- A restamp does not cull back faces. Zeroing an op would destroy geometry that a
  later frame cannot restore, and an op already culled by the static pass stays
  culled. Changing `transform-style` or `backface-visibility` still needs a
  relayout.
- Table-cell targets return `ErrNotRestampable`. The static path does not set
  `needsXformStamp` for a transform declared on a cell alone, so a restamp would
  paint a projection the page does not show. A table element itself works.
- `em` lengths in an override resolve against the element's resolved font size.
  There is no `rem` base threaded through, matching the parser's existing call
  site.
- The pass mutates the retained placement in place and is not goroutine-safe, the
  same contract as `RelayoutElement`.
- Pristine op geometry is retained only on ops that carry a projective transform,
  which is the rare 3D set. A 2D op still restamps from its live matrix.
- The host can restamp only an element with an id. A duplicate id resolves to the
  first box. An element without an id is counted and skipped.
- The returned patch does not refresh `Parent`/`World` on the boxes, so a 2D
  transform animation nested inside a restamped 3D subtree composes against the
  metadata from before the override. Give nested 2D motion its own pass or
  refresh the metadata.
- The host rebinds motion after each restamp, so `Stats.Rebinds` grows during a
  3D animation. That is correct but worth watching in a profile.
- `perspective(none)` inside a `transform` list is not parsed. Only the
  `perspective` property can express no perspective.

## Gate

Blinkless behavior tests pass against the Chrome reference, the display list is
unchanged when the fields are ignored, `RestampTransforms` matches a fresh layout
for 3D transforms and perspective, the pin is bumped in both modules, and
OwnFrame's transform, 3D, and perspective cards pass with an explicit origin and a
static transform.
