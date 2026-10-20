package viewtransition

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

func paintOps(d *ownframe.Display, indices []int, source, target Rect, alpha float64, classes []string, styles map[string]ClassStyle) []ownframe.DisplayOp {
	var scale float64 = 1
	for _, name := range classes {
		s := styles[name]
		if s.Scale != nil {
			scale *= math.Max(0, *s.Scale)
		}
	}
	alpha = styledAlpha(alpha, classes, styles)
	out := make([]ownframe.DisplayOp, 0, len(indices))
	for _, i := range indices {
		if i < 0 || i >= len(d.Ops) {
			continue
		}
		op := d.Ops[i].Clone()
		fit(&op, d.PixelPerPoint, source, target)
		if scale != 1 {
			scaleOp(&op, d.PixelPerPoint, target, scale)
		}
		op.SetPaintOpacity(clamp(op.Opacity() * alpha))
		out = append(out, op)
	}
	return out
}

func fit(op *ownframe.DisplayOp, ppp float64, from, to Rect) {
	if ppp <= 0 {
		ppp = 96.0 / 72.0
	}
	sx, sy := ratio(to.W, from.W), ratio(to.H, from.H)
	op.X = to.X/ppp + (op.X-from.X/ppp)*sx
	op.Y = to.Y/ppp + (op.Y-from.Y/ppp)*sy
	op.W *= sx
	op.H *= sy
	if op.Size != 0 {
		op.Size *= sy
	}
	if op.Width != 0 {
		op.Width *= (sx + sy) / 2
	}
}

func ratio(a, b float64) float64 {
	if b == 0 {
		return 1
	}
	return a / b
}

func scaleOp(op *ownframe.DisplayOp, ppp float64, box Rect, scale float64) {
	if ppp <= 0 {
		ppp = 96.0 / 72.0
	}
	cx, cy := (box.X+box.W/2)/ppp, (box.Y+box.H/2)/ppp
	op.X = cx + (op.X-cx)*scale
	op.Y = cy + (op.Y-cy)*scale
	op.W *= scale
	op.H *= scale
	op.Size *= scale
	op.Width *= scale
}

func clamp(v float64) float64 {
	if math.IsNaN(v) {
		return 0
	}
	return math.Max(0, math.Min(1, v))
}

func styledAlpha(alpha float64, classes []string, styles map[string]ClassStyle) float64 {
	for _, name := range classes {
		if v := styles[name].Opacity; v != nil {
			alpha *= clamp(*v)
		}
	}
	return clamp(alpha)
}
