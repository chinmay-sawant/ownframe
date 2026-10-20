package app

import "strings"

// setTheme applies the dark or light theme and the matching icon set.
func (a *App) setTheme(dark bool) error {
	if a.view.Dark == dark {
		return nil
	}

	a.view.Dark = dark

	if err := a.applyStyles(); err != nil {
		return err
	}

	registerImages(a.page, dark)

	return nil
}

// applyStyles sets the active menu's stylesheet over the shared ones.
func (a *App) applyStyles() error {
	return a.page.SetTheme(sectionCSS(a.view.Section) + themeSource(a.view.Dark))
}

// themeSource returns the extra stylesheet for the mode.
func themeSource(dark bool) string {
	if dark {
		return file("components/theme-dark.css")
	}

	return file("components/theme-light.css")
}

// flip toggles one flyout open or closed.
func flip(open, name string) string {
	if open == name {
		return ""
	}

	return name
}

// appLabel returns the name of the demo app in one app- action.
func appLabel(action string) string {
	id := strings.TrimPrefix(action, "app-")

	for _, tile := range appTiles() {
		if tile.ID == id {
			return tile.Name
		}
	}

	return "This app"
}
