package form

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.Action != "send" {
		return nil
	}

	a.view.Status = fmt.Sprintf(
		"email=%s remember=%t plan=%s color=%s file=%s",
		a.page.Form().FormValue("email"),
		a.page.Form().FormChecked("remember"),
		a.plan(),
		a.page.Form().FormValue("color"),
		a.page.Form().FormValue("file"),
	)
	a.page.SetData(a.view)

	return nil
}

func (a *App) plan() string {
	if a.page.Form().FormChecked("pro") {
		return a.page.Form().FormValue("pro")
	}

	if a.page.Form().FormChecked("free") {
		return a.page.Form().FormValue("free")
	}

	return ""
}
