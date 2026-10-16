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

// paintText shows the mood and the control labels.
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
	if a.st.dev {
		setCenteredText(a.pm.dev, "DEV ON", a.pm.devX)
	} else {
		setCenteredText(a.pm.dev, "DEV", a.pm.devX)
	}
}
