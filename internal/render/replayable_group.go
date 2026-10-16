package render

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// maxGroupDepth bounds the layer stack one group may nest.
const maxGroupDepth = 4

// maxXformPixels bounds one transformed op's scratch raster.
const maxXformPixels = 4 << 20

// replayableGroup accepts a group when every ancestor's mode is separable
// and no member crosses the window's fixed or pinned paint bands, where a
// buffered layer would split across repaints.
func replayableGroup(op *layout.DisplayOp) bool {
	for g, depth := op.Group(), 0; g != nil; g, depth = g.Parent, depth+1 {
		if depth >= maxGroupDepth || !separableMode(g.Mode) {
			return false
		}
		if g.Mask != nil && (op.Fixed || op.ZIndex != 0) {
			return false
		}

		if !masking.BlendNormal(g.Mode) && (op.Fixed || op.ZIndex != 0) {
			return false
		}
	}

	return true
}

// separableMode reports the mix-blend-modes the group compositor reproduces
// per pixel. hue, saturation, color, and luminosity need channel mixing.
func separableMode(mode string) bool {
	return masking.BlendSeparable(mode)
}

// finiteXform rejects matrices with a NaN or infinite entry.
func finiteXform(op *layout.DisplayOp) bool {
	m := op.Transform()
	for _, v := range [...]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}

	return true
}

// xformTooBig rejects a transformed op whose local scratch raster would
// exceed maxXformPixels CSS pixels.
func xformTooBig(op *layout.DisplayOp, ppp float64) bool {
	grow := 2*math.Max(op.Width, 1) + 4
	w, h := math.Abs(op.W)+grow, math.Abs(op.H)+grow

	return w*h*ppp*ppp > maxXformPixels
}
