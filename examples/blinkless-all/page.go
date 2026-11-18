package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// The page is a header, a tab bar, and one section. The title bar, the header with
// the catalog counts and the tab bar are fixed. Only the open section scrolls
// under them. A tab button carries
// data-action="tab-<id>"; the window loads the page routed to that action, and
// that page holds only the open tab's section. The other sections are not in
// the document, so nothing of them is laid out or painted.

const actionPrefix = "tab-"

// pageStyle is shared by every tab page.
const pageStyle = `
body{margin:0;padding:16px;background:#f6f4ee;color:#1f2937;font-family:sans-serif;font-size:14px}
.top{position:fixed;top:0;left:16px;right:16px;z-index:10;background:#f6f4ee;padding:8px 0 0 0;box-sizing:border-box;max-height:40vh;overflow:auto}
.topgap{position:static;visibility:hidden;opacity:0;margin:-16px 0 8px 0;overflow:hidden}
.top .head{padding:6px 12px;margin-bottom:8px}
.top .head p{margin:2px 0;font-size:12px;line-height:1.3}
.top .nums{gap:6px;margin:6px 0 0 0}
.top .num{padding:3px 8px;font-size:12px}
.top .num b{font-size:13px}
.top .tabs{margin-bottom:8px}
.top .tab{padding:6px 10px}
.sel{margin:0 0 6px 0;padding:5px 12px}
.sel{display:flex;justify-content:space-between;align-items:center;margin:0 0 8px 0;padding:6px 12px;background:#1f2937;color:#fff;border-radius:8px}
.sel-title{font-size:16px;font-weight:bold}
.sel-cur{font-size:14px}
.head{background:#fff;border:1px solid #d6d1c4;border-radius:8px;padding:12px 16px;margin-bottom:12px}
.head h1{margin:0 0 6px 0;font-size:20px}
.head p{margin:4px 0;line-height:1.4}
.nums{display:flex;flex-wrap:wrap;gap:8px;margin:8px 0}
.num{background:#f1efe7;border-radius:6px;padding:6px 10px}
.num b{font-size:16px}
.tabs{display:flex;flex-wrap:wrap;gap:6px;margin-bottom:12px}
.tab{padding:8px 12px;border:1px solid #c9c3b3;border-radius:6px;background:#ebe7da;cursor:pointer}
.tab.on{background:#1f2937;color:#fff;border-color:#1f2937}
.sec{background:#fff;border:1px solid #d6d1c4;border-radius:8px;padding:12px 16px}
.sec h2{margin:0 0 4px 0;font-size:17px}
.sec .sub{color:#6b7280;margin:0 0 10px 0}
.row{margin:0 0 14px 0;padding:0 0 12px 0;border-bottom:1px solid #eee9db}
.rowhead{display:flex;flex-wrap:wrap;gap:10px;align-items:flex-start}
.row .name{width:230px;font-family:monospace;font-size:13px}
.row .note{color:#4b5563;width:520px;line-height:1.35}
.demo{margin:8px 0 0 0;padding:10px 12px;background:#faf8f2;border:1px solid #e6e1d2;border-radius:6px}
.st{display:inline-block;padding:1px 7px;border-radius:10px;font-size:12px;margin-left:6px}
.st.implemented{background:#bbf7d0}.st.partial{background:#fde68a}.st.unsupported{background:#fecaca}
`

// pinned is the block that stays at the top of the window while the open
// section scrolls: the dark title bar, the catalog text with the counts, and the
// tab bar. The window paints position:fixed boxes pinned to the viewport and
// does not move position:sticky ones, so the block is fixed. A hidden copy
// without ids or actions stays in the flow and holds the space the fixed block
// covers, so the section starts below it. The block is capped at 40% of the
// window height and scrolls inside that cap, so a short window never has it
// cover the whole page.
func (a *App) pinned(open Tab) string {
	block := a.selected(open) + a.header() + a.tabBar(open.ID)
	gap := idAttr.ReplaceAllString(block, "")

	return `<div id="top" class="top">` + block + `</div><div class="top topgap">` + gap + `</div>`
}

// idAttr matches the id and data-action attributes the hidden copy drops.
var idAttr = regexp.MustCompile(` (id|data-action)="[^"]*"`)

