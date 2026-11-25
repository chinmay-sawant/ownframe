package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

// ddItem is one row inside a dropdown panel or context menu.
func ddItem(action, label, hint string) *ui.Element {
	row := ui.Button("").Attr("type", "button").Class("rb-dd-item").
		Attr("data-action", action).Attr("role", "menuitem").Attr("aria-label", label)
	row.Children(ui.Span().Class("rb-dd-label").Text(label))
	if hint != "" {
		row.Children(ui.Span().Class("rb-dd-hint").Text(hint))
	}
	return row
}

// ddSep is a hairline between dropdown groups.
func ddSep() *ui.Element { return ui.Div().Class("rb-dd-sep").Attr("role", "separator") }

// dropdown returns the trigger for a menu. The panel is not nested in the
// ribbon: blinkless counts absolutely positioned children toward a group's
// width, so the open panel mounts in overlays() at the trigger's anchor.
func (a *app) dropdown(id string, trigger *ui.Element, _ func() *ui.Element) *ui.Element {
	trigger.Attr("data-action", "xl:menu:"+id).Attr("aria-haspopup", "menu").
		Attr("aria-expanded", strconv.FormatBool(a.x.menu == id))
	return trigger
}

// menuPanel renders the open dropdown at its anchor, or nil when none is open.
func (a *app) menuPanel() *ui.Element {
	var panel *ui.Element
	switch a.x.menu {
	case "":
		return nil
	case "nf":
		panel = a.numberFormatMenu()
	case "fsize":
		panel = a.fontSizeMenu()
	case "ffamily":
		panel = a.fontFamilyMenu()
	case "border":
		panel = a.borderMenu()
	case "fill":
		panel = a.colorPicker("xl:fillc", "No Fill")
	case "fontc":
		panel = a.colorPicker("xl:fontc", "Automatic")
	case "cf":
		panel = a.condFormatMenu()
	case "cstyles":
		panel = a.cellStylesMenu()
	case "clear":
		panel = a.clearMenu()
	case "sortfilter":
		panel = a.sortFilterMenu()
	case "cells":
		panel = a.insertMenu()
	case "merge":
		panel = a.mergeMenu()
	case "freeze":
		panel = a.freezeMenu()
	case "chart":
		panel = a.chartMenu()
	case "file":
		panel = a.fileMenu()
	default:
		return nil
	}
	panel.ID("rb-menu-" + a.x.menu).
		Position(ui.PositionFixed).Left(ui.Px(a.x.anchorX)).Top(ui.Px(a.x.anchorY))
	return panel
}

func panelOf(items ...*ui.Element) *ui.Element {
	return ui.Div().Class("rb-dd-panel").Attr("role", "menu").Children(items...)
}

// numberFormats are the dropdown entries: label, Style.NumFmt value.
func numberFormats() []struct{ Label, Fmt string } {
	return []struct{ Label, Fmt string }{
		{"General", "General"},
		{"Number", "Number"},
		{"Currency", "Currency"},
		{"Accounting", "Accounting"},
		{"Short Date", "Date"},
		{"Long Date", "LongDate"},
		{"Time", "Time"},
		{"Percentage", "Percent"},
		{"Scientific", "Scientific"},
		{"Text", "Text"},
	}
}

func numberFormatLabel(f string) string {
	if f == "" {
		return "General"
	}
	for _, n := range numberFormats() {
		if n.Fmt == f {
			return n.Label
		}
	}
	return f
}

func (a *app) numberFormatMenu() *ui.Element {
	p := panelOf()
	for _, n := range numberFormats() {
		p.Children(ddItem("xl:nf:"+n.Fmt, n.Label, ""))
	}
	return p
}

func (a *app) fontSizeMenu() *ui.Element {
	p := panelOf()
	p.Attr("class", "rb-dd-panel rb-dd-scroll")
	for _, n := range []int{9, 10, 11, 12, 13, 14, 16, 18, 20, 24, 28, 32, 36} {
		p.Children(ddItem("xl:fsize:"+strconv.Itoa(n), strconv.Itoa(n), ""))
	}
	return p
}

// fontFamilies are the Font family dropdown entries.
func fontFamilies() []struct{ Label, Value string } {
	return []struct{ Label, Value string }{
		{"Default", ""}, {"Sans", "sans"}, {"Mono", "mono"}, {"Serif", "serif"},
	}
}

func fontFamilyLabel(v string) string {
	for _, f := range fontFamilies() {
		if f.Value == v {
			return f.Label
		}
	}
	return "Default"
}

func (a *app) fontFamilyMenu() *ui.Element {
	p := panelOf()
	for _, f := range fontFamilies() {
		p.Children(ddItem("xl:family:"+f.Value, f.Label, ""))
	}
	return p
}

