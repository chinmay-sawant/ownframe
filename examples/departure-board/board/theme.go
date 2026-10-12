package board

import (
	"fmt"
	"strings"
)

// sheet renders the eight custom properties the template reads from one
// colors row, in the order: bg, tile-bg, fg, tile-fg, accent, border,
// muted, radius.
func sheet(colors string) string {
	v := strings.Fields(colors)

	return ":root{--bg:" + v[0] + ";--tile-bg:" + v[1] + ";--fg:" + v[2] +
		";--tile-fg:" + v[3] + ";--accent:" + v[4] + ";--border:" + v[5] +
		";--muted:" + v[6] + ";--radius:" + v[7] + "}"
}

// livery builds one skin from its day and night color rows.
func livery(name, day, night string, accent [3]float64) Livery {
	return Livery{Name: name, CSS: sheet(day), Night: sheet(night), Accent: accent}
}

// applySheet stores the sheet of the current livery, dimmed when night is
// on. It does not redraw: Click draws after its handler and Tick redraws
// after a night flip.
func (a *App) applySheet() error {
	if a.livery.Name == "" {
		l, _ := liveryByName(DefaultLivery)
		a.livery = l
	}

	css := a.livery.CSS
	if a.night {
		css = a.livery.Night
	}

	return a.page.SetTheme(css)
}

// SetLivery switches the skin and remembers it in the view.
func (a *App) SetLivery(name string) error {
	l, ok := liveryByName(name)
	if !ok {
		return fmt.Errorf("unknown livery %q", name)
	}

	a.livery = l
	a.view.Livery = name

	return a.applySheet()
}

// SetNight turns the dim night sheets on or off and re-applies the sheet.
func (a *App) SetNight(on bool) error {
	a.night = on
	a.view.Night = on

	return a.applySheet()
}
