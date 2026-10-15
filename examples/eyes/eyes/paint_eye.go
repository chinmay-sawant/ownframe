package eyes

import "github.com/chinmay-sawant/ownframe"

// paintEye keeps one rounded eye and shapes it with top and bottom lids.
func (a *App) paintEye(d *ownframe.Display, eye, top, bot *ownframe.DisplayOp,
	b rect, rc, moodO float64, mood string, slant float64) {
	if eye == nil || b.h <= 0 {
		return
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	h := b.h
	if moodO > 1 {
		h = b.h * min(moodO, 1.25)
	}
	x := b.x + a.st.lookX
	y := b.y + a.st.lookY + (b.h-h)/2
	setOp(d, eye, x, y, b.w, h, 1)
	eye.Radius = min(rc*h/b.h, min(b.w, h)/2*0.98) * p
	mTop, mBot := 0.0, 0.0
	switch mood {
	case "squint":
		mTop, mBot = 0.5, 0.5
	case "angry":
		mTop, mBot = 1, 0.08
	}
	lid := (1 - min(moodO, 1)) * h
	topH := (1-a.st.bm)*h + lid*mTop
	botH := lid * mBot
	if topH+botH > h-6 && h > 6 {
		f := (h - 6) / (topH + botH)
		topH, botH = topH*f, botH*f
	}
	setLid(d, top, x-2, y-2, b.w+4, topH+2)
	setOpRotation(top, slant)
	setLid(d, bot, x-2, y+h-botH, b.w+4, botH+2)
}
