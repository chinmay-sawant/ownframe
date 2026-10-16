package eyes

import "github.com/chinmay-sawant/ownframe"

// paintEye keeps one rounded eye and shapes it with top and bottom lids.
func (a *App) paintEye(d *ownframe.Display, eye, top, bot *ownframe.DisplayOp,
	b rect, rc float64, style eyeStyle) {
	if eye == nil || b.h <= 0 {
		return
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	h := b.h * style.height
	x := b.x + a.st.lookX + style.gazeX
	y := b.y + a.st.lookY + style.gazeY + (b.h-h)/2
	setOp(d, eye, x, y, b.w, h, 1)
	eye.Radius = min(rc*h/b.h, min(b.w, h)/2*0.98) * p
	lid := (1 - min(style.open, 1)) * h
	blink := (1 - a.st.bm) * h
	topH := blink*.6 + lid*style.top
	botH := blink*.4 + lid*style.bottom
	if topH+botH > h-6 && h > 6 {
		f := (h - 6) / (topH + botH)
		topH, botH = topH*f, botH*f
	}
	setLid(d, top, x-2, y-2, b.w+4, topH+2)
	setOpRotation(top, style.slant)
	setLid(d, bot, x-2, y+h-botH, b.w+4, botH+2)
}
