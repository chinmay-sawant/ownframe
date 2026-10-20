package app

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the control under the click: the app rail, the flyouts, then
// the active menu's own actions.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	a.view.Note = ""
	section := a.view.Section

	switch {
	case railSections[box.Action] != "":
		a.view.Section = railSections[box.Action]
		a.view.Flyout = ""
	case box.Action == "rail-more":
		a.view.Flyout = flip(a.view.Flyout, "more")
	case box.Action == "me":
		a.view.Flyout = flip(a.view.Flyout, "profile")
	case box.Action == "flyout-close":
		a.view.Flyout = ""
	case strings.HasPrefix(box.Action, "profile-status-"):
		a.view.Presence = strings.TrimPrefix(box.Action, "profile-status-")
	case box.Action == "profile-theme-dark":
		if err := a.setTheme(true); err != nil {
			return err
		}
	case box.Action == "profile-theme-light":
		if err := a.setTheme(false); err != nil {
			return err
		}
	case box.Action == "profile-signout":
		a.view.Flyout = ""
		a.view.Note = "Signed out of the demo"
	case strings.HasPrefix(box.Action, "app-"):
		a.view.Note = appLabel(box.Action) + " is a demo tile"
	default:
		a.dispatch(ctx, box.Action)
	}

	if section != a.view.Section {
		if err := a.applyStyles(); err != nil {
			return err
		}
	}

	a.page.SetData(a.view)
	a.save()

	return nil
}
