package eyes

import "github.com/chinmay-sawant/ownframe"

func (a *App) bindText(d *ownframe.Display, boxes []ownframe.Box) {
	a.pm.mood = textAt(d, boxes, "t-mood")
	a.pm.dev = textAt(d, boxes, "dev-toggle")
	a.pm.mode = textAt(d, boxes, "mode-toggle")
	a.pm.devBG = fillAt(d, boxes, "dev-toggle")
	a.pm.modeBG = fillAt(d, boxes, "mode-toggle")
	a.pm.moodX = textCenterX(d, boxes, "t-mood")
	a.pm.devX = textCenterX(d, boxes, "dev-toggle")
	a.pm.modeX = textCenterX(d, boxes, "mode-toggle")
}
