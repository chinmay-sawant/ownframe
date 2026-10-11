package app

func (a *App) DarkTheme() bool { return a.dark.Load() }
