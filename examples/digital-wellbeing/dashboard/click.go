package dashboard

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) click(_ context.Context, box ownframe.Box) error {
	switch action := box.Action; {
	case action == "usage-access":
		a.mu.Lock()
		a.accessRequest = true
		a.mu.Unlock()
	case action == "choose-app":
		a.openPicker()
	case action == "close-picker":
		a.closePicker()
	case action == "back-dashboard":
		a.applyBack()
	case action == "week-previous":
		a.changeWeek(1)
	case action == "week-next":
		a.changeWeek(-1)
	case action == "week-insights":
		a.view.Screen = "insights"
	case action == "all-apps":
		a.view.Screen = "apps"
	case strings.HasPrefix(action, "app-"):
		a.selectApp(strings.TrimPrefix(action, "app-"))
	case strings.HasPrefix(action, "day-"):
		index, _ := strconv.Atoi(strings.TrimPrefix(action, "day-"))
		a.openDay(index)
	}
	a.refreshView()
	a.page.SetData(&a.view)
	return nil
}

func (a *App) selectApp(pkg string) {
	a.view.Selected = pkg
	a.view.Selecting = false
	a.view.Screen = "dashboard"
	a.setSelecting(false)
	a.clearUsage()
	a.setSelected(pkg)
}
