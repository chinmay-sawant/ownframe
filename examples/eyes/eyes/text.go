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
	setText(a.pm.mood, strings.ToUpper(mood))
	if a.st.automatic {
		setText(a.pm.mode, "AUTO")
	} else {
		setText(a.pm.mode, "STATIC")
	}
	if !a.st.automatic {
		setText(a.pm.hint, "STATIC - TAP EYES TO CHANGE")
	} else if a.tch {
		setText(a.pm.hint, "TAP TO CHANGE MOOD - SWIPE TO LOOK")
	} else {
		setText(a.pm.hint, "TAP EYES - DRAG LOOK - 1-4 SELECT - S SPIRAL")
	}
	if a.st.dev {
		setText(a.pm.dev, "DEV ON")
	} else {
		setText(a.pm.dev, "DEV")
	}
}
