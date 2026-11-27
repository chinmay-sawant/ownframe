package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// BenchmarkBlendSiblingLayerPromotion measures retained layer resolution when
// an unrelated blend-mode element exists elsewhere on the page.
func BenchmarkBlendSiblingLayerPromotion(b *testing.B) {
	blendGroup := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}
	d := &layout.Display{
		Width:          400,
		Height:         200,
		PointsPerPixel: 1,
		PixelPerPoint:  1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 80, H: 40, R: 1, G: 0, B: 0, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpFillRect, X: 20, Y: 20, W: 40, H: 20, R: 0, G: 1, B: 0, Alpha: 1},
			{ID: 3, Kind: layout.DisplayOpFillRect, X: 200, Y: 10, W: 60, H: 60, R: 0, G: 0, B: 1, Alpha: 1},
		},
		OpOwners: []int{0, 0, 1},
		Order:    []int{0, 1, 2},
		Boxes: []layout.Box{
			{ID: "card", X: 10, Y: 10, W: 80, H: 40, ParentIndex: -1, OperationIDs: []uint64{1, 2}},
			{ID: "badge", X: 200, Y: 10, W: 60, H: 60, ParentIndex: -1, OperationIDs: []uint64{3}},
		},
	}
	d.Ops[2].SetBlendGroup(blendGroup)
	spec := LayerSpec{
		ElementID: "card",
		BoxIndex:  0,
		Matrix:    layout.Matrix2D{A: 1, D: 1, E: 12},
		Base:      layout.Matrix2D{A: 1, D: 1},
		Alpha:     1,
	}

	b.ReportAllocs()
	var promoted int
	for b.Loop() {
		_, reason := resolveLayer(d, spec, 1, nil, nil)
		if reason == "" {
			promoted++
		}
	}
	b.ReportMetric(float64(promoted)/float64(b.N), "promoted/op")
}
