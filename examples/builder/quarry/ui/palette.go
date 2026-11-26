package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

type paletteState struct {
	query  string
	active int
	top    int // first visible row
}

type palItem struct {
	id    string
	label string
	hint  string
	kind  string
	run   func(ctx context.Context) error
	score int
}

// fuzzyScore scores query as an in-order subsequence of text; -1 means no match.
// Consecutive and word-start hits score higher.
func fuzzyScore(query, text string) int {
	if query == "" {
		return 0
	}
	q, t := []rune(strings.ToLower(query)), []rune(strings.ToLower(text))
	score, qi, prev := 0, 0, noPrevMatch
	for i, r := range t {
		if qi < len(q) && r == q[qi] {
			s := 1
			if i == prev+1 {
				s += 4
			}
			if i == 0 || !unicode.IsLetter(t[i-1]) && !unicode.IsDigit(t[i-1]) {
				s += 3
			}
			score += s
			prev = i
			qi++
		}
	}
	if qi < len(q) {
		return -1
	}
	return score*10 - len(t)
}

// paletteItems builds every command, then filters and ranks by the query.
func (a *App) paletteItems() []palItem {
	var all []palItem
	act := func(id, label, hint string, run func(ctx context.Context) error) {
		all = append(all, palItem{id: id, label: label, hint: hint, kind: "Action", run: run})
	}
	act("theme", "Toggle Dark/Light Theme", "Ctrl+Shift+T", func(ctx context.Context) error { return a.toggleTheme(ctx) })
	act("monitor", "Open Live Monitor", "", func(ctx context.Context) error {
		a.openMonitorTab()
		a.startMonitor()
		return a.redraw(ctx)
	})
	act("newquery", "New SQL Query", "Ctrl+N", func(ctx context.Context) error { a.newQueryTab(""); return a.redraw(ctx) })
	act("export", "Export Current Results", "", func(ctx context.Context) error { return a.openExport(ctx, false) })
	act("readonly", "Toggle Read-Only Mode", "", func(ctx context.Context) error { return a.toggleReadOnly(ctx) })
	act("connections", "Open Connection Manager", "", a.openConnManager)
	act("history", "Open Query History", "", a.openHistory)
	for _, c := range a.savedConnections() {
		c := c
		all = append(all, palItem{id: "conn-" + c.ID, label: "Connect: " + c.Name, hint: c.EnvTag, kind: "Connection",
			run: func(ctx context.Context) error {
				if err := a.connect(ctx, c.ConnectionConfig); err != nil {
					a.notify(err.Error())
				}
				return a.redraw(ctx)
			}})
	}
	for _, s := range a.schema {
		for _, g := range s.Children {
			for _, n := range g.Children {
				n := n
				if n.Kind != driver.NodeTable && n.Kind != driver.NodeView && n.Kind != driver.NodeMatView {
					continue
				}
				all = append(all, palItem{id: "tbl-" + nodeKey(n), label: n.QualifiedName(), hint: string(n.Kind), kind: "Table",
					run: func(ctx context.Context) error { return a.openTable(ctx, n) }})
			}
		}
	}
	if a.prefs != nil {
		snips, _ := a.prefs.ListSnippets(a.ctx)
		for _, sn := range snips {
			sn := sn
			all = append(all, palItem{id: fmt.Sprintf("snip-%d", sn.ID), label: "Snippet: " + sn.Name, kind: "Snippet",
				run: func(ctx context.Context) error { return a.loadSQL(ctx, sn.SQL) }})
		}
	}
	if a.pal.query == "" {
		return all
	}
	var out []palItem
	for _, it := range all {
		if s := fuzzyScore(a.pal.query, it.label); s >= 0 {
			it.score = s
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	return out
}

func (a *App) toggleReadOnly(ctx context.Context) error {
	if a.cfg.EnvTag == driver.EnvProd && a.cfg.ReadOnly {
		return a.askUnlockProd(ctx)
	}
	if err := a.reopenWritable(ctx, a.cfg.ReadOnly); err != nil {
		a.notify(err.Error())
	}
	return a.redraw(ctx)
}

func (a *App) openPalette(ctx context.Context) error {
	a.modal = "palette"
	a.pal = paletteState{}
	a.page.Form().RequestFocus("pal-input")
	return a.redraw(ctx)
}

func (a *App) paletteKey(ctx context.Context, key string) error {
	items := a.paletteItems()
	switch key {
	case "arrowdown":
		if len(items) > 0 {
			a.pal.active = (a.pal.active + 1) % len(items)
		}
	case "arrowup":
		if len(items) > 0 {
			a.pal.active = (a.pal.active - 1 + len(items)) % len(items)
		}
	case "pagedown":
		a.pal.active = min(len(items)-1, a.pal.active+paletteRowsShown)
	case "pageup":
		a.pal.active = max(0, a.pal.active-paletteRowsShown)
	case "enter":
		return a.paletteRun(ctx, a.pal.active)
	default:
		return nil
	}
	a.paletteReveal()
	return a.redraw(ctx)
}

func (a *App) paletteRun(ctx context.Context, i int) error {
	items := a.paletteItems()
	if i < 0 || i >= len(items) {
		return nil
	}
	a.modal = ""
	_ = a.page.Form().Focus(ctx, "")
	return items[i].run(ctx)
}

// paletteScroll moves the visible window by rows (wheel input).
func (a *App) paletteScroll(rows int) {
	n := len(a.paletteItems())
	a.pal.top = max(0, min(a.pal.top+rows, max(0, n-paletteRowsShown)))
	a.pal.active = max(a.pal.top, min(a.pal.active, a.pal.top+paletteRowsShown-1))
}

// paletteReveal scrolls the window just far enough to show the active row.
func (a *App) paletteReveal() {
	if a.pal.active < a.pal.top {
		a.pal.top = a.pal.active
	}
	if a.pal.active >= a.pal.top+paletteRowsShown {
		a.pal.top = a.pal.active - paletteRowsShown + 1
	}
}

func (a *App) paletteView() *ui.Element {
	items := a.paletteItems()
	a.pal.active = max(0, min(a.pal.active, len(items)-1))
	a.pal.top = max(0, min(a.pal.top, max(0, len(items)-paletteRowsShown)))
	end := min(len(items), a.pal.top+paletteRowsShown)
	rows := max(1, end-a.pal.top)
	rowsCol := ui.Div().Class("qr-pal-rows")
	list := ui.Div().ID("pal-list").Class("qr-palette-list").Attr("role", ariaRole("listbox")).Attr("aria-label", "Commands").
		Height(ui.Px(float64(rows * paletteRowH)))
	for i := a.pal.top; i < end; i++ {
		it := items[i]
		i := i
		cls := "qr-list-row"
		if i == a.pal.active {
			cls += " qr-list-row-active"
		}
		rowsCol.Children(ui.Div().ID(fmt.Sprintf("pal-%d", i)).Class(cls).Attr("role", ariaRole("option")).Children(
			ui.Span().Class("qr-badge").Text(it.kind),
			ui.Span().Class("qr-tree-label").Text(it.label),
			ui.Span().Class("qr-muted qr-nowrap").Text(it.hint),
		).OnClick(func(ctx context.Context) error { return a.paletteRun(ctx, i) }))
	}
	if len(items) == 0 {
		rowsCol.Children(ui.Div().ID("pal-empty").Class("qr-note qr-pad").Text("No matching commands."))
	}
	list.Children(rowsCol, a.scrollbarFlow("pal", "Command list scroll", len(items)*paletteRowH, rows*paletteRowH, a.pal.top*paletteRowH,
		func(o int) {
			a.pal.top = o / paletteRowH
			a.pal.active = max(a.pal.top, min(a.pal.active, a.pal.top+paletteRowsShown-1))
		}))
	// Panel height: padding, search row, gap, and exactly the visible rows.
	panelH := paletteSides*paletteInset + paletteSearchH + paletteGap + rows*paletteRowH + panelBorder
	return ui.Div().ID("modal-scrim").Class("qr-scrim").Children(
		ui.Div().ID("palette").Class("qr-modal qr-palette").Height(ui.Px(float64(panelH))).
			Attr("role", ariaRole("dialog")).Attr("aria-modal", "true").Attr("aria-label", "Command palette").Children(
			ui.Div().Class("qr-palette-search").Children(
				a.icon("search", toneMuted, iconMedium),
				ui.Input().ID("pal-input").Class("qr-palette-input").Attr("type", "text").Attr("placeholder", "Type a command, table, or connection").
					Attr("aria-label", "Command search").Attr("value", a.pal.query),
			),
			list,
		),
	).OnClick(a.closeModal)
}
