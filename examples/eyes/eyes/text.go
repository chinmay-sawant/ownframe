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
	setText(a.pm.mood, strings.ToUpper(moodName(a.st.mood)))
	if a.tch {
		setText(a.pm.hint, "TAP TO CHANGE MOOD - SWIPE TO LOOK")
	} else {
		setText(a.pm.hint, "CLICK FOR MOOD - DRAG TO LOOK - 1-5 MOODS")
	}
	if a.st.dev {
		setText(a.pm.dev, "DEV ON")
	} else {
		setText(a.pm.dev, "DEV")
	}
}
