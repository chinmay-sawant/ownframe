package cssmotion

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
		Opacity: f(a.Opacity, b.Opacity), Matrix: b.Matrix, HasMatrix: b.HasMatrix,
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
	atx, aty, asx, asy, arotate, askew := decomposeMatrix2D(a)
	btx, bty, bsx, bsy, brotate, bskew := decomposeMatrix2D(b)
	if math.Abs(asx) < matrixScaleFloor || math.Abs(asy) < matrixScaleFloor ||
		math.Abs(bsx) < matrixScaleFloor || math.Abs(bsy) < matrixScaleFloor {
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
		lerp(atx, btx, t), lerp(aty, bty, t),
		lerp(asx, bsx, t), lerp(asy, bsy, t),
		lerp(arotate, brotate, t), lerp(askew, bskew, t),
	)
}

// decomposeMatrix2D is the WebKit two-dimensional unmatrix: translation,
// scale, rotation, and one XY shear. ScaleX is signed positive and ScaleY
// carries the determinant sign.
func decomposeMatrix2D(m layout.Matrix2D) (tx, ty, sx, sy, rotation, skew float64) {
	rowX, rowY := m.A, m.B
	sx = math.Hypot(rowX, rowY)
	if sx != 0 {
		rowX, rowY = rowX/sx, rowY/sx
	} else {
		rowX, rowY = 1, 0
	}
	sy = -rowY*m.C + rowX*m.D
	if sy != 0 {
		skew = (rowX*m.C + rowY*m.D) / sy
	}
	return m.E, m.F, sx, sy, math.Atan2(rowY, rowX), math.Atan(skew)
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
