package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

func ribbonTabs() []string {
	return []string{"Home", "Insert", "Page Layout", "Formulas", "Data", "Review", "View"}
}

const defaultFontSize = 13

// rbBtn is a compact ribbon button; pressed shows a toggle state.
func rbBtn(action, label string, pressed bool) *ui.Element {
	cls := "rb-rb"
	if pressed {
		cls += " rb-rb-on"
	}
	b := ui.Button(label).Attr("type", "button").Class(cls).
		Attr("data-action", action).Attr("aria-label", label)
	if action != "" {
		b.ID("xl-rb-" + strings.ReplaceAll(action, ":", "-"))
	}
	if pressed {
		b.Attr("aria-pressed", "true")
	} else {
		b.Attr("aria-pressed", "false")
	}
	return b
}

func rbDrop(label string) *ui.Element {
	return ui.Button("").Attr("type", "button").Class("rb-rb rb-rb-dd").
		Attr("aria-label", label).Children(
		ui.Span().Class("rb-dd-label").Text(label),
		ui.Span().Class("rb-dd-arrow").Attr("aria-hidden", "true").Text("▾"),
	)
}

// rbTip is rbBtn with a tooltip and a spoken label distinct from the glyph.
func rbTip(action, glyph, tip string, pressed bool) *ui.Element {
	return rbBtn(action, glyph, pressed).Attr("title", tip).Attr("aria-label", tip)
}

// rbLarge is a tall button with a glyph above its label.
func rbLarge(action, glyph, label string) *ui.Element {
	return ui.Button("").Attr("type", "button").Class("rb-rb rb-rb-lg").
		Attr("data-action", action).Attr("aria-label", label).Children(
		ui.Span().Class("rb-rb-glyph").Text(glyph),
		ui.Span().Class("rb-rb-cap").Text(label),
	)
}

func rbRow(children ...*ui.Element) *ui.Element {
	return ui.Div().Class("rb-rb-row").Children(children...)
}

func rbCol(children ...*ui.Element) *ui.Element {
	return ui.Div().Class("rb-rb-col").Children(children...)
}

func rbGroup(name string, rows ...*ui.Element) *ui.Element {
	return ui.Div().Class("rb-rg").Attr("role", "group").Attr("aria-label", name).Children(
		ui.Div().Class("rb-rg-body").Children(rows...),
		ui.Span().Class("rb-rg-name").Text(name),
	)
}

// ribbon builds the tab strip and the active tab body.
func (a *app) ribbon() *ui.Element {
	tab := a.ribbonTab
	if tab == "" {
		tab = "Home"
	}
	strip := ui.Div().Class("rb-tabs").Attr("role", "tablist").Attr("aria-label", "Ribbon tabs")
	strip.Children(ui.Button("File").Attr("type", "button").Class("rb-tab rb-tab-file").
		Attr("data-action", "xl:menu:file").Attr("aria-haspopup", "menu").
		Attr("aria-expanded", strconv.FormatBool(a.x.menu == "file")))
	for _, t := range ribbonTabs() {
		cls := "rb-tab"
		sel := "false"
		if t == tab {
			cls += " rb-tab-on"
			sel = "true"
		}
		strip.Children(ui.Button(t).Attr("type", "button").Class(cls).Attr("role", "tab").
			Attr("aria-selected", sel).Attr("data-action", "xl:tab:"+t))
	}
	strip.Children(ui.Span().Class("rb-tabs-fill"))
	collapseGlyph, collapseTip := "^", "Collapse the Ribbon (Ctrl+F1)"
	if a.ribbonCollapsed {
		collapseGlyph, collapseTip = "v", "Expand the Ribbon (Ctrl+F1)"
	}
	strip.Children(rbTip("xl:ribbon:toggle", collapseGlyph, collapseTip, false).
		Attr("aria-expanded", strconv.FormatBool(!a.ribbonCollapsed)).Attr("id", "xl-ribbon-collapse"))
	strip.Children(rbTip("xl:undo", "Undo", "Undo (Ctrl+Z)", false).Class(a.dis("rb-rb rb-qa", !a.wb.CanUndo())))
	strip.Children(rbTip("xl:redo", "Redo", "Redo (Ctrl+Y)", false).Class(a.dis("rb-rb rb-qa", !a.wb.CanRedo())))
	root := ui.Div().Class("rb-ribbon").Attr("role", "toolbar").Attr("aria-label", "Ribbon").Children(strip)
	if a.ribbonCollapsed {
		return root
	}
	body := ui.Div().Class("rb-ribbon-body")
	switch tab {
	case "Insert":
		body.Children(a.insertTab()...)
	case "Page Layout":
		body.Children(a.pageLayoutTab()...)
	case "Formulas":
		body.Children(a.formulasTab()...)
	case "Data":
		body.Children(a.dataTab()...)
	case "Review":
		body.Children(a.reviewTab()...)
	case "View":
		body.Children(a.viewTab()...)
	default:
		body.Children(a.homeTab()...)
	}
	return root.Children(body)
}

