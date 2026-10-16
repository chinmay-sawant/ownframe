package eyes

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	hudIdle = 5 * time.Second
	hudFade = 300 * time.Millisecond
	hudMin  = .0001
)

func (a *App) wakeHUD(now time.Time) { a.st.hudUntil = now.Add(hudIdle) }

func (a *App) hudAlpha(now time.Time) float64 {
	remaining := a.st.hudUntil.Sub(now)
	if remaining >= hudFade {
		return 1
	}
	return clamp(float64(remaining)/float64(hudFade), 0, 1)
}

func (a *App) paintHUD(now time.Time) {
	alpha := a.hudAlpha(now)
	visible := alpha > 0
	if !visible {
		alpha = hudMin
		for _, op := range []*ownframe.DisplayOp{a.pm.dev, a.pm.mode} {
			setText(op, "")
		}
	}
	if !visible {
		clearTextOps(a.pm.hintRuns)
	}
	for _, op := range []*ownframe.DisplayOp{a.pm.dev, a.pm.mode,
		a.pm.devBG, a.pm.modeBG} {
		if op != nil {
			op.SetPaintOpacity(alpha)
		}
	}
	for _, op := range a.pm.hintRuns {
		op.SetPaintOpacity(alpha)
	}
}
