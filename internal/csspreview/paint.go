package csspreview

import "github.com/chinmay-sawant/ownframe"

type paint struct{ x, y, w, h, r, g, b, a, opacity float64 }

func capture(d *ownframe.Display, box ownframe.Box) []paint {
	out := []paint{}
	if d != nil {
		for i := range d.Ops {
			op := d.Ops[i]
			if inside(d, box, op) {
				out = append(out, read(&d.Ops[i]))
			}
		}
	}
	return out
}

func inside(d *ownframe.Display, b ownframe.Box, o ownframe.DisplayOp) bool {
	x, y := o.X+o.W/2, o.Y+o.H/2
	if o.Kind == ownframe.DisplayOpText {
		y = o.Y
	}
	p := d.PointsPerPixel
	return x >= b.X*p && x <= (b.X+b.W)*p && y >= b.Y*p && y <= (b.Y+b.H)*p
}

func read(o *ownframe.DisplayOp) paint {
	return paint{
		x: o.X, y: o.Y, w: o.W, h: o.H,
		r: o.R, g: o.G, b: o.B, a: 1, opacity: o.Opacity(),
	}
}

func write(o *ownframe.DisplayOp, v paint) {
	o.X, o.Y, o.W, o.H = v.x, v.y, v.w, v.h
	o.R, o.G, o.B, o.Alpha = v.r, v.g, v.b, v.a
	o.SetPaintOpacity(v.opacity)
}
