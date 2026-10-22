package eyes

import "github.com/chinmay-sawant/ownframe"

// paintEye keeps the rounded eye shell fixed and shapes its visible opening.
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
	y := b.y + a.st.lookY + style.gazeY + (b.h-h)/halfDivisor
	setOp(d, eye, x, y, b.w, h, 1)
	eye.Radius = min(rc*h/b.h, min(b.w, h)/halfDivisor*eyeRound) * p
	lid := (1 - min(style.open, 1)) * h
	blink := (1 - a.st.bm) * h
	topH := blink*lidTopShare + lid*style.top
	botH := blink*lidLowShare + lid*style.bottom
	if topH+botH > h-lidMinGap && h > lidMinGap {
		f := (h - lidMinGap) / (topH + botH)
		topH, botH = topH*f, botH*f
	}
	setLid(d, top, x-lidOutset, y-lidOutset, b.w+lidGrow, topH+lidOutset)
	setOpRotation(top, style.slant)
	setLid(d, bot, x-lidOutset, y+h-botH, b.w+lidGrow, botH+lidOutset)
}
