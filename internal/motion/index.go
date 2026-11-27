package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"slices"
)

// emptyIndex is the shared immutable index returned for nil or empty displays.
// It performs 0 allocations and its lookups are all empty.
var emptyIndex = &Index{} //nolint:gochecknoglobals // reason: shared immutable zero index returned by EmptyIndex, allocated once per process.

// EmptyIndex returns the shared immutable empty index with 0 allocations.
func EmptyIndex() *Index {
	return emptyIndex
}

// Index resolves operation ownership for one display list without scanning
// it per box. A bind builds one index for the display it resolves and shares
// it across every animated box and every transition capture, instead of
// scanning the whole operation list once per box per paint class.
//
// The index is read-only after construction. Sampling mutates operation
// paint but never adds, removes, renumbers, or reassigns operations, so one
// index serves every selection the bind makes against its display.
type Index struct {
	display *layout.Display
	// owners lists operation positions per owning box index, in operation
	// order. It is only consulted when ownersValid holds, the same condition
	// Select requires.
	owners [][]int
	// byID maps an operation id to its positions, in operation order.
	byID map[uint64][]int
	// ownersValid reports that every operation has a nearest owner, so an
	// owner-scoped selection can trust the owner map.
	ownersValid bool
}

// NewIndex builds the shared lookup for d. A nil display, or a display with no
// operations, returns a shared immutable empty index with 0 allocs.
func NewIndex(d *layout.Display) *Index {
	if d == nil || len(d.Ops) == 0 {
		return emptyIndex
	}
	x := &Index{display: d}
	x.byID = make(map[uint64][]int, len(d.Ops))
	for i := range d.Ops {
		id := d.Ops[i].ID
		x.byID[id] = append(x.byID[id], i)
	}
	if len(d.OpOwners) != len(d.Ops) {
		return x
	}
	x.ownersValid = true
	x.owners = make([][]int, len(d.Boxes))
	for i := range d.Ops {
		owner := d.OpOwners[i]
		if owner < 0 || owner >= len(d.Boxes) {
			continue
		}
		x.owners[owner] = append(x.owners[owner], i)
	}
	return x
}

// NewKeyframesIndex returns the shared index for a keyframe rule set and display.
// If len(keyframes) == 0, it returns the shared emptyIndex immediately.
func NewKeyframesIndex(keyframes []layout.KeyframesRule, d *layout.Display) *Index {
	if len(keyframes) == 0 {
		return emptyIndex
	}
	return NewIndex(d)
}

// Display returns the display this index was built for, or nil for an empty index.
func (x *Index) Display() *layout.Display {
	if x == nil {
		return nil
	}
	return x.display
}

// NumOps returns the number of operations in the indexed display.
func (x *Index) NumOps() int {
	if x == nil || x.display == nil {
		return 0
	}
	return len(x.display.Ops)
}

// Repoint updates the underlying display pointer when a new display with
// identical operation topology is committed.
func (x *Index) Repoint(d *layout.Display) {
	if x != nil && x != emptyIndex {
		x.display = d
	}
}

// Select returns the operation indices a property of one kind may write,
// exactly as Select does: owner-scoped kinds read the nearest-owner map and
// fall back to the filtered subtree without a trusted index, while a
// transform or an opacity keeps the whole subtree in operation order.
func (x *Index) Select(box layout.Box, boxIndex int, kind Kind) []int {
	if x == nil || x.display == nil {
		return nil
	}
	if kind == KindSubtree {
		return x.subtree(box)
	}
	if boxIndex >= 0 && x.ownersValid {
		var out []int
		if boxIndex < len(x.owners) {
			for _, index := range x.owners[boxIndex] {
				if matches(kind, &x.display.Ops[index]) {
					out = append(out, index)
				}
			}
		}
		return out
	}
	var out []int
	for _, index := range x.subtree(box) {
		if matches(kind, &x.display.Ops[index]) {
			out = append(out, index)
		}
	}
	return out
}

