package remote

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	a.host = strings.TrimSpace(a.page.FormValue("host"))
	if box.ID != "" && box.ID != "host" {
		a.flash(box.ID)
	}

	switch box.ID {
	case "theme":
		a.toggleTheme()
	case "mode":
		a.toggleMode()
	case "connect":
		a.connect()
	case "scan":
		a.view.Status = "Searching this Wi-Fi"
		a.later(a.doFind)
	case "tab-remote":
		a.show("remote", "Volume on the left, channels on the right.")
	case "tab-pad":
		a.show("pad", "Drag the pad to move the pointer. Tap it to click.")
	case "tab-nums":
		a.show("nums", "Numbers and the keys that are not on the main remote.")
	case "":
	default:
		a.press(box.ID)
	}

	a.setData()

	return nil
}

func (a *App) onKey(ctx context.Context, key string) error {
	if a.page.FocusID() == "sensitivity" {
		return a.sensitivityKey(ctx, key)
	}
	if a.page.FocusID() == "host" {
		if key == "enter" {
			a.host = strings.TrimSpace(a.page.FormValue("host"))
			a.connect()
			a.setData()
			return a.page.Redraw(ctx)
		}
		return nil
	}
	if id := a.page.FocusID(); id != "" && (key == "enter" || key == "space") {
		for _, box := range a.page.Boxes() {
			if box.ID == id {
				return a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2)
			}
		}
	}

	id := map[string]string{
		"arrowup": "up", "arrowdown": "down",
		"arrowleft": "left", "arrowright": "right",
		"enter": "ok", "escape": "back",
	}[key]
	if id == "" {
		return nil
	}

	a.host = strings.TrimSpace(a.page.FormValue("host"))
	a.press(id)
	a.setData()

	return a.page.Redraw(ctx)
}

func (a *App) toggleTheme() {
	a.dark = !a.dark
	css := lightTheme
	a.view.ThemeLabel = "Dark"
	a.view.Status = "Light theme"
	if a.dark {
		css = darkTheme
		a.view.ThemeLabel = "Light"
		a.view.Status = "Dark theme"
	}

	if err := a.page.SetTheme(css); err != nil {
		a.view.Status = err.Error()
	}
}