// dis returns the class string with a disabled marker.
func (a *app) dis(base string, off bool) string {
	if off {
		return base + " rb-rb-off"
	}
	return base
}

func (a *app) activeStyle() Style {
	return a.wb.Sheet().StyleAt(a.active.Row, a.active.Col)
}

func (a *app) homeTab() []*ui.Element {
	st := a.activeStyle()
	size := st.FontSize
	if size == 0 {
		size = defaultFontSize
	}
	sizeTrig := rbDrop(strconv.Itoa(size)).Class("rb-rb rb-rb-dd rb-rb-size").Attr("aria-label", "Font size")
	famTrig := rbDrop(fontFamilyLabel(st.FontFamily)).Class("rb-rb rb-rb-dd rb-rb-size").Attr("aria-label", "Font family")
	fillTrig := rbDrop("Fill").Attr("title", "Fill Color").Attr("aria-label", "Fill Color")
	fontTrig := rbDrop("A").Attr("title", "Font Color").Attr("aria-label", "Font Color")
	nfTrig := rbDrop(numberFormatLabel(st.NumFmt)).Class("rb-rb rb-rb-dd rb-rb-nf").Attr("aria-label", "Number format")
	return []*ui.Element{
		rbGroup("Clipboard",
			rbRow(rbLarge("xl:paste", "📋", "Paste"),
				rbCol(rbRow(rbTip("xl:cut", "Cut", "Cut (Ctrl+X)", false),
					rbTip("xl:copy", "Copy", "Copy (Ctrl+C)", false)),
					rbTip("xl:painter", "Painter", "Format Painter", a.x.painter).Class("rb-rb"+actClass(a.x.painter)))),
		),
		rbGroup("Font",
			rbRow(
				a.dropdown("ffamily", famTrig, a.fontFamilyMenu),
				a.dropdown("fsize", sizeTrig, a.fontSizeMenu),
				rbTip("xl:fsize:+", "A+", "Increase Font Size", false),
				rbTip("xl:fsize:-", "A-", "Decrease Font Size", false),
			),
			rbRow(
				rbTip("xl:bold", "B", "Bold (Ctrl+B)", st.Bold).Class(a.dis("rb-rb rb-bold", false)+actClass(st.Bold)),
				rbTip("xl:italic", "I", "Italic (Ctrl+I)", st.Italic).Class("rb-rb rb-italic"+actClass(st.Italic)),
				rbTip("xl:underline", "U", "Underline (Ctrl+U)", st.Underline).Class("rb-rb rb-underline"+actClass(st.Underline)),
				rbTip("xl:strike", "S", "Strikethrough", st.Strike).Class("rb-rb rb-strike"+actClass(st.Strike)),
				a.dropdown("border", rbDrop("Border").Attr("title", "Borders").Attr("aria-label", "Borders"), a.borderMenu),
				a.dropdown("fill", fillTrig, func() *ui.Element { return a.colorPicker("xl:fillc", "No Fill") }),
				a.dropdown("fontc", fontTrig, func() *ui.Element { return a.colorPicker("xl:fontc", "Automatic") }),
			),
		),
		rbGroup("Alignment",
			rbRow(
				rbTip("xl:align:left", "Left", "Align Left", st.Align == "left").Class("rb-rb"+actClass(st.Align == "left")),
				rbTip("xl:align:center", "Center", "Center", st.Align == "center").Class("rb-rb"+actClass(st.Align == "center")),
				rbTip("xl:align:right", "Right", "Align Right", st.Align == "right").Class("rb-rb"+actClass(st.Align == "right")),
				rbTip("xl:valign:top", "Top", "Align Top", st.VAlign == "top").Class("rb-rb"+actClass(st.VAlign == "top")),
				rbTip("xl:valign:middle", "Mid", "Align Middle", st.VAlign == "middle").Class("rb-rb"+actClass(st.VAlign == "middle")),
				rbTip("xl:valign:bottom", "Bot", "Align Bottom", st.VAlign == "bottom").Class("rb-rb"+actClass(st.VAlign == "bottom")),
			),
			rbRow(rbTip("xl:wrap", "Wrap", "Wrap Text", st.Wrap).Class("rb-rb"+actClass(st.Wrap)),
				a.dropdown("merge", rbDrop("Merge"), a.mergeMenu),
				rbTip("xl:indent:-", "<-", "Decrease Indent", false),
				rbTip("xl:indent:+", "->", "Increase Indent", false)),
		),
		rbGroup("Number",
			rbRow(a.dropdown("nf", nfTrig, a.numberFormatMenu)),
			rbRow(
				rbTip("xl:cur", "$", "Accounting Number Format", false),
				rbTip("xl:pct", "%", "Percent Style", false),
				rbTip("xl:comma", ",", "Comma Style", false),
				rbTip("xl:dec:+", ".0+", "Increase Decimal", false),
				rbTip("xl:dec:-", ".0-", "Decrease Decimal", false),
			),
		),
		rbGroup("Styles",
			a.dropdown("cf", rbDrop("Conditional Formatting"), a.condFormatMenu),
			a.dropdown("cstyles", rbDrop("Cell Styles"), a.cellStylesMenu),
		),
		rbGroup("Cells",
			rbRow(a.dropdown("cells", rbDrop("Delete / Insert"), a.insertMenu)),
		),
		rbGroup("Editing",
			rbRow(rbTip("xl:autosum", "Sum", "AutoSum", false), a.dropdown("clear", rbDrop("Clear"), a.clearMenu)),
			rbRow(a.dropdown("sortfilter", rbDrop("Sort & Filter"), a.sortFilterMenu), rbTip("xl:dlg:find", "Find", "Find & Replace (Ctrl+F)", false)),
		),
	}
}

