package test_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// ----------------------------------------------------------------------------
// Category 3: GPU Raster & Shader Replay Benchmarks
// ----------------------------------------------------------------------------

const blendShaderHTML = `
<style>
  .stage { position: relative; width: 400px; height: 400px; background: #000; }
  .layer-base { width: 200px; height: 200px; background: #ff4500; border-radius: 16px; }
  .layer-screen {
    position: absolute; top: 50px; left: 50px; width: 200px; height: 200px;
    background: #00bfff; border-radius: 16px; mix-blend-mode: screen;
  }
  .layer-multiply {
    position: absolute; top: 100px; left: 100px; width: 200px; height: 200px;
    background: #32cd32; border-radius: 16px; mix-blend-mode: multiply;
  }
</style>
<div class="stage">
  <div class="layer-base"></div>
  <div class="layer-screen"></div>
  <div class="layer-multiply"></div>
</div>
`

func BenchmarkGPU_Raster_Card_PNG(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: cardHTML, Width: 400, Height: 300})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Probe().PNG()
	}
}

func BenchmarkGPU_Raster_Dashboard_PNG(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: dashboardLayoutHTML, Width: 800, Height: 600})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Probe().PNG()
	}
}

func BenchmarkGPU_Shader_MixBlend_PNG(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: blendShaderHTML, Width: 400, Height: 400})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Probe().PNG()
	}
}

func BenchmarkGPU_DisplayList_VectorOps(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: blendShaderHTML, Width: 400, Height: 400})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := p.Probe().Display()
		if d == nil {
			b.Fatal("nil display")
		}
	}
}
