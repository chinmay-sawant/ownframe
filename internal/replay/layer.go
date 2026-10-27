package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

// Layer is one retained compositor layer: the operations of one element,
// rendered once into Tex and composited per frame through Matrix and Alpha.
//
// Box is the painted union of the owned operations in the captured base pose,
// in CSS pixels. The texture's origin is Box.Min. Matrix is in the display's
// point space, the same convention as DisplayOp.Transform: its linear terms
// are unitless and its translation is in layout points. It maps the captured
// base pose to the animated pose, so a moving layer never re-bakes.
//
// The texture holds each owned operation at its own opacity: the sampled group
// alpha is divided out at bake time and applied once at composite time, so a
// group of overlapping children darkens once instead of once per child. The
// content signature excludes Matrix, Alpha, and the operation transforms baked
// into the base pose. A transform or opacity track, whether an animation or a
// transition, therefore reuses the texture; a geometry, colour, text, image, or
// grid change re-bakes it.
type Layer struct {
	ElementID string
	Gen       uint64
	Box       image.Rectangle
	Tex       *ebiten.Image
	Sig       uint64
	Ops       []int
	// Alpha is the element's group opacity, applied once when the texture is
	// composited. It is not baked into the texture.
	Alpha   float64
	Matrix  layout.Matrix2D
	Mode    string
	LastUse uint64
	Valid   bool

	// firstRank is the layer's first position in Display.Order. The paint
	// walk composites the layer there and skips every owned op index.
	firstRank int

	// pose is the last composited page box, kept for old ∪ new damage.
	pose image.Rectangle

	// misses counts consecutive frames whose content signature changed. A
	// layer that re-bakes every frame costs more than painting it directly,
	// so it is demoted after a couple of misses.
	misses int
}

// The screen's display must stay drawable without layers: the full replay and
// viewport paths do not composite them. A page whose animation writes the
// sampled matrix into the display therefore passes the inverse matrix as Base,
// so the layer holds the base pose while the direct path still shows the
// current pose.
//
// LayerSpec is the promoted compositor layer request. The canonical definition
// lives in internal/host, where the window reads it as a screen role; this
// alias keeps the renderer and its callers on the same type without making the
// page import replay.
type LayerSpec = host.LayerSpec

// LayerSource is the optional screen role the window asks for this frame's
// promoted elements.
type LayerSource = host.LayerSource

// LayerStats is one snapshot of the layer cache. Promoted, Demoted, Rebaked,
// Evicted, and Reasons describe the last BeginLayers call; Active and Pixels
// describe the cache now. Bytes is Pixels times four bytes per pixel.
type LayerStats struct {
	Active   int
	Promoted int
	Demoted  int
	Rebaked  int
	Evicted  int
	Pixels   int
	Bytes    int
	Reasons  map[string]int
}

// Layers returns the layer cache counters.
func (renderer *Renderer) Layers() LayerStats {
	return renderer.layers.snapshot()
}

// layerOwnedCap sizes the common element's owned-operation list.
const layerOwnedCap = 8

// InvertLayerMatrix returns the inverse of a 2D layer matrix. It is exported
// so a page can derive LayerSpec.Base from the sampled operation matrix.
func InvertLayerMatrix(m layout.Matrix2D) (layout.Matrix2D, bool) {
	return host.InvertLayerMatrix(m)
}

// layerCandidate is one resolved LayerSpec: the operations it owns, its
// captured base box, its content signature, and its paint-order span.
type layerCandidate struct {
	spec                LayerSpec
	key                 layerKey
	owned               []int
	box                 image.Rectangle
	sig                 uint64
	alpha               float64
	mode                string
	matrix              layout.Matrix2D
	base                layout.Matrix2D
	firstRank, lastRank int
}

