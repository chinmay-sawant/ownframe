package main

import (
	"fmt"
	"strconv"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// findHitLimit caps the hits listed in the find dialog.
const (
	findHitLimit = 40
	dialogWidth  = 540.0 // modal width before zoom
	vlistDX      = 120.0 // validation list offset from the chart origin
	vlistDY      = 30.0
)

// opt reads a dialog option, falling back to def.
func (a *app) opt(key, def string) string {
	if v, ok := a.x.opt[key]; ok {
		return v
	}
	return def
}

func (a *app) setOpt(key, val string) {
	if a.x.opt == nil {
		a.x.opt = map[string]string{}
	}
	a.x.opt[key] = val
}

// input reads a dialog text field.
func (a *app) input(id string) string { return a.page.Form().FormValue(id) }

// optBtn is a radio-style option button bound to the opt store.
func (a *app) optBtn(key, val, label, def string) *ui.Element {
	on := a.opt(key, def) == val
	cls := "rb-opt"
	if on {
		cls += " rb-opt-on"
	}
	return ui.Button(label).Attr("type", "button").Class(cls).
		Attr("data-action", "xl:opt:"+key+":"+val).Attr("aria-pressed", strconv.FormatBool(on))
}

func (a *app) flagBtn(key, label string, def bool) *ui.Element {
	cur := a.opt(key, strconv.FormatBool(def)) == "true"
	cls := "rb-opt"
	if cur {
		cls += " rb-opt-on"
	}
	return ui.Button(label).Attr("type", "button").Class(cls).
		Attr("data-action", "xl:flag:"+key).Attr("aria-pressed", strconv.FormatBool(cur))
}

func (a *app) flag(key string, def bool) bool {
	return a.opt(key, strconv.FormatBool(def)) == "true"
}

func dlgBtn(action, label string, primary bool) *ui.Element {
	v := chadcn.ButtonOutlined
	cls := "of-button of-button--outlined of-button--small"
	if primary {
		v = chadcn.ButtonFilled
		cls = "of-button of-button--filled of-button--small rb-primary"
	}
	_ = v
	return ui.Button(label).Attr("type", "button").Class(cls).Attr("data-action", action)
}

func fieldRow(label string, control ...*ui.Element) *ui.Element {
	return ui.Div().Class("rb-field").Children(ui.Span().Class("rb-field-label").Text(label)).Children(control...)
}

func (a *app) modal(id, title, desc string, body []*ui.Element, footer ...*ui.Element) *ui.Element {
	return chadcn.Dialog(chadcn.DialogOptions{
		ID: "rb-dlg-" + id, Title: title, Description: desc,
		CloseAction: "xl:close", Open: true, Footer: footer, Width: ui.Px(a.zf(dialogWidth)),
	}, body...)
}

// dialog returns the open modal for a.overlay, or nil.
func (a *app) dialog() *ui.Element {
	switch a.overlay {
	case "find":
		return a.findDialog()
	case "paste":
		return a.pasteDialog()
	case "format":
		return a.formatDialog()
	case "validation":
		return a.validationDialog()
	case "sort":
		return a.sortDialog()
	case "condfmt":
		return a.condFormatDialog()
	case "comment":
		return a.commentDialog()
	case "open":
		return a.openDialog()
	case "save":
		return a.saveDialog()
	case "shortcuts":
		return a.shortcutsDialog()
	case "filter":
		return a.filterDialog()
	case "validation:list":
		return a.validationListMenu()
	}
	return nil
}

// validationListMenu is the in-cell dropdown for a List-validated cell. The
// grid opens it from the cell arrow; picking a row writes the value.
func (a *app) validationListMenu() *ui.Element {
	items := a.wb.Sheet().ValidationList(a.active.Row, a.active.Col)
	panel := ui.Div().Class("rb-dd-panel rb-vlist").Attr("role", "listbox").Attr("aria-label", "Allowed values")
	for i, it := range items {
		panel.Children(ddItem("xl:vlist:"+strconv.Itoa(i), it, ""))
	}
	return ui.Div().Class("rb-ctx-layer").Attr("data-action", "xl:close").Children(
		panel.Position(ui.PositionFixed).Left(ui.Px(a.zf(chartOriginX + vlistDX))).Top(ui.Px(a.zf(chartOriginY + vlistDY))))
}

func (a *app) findDialog() *ui.Element {
	body := []*ui.Element{
		fieldRow("Find what", chadcn.Input("rb-find-q", a.opt("findq", ""))),
		fieldRow("Replace with", chadcn.Input("rb-find-r", a.opt("findr", ""))),
		fieldRow("Options",
			a.flagBtn("matchcase", "Match case", false),
			a.flagBtn("entire", "Match entire cell contents", false),
		),
		fieldRow("Within", a.optBtn("scope", "sheet", "Sheet", "sheet"), a.optBtn("scope", "book", "Workbook", "sheet")),
		fieldRow("Look in", a.optBtn("lookin", "formulas", "Formulas", "formulas"), a.optBtn("lookin", "values", "Values", "formulas")),
	}
	if a.x.findMsg != "" {
		body = append(body, ui.Div().Class("app-muted rb-find-msg").Attr("role", "status").Text(a.x.findMsg))
	}
	list := ui.Div().Class("rb-find-list").Attr("role", "list")
	for i, h := range a.x.findHits {
		if i >= findHitLimit {
			break
		}
		list.Children(ui.Button(h.Sheet+"!"+h.Ref.A1()+"  "+h.Text).Attr("type", "button").Class("rb-find-hit").
			Attr("data-action", "xl:find:go:"+strconv.Itoa(i)))
	}
	if len(a.x.findHits) > 0 {
		body = append(body, list)
	}
	return a.modal("find", "Find and Replace", "Search the current sheet or the whole workbook.", body,
		dlgBtn("xl:find:prev", "Find Previous", false), dlgBtn("xl:find:next", "Find Next", false),
		dlgBtn("xl:find:all", "Find All", false), dlgBtn("xl:find:replace", "Replace", false),
		dlgBtn("xl:find:replaceall", "Replace All", true), dlgBtn("xl:close", "Close", false))
}

func (a *app) pasteDialog() *ui.Element {
	body := []*ui.Element{
		fieldRow("Paste",
			a.optBtn("ps", "all", "All", "all"),
			a.optBtn("ps", "values", "Values only", "all"),
			a.optBtn("ps", "formulas", "Formulas only", "all"),
			a.optBtn("ps", "formats", "Formats only", "all"),
		),
		fieldRow("Operation", a.flagBtn("pstranspose", "Transpose", false), a.flagBtn("psskip", "Skip blanks", false)),
		ui.Div().Class("app-muted").Text("Reads the clipboard as tab-separated text and pastes at the active cell."),
	}
	return a.modal("paste", "Paste Special", "", body,
		dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:paste:special", "OK", true))
}

func formatTabs() []string {
	return []string{"Number", "Alignment", "Font", "Border", "Fill"}
}

// fontSizes are the Format Cells font size choices.
func fontSizes() []int { return []int{11, defaultFontSize, 16, 20, 28} }

// fontSizeOpts builds the size option buttons, marking the current size.
func fontSizeOpts(cur int, opt func(act, label string, on bool) *ui.Element) []*ui.Element {
	var out []*ui.Element
	for _, n := range fontSizes() {
		v := strconv.Itoa(n)
		out = append(out, opt("size:"+v, v, cur == n))
	}
	return out
}

func (a *app) formatDialog() *ui.Element {
	d := a.x.fmtDraft
	tab := a.x.fmtTab
	if tab == "" {
		tab = "Number"
	}
	tabs := ui.Div().Class("rb-fc-tabs").Attr("role", "tablist")
	for _, t := range formatTabs() {
		cls := "rb-opt"
		sel := "false"
		if t == tab {
			cls += " rb-opt-on"
			sel = "true"
		}
		tabs.Children(ui.Button(t).Attr("type", "button").Class(cls).Attr("role", "tab").
			Attr("aria-selected", sel).Attr("data-action", "xl:fc:tab:"+t))
	}
	fcOpt := func(act, label string, on bool) *ui.Element {
		cls := "rb-opt"
		if on {
			cls += " rb-opt-on"
		}
		return ui.Button(label).Attr("type", "button").Class(cls).
			Attr("data-action", "xl:fc:"+act).Attr("aria-pressed", strconv.FormatBool(on))
	}
	pane := ui.Div().Class("rb-fc-pane")
	switch tab {
	case "Alignment":
		pane.Children(
			fieldRow("Horizontal", fcOpt("align:", "General", d.Align == ""), fcOpt("align:left", "Left", d.Align == "left"),
				fcOpt("align:center", "Center", d.Align == "center"), fcOpt("align:right", "Right", d.Align == "right")),
			fieldRow("Vertical", fcOpt("valign:", "Bottom", d.VAlign == "" || d.VAlign == "bottom"), fcOpt("valign:middle", "Middle", d.VAlign == "middle"),
				fcOpt("valign:top", "Top", d.VAlign == "top")),
			fieldRow("Indent", fcOpt("indent:-", "-", false), ui.Span().Class("rb-dec-n").Text(strconv.Itoa(d.Indent)), fcOpt("indent:+", "+", false)),
			fieldRow("Text control", fcOpt("wrap", "Wrap text", d.Wrap)),
		)
	case "Font":
		pane.Children(
			fieldRow("Family", fcOpt("family:", "Default", d.FontFamily == ""), fcOpt("family:sans", "Sans", d.FontFamily == "sans"),
				fcOpt("family:mono", "Mono", d.FontFamily == "mono"), fcOpt("family:serif", "Serif", d.FontFamily == "serif")),
			fieldRow("Style", fcOpt("bold", "Bold", d.Bold), fcOpt("italic", "Italic", d.Italic),
				fcOpt("underline", "Underline", d.Underline), fcOpt("strike", "Strikethrough", d.Strike)),
			fieldRow("Size", fontSizeOpts(d.FontSize, fcOpt)...),
			fieldRow("Color", fcSwatches("xl:fc:font")...),
		)
	case "Border":
		row := fieldRow("Presets")
		cur := a.opt("fcborder", "")
		for _, b := range borderMenuItems() {
			row.Children(fcOpt("border:"+b.Action, b.Label, cur == b.Action))
		}
		pane.Children(row)
	case "Fill":
		pane.Children(fieldRow("Background", append([]*ui.Element{fcOpt("fill:", "No Fill", d.BG == "")}, fcSwatches("xl:fc:fill")...)...))
	default:
		row := fieldRow("Category")
		for _, n := range numberFormats() {
			row.Children(fcOpt("nf:"+n.Fmt, n.Label, d.NumFmt == n.Fmt || (d.NumFmt == "" && n.Fmt == "General")))
		}
		pane.Children(row, fieldRow("Decimal places",
			fcOpt("dec:-", "-", false), ui.Span().Class("rb-dec-n").Text(strconv.Itoa(d.Decimals)), fcOpt("dec:+", "+", false)))
	}
	preview := ui.Div().Class("rb-fc-preview").Text("Sample 1234.5")
	if d.BG != "" {
		preview.BackgroundColor(ui.Hex(d.BG))
	}
	return a.modal("format", "Format Cells", "Changes apply to "+a.sel.String()+" when you press OK.",
		[]*ui.Element{tabs, pane, preview},
		dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:fc:ok", "OK", true))
}

// fcSwatches is a compact color row for the Format Cells dialog.
func fcSwatches(prefix string) []*ui.Element {
	out := make([]*ui.Element, 0, len(officeStandardColors())+len(officeThemeColors()))
	for _, c := range officeStandardColors() {
		out = append(out, swatch(prefix, c))
	}
	for _, c := range officeThemeColors()[2:] {
		out = append(out, swatch(prefix, c))
	}
	return out
}

func (a *app) validationDialog() *ui.Element {
	body := []*ui.Element{
		fieldRow("Allow",
			a.optBtn("vtype", "list", "List", "list"), a.optBtn("vtype", "whole", "Whole number", "list"),
			a.optBtn("vtype", "decimal", "Decimal", "list"), a.optBtn("vtype", "length", "Text length", "list")),
	}
	if a.opt("vtype", "list") == "list" {
		body = append(body, fieldRow("Source (comma separated)", chadcn.Input("rb-val-list", a.opt("vlist", "Yes,No"))))
	} else {
		body = append(body, fieldRow("Minimum", chadcn.Input("rb-val-min", a.opt("vmin", "0"))),
			fieldRow("Maximum", chadcn.Input("rb-val-max", a.opt("vmax", "100"))))
	}
	body = append(body, ui.Div().Class("app-muted").Text("Applies to "+a.sel.String()+". Invalid entries are rejected."))
	return a.modal("validation", "Data Validation", "", body,
		dlgBtn("xl:val:clear", "Clear", false), dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:val:ok", "OK", true))
}

func (a *app) sortDialog() *ui.Element {
	body := []*ui.Element{
		fieldRow("Sort by column", chadcn.Input("rb-sort-k1", a.opt("sk1", ColName(a.active.Col))),
			a.optBtn("so1", "asc", "A to Z", "asc"), a.optBtn("so1", "desc", "Z to A", "asc")),
		fieldRow("Then by column", chadcn.Input("rb-sort-k2", a.opt("sk2", "")),
			a.optBtn("so2", "asc", "A to Z", "asc"), a.optBtn("so2", "desc", "Z to A", "asc")),
		fieldRow("Options", a.flagBtn("sorthdr", "My data has headers", true)),
		ui.Div().Class("app-muted").Text("Sorts " + a.sortRange().String() + ". Leave the second column empty for a single-key sort."),
	}
	return a.modal("sort", "Sort", "", body, dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:sort:apply", "OK", true))
}

func condKinds() []struct{ Kind, Label string } {
	return []struct{ Kind, Label string }{
		{"gt", "Greater than"}, {"lt", "Less than"}, {"between", "Between"}, {"eq", "Equal to"},
		{"text", "Text contains"}, {"dupes", "Duplicates"}, {"top", "Top N"}, {"bottom", "Bottom N"}, {"above", "Above average"}, {"below", "Below average"}, {"databar", "Data bar"}, {"scale3", "3-color scale"},
	}
}

func (a *app) condFormatDialog() *ui.Element {
	kinds := fieldRow("Rule")
	for _, k := range condKinds() {
		kinds.Children(a.optBtn("cfkind", k.Kind, k.Label, "gt"))
	}
	body := []*ui.Element{
		kinds,
		fieldRow("Value", chadcn.Input("rb-cf-a", a.opt("cfa", "1000")), chadcn.Input("rb-cf-b", a.opt("cfb", ""))),
		fieldRow("Format with", a.optBtn("cfpreset", "red", "Red", "red"), a.optBtn("cfpreset", "yellow", "Yellow", "red"), a.optBtn("cfpreset", "green", "Green", "red")),
	}
	list := ui.Div().Class("rb-rules").Attr("role", "list").Attr("aria-label", "Rules")
	for i, r := range a.wb.Sheet().CondRules() {
		state := "on"
		if r.Off {
			state = "off"
		}
		n := strconv.Itoa(i)
		list.Children(ui.Div().Class("rb-rule").Attr("role", "listitem").Children(
			ui.Span().Class("rb-rule-text").Text(fmt.Sprintf("%s %s %s %s [%s]", r.Rule.Kind, r.Rule.A, r.Rule.B, r.Range.String(), state)),
			ui.Button("Up").Attr("type", "button").Class("rb-opt").Attr("data-action", "xl:cfr:up:"+n),
			ui.Button("Toggle").Attr("type", "button").Class("rb-opt").Attr("data-action", "xl:cfr:toggle:"+n),
			ui.Button("Delete").Attr("type", "button").Class("rb-opt").Attr("data-action", "xl:cfr:del:"+n),
		))
	}
	body = append(body, ui.Div().Class("rb-picker-title").Text("Rules on this sheet"), list)
	return a.modal("condfmt", "Conditional Formatting Rules", "New rules apply to "+a.sel.String()+".", body,
		dlgBtn("xl:close", "Close", false), dlgBtn("xl:cf:apply", "Add Rule", true))
}

func (a *app) commentDialog() *ui.Element {
	cur := a.wb.Sheet().Comment(a.active.Row, a.active.Col)
	return a.modal("comment", "Comment on "+a.active.A1(), "", []*ui.Element{
		fieldRow("Note", chadcn.Input("rb-comment-text", cur)),
	}, dlgBtn("xl:comment:delete", "Delete", false), dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:comment:ok", "OK", true))
}

func (a *app) openDialog() *ui.Element {
	return a.modal("open", "Open", "Open a .json, .xlsx or .csv file (CSV is added as a new sheet).", []*ui.Element{
		fieldRow("File path", chadcn.Input("rb-open-path", a.opt("openpath", ""))),
		a.statusLine(),
	}, dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:file:open", "Open", true))
}

func (a *app) saveDialog() *ui.Element {
	return a.modal("save", "Save or Export", "", []*ui.Element{
		fieldRow("File path", chadcn.Input("rb-save-path", a.opt("savepath", "workbook.json"))),
		fieldRow("Format", a.optBtn("savefmt", "json", "Workbook JSON", "json"), a.optBtn("savefmt", "xlsx", "Excel .xlsx", "json"),
			a.optBtn("savefmt", "csv", "Active sheet CSV", "json")),
		a.statusLine(),
	}, dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:file:save", "Save", true))
}

func (a *app) statusLine() *ui.Element {
	return ui.Div().Class("app-muted rb-find-msg").Attr("role", "status").Text(a.x.msg)
}

// shortcutGroup is one titled block of the Keyboard Shortcuts dialog.
type shortcutGroup struct {
	Name  string
	Items [][2]string
}

func shortcutGroups() []shortcutGroup {
	return []shortcutGroup{
		{"Navigation", [][2]string{{"Arrow keys", "Move active cell"}, {"Ctrl+Home", "Go to A1"}, {"Tab / Enter", "Next cell"}}},
		{"Selection", [][2]string{{"Shift+Arrows", "Extend selection"}, {"Ctrl+A", "Select all"}}},
		{"Editing", [][2]string{{"F2", "Edit cell"}, {"Ctrl+C / X / V", "Copy, cut, paste"}, {"Ctrl+Z / Y", "Undo, redo"}, {"Ctrl+Alt+V", "Paste Special"}}},
		{"Formatting", [][2]string{{"Ctrl+B", "Bold"}, {"Ctrl+I", "Italic"}, {"Ctrl+U", "Underline"}, {"Ctrl+1", "Format Cells"}}},
		{"Data", [][2]string{{"Ctrl+F / Ctrl+H", "Find and Replace"}, {"Ctrl+Shift+L", "Toggle filter"}, {"Shift+F2", "New comment"}, {"Ctrl+/", "This list"}}},
	}
}

func (a *app) shortcutsDialog() *ui.Element {
	var body []*ui.Element
	for _, g := range shortcutGroups() {
		sec := ui.Div().Class("rb-sc-group").Children(ui.Div().Class("rb-picker-title").Text(g.Name))
		for _, it := range g.Items {
			sec.Children(ui.Div().Class("rb-sc-row").Children(chadcn.Kbd(it[0]), ui.Span().Text(it[1])))
		}
		body = append(body, sec)
	}
	return a.modal("shortcuts", "Keyboard Shortcuts", "", body, dlgBtn("xl:close", "Close", true))
}

func (a *app) filterDialog() *ui.Element {
	col := a.x.filterCol
	query := a.opt("fsearch", "")
	body := []*ui.Element{
		ui.Div().Class("rb-filter-sort").Children(
			dlgBtn("xl:filter:sort:asc", "Sort A to Z", false),
			dlgBtn("xl:filter:sort:desc", "Sort Z to A", false),
		),
		chadcn.Input("rb-fsearch", query).Attr("placeholder", "Search").Attr("aria-label", "Search values in column "+ColName(col)),
	}
	list := ui.Div().Class("rb-filter-list").Attr("role", "list").Attr("aria-label", "Values in column "+ColName(col))
	for i, v := range a.x.filterVals {
		if !filterMatches(v, query) {
			continue
		}
		on := a.x.filterAllowed[v]
		label := v
		if label == "" {
			label = "(Blanks)"
		}
		list.Children(chadcn.CheckboxWithOptions(label, on, chadcn.CheckboxOptions{ID: "rb-fv-" + strconv.Itoa(i)}).
			Attr("data-action", "xl:filter:val:"+strconv.Itoa(i)))
	}
	body = append(body, list)
	return a.modal("filter", "Filter column "+ColName(col), "", body,
		dlgBtn("xl:filter:selall", "Select All", false), dlgBtn("xl:filter:clearcol", "Clear Filter", false),
		dlgBtn("xl:close", "Cancel", false), dlgBtn("xl:filter:apply", "OK", true))
}
