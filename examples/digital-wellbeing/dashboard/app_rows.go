package dashboard

import "strings"

func appRows(apps []AppUsage) []AppUsage {
	max := 0
	for _, app := range apps {
		if app.Minutes > max {
			max = app.Minutes
		}
	}
	for i := range apps {
		apps[i].Time = duration(apps[i].Minutes)
		if max > 0 {
			apps[i].Percent = apps[i].Minutes * 100 / max
		}
	}
	return apps
}

func (a *App) usageRows(apps []AppUsage) []AppUsage {
	icons := make(map[string]string, len(a.view.Apps))
	for _, app := range a.view.Apps {
		icons[app.Package] = app.IconSource
	}
	for i := range apps {
		apps[i].IconSource = icons[apps[i].Package]
		apps[i].Letter = "•"
		if apps[i].Label != "" {
			apps[i].Letter = strings.ToUpper(string([]rune(apps[i].Label)[0]))
		}
	}
	return appRows(apps)
}