// Header describes the catalog counts and the family idea.
func (a *App) header() string {
	cat := a.cat
	total := cat.Total()
	impl := cat.Count(StatusImplemented)

	var names []string
	for _, t := range a.tabs {
		names = append(names, t.Names()...)
	}

	tally := cat.Tally(names)

	var b strings.Builder

	b.WriteString(`<div class="head">`)
	fmt.Fprintf(&b, `<p>The catalog lists %d CSS properties. %d are implemented (%.1f%%). `+
		`This window covers the %d in six groups that were not yet implemented: `+
		`%d implemented now, %d partial, %d unsupported.</p>`,
		total, impl, 100*float64(impl)/float64(total), len(names),
		tally[StatusImplemented], tally[StatusPartial], tally[StatusUnsupported])
	b.WriteString(`<p>A family is a group of properties that one example covers. ` +
		`Two example boxes can stand for eight properties, so the tab count is the number of ` +
		`properties, not the number of boxes. Each row says which ones it also covers.</p>`)
	b.WriteString(`<div class="nums">`)

	for _, t := range a.tabs {
		tl := cat.Tally(t.Names())
		fmt.Fprintf(&b, `<div class="num"><b>%d</b> %s: %d implemented, %d partial, %d unsupported</div>`,
			t.Count(), html.EscapeString(t.Title), tl[StatusImplemented], tl[StatusPartial], tl[StatusUnsupported])
	}

	b.WriteString(`</div></div>`)

	return b.String()
}

// selected is the slim bar above the header: the page title on the left and
// the open tab on the right.
func (a *App) selected(open Tab) string {
	return `<div id="selected" class="sel"><span class="sel-title">blinkless: properties still to validate</span>` +
		`<span class="sel-cur">Selected: ` + html.EscapeString(open.Title) + fmt.Sprintf(` (%d)`, open.Count()) + `</span></div>`
}

// tabBar draws the buttons. The open tab is dark. Each label shows its count.
func (a *App) tabBar(open string) string {
	var b strings.Builder

	b.WriteString(`<div class="tabs">`)

	for _, t := range a.tabs {
		class := "tab"
		if t.ID == open {
			class += " on"
		}

		fmt.Fprintf(&b, `<div id="tab-btn-%s" class="%s" data-action="%s%s">%s (%d)</div>`,
			t.ID, class, actionPrefix, t.ID, html.EscapeString(t.Title), t.Count())
	}

	b.WriteString(`</div>`)

	return b.String()
}

// section wraps the open tab's body with its title and one row per example.
func (a *App) section(t Tab) string {
	var b strings.Builder

	fmt.Fprintf(&b, `<div class="sec" id="sec-%s"><h2>%s</h2><p class="sub">%d properties, %d examples</p>`,
		t.ID, html.EscapeString(t.Title), t.Count(), len(t.Props))

	if t.Head != nil {
		b.WriteString(t.Head())
	}

	for _, p := range t.Props {
		st, _ := a.cat.Status(p.Name)
		covers := ""

		if len(p.Covers) > 0 {
			covers = " also covers " + html.EscapeString(strings.Join(p.Covers, ", "))
		}

		fmt.Fprintf(&b, `<div class="row" id="row-%s"><div class="rowhead"><div class="name">%s<span class="st %s">%s</span></div>`+
			`<div class="note">%s%s</div></div>`,
			html.EscapeString(p.Name), html.EscapeString(p.Name), strings.ReplaceAll(st, " ", "-"), html.EscapeString(st),
			html.EscapeString(p.Note), covers)

		if p.Demo != nil {
			b.WriteString(`<div class="demo" id="demo-` + html.EscapeString(p.Name) + `">` + p.Demo() + `</div>`)
		}

		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	return b.String()
}

// Page returns the whole document with tab id open.
func (a *App) Page(id string) string {
	var open Tab

	for _, t := range a.tabs {
		if t.ID == id {
			open = t
		}
	}

	return `<!DOCTYPE html><html><head><style>` + pageStyle + `</style></head><body>` +
		a.pinned(open) + a.section(open) + `</body></html>`
}
