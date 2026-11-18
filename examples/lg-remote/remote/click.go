package remote

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	a.host = strings.TrimSpace(a.page.Form().FormValue("host"))
	if box.ID != "" && box.ID != "host" && box.ID != typingField {
		a.flash(box.ID)
	}

	switch box.ID {
	case "theme":
		a.toggleTheme()
	case "settings", "settings-done":
		a.openSettings()
	case "set-dark":
		a.setDark(true)
	case "set-light":
		a.setDark(false)
	case "accent-pink", "accent-red", "accent-orange", "accent-green", "accent-blue", "accent-purple":
		a.setAccent(strings.TrimPrefix(box.ID, "accent-"))
	case "layout-classic":
		a.setLayout(layoutClassic)
	case "layout-comfort":
		a.setLayout(layoutComfort)
	case "tab-keys":
		a.openKeys(ctx)
	case "kb-back":
		a.press("back")
	case "kb-delete", "kb-enter", "kb-clear":
		a.keysPress(box.ID)
	case "tab-media":
		a.show("media", "Playback, guide, colour keys, and apps.")
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
	case typingField:
		// A touch tap may not focus the field, so ask again next frame.
		a.focusNext = typingField
	default:
		a.press(box.ID)
	}

	a.setData()

	return nil
}

func (a *App) onKey(ctx context.Context, key string) error {
	if a.page.Form().FocusID() == "sensitivity" {
		return a.sensitivityKey(ctx, key)
	}
	if a.page.Form().FocusID() == "host" {
		if key == "enter" {
			a.host = strings.TrimSpace(a.page.Form().FormValue("host"))
			a.connect()
			a.setData()
			return a.page.Redraw(ctx)
		}
		return nil
	}
	if id := a.page.Form().FocusID(); id != "" && (key == "enter" || key == "space") {
		for _, box := range a.page.Boxes() {
			if box.ID == id {
				return a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2)
			}
		}
	}

	if a.view.Panel == "settings" && key == "escape" {
		a.closeSettings()
		a.setData()

		return a.page.Redraw(ctx)
	}

	id := map[string]string{
		"arrowup": "up", "arrowdown": "down",
		"arrowleft": "left", "arrowright": "right",
		"enter": "ok", "escape": "back",
	}[key]
	if id == "" {
		return nil
	}

	a.host = strings.TrimSpace(a.page.Form().FormValue("host"))
	a.press(id)
	a.setData()

	return a.page.Redraw(ctx)
}

func (a *App) setDark(dark bool) {
	if a.dark != dark {
		a.toggleTheme()
	}
}

func (a *App) toggleTheme() {
	a.dark = !a.dark
	css := a.themeCSS()
	a.view.ThemeLabel = "Dark"
	a.view.Status = "Light theme"
	if a.dark {
		a.view.ThemeLabel = "Light"
		a.view.Status = "Dark theme"
	}

	if err := a.page.SetTheme(css); err != nil {
		a.view.Status = err.Error()
	}
}

func (a *App) setAccent(id string) {
	pick := findAccent(id)
	a.view.Accent = pick.ID
	a.view.Status = pick.Label + " accent"
	if err := a.page.SetTheme(a.themeCSS()); err != nil {
		a.view.Status = err.Error()
	}
}
