package replay

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// makeLayerBenchDisplay constructs a layout.Display with totalOps display ops
// and 4 promoted layer boxes (cards). If nonContiguous is true, the cards'
// owned operations are interleaved with non-owned background operations so that
// lastRank - firstRank + 1 != len(owned), exercising layerOrderSafe's map allocation.
func makeLayerBenchDisplay(totalOps int, nonContiguous bool) (*layout.Display, []LayerSpec) {
	if totalOps < 20 {
		totalOps = 20
	}

	const (
		numLayers   = 4
		opsPerLayer = 4
		layerWidth  = 100
		layerHeight = 100
	)

	d := &layout.Display{
		Width:          1200,
		Height:         1200,
		PointsPerPixel: 1,
		PixelPerPoint:  1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 1200, H: 1200, ParentIndex: -1},
		},
		Ops:      make([]layout.DisplayOp, 0, totalOps),
		OpOwners: make([]int, 0, totalOps),
		Order:    make([]int, 0, totalOps),
	}

	specs := make([]LayerSpec, numLayers)
	for i := 0; i < numLayers; i++ {
		boxX := float64(50 + i*150)
		boxY := float64(50)
		boxID := fmt.Sprintf("card%d", i)
		d.Boxes = append(d.Boxes, layout.Box{
			ID:          boxID,
			X:           boxX,
			Y:           boxY,
			W:           layerWidth,
			H:           layerHeight,
			ParentIndex: 0,
		})
		specs[i] = LayerSpec{
			ElementID: boxID,
			BoxIndex:  i + 1,
			Matrix:    layerTranslate(float64(4+i*4), 0),
			Alpha:     1,
		}
	}

	// Create 4 ops for each of the 4 layers (16 ops total)
	type layerOpInfo struct {
		op    layout.DisplayOp
		owner int
	}
	var layerOps []layerOpInfo
	for i := 0; i < numLayers; i++ {
		owner := i + 1
		boxX := float64(50 + i*150)
		boxY := float64(50)
		layerOps = append(layerOps,
			// Op 0: Background fill
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX, Y: boxY, W: layerWidth, H: layerHeight,
					R: 0.1, G: 0.2, B: float64(i+1) * 0.2, Alpha: 1,
				},
				owner: owner,
			},
			// Op 1: Header fill
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 5, Y: boxY + 5, W: layerWidth - 10, H: 25,
					R: 0.8, G: 0.8, B: 0.8, Alpha: 1,
				},
				owner: owner,
			},
			// Op 2: Inner content
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 10, Y: boxY + 35, W: layerWidth - 20, H: 30,
					R: 0.3, G: 0.6, B: 0.9, Alpha: 1,
				},
				owner: owner,
			},
			// Op 3: Badge
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 10, Y: boxY + 70, W: 30, H: 20,
					R: 0.9, G: 0.3, B: 0.2, Alpha: 1,
				},
				owner: owner,
			},
		)
	}

	remainingOps := totalOps - len(layerOps)

	if !nonContiguous {
		// Contiguous layout: all layer ops placed first consecutively
		for _, lo := range layerOps {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, lo.op)
			d.OpOwners = append(d.OpOwners, lo.owner)
			d.Order = append(d.Order, idx)
		}
		// Remaining page background/other ops
		for i := 0; i < remainingOps; i++ {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, layout.DisplayOp{
				Kind: layout.DisplayOpFillRect,
				X:    float64(10 + (i%20)*50),
				Y:    float64(300 + (i/20)*30),
				W:    40, H: 20,
				R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
			})
			d.OpOwners = append(d.OpOwners, 0)
			d.Order = append(d.Order, idx)
		}
	} else {
		// Non-contiguous layout: interleave background ops between layer ops so
		// that each layer's ops are separated in d.Order by non-overlapping page ops.
		bgIndex := 0
		for layerIdx := 0; layerIdx < numLayers; layerIdx++ {
			// Add 2 ops for this layer
			for k := 0; k < 2; k++ {
				lo := layerOps[layerIdx*opsPerLayer+k]
				idx := len(d.Ops)
				d.Ops = append(d.Ops, lo.op)
				d.OpOwners = append(d.OpOwners, lo.owner)
				d.Order = append(d.Order, idx)
			}
			// Insert an intervening background op outside the cards (e.g. Y >= 300)
			if bgIndex < remainingOps {
				idx := len(d.Ops)
				d.Ops = append(d.Ops, layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    float64(10 + (bgIndex%20)*50),
					Y:    float64(300 + (bgIndex/20)*30),
					W:    40, H: 20,
					R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
				})
				d.OpOwners = append(d.OpOwners, 0)
				d.Order = append(d.Order, idx)
				bgIndex++
			}
			// Add remaining 2 ops for this layer
			for k := 2; k < 4; k++ {
				lo := layerOps[layerIdx*opsPerLayer+k]
				idx := len(d.Ops)
				d.Ops = append(d.Ops, lo.op)
				d.OpOwners = append(d.OpOwners, lo.owner)
				d.Order = append(d.Order, idx)
			}
		}
		// Append remaining background ops
		for ; bgIndex < remainingOps; bgIndex++ {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, layout.DisplayOp{
				Kind: layout.DisplayOpFillRect,
				X:    float64(10 + (bgIndex%20)*50),
				Y:    float64(300 + (bgIndex/20)*30),
				W:    40, H: 20,
				R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
			})
			d.OpOwners = append(d.OpOwners, 0)
			d.Order = append(d.Order, idx)
		}
	}

	return d, specs
}

