package ui

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Icons are inline SVG documents registered with Page.SetImage and drawn with
// <img>, because the builder has no <svg> tag and an <img> cannot inherit
// currentColor. Each icon is therefore registered once per tone and theme
// under a source name that encodes both.

// iconTone names the colour role of an icon; the hex values follow the
// --qr-* tokens of each theme.
type iconTone string

const (
	toneText    iconTone = "text"
	toneMuted   iconTone = "muted"
	toneBrand   iconTone = "brand"
	toneOnBrand iconTone = "onbrand"
	toneOK      iconTone = "ok"
	toneWarn    iconTone = "warn"
	toneErr     iconTone = "err"
)

func toneHex(t iconTone, dark bool) string {
	type pair struct{ dark, light string }
	m := map[iconTone]pair{
		toneText:    {"#ededed", "#171717"},
		toneMuted:   {"#8f8f8f", "#6b6b6b"},
		toneBrand:   {"#3ecf8e", "#1f9d6b"},
		toneOnBrand: {"#0b1f15", "#ffffff"},
		toneOK:      {"#3ecf8e", "#1a9a68"},
		toneWarn:    {"#e5b340", "#b7791f"},
		toneErr:     {"#e5645f", "#d1453b"},
	}
	p, ok := m[t]
	if !ok {
		p = m[toneText]
	}
	if dark {
		return p.dark
	}
	return p.light
}

// iconBody returns the SVG shapes (24x24 grid, stroke only) for an icon name.
func iconBody(name string) string {
	line := func(x1, y1, x2, y2 string) string {
		return fmt.Sprintf(`<line x1="%s" y1="%s" x2="%s" y2="%s"/>`, x1, y1, x2, y2)
	}
	poly := func(pts string) string { return `<polyline points="` + pts + `"/>` }
	path := func(d string) string { return `<path d="` + d + `"/>` }
	switch name {
	case "database", "connections":
		return `<ellipse cx="12" cy="5" rx="9" ry="3"/>` + path("M3 5V19A9 3 0 0 0 21 19V5") + path("M3 12A9 3 0 0 0 21 12")
	case "layers", "explorer":
		return poly("12 2 2 7 12 12 22 7 12 2") + poly("2 17 12 22 22 17") + poly("2 12 12 17 22 12")
	case "activity", "monitor":
		return poly("22 12 18 12 15 21 9 3 6 12 2 12")
	case "clock", "history":
		return `<circle cx="12" cy="12" r="10"/>` + poly("12 6 12 12 16 14")
	case "settings":
		return path("M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z") +
			`<circle cx="12" cy="12" r="3"/>`
	case "sun":
		return `<circle cx="12" cy="12" r="4"/>` + line("12", "2", "12", "4") + line("12", "20", "12", "22") +
			line("4.9", "4.9", "6.3", "6.3") + line("17.7", "17.7", "19.1", "19.1") + line("2", "12", "4", "12") + line("20", "12", "22", "12") +
			line("4.9", "19.1", "6.3", "17.7") + line("17.7", "6.3", "19.1", "4.9")
	case "moon":
		return path("M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z")
	case "refresh":
		return path("M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8") + path("M21 3v5h-5") +
			path("M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16") + path("M8 16H3v5")
	case "chevron-right":
		return poly("9 18 15 12 9 6")
	case "chevron-left":
		return poly("15 18 9 12 15 6")
	case "chevron-up":
		return poly("18 15 12 9 6 15")
	case "chevron-down":
		return poly("6 9 12 15 18 9")
	case "chevrons-left":
		return poly("11 17 6 12 11 7") + poly("18 17 13 12 18 7")
	case "chevrons-right":
		return poly("13 17 18 12 13 7") + poly("6 17 11 12 6 7")
	case "x", "close":
		return line("18", "6", "6", "18") + line("6", "6", "18", "18")
	case "plus":
		return line("12", "5", "12", "19") + line("5", "12", "19", "12")
	case "minus", "minimize":
		return line("5", "12", "19", "12")
	case "maximize":
		return `<rect x="5" y="5" width="14" height="14" rx="1.5"/>`
	case "restore":
		return `<rect x="4" y="8" width="12" height="12" rx="1.5"/>` + path("M8 8V5.5A1.5 1.5 0 0 1 9.5 4H18.5A1.5 1.5 0 0 1 20 5.5V14.5A1.5 1.5 0 0 1 18.5 16H16")
	case "play":
		return poly("6 3 20 12 6 21 6 3")
	case "square":
		return `<rect x="5" y="5" width="14" height="14" rx="2"/>`
	case "search":
		return `<circle cx="11" cy="11" r="7"/>` + line("21", "21", "16.65", "16.65")
	case "zap", "trigger":
		return poly("13 2 3 14 12 14 11 22 21 10 12 10 13 2")
	case "align-left", "format":
		return line("21", "6", "3", "6") + line("15", "12", "3", "12") + line("17", "18", "3", "18")
	case "save":
		return path("M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z") + poly("17 21 17 13 7 13 7 21") + poly("7 3 7 8 15 8")
	case "download":
		return path("M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4") + poly("7 10 12 15 17 10") + line("12", "15", "12", "3")
	case "upload":
		return path("M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4") + poly("17 8 12 3 7 8") + line("12", "3", "12", "15")
	case "table":
		return `<rect x="3" y="3" width="18" height="18" rx="2"/>` + line("3", "9", "21", "9") + line("3", "15", "21", "15") + line("9", "3", "9", "21")
	case "eye", "view":
		return path("M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z") + `<circle cx="12" cy="12" r="3"/>`
	case "hash", "index":
		return line("4", "9", "20", "9") + line("4", "15", "20", "15") + line("10", "3", "8", "21") + line("16", "3", "14", "21")
	case "code", "function":
		return poly("16 18 22 12 16 6") + poly("8 6 2 12 8 18")
	case "folder", "schema":
		return path("M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z")
	case "lock":
		return `<rect x="3" y="11" width="18" height="11" rx="2"/>` + path("M7 11V7a5 5 0 0 1 10 0v4")
	case "panel-right":
		return `<rect x="3" y="3" width="18" height="18" rx="2"/>` + line("15", "3", "15", "21")
	case "check":
		return poly("20 6 9 17 4 12")
	case "pause":
		return line("8", "5", "8", "19") + line("16", "5", "16", "19")
	case "copy":
		return `<rect x="9" y="9" width="12" height="12" rx="2"/>` + path("M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1")
	case "file-code", "query":
		return path("M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z") + poly("14 2 14 8 20 8") + poly("10 13 8 15 10 17") + poly("14 13 16 15 14 17")
	case "list":
		return line("8", "6", "21", "6") + line("8", "12", "21", "12") + line("8", "18", "21", "18") + line("3", "6", "3.01", "6") + line("3", "12", "3.01", "12") + line("3", "18", "3.01", "18")
	case "trash":
		return poly("3 6 5 6 21 6") + path("M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6") + path("M10 11v6M14 11v6M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2")
	case "filter":
		return poly("22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3")
	case "star":
		return poly("12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2")
	}
	return `<circle cx="12" cy="12" r="4"/>`
}

