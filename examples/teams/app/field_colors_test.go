package app

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// fieldOps returns the operations inside the compose field box.
func fieldOps(app *App) []layout.DisplayOp {
	box, ok := findBox(app.Boxes(), "chat-compose")
	d := app.Page().Display()
	if !ok || d == nil {
		return nil
	}

	left := box.X * d.PointsPerPixel
	right := (box.X + box.W) * d.PointsPerPixel
	out := []layout.DisplayOp{}

	for i := range d.Ops {
		op := d.Ops[i]
		if op.X >= left && op.X+op.W <= right {
			out = append(out, op)
		}
	}

	return out
}

// fieldColors returns the caret column and the highlight fill inside the
// compose field, which is what a select-all paints there.
func fieldColors(app *App) (caret, selection [3]float64) {
	for _, op := range fieldOps(app) {
		if op.W < 2 && op.H > 5 && op.H < 20 {
			caret = [3]float64{op.R, op.G, op.B}
		}

		if op.Kind == layout.DisplayOpFillRect && op.W > 4 && op.H > 4 {
			selection = [3]float64{op.R, op.G, op.B}
		}
	}

	return caret, selection
}

// assertColor fails unless got is within a rounding step of want.
func assertColor(t *testing.T, name string, got, want [3]float64) {
	t.Helper()

	for i := range got {
		if math.Abs(got[i]-want[i]) > 0.01 {
			t.Fatalf("%s = %.3f,%.3f,%.3f, want %.3f,%.3f,%.3f",
				name, got[0], got[1], got[2], want[0], want[1], want[2])
		}
	}
}

func rgb(r, g, b byte) [3]float64 {
	return [3]float64{float64(r) / 255, float64(g) / 255, float64(b) / 255}
}
