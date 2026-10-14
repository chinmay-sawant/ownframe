package dashboard

func (a *App) refreshView() {
	switch {
	case !a.view.Access:
		a.view.Screen = "welcome"
	case a.view.Selecting:
		a.view.Screen = "picker"
	case a.view.Screen != "insights" && a.view.Screen != "day" && a.view.Screen != "apps":
		a.view.Screen = "dashboard"
	}
	a.setScreen(a.view.Screen)
}

func (a *App) openPicker() {
	a.returnScreen = a.view.Screen
	a.view.Selecting = true
	a.setSelecting(true)
}

func (a *App) closePicker() {
	a.view.Selecting = false
	a.view.Screen = a.returnScreen
	a.setSelecting(false)
}

func (a *App) changeWeek(delta int) {
	offset := a.WeekOffset() + delta
	if offset < 0 || offset > 52 {
		return
	}
	a.setWeekOffset(offset)
	a.view.WeekOffset = offset
	a.clearUsage()
}

func (a *App) openDay(index int) {
	if index < 0 || index >= len(a.view.Days) {
		return
	}
	a.view.SelectedDay = a.view.Days[index]
	a.view.HasSelectedDay = true
	for i := range a.view.Days {
		a.view.Days[i].Selected = i == index
	}
	a.view.Screen = "day"
}

func (a *App) applyBack() bool {
	if a.view.Selecting {
		a.closePicker()
		return true
	}
	if a.view.Screen == "day" || a.view.Screen == "insights" || a.view.Screen == "apps" {
		a.view.Screen = "dashboard"
		return true
	}
	return false
}
