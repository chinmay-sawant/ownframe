package eyes

import "context"

func (a *App) onKeyDown(_ context.Context, k string) error {
	a.wakeHUD(a.now())
	switch k {
	case "1", "2", "3", "4":
		a.selectMood(int(k[0]-'1')%len(poses), a.now())
	case "s":
		if a.st.automatic {
			a.Shake()
		} else {
			a.selectMood(spiralMood, a.now())
		}
	case "space", "enter":
		a.st.blinkOn = true
		a.st.blinkP = 0
	case "arrowleft":
		a.st.tLookX = -keyLookX
	case "arrowright":
		a.st.tLookX = keyLookX
	case "arrowup":
		a.st.tLookY = -keyLookY
	case "arrowdown":
		a.st.tLookY = keyLookY
	}
	return nil
}
