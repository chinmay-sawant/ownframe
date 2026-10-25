package cssmotion

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestReadCarriesTheBakedMatrix(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}
	want := layout.Matrix2D{A: 0, B: 1, C: -1, D: 0, E: 10, F: 20}
	op.SetXform(want)

	paint := Read(op)
	if !paint.HasMatrix {
		t.Fatal("a transformed operation was read without a matrix")
	}
	if paint.Matrix != want {
		t.Fatalf("matrix = %+v, want %+v", paint.Matrix, want)
	}
}

func TestReadWithoutTransformReportsNoMatrix(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}

	paint := Read(op)
	if paint.HasMatrix {
		t.Fatal("an untransformed operation was read with a matrix")
	}
	if paint.Matrix != layout.IdentityMatrix() {
		t.Fatalf("matrix = %+v, want the identity", paint.Matrix)
	}
}

func TestWriteInstallsAMatrixOnlyWhenSet(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}
	static := layout.Matrix2D{A: 1, B: 0, C: 0, D: 1, E: 3, F: 2}
	op.SetXform(static)

	paint := Paint{W: 20, H: 10, Alpha: 1, Opacity: 1}
	Write(op, paint)
	if got := op.Transform(); got != static {
		t.Fatalf("a paint without a matrix replaced the transform with %+v", got)
	}

	want := layout.Matrix2D{A: 2, B: 0, C: 0, D: 2, E: 8, F: 6}
	paint.Matrix, paint.HasMatrix = want, true
	Write(op, paint)
	if got := op.Transform(); got != want {
		t.Fatalf("matrix = %+v, want %+v", got, want)
	}
	if op.W != 20 || op.H != 10 {
		t.Fatalf("paint = %vx%v, want 20x10", op.W, op.H)
	}
}

// TestWriteLeavesAlphaAndSizeAlone pins the phase 03 write rule: a keyframed
// paint owns no size or alpha field, so a masked write never changes them. A
// translucent operation must keep its alpha while its transform animates.
func TestWriteLeavesAlphaAndSizeAlone(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, R: .2, G: .3, B: .4, Alpha: .4}

	paint := Paint{X: 10, Y: 11, W: 99, H: 99, Alpha: 1, Opacity: .5,
		Matrix: layout.Matrix2D{A: 1, D: 1, E: 3}, HasMatrix: true, Mask: MaskX | MaskY | MaskMatrix}
	Write(op, paint)
	if op.Alpha != .4 {
		t.Fatalf("write changed alpha to %v, want the translucent .4", op.Alpha)
	}
	if op.W != 6 || op.H != 7 {
		t.Fatalf("write changed the size to %vx%v, want 6x7", op.W, op.H)
	}
	if op.X != 10 || op.Y != 11 {
		t.Fatalf("masked position write = %v,%v, want 10,11", op.X, op.Y)
	}
	if op.R != .2 || op.G != .3 || op.B != .4 {
		t.Fatalf("a geometry write changed the colour to %v,%v,%v", op.R, op.G, op.B)
	}
	if got := Read(op).Opacity; got != 1 {
		t.Fatalf("a geometry write changed the element opacity to %v, want the untouched 1", got)
	}
}

// TestReadKeepsTheCompositeOpacityOutOfThePaint checks that a translucent
// operation round-trips: Read reports the element opacity alone, so writing it
// back does not multiply the alpha a second time.
func TestReadKeepsTheCompositeOpacityOutOfThePaint(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 1, Y: 2, W: 3, H: 4, Alpha: .5}
	op.SetPaintOpacity(.5)

	paint := Read(op)
	if paint.Alpha != .5 {
		t.Fatalf("read alpha = %v, want the operation's .5", paint.Alpha)
	}
	if paint.Opacity != .5 {
		t.Fatalf("read opacity = %v, want the element's .5, not the composited .25", paint.Opacity)
	}
	Write(op, paint)
	if got := op.Opacity(); got != .25 {
		t.Fatalf("round trip composite opacity = %v, want .25", got)
	}
}

func TestBlendKeepsTheDestinationMatrix(t *testing.T) {
	from := Paint{X: 0, Alpha: 1, Opacity: 1}
	to := Paint{X: 10, Alpha: 1, Opacity: 1,
		Matrix: layout.Matrix2D{A: 1, D: 1, E: 20, F: 30}, HasMatrix: true}

	got := Blend(from, to, .5)
	if !got.HasMatrix || got.Matrix != to.Matrix {
		t.Fatalf("blended matrix = %+v (set %v), want the destination %+v", got.Matrix, got.HasMatrix, to.Matrix)
	}
	if got.X != 5 {
		t.Fatalf("blended X = %v, want 5", got.X)
	}
}
