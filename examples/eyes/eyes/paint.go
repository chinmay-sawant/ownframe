package eyes

import "github.com/chinmay-sawant/ownframe"

// paint moves the cached eye ops to match the simulation.
func (a *App) paint() {
	d := a.page.Display()
	if d == nil {
		return
	}
	if a.page.Generation() != a.bnd {
		a.bind(d)
	}
	// Rebase when layout moved (resize) but generation stayed.
	a.rebase(d)
	m := moodName(a.st.mood)
	a.paintEye(d, a.pm.left, a.pm.lidLT, a.pm.lidLB, a.pm.baseL, a.pm.rcL, a.st.moodL, m)
	a.paintEye(d, a.pm.right, a.pm.lidRT, a.pm.lidRB, a.pm.baseR, a.pm.rcR, a.st.moodR, m)
	a.paintText()
}

func (a *App) rebase(d *ownframe.Display) {
	if a.pm.left != nil && a.pm.baseL.w == 0 {
		a.pm.baseL = baseOf(d, a.pm.left)
		a.pm.rcL = baseRCss(d, a.pm.left)
	}
	if a.pm.right != nil && a.pm.baseR.w == 0 {
		a.pm.baseR = baseOf(d, a.pm.right)
		a.pm.rcR = baseRCss(d, a.pm.right)
	}
}

// paintEye draws the full capsule plus two background lids. The eye keeps
// its size and one uniform radius; the lids crop it into a dome, a cup, or
// a slit, so squints read as lids and never pinch into a fish.
func (a *App) paintEye(d *ownframe.Display, eye, top, bot *ownframe.DisplayOp, b rect, rc, moodO float64, mood string) {
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
	nr := min(rc*h/b.h, min(b.w, h)/2*0.98)
	eye.Radius = nr * p
	mTop, mBot := 0.0, 0.0
	switch mood {
	case "happy":
		mBot = 1
	case "sleepy", "curious":
		mTop, mBot = 0.85, 0.15
	}
	lid := (1 - min(moodO, 1)) * h
	topH := (1-a.st.bm)*h + lid*mTop
	botH := lid * mBot
	if topH+botH > h-6 && h > 6 {
		f := (h - 6) / (topH + botH)
		topH, botH = topH*f, botH*f
	}
	setLid(d, top, x-2, y-2, b.w+4, topH+2)
	setLid(d, bot, x-2, y+h-botH, b.w+4, botH+2)
}