func (a *app) clearMenu() *ui.Element {
	return panelOf(
		ddItem("xl:clear:all", "Clear All", ""),
		ddItem("xl:clear:formats", "Clear Formats", ""),
		ddItem("xl:clear:contents", "Clear Contents", "Del"),
		ddItem("xl:clear:comments", "Clear Comments", ""),
	)
}

func (a *app) sortFilterMenu() *ui.Element {
	return panelOf(
		ddItem("xl:sort:asc", "Sort A to Z", ""),
		ddItem("xl:sort:desc", "Sort Z to A", ""),
		ddItem("xl:dlg:sort", "Custom Sort...", ""),
		ddSep(),
		ddItem("xl:filter:toggle", "Filter", "Ctrl+Shift+L"),
		ddItem("xl:filter:clear", "Clear Filter", ""),
	)
}

func (a *app) condFormatMenu() *ui.Element {
	return panelOf(
		ddItem("xl:cf:gt", "Greater Than...", ""),
		ddItem("xl:cf:lt", "Less Than...", ""),
		ddItem("xl:cf:between", "Between...", ""),
		ddItem("xl:cf:eq", "Equal To...", ""),
		ddItem("xl:cf:text", "Text that Contains...", ""),
		ddItem("xl:cf:dupes", "Duplicate Values", ""),
		ddSep(),
		ddItem("xl:cf:top", "Top 10 Items", ""),
		ddItem("xl:cf:bottom", "Bottom 10 Items", ""),
		ddItem("xl:cf:above", "Above Average", ""),
		ddItem("xl:cf:below", "Below Average", ""),
		ddItem("xl:cf:databar", "Data Bars", ""),
		ddItem("xl:cf:scale3", "Color Scales", ""),
		ddSep(),
		ddItem("xl:cf:clear", "Clear Rules from Selection", ""),
		ddItem("xl:dlg:condfmt", "Manage Rules...", ""),
	)
}

// cellStyles is the Cell Styles gallery.
// cellStyle is one Cell Styles gallery entry.
type cellStyle struct {
	Name   string
	FG, BG string
	Bold   bool
	Size   int
}

func cellStyles() []cellStyle {
	return []cellStyle{
		{"Good", "#006100", "#c6efce", false, 0},
		{"Bad", "#9c0006", "#ffc7ce", false, 0},
		{"Neutral", "#9c6500", "#ffeb9c", false, 0},
		{"Heading 1", "#44546a", "", true, 18},
		{"Heading 2", "#44546a", "", true, 15},
		{"Heading 3", "#44546a", "", true, 13},
		{"Total", "", "", true, 0},
		{"Accent 1", "#ffffff", "#4472c4", false, 0},
		{"Accent 2", "#ffffff", "#ed7d31", false, 0},
		{"Accent 3", "#000000", "#a5a5a5", false, 0},
		{"Accent 4", "#000000", "#ffc000", false, 0},
	}
}

func (a *app) cellStylesMenu() *ui.Element {
	p := ui.Div().Class("rb-dd-panel rb-styles").Attr("role", "menu")
	for _, cs := range cellStyles() {
		b := ui.Button(cs.Name).Attr("type", "button").Class("rb-style-chip").
			Attr("data-action", "xl:cstyle:"+cs.Name).Attr("role", "menuitem")
		if cs.BG != "" {
			b.BackgroundColor(ui.Hex(cs.BG))
		}
		if cs.FG != "" {
			b.Color(ui.Hex(cs.FG))
		}
		p.Children(b)
	}
	return p
}

func (a *app) insertMenu() *ui.Element {
	return panelOf(
		ddItem("xl:cells:insrow", "Insert Sheet Rows", ""),
		ddItem("xl:cells:inscol", "Insert Sheet Columns", ""),
		ddItem("xl:cells:delrow", "Delete Sheet Rows", ""),
		ddItem("xl:cells:delcol", "Delete Sheet Columns", ""),
		ddSep(),
		ddItem("xl:cells:clearsel", "Delete Cell Contents", ""),
		ddItem("xl:dlg:comment", "Insert Comment", "Shift+F2"),
	)
}

func (a *app) mergeMenu() *ui.Element {
	return panelOf(
		ddItem("xl:merge:center", "Merge & Center", ""),
		ddItem("xl:merge:across", "Merge Cells", ""),
		ddItem("xl:merge:unmerge", "Unmerge Cells", ""),
	)
}

func (a *app) freezeMenu() *ui.Element {
	return panelOf(
		ddItem("xl:freeze:both", "Freeze Panes", ""),
		ddItem("xl:freeze:row", "Freeze Top Row", ""),
		ddItem("xl:freeze:col", "Freeze First Column", ""),
		ddItem("xl:freeze:none", "Unfreeze Panes", ""),
	)
}

