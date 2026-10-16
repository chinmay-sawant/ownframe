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
	if !a.st.automatic {
		setCenteredText(a.pm.hint, "STATIC - TAP EYES TO CHANGE", a.pm.hintX)
	} else if a.tch {
		setCenteredText(a.pm.hint, "TAP TO CHANGE MOOD - SWIPE TO LOOK", a.pm.hintX)
	} else {
		setCenteredText(a.pm.hint, "TAP EYES - DRAG LOOK - 1-4 SELECT - S SPIRAL", a.pm.hintX)
	}
	if a.st.dev {
		setCenteredText(a.pm.dev, "DEV ON", a.pm.devX)
	} else {
		setCenteredText(a.pm.dev, "DEV", a.pm.devX)
	}
}
