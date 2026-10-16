package pet

import (
	"fmt"
	"strings"
	"time"
)

// View is passed to the HTML template renderer.
type View struct {
	Energy         int
	Hunger         int
	Happiness      int
	Mood           string
	MoodUpper      string
	MoodEmoji      string
	Speech         string
	EnergyWidth    string
	HungerWidth    string
	HappinessWidth string
	IsSleeping     bool
	StatusHint     string
}

func (a *App) refreshView(now time.Time) {
	e := int(a.state.Energy + 0.5)
	h := int(a.state.Hunger + 0.5)
	hp := int(a.state.Happiness + 0.5)

	speech := a.state.Speech
	if now.After(a.state.SpeechUntil) {
		speech = statusHint(a.state)
	}

	a.view = View{
		Energy:         e,
		Hunger:         h,
		Happiness:      hp,
		Mood:           a.state.Mood,
		MoodUpper:      strings.ToUpper(a.state.Mood),
		MoodEmoji:      moodEmoji(a.state.Mood),
		Speech:         speech,
		EnergyWidth:    fmt.Sprintf("%d%%", e),
		HungerWidth:    fmt.Sprintf("%d%%", h),
		HappinessWidth: fmt.Sprintf("%d%%", hp),
		IsSleeping:     a.state.Sleeping,
		StatusHint:     statusHint(a.state),
	}
	a.page.SetData(&a.view)
}
