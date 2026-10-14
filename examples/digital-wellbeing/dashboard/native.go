package dashboard

// TakeAccessRequest lets Android consume a tap that opens system settings.
func (a *App) TakeAccessRequest() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	want := a.accessRequest
	a.accessRequest = false
	return want
}

// SelectedPackage returns the package currently shown in the chart.
func (a *App) SelectedPackage() string {
	a.selectionMu.RLock()
	defer a.selectionMu.RUnlock()
	return a.selected
}

// WeekOffset returns how many seven-day ranges back the page is showing.
func (a *App) WeekOffset() int {
	a.selectionMu.RLock()
	defer a.selectionMu.RUnlock()
	return a.weekOffset
}

func (a *App) setSelected(pkg string) {
	a.selectionMu.Lock()
	a.selected = pkg
	a.selectionMu.Unlock()
}

func (a *App) setSelecting(selecting bool) {
	a.selectionMu.Lock()
	a.selecting = selecting
	a.selectionMu.Unlock()
}

func (a *App) setWeekOffset(offset int) {
	a.selectionMu.Lock()
	a.weekOffset = offset
	a.selectionMu.Unlock()
}

func (a *App) setScreen(screen string) {
	a.selectionMu.Lock()
	a.currentScreen = screen
	a.selectionMu.Unlock()
}

// RequestBack handles in-app history before allowing Android to close Mochi.
func (a *App) RequestBack() bool {
	a.selectionMu.RLock()
	selecting, screen := a.selecting, a.currentScreen
	a.selectionMu.RUnlock()
	if !selecting && screen != "day" && screen != "insights" && screen != "apps" {
		return false
	}
	a.mu.Lock()
	a.backRequest = true
	a.mu.Unlock()
	return true
}