// resolveLayer applies the narrow promotion predicate and captures the
// element's operations in paint order. reason is non-empty on rejection.
func resolveLayer(d *layout.Display, spec LayerSpec, gen uint64, claimed []bool, ranks []int, owners ...[]int) (layerCandidate, string) {
	var out layerCandidate
	if d == nil {
		return out, "no display"
	}
	if spec.BoxIndex < 0 || spec.BoxIndex >= len(d.Boxes) {
		return out, "missing element box"
	}
	box := &d.Boxes[spec.BoxIndex]
	if box.Transform.HasTransform3D || box.Transform.HasPerspective {
		return out, "3d transform"
	}
	if hasGroups(d) {
		// The direct layer walk paints group members without buffering them.
		// Until layer painting runs inside a group buffer, a page with any
		// blend or isolation group keeps the direct path.
		return out, "display has blend or isolation groups"
	}

	matrix := normalMatrix(spec.Matrix)
	alpha := clampLayerAlpha(spec.Alpha)
	if !normalMode(spec.Mode) && !layerTranslation(matrix) {
		return out, "blend mode with a matrix"
	}
	if !normalMode(spec.Mode) && alpha < 1 {
		// The shared blend shader takes its source alpha from the layer
		// texture; a group alpha would need an extra shader uniform. Keep
		// such an element direct until the blend program takes Alpha.
		return out, "blend mode with group alpha"
	}

	var ownerBuf []int
	if len(owners) > 0 {
		ownerBuf = owners[0]
	}
	owned := layerOwnedOps(d, spec.BoxIndex, ownerBuf)
	if len(owned) == 0 {
		return out, "no owned ops"
	}
	if len(owned) < 2 && !singleRichOp(&d.Ops[owned[0]]) {
		return out, "fewer than two owned ops"
	}

	base := normalMatrix(spec.Base)
	ppt := layerPointsPerPixel(d)
	boxRect := image.Rectangle{}
	sig := fnvOffsetBasis
	firstRank, lastRank := len(d.Order), -1
	if len(ranks) < len(d.Ops) {
		ranks = layerRanks(d)
	}

	for _, index := range owned {
		if index < 0 || index >= len(d.Ops) {
			return out, "owned op outside display"
		}
		if claimed != nil && claimed[index] {
			return out, "overlaps another layer"
		}
		op := &d.Ops[index]
		if op.Fixed || op.ZIndexSet && op.ZIndex != 0 {
			// ZIndexSet at level zero is a stacking-context marker, which a
			// transform or an opacity below 1 creates. A transition target
			// carries that marker in the committed display, so only a non-zero
			// level counts as the z-indexed band the layer cannot cross.
			return out, "fixed or z-indexed op"
		}
		if op.Group() != nil {
			return out, "op inside a blend group"
		}
		if _, _, _, _, clipped := op.OverflowClip(); clipped {
			// A layer bakes the clip into the texture, but the composite
			// pose moves it with the animation. Keep such subtrees direct
			// so the replay scissor holds the ancestor clip still.
			return out, "clipped op"
		}
		b, ok := layerOpBox(op, base, ppt)
		if !ok || b.Empty() {
			return out, "unbounded op"
		}
		boxRect = boxRect.Union(b)
		sig = layerSignOp(sig, op, alpha)
		rank := -1
		if index >= 0 && index < len(ranks) {
			rank = ranks[index]
		}
		if rank < 0 {
			continue
		}
		firstRank = min(firstRank, rank)
		lastRank = max(lastRank, rank)
	}
	if firstRank == len(d.Order) || lastRank < 0 {
		return out, "no painted ops"
	}
	if !layerOrderSafe(d, owned, firstRank, lastRank, boxRect, matrix, ppt) {
		// A layer paints as one unit at its first position. An operation
		// painted between its owned ranks that could cover it would land in
		// the wrong order, so keep such a subtree direct.
		return out, "non-contiguous ops"
	}

	boxRect = boxRect.Intersect(contentRect(d))
	if boxRect.Empty() {
		return out, "empty layer box"
	}
	if boxRect.Dx()*boxRect.Dy() > layerMaxPixels {
		return out, "layer larger than the pixel cap"
	}
	if limit := ebiten.MaxImageSize(); limit > 0 && (boxRect.Dx() > limit || boxRect.Dy() > limit) {
		return out, "layer larger than the max image size"
	}

	out = layerCandidate{
		spec: spec, key: layerKey{id: spec.ElementID, box: spec.BoxIndex, gen: gen},
		owned: owned, box: boxRect, sig: sig, alpha: alpha, mode: spec.Mode,
		matrix: matrix, base: base, firstRank: firstRank, lastRank: lastRank,
	}

	return out, ""
}