func TestLayerBenchmarkSetup(t *testing.T) {
	for _, count := range []int{100, 500, 1000} {
		for _, nonContig := range []bool{false, true} {
			d, specs := makeLayerBenchDisplay(count, nonContig)
			var renderer Renderer
			defer renderer.Close()

			damage := renderer.BeginLayers(d, 1, specs)
			if len(damage) == 0 {
				t.Fatalf("count=%d nonContig=%v: expected damage", count, nonContig)
			}
			stats := renderer.Layers()
			if stats.Active != 4 {
				t.Fatalf("count=%d nonContig=%v: expected 4 active layers, got %d (stats: %+v)", count, nonContig, stats.Active, stats)
			}
		}
	}
}

// BenchmarkLayerBeginLayers measures steady-state per-frame allocations and latency
// of BeginLayers during animation with 4 retained layers over 100, 500, and 1000 operations.
func BenchmarkLayerBeginLayers(b *testing.B) {
	for _, count := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("ops_%d/contiguous", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, false)
			var renderer Renderer
			defer renderer.Close()

			// Warm up: initial promotion & bake
			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate per-frame 2D transform animation: matrix translation moves each frame
				specs[0].Matrix = layerTranslate(float64(4+(i&31)), 0)
				renderer.BeginLayers(d, 1, specs)
			}
		})

		b.Run(fmt.Sprintf("ops_%d/non_contiguous", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, true)
			var renderer Renderer
			defer renderer.Close()

			// Warm up: initial promotion & bake
			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate per-frame 2D transform animation: matrix translation moves each frame
				specs[0].Matrix = layerTranslate(float64(4+(i&31)), 0)
				renderer.BeginLayers(d, 1, specs)
			}
		})
	}
}

// BenchmarkLayerDrawLayers measures allocation overhead when compositing retained
// layers onto the destination buffer (DrawLayers).
func BenchmarkLayerDrawLayers(b *testing.B) {
	for _, count := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("ops_%d/normal", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, false)
			var renderer Renderer
			defer renderer.Close()

			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			dst := ebiten.NewImage(d.Width, d.Height)
			defer dst.Deallocate()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				renderer.DrawLayers(dst, d, 0, 0)
			}
		})
	}

	b.Run("blend_mode", func(b *testing.B) {
		d, specs := makeLayerBenchDisplay(100, false)
		specs[0].Mode = "multiply"
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, specs)
		if renderer.Layers().Active != 4 {
			b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
		}

		dst := ebiten.NewImage(d.Width, d.Height)
		defer dst.Deallocate()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.DrawLayers(dst, d, 0, 0)
		}
	})
}