func actClass(on bool) string {
	if on {
		return " rb-rb-on"
	}
	return ""
}

func (a *app) insertTab() []*ui.Element {
	return []*ui.Element{
		rbGroup("Charts",
			rbRow(a.dropdown("chart", rbDrop("Insert Chart"), a.chartMenu)),
			rbRow(rbTip("xl:chart:new:column", "Column", "Column chart from selection", false),
				rbTip("xl:chart:new:bar", "Bar", "Bar chart from selection", false),
				rbTip("xl:chart:new:line", "Line", "Line chart from selection", false),
				rbTip("xl:chart:new:area", "Area", "Area chart from selection", false),
				rbTip("xl:chart:new:pie", "Pie", "Pie chart from selection", false)),
		),
		rbGroup("Comments",
			rbRow(rbTip("xl:dlg:comment", "New Comment", "New Comment (Shift+F2)", false)),
		),
		rbGroup("Symbols",
			rbRow(rbTip("xl:dlg:validation", "Data Validation", "Data Validation", false)),
		),
	}
}

func (a *app) pageLayoutTab() []*ui.Element {
	return []*ui.Element{
		rbGroup("Themes",
			rbRow(rbTip("xl:theme", themeLabel(a.dark), "Switch between Excel Light and Excel Dark", false)),
		),
		rbGroup("Sheet Options",
			rbRow(rbTip("xl:view:grid", "Gridlines", "Show gridlines", !a.x.hideGrid).Class("rb-rb"+actClass(!a.x.hideGrid)),
				rbTip("xl:view:heads", "Headings", "Show row and column headings", !a.x.hideHeads).Class("rb-rb"+actClass(!a.x.hideHeads))),
		),
	}
}