// layerOrderSafe reports whether a layer's owned operations may paint as one
// unit at their first rank. An operation painted between the first and last
// owned ranks is out of order only when the layer could cover it: if its box
// stays outside the layer's captured and animated boxes, the two paints the
// same pixels either way.
func layerOrderSafe(d *layout.Display, owned []int, firstRank, lastRank int, box image.Rectangle,
	matrix layout.Matrix2D, ppt float64) bool {
	if lastRank-firstRank+1 == len(owned) {
		return true
	}
	if d == nil || box.Empty() {
		return false
	}

	isOwned := make(map[int]struct{}, len(owned))
	for _, index := range owned {
		isOwned[index] = struct{}{}
	}

	animated := layerMatrixBox(box, matrix, ppt)
	for rank := firstRank; rank <= lastRank && rank < len(d.Order); rank++ {
		index := d.Order[rank]
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		if _, ok := isOwned[index]; ok {
			continue
		}
		op := &d.Ops[index]
		if _, hasProj := op.ProjectiveTransformValue(); op.Group() != nil || hasProj {
			return false
		}
		b, ok := layerOpBox(op, layout.IdentityMatrix(), ppt)
		if !ok || b.Empty() || b.Overlaps(box) || b.Overlaps(animated) {
			return false
		}
	}

	return true
}

// pose returns the candidate's composited page box, matching layerPose.
func (cand *layerCandidate) pose(d *layout.Display) image.Rectangle {
	return layerMatrixBox(cand.box, cand.matrix, layerPixelPerPoint(d))
}

// layerMatrixBox maps a page box through a layer matrix and rounds the result
// outward, matching layerPose.
func layerMatrixBox(b image.Rectangle, m layout.Matrix2D, ppp float64) image.Rectangle {
	box := image.Rectangle{}

	for _, p := range [4][2]float64{
		{float64(b.Min.X), float64(b.Min.Y)},
		{float64(b.Max.X), float64(b.Min.Y)},
		{float64(b.Max.X), float64(b.Max.Y)},
		{float64(b.Min.X), float64(b.Max.Y)},
	} {
		x := m.A*p[0] + m.C*p[1] + m.E*ppp
		y := m.B*p[0] + m.D*p[1] + m.F*ppp
		box = box.Union(image.Rect(int(math.Floor(x))-1, int(math.Floor(y))-1,
			int(math.Ceil(x))+1, int(math.Ceil(y))+1))
	}

	return box
}

// layerRanks maps each operation index to its position in Display.Order, or
// -1 when the op never paints.
func layerRanks(d *layout.Display, buf ...[]int) []int {
	var ranks []int
	if len(buf) > 0 {
		ranks = buf[0]
	}
	if cap(ranks) < len(d.Ops) {
		ranks = make([]int, len(d.Ops))
	} else {
		ranks = ranks[:len(d.Ops)]
	}
	for i := range ranks {
		ranks[i] = -1
	}
	for rank, index := range d.Order {
		if index >= 0 && index < len(ranks) {
			ranks[index] = rank
		}
	}

	return ranks
}

// layerOwners returns the nearest owner index for each operation in d.Ops.
func layerOwners(d *layout.Display, buf ...[]int) []int {
	var owners []int
	if len(buf) > 0 {
		owners = buf[0]
	}
	if cap(owners) < len(d.Ops) {
		owners = make([]int, len(d.Ops))
	} else {
		owners = owners[:len(d.Ops)]
	}
	if len(d.OpOwners) == len(d.Ops) {
		copy(owners, d.OpOwners)
	} else {
		for i := range d.Ops {
			owners[i] = layerOwnerFromIDs(d, d.Ops[i].ID)
		}
	}

	return owners
}

// layerOwnedOps returns the element's owned operation indices in paint order.
// It mirrors the viewtransition capture helper: the nearest owner from
// OpOwners, with the OperationIDs fallback for a display that predates it.
func layerOwnedOps(d *layout.Display, root int, optionalOwners ...[]int) []int {
	var owners []int
	if len(optionalOwners) > 0 && len(optionalOwners[0]) >= len(d.Ops) {
		owners = optionalOwners[0]
	} else {
		owners = layerOwners(d)
	}

	owned := make([]int, 0, layerOwnedCap)
	for _, index := range d.Order {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		if layerBoxInSubtree(d.Boxes, owners[index], root) {
			owned = append(owned, index)
		}
	}

	return owned
}