// subtree returns the operation positions the box's subtree carries, in
// operation order, matching OperationIndices including its geometric fallback
// for boxes that predate operation ids.
func (x *Index) subtree(box layout.Box) []int {
	if x == nil || x.display == nil {
		return nil
	}
	d := x.display
	if box.OperationIDs == nil {
		return geometricIndices(d, box)
	}
	var out []int
	for _, id := range box.OperationIDs {
		out = append(out, x.byID[id]...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// ImageIndex returns the raster an inline SVG paints: the last image
// operation the box owns, with the same owner, id-set, and geometric tiers
// the SVG binder uses. An empty slice means the element painted no SVG.
func (x *Index) ImageIndex(box layout.Box, boxIndex int) []int {
	if x == nil || x.display == nil {
		return nil
	}
	d := x.display
	if boxIndex >= 0 && x.ownersValid && boxIndex < len(x.owners) {
		last := -1
		for _, index := range x.owners[boxIndex] {
			if d.Ops[index].Kind == layout.DisplayOpImage {
				last = index
			}
		}
		if last >= 0 {
			return []int{last}
		}
	}
	if len(box.OperationIDs) > 0 {
		last := -1
		for _, id := range box.OperationIDs {
			for _, index := range x.byID[id] {
				if d.Ops[index].Kind == layout.DisplayOpImage && index > last {
					last = index
				}
			}
		}
		if last >= 0 {
			return []int{last}
		}
		return nil
	}
	last := -1
	for i := range d.Ops {
		if d.Ops[i].Kind != layout.DisplayOpImage {
			continue
		}
		if !Contains(d, box, d.Ops[i]) {
			continue
		}
		last = i
	}
	if last >= 0 {
		return []int{last}
	}
	return nil
}

// Kind names the operations one paint-property family writes. A colour belongs
// to the element that paints it, so selection is by nearest owner; a transform
// or an opacity moves or fades the element's whole subtree.
type Kind uint8

// The operation sets a property family may write.
const (
	// KindFill is a background, accent, or rule fill: the owner's FillRect
	// operations.
	KindFill Kind = iota + 1
	// KindInk is text ink: the owner's Text operations.
	KindInk
	// KindStroke is a border or an outline: the owner's StrokeRect and Line
	// operations.
	KindStroke
	// KindRect is a box's own rounded-rectangle fill and stroke: the owner's
	// FillRect and StrokeRect operations (used for uniform border-radius).
	KindRect
	// KindSubtree is a transform or an opacity: every operation the box's
	// subtree carries, because a moved parent moves its children.
	KindSubtree
)

// Select returns the operation indices a property of one kind may write.
// boxIndex is the box's position in d.Boxes, or -1 when the caller cannot
// trust it; an owner-scoped kind then falls back to the box's subtree.
//
// A colour only writes the operations whose nearest owner is this box
// (d.OpOwners), so a parent's background colour cannot recolour a child's card
// and cannot touch its own label text. Transform and opacity keep the subtree
// the box's OperationIDs describe.
func Select(d *layout.Display, box layout.Box, boxIndex int, kind Kind) []int {
	if d == nil {
		return nil
	}
	if kind == KindSubtree {
		return OperationIndices(d, box)
	}
	if boxIndex >= 0 && len(d.OpOwners) == len(d.Ops) {
		var out []int
		for i := range d.Ops {
			if d.OpOwners[i] == boxIndex && matches(kind, &d.Ops[i]) {
				out = append(out, i)
			}
		}
		return out
	}
	// No trusted owner index: keep the property on the box's subtree, filtered
	// to the kinds it paints. The caller counts the fallback.
	var out []int
	for _, index := range OperationIndices(d, box) {
		if matches(kind, &d.Ops[index]) {
			out = append(out, index)
		}
	}
	return out
}

// matches reports whether one operation belongs to a property kind. A kind
// outside these sets is not a paint field a keyframe can write: an image and a
// grid run carry no colour, and a background never paints text.
//
// A corner-shape path carries both a fill and a border under one op kind, so
// the path's Stroke flag decides: a background recolours the fill path and a
// border or outline recolours the stroke path. Without this, a corner-shaped
// element would keep its declared colours while its rectangular neighbours
// moved.
func matches(kind Kind, op *layout.DisplayOp) bool {
	if op.Kind == layout.DisplayOpPath {
		corner, ok := op.CornerPathValue()
		if !ok || corner == nil {
			return false
		}

		switch kind {
		case KindFill:
			return !corner.Stroke
		case KindStroke:
			return corner.Stroke
		}

		return false
	}

	switch kind {
	case KindFill:
		return op.Kind == layout.DisplayOpFillRect
	case KindInk:
		return op.Kind == layout.DisplayOpText
	case KindStroke:
		return op.Kind == layout.DisplayOpStrokeRect || op.Kind == layout.DisplayOpLine
	case KindRect:
		return op.Kind == layout.DisplayOpFillRect || op.Kind == layout.DisplayOpStrokeRect
	}
	return false
}

// OperationIndices captures operations owned by a box. Older Blinkless boxes
// without ownership IDs use the geometric fallback in Contains.
func OperationIndices(d *layout.Display, box layout.Box) []int {
	if d == nil {
		return nil
	}
	if box.OperationIDs == nil {
		return geometricIndices(d, box)
	}
	ids := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		ids[id] = struct{}{}
	}
	var out []int
	for i := range d.Ops {
		if _, ok := ids[d.Ops[i].ID]; ok {
			out = append(out, i)
		}
	}
	return out
}

func geometricIndices(d *layout.Display, box layout.Box) []int {
	var out []int
	for i := range d.Ops {
		if Contains(d, box, d.Ops[i]) {
			out = append(out, i)
		}
	}
	return out
}

// Capture returns paint values for operations inside a layout box.
func Capture(d *layout.Display, box layout.Box) []Paint {
	var out []Paint
	for _, op := range Operations(d, box) {
		out = append(out, Read(op))
	}
	return out
}

// Operations returns pointers to display operations whose centers lie in box.
func Operations(d *layout.Display, box layout.Box) []*layout.DisplayOp {
	return Resolve(d, OperationIndices(d, box))
}

// Resolve finds the current display operations at previously captured indices.
func Resolve(d *layout.Display, indices []int) []*layout.DisplayOp {
	var out []*layout.DisplayOp
	if d == nil {
		return out
	}
	for _, index := range indices {
		if index >= 0 && index < len(d.Ops) {
			out = append(out, &d.Ops[index])
		}
	}
	return out
}

// Contains reports whether an operation's center lies inside a layout box.
func Contains(d *layout.Display, box layout.Box, op layout.DisplayOp) bool {
	x, y := op.X+op.W/halfDivisor, op.Y+op.H/halfDivisor
	if op.Kind == layout.DisplayOpText {
		y = op.Y
	}
	p := d.PointsPerPixel
	return x >= box.X*p && x <= (box.X+box.W)*p &&
		y >= box.Y*p && y <= (box.Y+box.H)*p
}

// Mask names the paint fields a keyframed property writes. A field outside the
// mask keeps the operation's own value, so a transform-only animation never
// rewrites the colour, and a colour-only animation never rewrites the matrix.
//
// W, H, and Alpha are maskable for a transition that captured a whole
// operation, but no keyframed paint property owns them in this phase: a size
// change is layout state, and a colour's own alpha is already folded into the
// operation's Alpha by the engine. The keyframe binder therefore leaves those
// bits clear, and a transform-only animation cannot turn a translucent
// operation opaque.
type Mask uint16

// One bit per field a paint write can change.
const (
	MaskX Mask = 1 << iota
	MaskY
	MaskW
	MaskH
	MaskMatrix
	MaskOpacity
	MaskRGB
	MaskAlpha
	MaskRadius
)

// MaskAll is every field a paint snapshot can carry. A zero Paint mask means
// the same thing: a full snapshot. Read returns one, so a transition that
// captured an operation writes every field it captured. A keyframed target
// narrows the mask to the fields its keyframes declare, and a paint-only target
// leaves W and H alone.
const MaskAll = MaskX | MaskY | MaskW | MaskH | MaskMatrix | MaskOpacity | MaskRGB | MaskAlpha | MaskRadius

// Paint is the set of display-list values sampled by CSS motion.
type Paint struct {
	X, Y, W, H      float64
	R, G, B, Alpha  float64
	Opacity         float64
	Radius, RadiusY float64
	// Matrix is the operation's baked 2D transform. HasMatrix reports whether
	// the operation carried one when the paint was captured, and whether a
	// write installs Matrix. A paint without a matrix leaves the operation's
	// transform untouched, so a scalar-only animation never disturbs a static
	// transform it does not own.
	Matrix    layout.Matrix2D
	HasMatrix bool
	// Mask names the fields a write may change. Zero is a full snapshot, the
	// meaning Read gives its result; a keyframed target narrows it so two
	// animations with disjoint properties compose instead of overwriting each
	// other.
	Mask Mask
	// NeedsRestamp is set by a sampler that could not express its value as a
	// 2D matrix. It is a per-frame signal, not a paint value: the operation is
	// left alone and the next phase runs the engine's 3D restamp pass.
	NeedsRestamp bool
}

// Writes reports whether the paint's mask allows m. A zero mask is a full
// snapshot, so every field is writable.
func (p Paint) Writes(m Mask) bool {
	return p.Mask == 0 || p.Mask&m != 0
}

// Read snapshots the paint values from an operation.
func Read(op *layout.DisplayOp) Paint {
	p := Paint{X: op.X, Y: op.Y, W: op.W, H: op.H,
		R: op.R, G: op.G, B: op.B, Alpha: op.Alpha, Opacity: paintOpacity(op),
		Radius: op.Radius, RadiusY: op.RadiusY, Mask: MaskAll}
	if op.XformSet {
		p.Matrix = op.Transform()
		p.HasMatrix = true
	} else {
		p.Matrix = layout.IdentityMatrix()
	}
	return p
}

// Write applies the masked fields of p to the operation. A keyframed target
// never asks for W or H, so a transform-only animation leaves a translucent
// operation's alpha, radius, and size alone.
func Write(op *layout.DisplayOp, p Paint) {
	if p.Writes(MaskX) {
		op.X = p.X
	}
	if p.Writes(MaskY) {
		op.Y = p.Y
	}
	if p.Writes(MaskW) {
		op.W = p.W
	}
	if p.Writes(MaskH) {
		op.H = p.H
	}
	if p.Writes(MaskRGB) {
		op.R, op.G, op.B = p.R, p.G, p.B
	}
	if p.Writes(MaskAlpha) {
		op.Alpha = p.Alpha
	}
	if p.Writes(MaskOpacity) {
		op.SetPaintOpacity(p.Opacity)
	}
	if p.Writes(MaskRadius) {
		op.Radius = p.Radius
		op.RadiusY = p.RadiusY
	}
	if p.Writes(MaskMatrix) && p.HasMatrix {
		op.SetXform(p.Matrix)
	}
}

// Blend linearly interpolates every sampled paint value. A pair where both
// endpoints carry a matrix interpolates it through the two-dimensional
// decomposition, so a rotate or scale transition follows the same path a
// keyframe does. When only one side carries a matrix the destination matrix is
// kept, which is the keyframe path that writes a full paint from one endpoint.
// The mask is the union of the endpoints, so a paint either endpoint may write
// stays writable after the blend.
func Blend(a, b Paint, t float64) Paint {
	f := func(x, y float64) float64 { return x + (y-x)*t }
	out := Paint{X: f(a.X, b.X), Y: f(a.Y, b.Y), W: f(a.W, b.W), H: f(a.H, b.H),
		R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), Alpha: f(a.Alpha, b.Alpha),
		Opacity: f(a.Opacity, b.Opacity), Radius: f(a.Radius, b.Radius),
		RadiusY: f(a.RadiusY, b.RadiusY), Matrix: b.Matrix, HasMatrix: b.HasMatrix,
		Mask: a.Mask | b.Mask}
	if a.HasMatrix && b.HasMatrix {
		out.Matrix = blendMatrix2D(a.Matrix, b.Matrix, t)
	}
	return out
}

// matrixScaleFloor is the scale below which a two-dimensional matrix is
// treated as singular and blended component by component. Decomposing a
// singular matrix cannot recover a rotation or a shear.
const matrixScaleFloor = 1e-12

// blendMatrix2D interpolates two affine matrices through translation, scale,
// rotation, and one XY shear. The reconstruction is T * R * SkewX * S, the
// same order the animation compositor uses.
func blendMatrix2D(a, b layout.Matrix2D, t float64) layout.Matrix2D {
	ta := decomposeMatrix2D(a)
	tb := decomposeMatrix2D(b)
	if math.Abs(ta.Sx) < matrixScaleFloor || math.Abs(ta.Sy) < matrixScaleFloor ||
		math.Abs(tb.Sx) < matrixScaleFloor || math.Abs(tb.Sy) < matrixScaleFloor {
		return layout.Matrix2D{
			A: lerp(a.A, b.A, t),
			B: lerp(a.B, b.B, t),
			C: lerp(a.C, b.C, t),
			D: lerp(a.D, b.D, t),
			E: lerp(a.E, b.E, t),
			F: lerp(a.F, b.F, t),
		}
	}

	return composeMatrix2D(
		lerp(ta.Tx, tb.Tx, t), lerp(ta.Ty, tb.Ty, t),
		lerp(ta.Sx, tb.Sx, t), lerp(ta.Sy, tb.Sy, t),
		lerp(ta.Rotate, tb.Rotate, t), lerp(ta.SkewX, tb.SkewX, t),
	)
}

// composeMatrix2D rebuilds the matrix a decomposition describes.
func composeMatrix2D(tx, ty, sx, sy, rotation, skew float64) layout.Matrix2D {
	sin, cos := math.Sincos(rotation)
	m := layout.Matrix2D{A: 1, D: 1, E: tx, F: ty}
	m = m.Mul(layout.Matrix2D{A: cos, B: sin, C: -sin, D: cos})
	m = m.Mul(layout.Matrix2D{A: 1, D: 1, C: math.Tan(skew)})
	return m.Mul(layout.Matrix2D{A: sx, D: sy})
}

// lerp interpolates one scalar.
func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// paintOpacity returns the operation's element opacity without folding in its
// ALPHA. The public Opacity accessor multiplies the two for the painters, but a
// write stores the element opacity alone, so reading the product back would
// multiply a translucent operation twice on the next write.
func paintOpacity(op *layout.DisplayOp) float64 {
	opacity := op.Opacity()
	if op.Alpha > 0 && op.Alpha < 1 && opacity > 0 {
		opacity /= op.Alpha
	}
	if opacity > 1 {
		return 1
	}
	return opacity
}
