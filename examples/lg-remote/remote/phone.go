package remote

import _ "embed"

//go:embed phone.html
var phoneHTML string

//go:embed phone.css
var phoneCSS string

// WithPhone turns on the phone remote, the pad, and Bluetooth.
// Desktop builds leave this off and keep the original page.
func WithPhone(on bool) Option {
	return func(a *App) {
		a.phone = on
		a.view.ShowBluetooth = on
		if on && a.view.Panel == "" {
			a.view.Panel = "remote"
		}
	}
}

const (
	layoutClassic = "classic"
	layoutComfort = "comfort"
)

// openSettings shows the settings panel, or returns to the panel it replaced.
func (a *App) openSettings() {
	if a.view.Panel == "settings" {
		a.closeSettings()
		return
	}
	a.view.Prev = a.view.Panel
	a.show("settings", "Theme and layout.")
}

func (a *App) closeSettings() {
	panel := a.view.Prev
	if panel == "" || (panel == "media" && a.view.Layout != layoutComfort) {
		panel = "remote"
	}
	a.show(panel, "The phone and the TV use the same Wi-Fi.")
}

func (a *App) setLayout(layout string) {
	if a.view.Layout == layout {
		return
	}
	a.view.Layout = layout
	a.view.Status = "Classic layout"
	if layout == layoutComfort {
		a.view.Status = "Comfort layout"
	}
}

func (a *App) show(panel, hint string) {
	a.view.Panel = panel
	a.view.Hint = hint
	if a.page == nil {
		return
	}

	a.page.Scroll().SetAllowScroll(!a.phone)
	a.page.ScrollTo(0, 0)
}
