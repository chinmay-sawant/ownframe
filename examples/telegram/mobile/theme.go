package mobile

// DarkTheme reports whether the page shows the dark theme.
func DarkTheme() bool {
	if app == nil {
		return false
	}

	return app.DarkTheme()
}
