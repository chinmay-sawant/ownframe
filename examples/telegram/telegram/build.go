package telegram

import "strings"

// shellHead opens the document and starts the shared stylesheet.
const shellHead = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Telegram</title><style>`

// shellBody opens the app after the styles. The inline padding keeps the
// page clear of the phone's system bars; InsetTop and InsetBottom arrive
// from the activity.
const shellBody = `</style></head><body><div class="app{{if .Active}} app-thread{{else if .Phone}} app-phone{{end}}" style="padding-top:{{.InsetTop}}px;padding-bottom:{{if and .Phone (not .Active)}}calc({{.InsetBottom}}px + 60px){{else}}{{.InsetBottom}}px{{end}}">`

// shellFoot closes the app and the document.
const shellFoot = `</div></body></html>`

// styles are concatenated in this order.
var styles = []string{"base.css", "list.css", "thread.css", "settings.css"}

// buildHTML assembles the page from the component fragments.
func buildHTML() string {
	var b strings.Builder

	b.WriteString(shellHead)

	for _, name := range styles {
		b.WriteString(file("components/" + name))
	}

	b.WriteString(shellBody)
	b.WriteString(`{{if .Active}}`)
	b.WriteString(file("components/thread.html"))
	b.WriteString(`{{else}}`)
	b.WriteString(`{{if eq .Tab "contacts"}}`)
	b.WriteString(file("components/contacts.html"))
	b.WriteString(`{{else if eq .Tab "settings"}}`)
	b.WriteString(file("components/settings.html"))
	b.WriteString(`{{else}}`)
	b.WriteString(file("components/chats.html"))
	b.WriteString(`{{end}}`)
	b.WriteString(file("components/tabs.html"))
	b.WriteString(`{{end}}`)
	b.WriteString(shellFoot)

	return b.String()
}
