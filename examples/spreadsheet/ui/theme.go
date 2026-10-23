package ui

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// toggleTheme switches the light and dark sheets and persists the choice.
func (a *App) toggleTheme() {
	a.dark = !a.dark
	a.status = "theme"
	a.work.post(job{kind: jobPref, key: "theme", value: a.themeName()})
}

// themeName is the stored and printed theme name.
func (a *App) themeName() string {
	if a.dark {
		return "dark"
	}

	return "light"
}

// onChange reads a dialog field after the page changed it.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "csvpath":
		if path := a.page.Form().FormValue("csvpath"); path != "" {
			a.loadCSV(path)
		}
	case "csvsave":
		a.csv.path = a.page.Form().FormValue("csvsave")
	}

	return nil
}
