package cssmotion

import "github.com/chinmay-sawant/blinkless/layout"

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
)

// MaskAll is every field a paint snapshot can carry. A zero Paint mask means
// the same thing: a full snapshot. Read returns one, so a transition that
// captured an operation writes every field it captured. A keyframed target
// narrows the mask to the fields its keyframes declare, and a paint-only target
// leaves W, H, and Alpha alone.
const MaskAll = MaskX | MaskY | MaskW | MaskH | MaskMatrix | MaskOpacity | MaskRGB | MaskAlpha

// Paint is the set of display-list values sampled by CSS motion.
type Paint struct {
	X, Y, W, H     float64
	R, G, B, Alpha float64
	Opacity        float64
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
		R: op.R, G: op.G, B: op.B, Alpha: op.Alpha, Opacity: paintOpacity(op), Mask: MaskAll}
	if op.XformSet {
		p.Matrix = op.Transform()
		p.HasMatrix = true
	} else {
		p.Matrix = layout.IdentityMatrix()
	}
	return p
}

// Write applies the masked fields of p to the operation. A keyframed target
// never asks for W, H, or Alpha, so a transform-only animation leaves a
// translucent operation's alpha and size alone.
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
	if p.Writes(MaskMatrix) && p.HasMatrix {
		op.SetXform(p.Matrix)
	}
}

// Blend linearly interpolates every sampled paint value. The matrix is not
// interpolated yet: the destination matrix is kept until transitions grow
// their own matrix decomposition path. The mask is the union of the endpoints,
// so a paint either endpoint may write stays writable after the blend.
func Blend(a, b Paint, t float64) Paint {
	f := func(x, y float64) float64 { return x + (y-x)*t }
	return Paint{X: f(a.X, b.X), Y: f(a.Y, b.Y), W: f(a.W, b.W), H: f(a.H, b.H),
		R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), Alpha: f(a.Alpha, b.Alpha),
		Opacity: f(a.Opacity, b.Opacity), Matrix: b.Matrix, HasMatrix: b.HasMatrix,
		Mask: a.Mask | b.Mask}
}

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
