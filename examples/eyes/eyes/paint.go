package eyes

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

// paint moves the cached eye ops to match the simulation.
func (a *App) paint() {
	d := a.page.Testing().Display()
	if d == nil {
		return
	}
	if a.page.Testing().Generation() != a.bnd {
		a.bind(d)
	}
	// Rebase when layout moved (resize) but generation stayed.
	a.rebase(d)
	m := moodName(a.st.mood)
	eyeAlpha := max(eyeAlphaMin, 1-a.st.spiralMix)
	left, right := a.st.styleL, a.st.styleR
	now := a.now()
	if m == "angry" {
		lean := angrySwayAmp * math.Sin(now.Sub(a.st.birth).Seconds()*angrySwayHz)
		left.slant += lean
		right.slant -= lean
	}
	pose := poseFor(a.st.mood)
	if pose.wink != winkNone {
		open := 1 - lidClose*winkPulse(now.Sub(a.st.poseAt))
		if pose.wink == winkScreenLeft {
			left.open = open
		} else {
			right.open = open
		}
	}
	a.paintEye(d, a.pm.left, a.pm.lidLT, a.pm.lidLB, a.pm.baseL, a.pm.rcL, left)
	a.paintEye(d, a.pm.right, a.pm.lidRT, a.pm.lidRB, a.pm.baseR, a.pm.rcR, right)
	for _, op := range []*ownframe.DisplayOp{
		a.pm.left, a.pm.right, a.pm.lidLT, a.pm.lidLB, a.pm.lidRT, a.pm.lidRB,
	} {
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
		a.pm.rcL = baseRCss(d, a.pm.left)
	}
	if a.pm.right != nil && a.pm.baseR.w == 0 {
		a.pm.baseR = baseOf(d, a.pm.right)
		a.pm.rcR = baseRCss(d, a.pm.right)
	}
}
