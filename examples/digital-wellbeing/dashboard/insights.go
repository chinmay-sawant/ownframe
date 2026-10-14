package dashboard

func (a *App) applyWeeklyInsights(usage Usage) {
	a.view.InsightWeek = duration(usage.AllMinutes)
	a.view.InsightAverage = duration(usage.AllMinutes / 7)
	a.view.WeekdayTotal = duration(usage.WeekdayMinutes)
	a.view.WeekendTotal = duration(usage.WeekendMinutes)
	a.view.AllApps = a.usageRows(usage.TopApps)
	a.view.TopApps = a.view.AllApps
	if len(a.view.TopApps) > 5 {
		a.view.TopApps = a.view.TopApps[:5]
	}
	a.view.TopApp, a.view.TopAppTime = "No app data yet", "0m"
	if len(a.view.TopApps) > 0 {
		a.view.TopApp = a.view.TopApps[0].Label
		a.view.TopAppTime = a.view.TopApps[0].Time
	}
}