// layerBoxInSubtree reports whether box is root or one of root's descendants.
func layerBoxInSubtree(boxes []layout.Box, box, root int) bool {
	for steps := 0; box >= 0 && box < len(boxes) && steps <= len(boxes); steps++ {
		if box == root {
			return true
		}
		box = boxes[box].ParentIndex
	}

	return false
}

// layerOwnerFromIDs finds the deepest box whose OperationIDs carries id.
func layerOwnerFromIDs(d *layout.Display, id uint64) int {
	best, bestDepth := -1, -1
	for i := range d.Boxes {
		if !layerHasID(d.Boxes[i].OperationIDs, id) {
			continue
		}
		depth := layerBoxDepth(d.Boxes, i)
		if depth > bestDepth {
			best, bestDepth = i, depth
		}
	}

	return best
}

func layerHasID(ids []uint64, id uint64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}

	return false
}

func layerBoxDepth(boxes []layout.Box, index int) int {
	depth := 0
	for p, steps := boxes[index].ParentIndex, 0; p >= 0 && p < len(boxes) && steps < len(boxes); steps++ {
		depth++
		p = boxes[p].ParentIndex
	}

	return depth
}

// layerOpBox bounds one op in its captured base pose. base maps the op's
// current baked transform back to the pose the layer texture holds.
//
// The bound temporarily borrows the op's transform for the base pose: a
// DisplayOp keeps its rare payload behind a shared pointer, so writing the
// matrix on a struct copy would reach the live display. The previous value is
// restored before this function returns.
func layerOpBox(op *layout.DisplayOp, base layout.Matrix2D, ppt float64) (image.Rectangle, bool) {
	if base.IsIdentity() {
		clone := *op
		if clone.Kind == layout.DisplayOpImage && clone.XformSet {
			// xformPointsBox predates transformed images: an image places
			// itself through its own matrix in drawImage. Bound the layout
			// rectangle through that matrix here, or every rotated card with an
			// image would demote.
			return layerMappedBox(&clone, ppt)
		}

		return opBounds(&clone, ppt)
	}

	saved, savedSet := op.Transform(), op.XformSet
	current := layout.IdentityMatrix()
	if savedSet {
		current = saved
	}
	op.SetXform(base.Mul(current))
	defer func() {
		op.SetXform(saved)
		op.XformSet = savedSet
	}()

	clone := *op
	if clone.Kind == layout.DisplayOpImage && clone.XformSet {
		return layerMappedBox(&clone, ppt)
	}

	return opBounds(&clone, ppt)
}

// layerMappedBox maps an op's layout rectangle through its transform, which is
// what an image paints.
func layerMappedBox(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	x0, y0, x1, y1 := op.X, op.Y, op.X+op.W, op.Y+op.H
	m := op.Transform()
	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		x, y := m.Apply(p[0], p[1])
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	pad := float64(boundsPad)

	return image.Rect(
		int(math.Floor(loX/ppt-pad)), int(math.Floor(loY/ppt-pad)),
		int(math.Ceil(hiX/ppt+pad)), int(math.Ceil(hiY/ppt+pad)),
	), true
}

// layerSignOp folds one operation into a layer signature. The op's matrix is
// excluded: it changes every frame on the animation path, and the layer
// composites it instead of baking it.
//
// The sampled group opacity is excluded too, because the bake divides it out
// and the composite applies it. The field the bake will actually paint is
// folded instead, so a descendant that keeps an opacity of its own still
// re-bakes when that opacity changes. The op is borrowed for the hash and
// restored afterwards: DisplayOp shares its rare payload, so writing the
// element opacity on a struct copy would reach the live display.
func layerSignOp(h uint64, op *layout.DisplayOp, groupAlpha float64) uint64 {
	saved := layerElementOpacity(op)
	defer op.SetPaintOpacity(saved)
	op.SetPaintOpacity(layerBakeOpacity(op, groupAlpha))

	clone := *op
	clone.XformSet = false

	return sigOp(h, &clone)
}

