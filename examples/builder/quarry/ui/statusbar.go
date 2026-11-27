package ui

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

// statusBar is the 24 px bottom bar: connection state, engine version, active
// database, read-only badge, last query latency and rows, and the theme.
func (a *App) statusBar() *ui.Element {
	dot := "qr-dot qr-dot-ok"
	state := "Connected"
	if a.drv == nil {
		dot, state = "qr-dot qr-dot-err", "Disconnected"
	}
	bar := ui.Tag("footer").ID("statusbar").Class("qr-status").Attr("aria-label", "Status bar")
	bar.Children(ui.Div().Class("qr-status-item").Children(ui.Span().ID("st-dot").Class(dot), ui.Span().ID("st-state").Text(state)))
	if a.version != "" {
		bar.Children(ui.Span().ID("st-version").Class("qr-status-item").Text(a.version))
	}
	if a.drv != nil {
		db := a.cfg.Database
		if a.cfg.Engine == driver.EngineSQLite {
			db = shortPath(a.cfg.FilePath)
		}
		bar.Children(ui.Span().ID("st-db").Class("qr-status-item").Text(db))
	}
	if a.cfg.ReadOnly {
		bar.Children(ui.Span().ID("st-readonly").Class("qr-badge qr-badge-warn").Text("READ-ONLY"))
	}
	if a.lastQuery.Set {
		txt := fmt.Sprintf("%s, %d rows", fmtDur(a.lastQuery.Dur), a.lastQuery.Rows)
		if a.lastQuery.Err {
			txt = "error after " + fmtDur(a.lastQuery.Dur)
		}
		bar.Children(ui.Span().ID("st-last").Class("qr-status-item").Text("Last query: " + txt))
	}
	if t := a.activeTabPtr(); t != nil && t.grid != nil {
		bar.Children(ui.Span().ID("st-rows").Class("qr-status-item").Text(fmt.Sprintf("%d rows shown", len(t.grid.rows))))
	}
	bar.Children(ui.Div().Class("qr-spacer"))
	theme := "QuarryDark"
	if !a.dark {
		theme = "QuarryLight"
	}
	bar.Children(ui.Span().ID("st-theme").Class("qr-status-item").Text(theme))
	return bar
}

func shortPath(p string) string {
	if len(p) <= pathShown {
		return p
	}
	return "..." + p[len(p)-37:]
}
