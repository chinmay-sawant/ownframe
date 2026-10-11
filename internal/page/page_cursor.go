package page

import "github.com/chinmay-sawant/ownframe/internal/host"

// CursorShape returns the shape for the hovered box: an I-beam over a text
// control, a hand over a link, a button, or another control, and the default
// arrow everywhere else.
func (p *Page) CursorShape() host.Shape {
	if p.hover != "" {
		if c, ok := p.control(p.hover); ok {
			if canEdit(c) {
				return host.ShapeText
			}

			return host.ShapePointer
		}

		b := p.boxByID(p.hover)
		switch b.Tag {
		case "a", "button":
			return host.ShapePointer
		}

		if b.Action != "" {
			return host.ShapePointer
		}
	}

	if p.linkAt(p.hoverX, p.hoverY) {
		return host.ShapePointer
	}

	return host.ShapeDefault
}

// linkAt reports whether the point is over a link's text run.
func (p *Page) linkAt(x, y float64) bool {
	d := p.display
	if d == nil || d.PointsPerPixel <= 0 {
		return false
	}

	pt := d.PointsPerPixel
	xPt, yPt := x*pt, y*pt
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != DisplayOpLinkURI {
			continue
		}

		if xPt >= op.X && xPt <= op.X+op.W && yPt >= op.Y && yPt <= op.Y+op.H {
			return true
		}
	}

	return false
}