// singleRichOp reports the one operation kinds a single-op layer may own.
func singleRichOp(op *layout.DisplayOp) bool {
	return op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpImage
}

// normalMatrix treats a zero matrix as the identity. The zero value is the
// convenient default for a spec that carries no animation.
func normalMatrix(m layout.Matrix2D) layout.Matrix2D {
	if m.A == 0 && m.B == 0 && m.C == 0 && m.D == 0 && m.E == 0 && m.F == 0 {
		return layout.IdentityMatrix()
	}

	return m
}

// layerTranslation reports a matrix with no linear part, the only shape the
// blend shader path can composite.
func layerTranslation(m layout.Matrix2D) bool {
	return m.A == 1 && m.B == 0 && m.C == 0 && m.D == 1
}

// layerAlphaEpsilon is the tolerance within which an operation's element
// opacity counts as the layer's sampled group opacity rather than an opacity of
// its own. The runtime writes one sampled value to every animated target, so a
// matched value is the group alpha and is divided out; a different value is a
// descendant's own opacity and must survive the bake.
const layerAlphaEpsilon = 1e-6

// layerElementOpacity returns the operation's element opacity without folding
// in its source alpha, which is the value the engine's paint-opacity field
// holds. A value the engine treats as unset reads as 1. It mirrors the motion
// driver's own read, so the bake and the sampler agree on what the group alpha
// is.
func layerElementOpacity(op *layout.DisplayOp) float64 {
	if op == nil {
		return 1
	}

	opacity := op.Opacity()
	if op.Alpha > 0 && op.Alpha < 1 && opacity > 0 {
		opacity /= op.Alpha
	}
	if math.IsNaN(opacity) || opacity > 1 {
		return 1
	}

	return opacity
}

// layerBakeOpacity returns the element opacity the layer bake writes for one
// owned operation. The layer texture must hold the operation's own opacity,
// because the composite multiplies the whole texture by the sampled group alpha
// once.
//
// The runtime writes the sampled group alpha into every animated operation, so
// an operation whose element opacity matches it has that value removed and
// bakes at 1. Any other operation keeps its own opacity, which leaves a
// descendant's independent opacity in the texture and lets the group alpha
// flatten it at composite time. Alpha 1 is already the unchanged bake. Alpha 0
// composites nothing, so the opacities are left alone; the same texture is
// correct again when the animation leaves zero, because the runtime clears the
// paint opacity at zero and the bake already holds the group-free value.
func layerBakeOpacity(op *layout.DisplayOp, groupAlpha float64) float64 {
	own := layerElementOpacity(op)
	if groupAlpha <= 0 || groupAlpha >= 1 {
		return own
	}
	if math.Abs(own-groupAlpha) <= layerAlphaEpsilon {
		return 1
	}

	return own
}

// clampLayerAlpha keeps an alpha inside 0..1. NaN reads as opaque.
func clampLayerAlpha(alpha float64) float64 {
	if math.IsNaN(alpha) {
		return 1
	}
	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}

	return alpha
}

// layerPointsPerPixel returns the point-to-pixel factor, or 1 for a hand-built
// display that left it zero.
func layerPointsPerPixel(d *layout.Display) float64 {
	if d != nil && d.PointsPerPixel > 0 {
		return d.PointsPerPixel
	}

	return 1
}

// layerPixelPerPoint is the inverse factor used to place a layer matrix in page
// pixels.
func layerPixelPerPoint(d *layout.Display) float64 {
	if d != nil && d.PixelPerPoint > 0 {
		return d.PixelPerPoint
	}

	return 1
}

// layerPose returns the layer's composited page box after its animation
// matrix, in CSS pixels, rounded outward.
func layerPose(layer *Layer, d *layout.Display) image.Rectangle {
	if layer == nil || layer.Tex == nil {
		return image.Rectangle{}
	}

	return layerMatrixBox(layer.Box, layer.Matrix, layerPixelPerPoint(d))
}

// layerTouches reports whether the layer's last composited box meets rect.
func layerTouches(layer *Layer, rect image.Rectangle) bool {
	return layer != nil && !layer.pose.Intersect(rect).Empty()
}
