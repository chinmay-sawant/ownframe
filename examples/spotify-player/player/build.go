package player

import "strings"

// shellHead opens the document and starts the shared stylesheet.
const shellHead = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Spotify — Player</title><style>`

// shellBody opens the layout after the styles.
const shellBody = `</style></head><body><div class="frame"><div class="page">`

// shellFoot closes the layout and the document.
const shellFoot = `</div></div></body></html>`

// styles are concatenated in this order.
var styles = []string{
	"base.css", "sidebar.css", "topbar.css", "greeting.css",
	"shelf.css", "tracklist.css", "profile.css", "nowbar.css",
	"search.css", "library.css", "liked.css", "browse.css",
	"radio.css", "queue.css", "motion.css",
}

// screens are the nav values that swap the main panel for one component.
var screens = []string{"search", "library", "liked", "browse", "radio", "queue"}

// buildHTML assembles the page from the component fragments.
func buildHTML() string {
	var b strings.Builder

	b.WriteString(shellHead)

	for _, name := range styles {
		b.WriteString(file("components/" + name))
	}

	b.WriteString(shellBody)
	b.WriteString(file("components/sidebar.html"))
	b.WriteString(`<main class="main panel">`)
	b.WriteString(file("components/topbar.html"))
	b.WriteString(`{{if eq .Nav "home"}}`)
	b.WriteString(file("components/greeting.html"))
	b.WriteString(file("components/shelf.html"))
	b.WriteString(file("components/tracklist.html"))
	b.WriteString(`{{end}}`)
	b.WriteString(`{{if eq .Nav "profile"}}`)
	b.WriteString(file("components/profile.html"))
	b.WriteString(`{{end}}`)

	for _, name := range screens {
		b.WriteString(`{{if eq .Nav "` + name + `"}}`)
		b.WriteString(file("components/" + name + ".html"))
		b.WriteString(`{{end}}`)
	}

	b.WriteString(`</main>`)
	b.WriteString(file("components/nowbar.html"))
	b.WriteString(shellFoot)

	return b.String()
}
