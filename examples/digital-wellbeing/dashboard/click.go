package dashboard

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) click(_ context.Context, box ownframe.Box) error {
	switch {
	case box.Action == "usage-access":
		a.mu.Lock()
		a.accessRequest = true
		a.mu.Unlock()
	case box.Action == "choose-app":
		a.view.Selecting = true
		a.setSelecting(true)
	case box.Action == "close-picker":
		a.view.Selecting = false
		a.setSelecting(false)
	case strings.HasPrefix(box.Action, "app-"):
		a.view.Selected = strings.TrimPrefix(box.Action, "app-")
		a.view.Selecting = false
		a.setSelecting(false)
		a.clearUsage()
		a.setSelected(a.view.Selected)
	}
	a.refreshView()
	a.page.SetData(&a.view)
	return nil
}

// TakeAccessRequest lets Android consume a tap that opens system settings.
func (a *App) TakeAccessRequest() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	want := a.accessRequest
	a.accessRequest = false
	return want
}

// SelectedPackage returns the one app whose usage is displayed.
func (a *App) SelectedPackage() string {
	a.selectionMu.RLock()
	defer a.selectionMu.RUnlock()
	return a.selected
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

// RequestBack closes the app picker, or lets Android close the activity.
func (a *App) RequestBack() bool {
	a.selectionMu.RLock()
	selecting := a.selecting
	a.selectionMu.RUnlock()
	if !selecting {
		return false
	}
	a.mu.Lock()
	a.backRequest = true
	a.mu.Unlock()
	return true
}
