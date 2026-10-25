package replay

import (
	"bytes"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Compare the engine's projected-baseline contract with a known local run
// transformed through the same matrix. This runs inside the GPU parity loop.
func checkProjectedText() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}
	var renderer Renderer
	defer renderer.Close()
	actual, expected := newParityImage(320, 240), newParityImage(320, 240)
	defer actual.Deallocate()
	defer expected.Deallocate()
	for _, matrix := range []layout.Matrix2D{
		{A: 0.5, D: 1, E: 10},
		{A: 1, D: 0.5, F: 10},
		{A: 1.2, D: 1.2},
	} {
		local := ops[0]
		local.X, local.Y = 30, 40
		local.SetXform(matrix)
		local.XformSet = true
		projected := local
		projected.X, projected.Y = matrix.Apply(local.X, local.Y)
		projected.XformSet = false
		actual.Clear()
		expected.Clear()
		renderer.drawOp(actual, &projected, 0, 0)
		if projected.Transform() != matrix {
			return fmt.Errorf("replaying projected text mutated the display matrix")
		}
		renderer.drawOp(expected, &local, 0, 0)
		a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
		actual.ReadPixels(a)
		expected.ReadPixels(b)
		if !bytes.Equal(a, b) {
			return fmt.Errorf("projected text differs from transformed local text for %+v", matrix)
		}
		got, _ := opBounds(&projected, 1/pxPerPt)
		want, _ := opBounds(&local, 1/pxPerPt)
		if got != want {
			return fmt.Errorf("projected text bounds = %v, want %v", got, want)
		}
	}
	return nil
}
