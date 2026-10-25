package main

import (
	"strings"
)

// shellHead opens the document and starts the shared stylesheet.
const shellHead = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Meridian — Trading Terminal</title><style>`

// shellBody opens the app shell after the styles.
const shellBody = `</style></head><body class="screen-{{.Screen}}"><div class="app" id="app-shell">`

// buildHTML assembles the terminal from its component fragments.
func buildHTML() string {
	var b strings.Builder

	b.WriteString(shellHead)

	for _, name := range []string{
		"base.css", "motion.css",
		"topbar.css", "rail.css",
		"watchlist.css", "chart.css", "quote.css", "blotter.css",
		"statusbar.css", "markets.css", "portfolio.css",
		"orders.css", "news.css", "screener.css", "settings.css",
		"layout-wide.css",
	} {
		b.WriteString(file("components/" + name))
	}

	b.WriteString(shellBody)
	b.WriteString(file("components/topbar.html"))
	b.WriteString(file("components/rail.html"))
	b.WriteString(`<main class="workspace">`)

	b.WriteString(`{{if eq .Screen "watchlist"}}`)
	b.WriteString(file("components/watchlist.html"))
	b.WriteString(`<section class="center-column" id="center-column">`)
	b.WriteString(file("components/chart.html"))
	b.WriteString(file("components/blotter.html"))
	b.WriteString(`</section>`)
	b.WriteString(file("components/quote.html"))
	b.WriteString(`{{end}}`)

	for _, screen := range []string{"markets", "portfolio", "orders", "news", "screener", "settings"} {
		b.WriteString(`{{if eq .Screen "` + screen + `"}}`)
		b.WriteString(file("components/" + screen + ".html"))
		b.WriteString(`{{end}}`)
	}

	b.WriteString(`</main>`)
	b.WriteString(file("components/statusbar.html"))
	b.WriteString(`</div>`)
	b.WriteString(toastHTML)
	b.WriteString(`</body></html>`)

	return b.String()
}

// toastHTML is always in the layout so the show class can transition.
const toastHTML = `{{if .Toast.Show}}<div class="toast-layer" id="toast-layer"><div class="toast {{.Toast.Kind}} show" id="toast"><span class="toast-icon"></span><div class="toast-copy"><strong>{{.Toast.Title}}</strong><span>{{.Toast.Message}}</span></div><span class="toast-close" data-action="toast-close">×</span></div></div>{{end}}`
