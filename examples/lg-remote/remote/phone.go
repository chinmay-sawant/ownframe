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

func (a *App) show(panel, hint string) {
	a.view.Panel = panel
	a.view.Hint = hint
	if a.page == nil {
		return
	}

	a.page.Scroll().SetAllowScroll(!a.phone)
	a.page.ScrollTo(0, 0)
}
