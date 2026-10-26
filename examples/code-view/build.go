package main

import "strings"

// shellHead opens the document and starts the shared stylesheet.
const shellHead = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Code Review</title><style>`

// shellBody opens the app shell after the styles.
const shellBody = `</style></head><body class="mode-{{.Mode}}{{if .Details}} with-details{{end}}"><div class="app" id="app-shell">`

// buildHTML assembles the code view from its component fragments.
func buildHTML() string {
	var b strings.Builder

	b.WriteString(shellHead)

	for _, name := range []string{
		"base.css", "chrome.css", "sidebar.css",
		"editor.css", "diff.css", "history.css", "details.css", "palette.css", "preview.css", "motion.css",
	} {
		b.WriteString(file("components/" + name))
	}

	b.WriteString(shellBody)
	b.WriteString(file("components/titlebar.html"))
	b.WriteString(file("components/activity.html"))
	b.WriteString(file("components/sidebar.html"))
	b.WriteString(file("components/editor.html"))
	b.WriteString(file("components/details.html"))
	b.WriteString(file("components/statusbar.html"))
	b.WriteString(file("components/palette.html"))
	b.WriteString(`</div></body></html>`)

	return b.String()
}