func themeLabel(dark bool) string {
	if dark {
		return "Dark theme"
	}
	return "Light theme"
}

func (a *app) formulasTab() []*ui.Element {
	fn := func(name string) *ui.Element {
		return rbTip("xl:fn:"+name, name, "Insert "+name, false)
	}
	return []*ui.Element{
		rbGroup("Function Library",
			rbRow(rbTip("xl:autosum", "AutoSum", "AutoSum", false), fn("AVERAGE"), fn("COUNT"), fn("MAX"), fn("MIN")),
			rbRow(fn("IF"), fn("VLOOKUP"), fn("CONCAT"), fn("ROUND"), fn("TODAY")),
		),
		rbGroup("Formula Auditing",
			rbRow(rbTip("xl:view:formulas", "Show Formulas", "Show Formulas (Ctrl+`)", a.x.showFormulas).Class("rb-rb"+actClass(a.x.showFormulas))),
		),
	}
}

func (a *app) dataTab() []*ui.Element {
	return []*ui.Element{
		rbGroup("Get & Transform",
			rbRow(rbTip("xl:dlg:open", "Open / Import", "Open or import a file", false)),
			rbRow(rbTip("xl:dlg:save", "Export", "Save or export", false)),
		),
		rbGroup("Sort & Filter",
			rbRow(rbTip("xl:sort:asc", "A-Z", "Sort A to Z", false), rbTip("xl:sort:desc", "Z-A", "Sort Z to A", false),
				rbTip("xl:dlg:sort", "Sort", "Custom Sort", false)),
			rbRow(rbTip("xl:filter:toggle", "Filter", "Filter (Ctrl+Shift+L)", a.x.filterOn).Class("rb-rb"+actClass(a.x.filterOn)),
				rbTip("xl:filter:open:"+strconv.Itoa(a.active.Col), "Column", "Filter the active column", false),
				rbTip("xl:filter:clear", "Clear", "Clear filter", false)),
		),
		rbGroup("Data Tools",
			rbRow(rbTip("xl:dlg:validation", "Data Validation", "Data Validation", false)),
		),
	}
}

func (a *app) reviewTab() []*ui.Element {
	return []*ui.Element{
		rbGroup("Comments",
			rbRow(rbTip("xl:dlg:comment", "New Comment", "New Comment (Shift+F2)", false),
				rbTip("xl:comment:delete", "Delete", "Delete Comment", false)),
			rbRow(rbTip("xl:comment:prev", "Previous", "Previous Comment", false),
				rbTip("xl:comment:next", "Next", "Next Comment", false)),
		),
	}
}

func (a *app) viewTab() []*ui.Element {
	return []*ui.Element{
		rbGroup("Show",
			rbRow(rbTip("xl:view:grid", "Gridlines", "Show gridlines", !a.x.hideGrid).Class("rb-rb"+actClass(!a.x.hideGrid)),
				rbTip("xl:view:heads", "Headings", "Show headings", !a.x.hideHeads).Class("rb-rb"+actClass(!a.x.hideHeads))),
			rbRow(rbTip("xl:view:fbar", "Formula Bar", "Show formula bar", !a.x.hideFormulaBar).Class("rb-rb"+actClass(!a.x.hideFormulaBar)),
				rbTip("xl:view:formulas", "Show Formulas", "Show Formulas (Ctrl+`)", a.x.showFormulas).Class("rb-rb"+actClass(a.x.showFormulas))),
		),
		rbGroup("Zoom",
			rbRow(rbTip("xl:zoom:in", "Zoom +", "Zoom in", false), rbTip("xl:zoom:out", "Zoom -", "Zoom out", false)),
			rbRow(rbTip("xl:zoom:reset", "100%", "Reset zoom to 100%", false),
				rbTip("xl:zoom:sel", "To Selection", "Zoom to selection", false)),
		),
		rbGroup("Window",
			rbRow(a.dropdown("freeze", rbDrop("Freeze Panes"), a.freezeMenu)),
			rbRow(rbTip("xl:theme", themeLabel(a.dark), "Toggle Light and Dark theme", false)),
		),
	}
}
