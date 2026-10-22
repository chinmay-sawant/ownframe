package dashboard

import "time"

func (a *App) clearUsage() {
	offset := a.WeekOffset()
	a.view.WeekOffset = offset
	start := time.Now().AddDate(0, 0, -weekSpanDays-daysPerWeek*offset)
	a.view.Days = make([]Day, daysPerWeek)
	for i := range a.view.Days {
		day := start.AddDate(0, 0, i)
		a.view.Days[i] = Day{Index: i, Label: day.Format("Mon"), Date: day.Format("Jan 2"), Time: "0m"}
	}
	end := start.AddDate(0, 0, weekSpanDays)
	a.view.WeekRange = start.Format("Jan 2") + " – " + end.Format("Jan 2")
	a.view.WeekHeading = "This week"
	if offset > 0 {
		a.view.WeekHeading = "Week of " + start.Format("Jan 2")
	}
	a.view.CanPrevious, a.view.CanNext = offset < weeksPerYear, offset > 0
	a.view.Today, a.view.Week, a.view.Average = "—", "—", "—"
	a.view.InsightWeek, a.view.InsightAverage = "—", "—"
	a.view.WeekdayTotal, a.view.WeekendTotal = "—", "—"
	a.view.TopApps, a.view.AllApps = nil, nil
	a.view.SelectedDay, a.view.HasSelectedDay = Day{}, false
	a.view.TopApp, a.view.TopAppTime = "No app data yet", "—"
	if a.view.Selected == "" {
		a.view.SelectedName = "All apps"
	} else {
		a.view.SelectedName = "Loading app stats"
	}
	a.view.HeroEyebrow, a.view.HeroCaption = "Today across all apps", "Time in the foreground today"
	if a.view.Selected != "" {
		a.view.HeroEyebrow = "Today with " + a.view.SelectedName
	}
	if offset > 0 {
		a.view.HeroEyebrow = "Week ending " + end.Format("Jan 2")
		a.view.HeroCaption = "Time in the foreground on the final day"
	}
	a.view.GoalMessage = "Mochi is finding your recent screen time"
}
