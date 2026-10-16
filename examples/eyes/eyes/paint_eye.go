package eyes

import "github.com/chinmay-sawant/ownframe"

// paintEye changes eye height without drawing brow-like lid masks.
func (a *App) paintEye(d *ownframe.Display, eye *ownframe.DisplayOp,
	b rect, style eyeStyle) {
	if eye == nil || b.h <= 0 {
		return
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	opening := clamp(style.open*a.st.bm, .02, 1.2)
	h := b.h * style.height * opening
	x := b.x + a.st.lookX + style.gazeX
	y := b.y + a.st.lookY + style.gazeY + (b.h-h)/2
	setOp(d, eye, x, y, b.w, h, 1)
	eye.Radius = min(b.w, h) / 2 * 0.98 * p
}
