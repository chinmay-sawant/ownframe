package eyes

import "context"

func (a *App) onKeyDown(_ context.Context, k string) error {
	switch k {
	case "1", "2", "3", "4":
		a.selectMood(int(k[0]-'1')%len(moods), a.now())
	case "s":
		a.Shake()
	case "space", "enter":
		a.st.blinkOn = true
		a.st.blinkP = 0
	case "arrowleft":
		a.st.tLookX = -44
	case "arrowright":
		a.st.tLookX = 44
	case "arrowup":
		a.st.tLookY = -26
	case "arrowdown":
		a.st.tLookY = 26
	}
	return nil
}
