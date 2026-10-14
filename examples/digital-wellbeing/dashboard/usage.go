package dashboard

import "time"

func (a *App) clearUsage() {
	if a.view.Selected == "" {
		a.view.SelectedName = "All apps"
	} else {
		a.view.SelectedName = "Loading app stats"
	}
	a.view.Days = make([]Day, 7)
	today := time.Now()
	for i := range a.view.Days {
		day := today.AddDate(0, 0, i-6)
		a.view.Days[i] = Day{Label: day.Format("Mon"), Time: "0m"}
	}
	a.view.Today, a.view.Week, a.view.Average = "—", "—", "—"
	a.view.GoalMessage = "Mochi is finding your recent screen time"
}

func (a *App) applyUsage(usage Usage) {
	a.view.SelectedName = usage.Label
	a.view.Days = make([]Day, 0, len(usage.Days))
	maxMinutes, total := 0, 0
	for _, day := range usage.Days {
		if day.Minutes > maxMinutes {
			maxMinutes = day.Minutes
		}
		total += day.Minutes
	}
	if maxMinutes < 60 {
		maxMinutes = 60
	}
	for _, day := range usage.Days {
		level := day.Minutes * 6 / maxMinutes
		if day.Minutes > 0 && level == 0 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		a.view.Days = append(a.view.Days, Day{
			Label: day.Label, Minutes: day.Minutes,
			Time: duration(day.Minutes), Level: level,
		})
	}
	today := 0
	if len(usage.Days) > 0 {
		today = usage.Days[len(usage.Days)-1].Minutes
	}
	a.view.Today = duration(today)
	a.view.Week = duration(total)
	if len(usage.Days) > 0 {
		a.view.Average = duration(total / len(usage.Days))
	} else {
		a.view.Average = "0m"
	}
	if today < 120 {
		a.view.GoalMessage = duration(120-today) + " to your gentle 2 hour goal"
	} else {
		a.view.GoalMessage = "You reached your gentle 2 hour goal"
	}
}

func duration(minutes int) string {
	if minutes < 60 {
		return itoa(minutes) + "m"
	}
	return itoa(minutes/60) + "h " + twoDigits(minutes%60) + "m"
}
