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
	eyeAlpha := max(0.0001, 1-a.st.spiralMix)
	left, right := a.st.styleL, a.st.styleR
	now := a.now()
	pose := poseFor(a.st.mood)
	if pose.wink != winkNone {
		open := 1 - .98*winkPulse(now.Sub(a.st.poseAt))
		if pose.wink == winkScreenLeft {
			left.open = open
		} else {
			right.open = open
		}
	}
	a.paintEye(d, a.pm.left, a.pm.baseL, left)
	a.paintEye(d, a.pm.right, a.pm.baseR, right)
	for _, op := range []*ownframe.DisplayOp{a.pm.left, a.pm.right} {
		if op != nil {
			op.Alpha = eyeAlpha
		}
	}
	a.paintSpiral(d, a.pm.baseL, a.pm.spiralL, a.st.spiralMix, a.st.spiral)
	a.paintSpiral(d, a.pm.baseR, a.pm.spiralR, a.st.spiralMix, -a.st.spiral)
	a.paintText()
	a.paintHUD(now)
}

func (a *App) rebase(d *ownframe.Display) {
	if a.pm.left != nil && a.pm.baseL.w == 0 {
		a.pm.baseL = baseOf(d, a.pm.left)
	}
	if a.pm.right != nil && a.pm.baseR.w == 0 {
		a.pm.baseR = baseOf(d, a.pm.right)
	}
}
