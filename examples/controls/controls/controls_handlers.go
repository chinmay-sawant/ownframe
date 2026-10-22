package controls

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// onClick refreshes the status line. A click on a checkbox, a radio, or a
// select changes it before Change runs; a click on a text field only moves
// the focus, so Click refreshes the line too.
func (a *App) onClick(_ context.Context, _ ownframe.Box) error {
	a.refresh()

	return nil
}

// onChange refreshes the status line after a control changed.
func (a *App) onChange(_ context.Context, _ ownframe.Box) error {
	a.refresh()

	return nil
}

func (a *App) refresh() {
	a.view.Status = fmt.Sprintf(
		"focus=%s name=%s agree=%t plan=%s color=%s doc=%s",
		a.page.Edit().FocusedField(),
		a.page.Form().FormValue("name"),
		a.page.Form().FormChecked("agree"),
		a.plan(),
		a.page.Form().FormValue("color"),
		a.page.Form().FormValue("doc"),
	)
	a.page.SetData(a.view)
}

// plan names the checked radio with name=plan, or "" when none is checked.
func (a *App) plan() string {
	if a.page.Form().FormChecked("pro") {
		return "pro"
	}

	if a.page.Form().FormChecked("free") {
		return "free"
	}

	return ""
}
