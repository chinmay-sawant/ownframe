package eyes

import (
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func setText(op *ownframe.DisplayOp, s string) {
	if op != nil && op.Text != s {
		op.Text = s
	}
}

func clearTextOps(ops []*ownframe.DisplayOp) {
	for _, op := range ops {
		setText(op, "")
	}
}

// paintText shows the mood, the hint, and the devtools switch label.
func (a *App) paintText() {
	mood := moodName(a.st.mood)
	if a.st.spiralMix > 0 {
		mood = "spiral"
	}
	setCenteredText(a.pm.mood, strings.ToUpper(mood), a.pm.moodX)
	if a.st.automatic {
		setCenteredText(a.pm.mode, "AUTO", a.pm.modeX)
	} else {
		setCenteredText(a.pm.mode, "STATIC", a.pm.modeX)
	}
	hint := "TAP ◀ PREVIOUS · TAP ▶ NEXT · DOUBLE TAP SAME SIDE"
	if !a.tch && a.st.automatic {
		hint = "CLICK ◀ PREV · CLICK ▶ NEXT · DRAG LOOK · 1-4 MOODS · S SPIRAL"
	}
	setCenteredText(a.pm.hint, hint, a.pm.hintX)
	if len(a.pm.hintRuns) > 1 {
		clearTextOps(a.pm.hintRuns[1:])
	}
	if a.st.dev {
		setCenteredText(a.pm.dev, "DEV ON", a.pm.devX)
	} else {
		setCenteredText(a.pm.dev, "DEV", a.pm.devX)
	}
}
