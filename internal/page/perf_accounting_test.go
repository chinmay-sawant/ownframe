package page

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPerfDisjointRegions(t *testing.T) {
	p := &Page{}
	d := &Display{PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3, Ops: []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 150, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 75, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 165, H: 15},
	}}
	rects := []image.Rectangle{image.Rect(0, 0, 20, 20), image.Rect(200, 0, 220, 20)}
	p.recordPartial(2, rects, d)
	if p.stats.dirtyRegions != 2 || p.stats.dirtyArea != 800 || p.stats.dirtyOps != 3 {
		t.Fatalf("regions=%d area=%d ops=%d", p.stats.dirtyRegions, p.stats.dirtyArea, p.stats.dirtyOps)
	}
	p.recordPartial(0, nil, d)
	if p.stats.dirtyRegions != 0 || p.stats.dirtyArea != 0 || p.stats.dirtyOps != 0 {
		t.Fatal("empty damage retained counters")
	}
}