// iconSVG renders the stroke icon in colour hex at 2x resolution for crisp downscaling.
func iconSVG(name, hex string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">`, hex)
	b.WriteString(iconBody(name))
	b.WriteString(`</svg>`)
	return []byte(b.String())
}

// logoSVG is the Quarry mark: a rounded tile holding stepped quarry terraces.
func logoSVG(dark bool) []byte {
	bg, fg := "#3ecf8e", "#0b1f15"
	if !dark {
		bg, fg = "#1f9d6b", "#ffffff"
	}
	return []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="48" height="48">` +
		`<rect x="0" y="0" width="24" height="24" rx="6" fill="` + bg + `"/>` +
		`<path d="M5 8H11V12H15V16H19V19H5Z" fill="` + fg + `"/>` +
		`<rect x="5" y="5" width="6" height="1.8" rx="0.9" fill="` + fg + `"/></svg>`)
}

// icon returns an <img> showing the named icon at size px.
func (a *App) icon(name string, tone iconTone, size int) *ui.Element {
	src := a.registerImage(fmt.Sprintf("qi-%s-%s", name, tone), func(dark bool) []byte {
		return iconSVG(name, toneHex(tone, dark))
	})
	return ui.Tag("img").Class("qr-icon").Attr("src", src).Attr("alt", "").Attr("aria-hidden", "true").
		Width(ui.Px(float64(size))).Height(ui.Px(float64(size)))
}

// logo returns the Quarry mark at size px.
func (a *App) logo(size int) *ui.Element {
	src := a.registerImage("qi-logo", logoSVG)
	return ui.Tag("img").Class("qr-logo").Attr("src", src).Attr("alt", "Quarry").
		Width(ui.Px(float64(size))).Height(ui.Px(float64(size)))
}

// registerImage queues an image for the current theme and returns its source
// name. flushImages hands queued images to the page.
func (a *App) registerImage(base string, build func(dark bool) []byte) string {
	src := base + "-l"
	if a.dark {
		src = base + "-d"
	}
	if _, ok := a.imageSent[src]; !ok {
		if a.imagePending == nil {
			a.imagePending = map[string][]byte{}
		}
		a.imagePending[src] = build(a.dark)
		if a.imageSent == nil {
			a.imageSent = map[string]bool{}
		}
		a.imageSent[src] = false
	}
	return src
}

func (a *App) flushImages() {
	if a.page == nil {
		return
	}
	for src, data := range a.imagePending {
		a.page.SetImage(src, data)
		a.imageSent[src] = true
	}
	for k := range a.imagePending {
		delete(a.imagePending, k)
	}
}

// ariaRole maps the ARIA role a widget really has to the nearest role the
// engine's accessibility tree supports (main, navigation, group, heading,
// paragraph, img, link, button, textbox, checkbox, radio, combobox, list,
// listitem). Grid, row, cell, tab, tree, listbox, dialog, menu and status are
// not in that set and are silently ignored if written as is.
func ariaRole(want string) string {
	switch want {
	case "grid", "table", "tree", "tablist", "listbox", "menu":
		return "list"
	case "row", "option", "treeitem", "menuitem":
		return "listitem"
	case "tab":
		return "button"
	}
	return "group" // cell, columnheader, gridcell, dialog, alert, status, complementary
}