func (a *app) chartMenu() *ui.Element {
	p := panelOf()
	for _, k := range chartKinds() {
		p.Children(ddItem("xl:chart:new:"+k.Kind, k.Label+" Chart", ""))
	}
	return p
}

func (a *app) fileMenu() *ui.Element {
	return panelOf(
		ddItem("xl:file:new", "New", "Ctrl+N"),
		ddItem("xl:dlg:open", "Open...", "Ctrl+O"),
		ddItem("xl:dlg:save", "Save / Export...", "Ctrl+S"),
		ddItem("xl:dlg:shortcuts", "Keyboard Shortcuts", "Ctrl+/"),
	)
}

// openContextMenu opens the right-click menu at window coordinates. kind is
// "cell", "row", "col" or "tab". The host (grid.go) calls it from its
// right-click handler.
func (a *app) openContextMenu(kind string, x, y float64) {
	a.x.ctxKind, a.x.ctxX, a.x.ctxY = kind, x, y
	a.setOverlay("ctx")
	a.redraw()
}

// ctxItem is one context menu row. An empty Action is a separator.
type ctxItem struct{ Action, Label, Hint string }

func ctxItems(kind string) []ctxItem {
	switch kind {
	case "row", "col":
		return []ctxItem{
			{"xl:ctx:cut", "Cut", "Ctrl+X"}, {"xl:ctx:copy", "Copy", "Ctrl+C"}, {"xl:ctx:paste", "Paste", "Ctrl+V"},
			{"xl:ctx:insert", "Insert", ""}, {"xl:ctx:delete", "Delete", ""},
			{"xl:ctx:hide", "Hide", ""}, {"xl:ctx:unhide", "Unhide", ""},
			{"xl:ctx:autofit", "Auto-Fit", ""}, {"xl:ctx:clear", "Clear Contents", ""},
		}
	case "tab":
		return []ctxItem{
			{"xl:ctx:addsheet", "Insert Sheet", ""}, {"xl:ctx:deletesheet", "Delete", ""},
		}
	}
	return []ctxItem{
		{"xl:ctx:cut", "Cut", "Ctrl+X"}, {"xl:ctx:copy", "Copy", "Ctrl+C"}, {"xl:ctx:paste", "Paste", "Ctrl+V"},
		{"xl:dlg:paste", "Paste Special...", "Ctrl+Alt+V"},
		{"xl:cells:insrow", "Insert Row", ""}, {"xl:cells:delrow", "Delete Row", ""},
		{"xl:ctx:clear", "Clear Contents", "Del"},
		{"xl:dlg:format", "Format Cells...", "Ctrl+1"},
		{"xl:dlg:comment", "Insert Comment", "Shift+F2"},
		{"xl:dlg:validation", "Data Validation...", ""},
	}
}

// ctxKindFor maps a hit-tested element id to a context menu kind.
func ctxKindFor(id string) string {
	switch {
	case strings.HasPrefix(id, "xl-gch-"):
		return "col"
	case strings.HasPrefix(id, "xl-grh-"):
		return "row"
	case strings.HasPrefix(id, "xl-tab-"):
		return "tab"
	case isGridID(id):
		return "cell"
	}
	return ""
}

// contextMenuFor reports the right-click rows for one element id. It is the
// ownframe.Handlers.ContextMenuFor hook; outside the grid, header and tab
// strip it returns nil so the built-in editing menu stays.
func (a *app) contextMenuFor(id string) []ownframe.MenuItem {
	kind := ctxKindFor(id)
	if kind == "" {
		return nil
	}
	a.x.ctxKind = kind
	items := ctxItems(kind)
	out := make([]ownframe.MenuItem, 0, len(items))
	for _, it := range items {
		out = append(out, ownframe.MenuItem{ID: it.Action, Label: it.Label, Enabled: true, Hint: it.Hint})
	}
	return out
}

// menuAction runs the chosen row. It is the Handlers.MenuAction hook.
func (a *app) menuAction(ctx context.Context, id string) error {
	a.runActionCtx(ctx, id)
	return nil
}

// contextMenu renders the open in-page context menu at its saved position.
func (a *app) contextMenu() *ui.Element {
	if a.overlay != "ctx" {
		return nil
	}
	var items []*ui.Element
	for _, it := range ctxItems(a.x.ctxKind) {
		items = append(items, ddItem(it.Action, it.Label, it.Hint))
	}
	return ui.Div().Class("rb-ctx-layer").Attr("data-action", "xl:close").Children(
		ui.Div().Class("rb-dd-panel rb-ctx").Attr("role", "menu").Attr("aria-label", "Context menu").
			Attr("id", "rb-ctx-menu").
			Position(ui.PositionFixed).Left(ui.Px(a.x.ctxX)).Top(ui.Px(a.x.ctxY)).
			Children(items...),
	)
}
