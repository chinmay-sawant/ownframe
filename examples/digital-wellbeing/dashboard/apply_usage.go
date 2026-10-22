package dashboard

func (a *App) applyUsage(usage Usage) {
	a.view.SelectedName = usage.Label
	a.view.WeekRange = usage.WeekRange
	maxMinutes, total := 0, 0
	for _, day := range usage.Days {
		if day.Minutes > maxMinutes {
			maxMinutes = day.Minutes
		}
		total += day.Minutes
	}
	if maxMinutes < minutesPerHour {
		maxMinutes = minutesPerHour
	}
	a.view.Days = make([]Day, 0, len(usage.Days))
	cumulative := 0
	for i, day := range usage.Days {
		cumulative += day.Minutes
		level := day.Minutes * barLevels / maxMinutes
		if day.Minutes > 0 && level == 0 {
			level = 1
		}
		if level > barLevels {
			level = barLevels
		}
		a.view.Days = append(a.view.Days, Day{Index: i, Label: day.Label,
			Date: day.Date, Minutes: day.Minutes, Time: duration(day.Minutes),
			Average: duration(cumulative / (i + 1)), Level: level,
			Selected: a.view.HasSelectedDay && i == a.view.SelectedDay.Index,
			Apps:     a.usageRows(day.Apps)})
	}
	a.applyWeeklyInsights(usage)
	if len(a.view.Days) > 0 {
		last := a.view.Days[len(a.view.Days)-1]
		a.view.Today = last.Time
		if usage.WeekOffset > 0 {
			a.view.HeroEyebrow = "Week ending " + last.Date
			a.view.HeroCaption = "Time in the foreground on the final day"
		} else if a.view.Selected != "" {
			a.view.HeroEyebrow = "Today with " + usage.Label
		}
	}
	a.view.Week = duration(total)
	a.view.Average = duration(total / daysPerWeek)
	if a.view.SelectedDay.Index >= 0 && a.view.SelectedDay.Index < len(a.view.Days) && a.view.Screen == "day" {
		a.view.SelectedDay = a.view.Days[a.view.SelectedDay.Index]
	}
	if usage.WeekOffset == 0 && len(a.view.Days) > 0 {
		today := a.view.Days[len(a.view.Days)-1].Minutes
		if today < goalMinutes {
			a.view.GoalMessage = duration(goalMinutes-today) + " to your gentle 2 hour goal"
		} else {
			a.view.GoalMessage = "You reached your gentle 2 hour goal"
		}
	}
}
