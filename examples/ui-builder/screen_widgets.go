package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// widgetsState holds the small interactive state for the widgets gallery tab.
type widgetsState struct {
	variant      string // selected button variant showcase
	badges       int    // badge counter value
	toggled      bool   // toggle switch state
	picked       string // selected nav item
	inputVal     string // text input value
	search       string // search field value
	checkNotify  bool   // checkbox row: notifications
	checkCompact bool   // checkbox row: compact mode
	radio        string // radio row selection
	selectOpt    string // select-style option selection
	steps        int    // stepper value 0..10
	progress     int    // progress bar percent 0..100
	tabSel       string // tabs row selection
	pageIdx      int    // pagination current page 1..4
	crumb        string // breadcrumb selection
	alertOff     bool   // dismissible demo banner hidden
	avatar       string // selected avatar initials
	tableRow     int    // selected data-table row 0..2
}

func defaultWidgetsState() widgetsState {
	return widgetsState{
		variant:      "primary",
		badges:       3,
		toggled:      true,
		picked:       "home",
		inputVal:     "Hello widgets",
		search:       "ownframe",
		checkNotify:  true,
		checkCompact: false,
		radio:        "medium",
		selectOpt:    "pro",
		steps:        4,
		progress:     62,
		tabSel:       "overview",
		pageIdx:      2,
		crumb:        "widgets",
		alertOff:     false,
		avatar:       "AK",
		tableRow:     1,
	}
}

func (a *builderApp) widgetsPanel() *ui.Element {
	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("12px").
		BackgroundColor("#111827").
		Children(
			a.widgetsHeader(),
			widgetsRow(a.buttonsCard(), a.badgesCard()),
			widgetsRow(a.toggleCard(), a.widgetsNavCard()),
			widgetsRow(a.inputCard(), a.searchCard()),
			widgetsRow(a.checkboxCard(), a.radioCard()),
			widgetsRow(a.selectCard(), a.stepperCard()),
			widgetsRow(a.progressCard(), a.tabsCard()),
			widgetsRow(a.paginationCard(), a.breadcrumbCard()),
			widgetsRow(a.alertCard(), a.avatarCard()),
			widgetsRow(a.tableCard(), a.spotlightCard()),
		)
}

func widgetsRow(cards ...*ui.Element) *ui.Element {
	return ui.Div().Display("flex").Gap("12px").Children(cards...)
}

func widgetShell(title, desc string, body ...*ui.Element) *ui.Element {
	head := []*ui.Element{
		ui.Text(title).FontWeight("bold").Color("#f4f7ff"),
		ui.P().Text(desc).Margin("6px 0 0").FontSize("12px").Color("#a9bbd1"),
	}
	return ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		Padding("14px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(append(head, body...)...)
}

func widgetsBtn(label, action string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		BackgroundColor(bg).Color(fg).CSS("border", border).
		CSS("flex-shrink", "0").
		Padding("8px 12px").BorderRadius("8px").FontSize("12px").FontWeight("bold")
}

func widgetLabel(text string) *ui.Element {
	return ui.Text(strings.ToUpper(text)).FontSize("11px").FontWeight("bold").Color("#8fa3bd")
}

func widgetDivider() *ui.Element {
	return ui.Div().Height("1px").CSS("margin", "10px 0").BackgroundColor("#34445c")
}

func (a *builderApp) widgetsHeader() *ui.Element {
	return ui.Div().Children(
		ui.Text("Widgets gallery").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
		ui.P().Text("Buttons, inputs, toggles, tabs, tables, and cards from the Go builder. Every control below is wired to Go.").
			Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
	)
}

func (a *builderApp) buttonsCard() *ui.Element {
	return widgetShell("Buttons", fmt.Sprintf("Selected variant: %s. Click a variant to preview it.", a.widgets.variant),
		ui.Div().Display("flex").Gap("10px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(
			a.variantButton("Primary", "primary"),
			a.variantButton("Secondary", "secondary"),
			a.variantButton("Danger", "danger"),
			a.variantButton("Ghost", "ghost"),
		),
	)
}

func (a *builderApp) variantButton(label, variant string) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	switch variant {
	case "primary":
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	case "danger":
		bg, fg, border = "#b3363f", "#ffffff", "1px solid #b3363f"
	case "ghost":
		bg, fg, border = "#202d42", "#dce6f6", "1px solid #42536b"
	}
	if a.widgets.variant == variant {
		border = "2px solid #9db9f5"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:variant:"+variant).
		BackgroundColor(bg).Color(fg).CSS("border", border).
		Padding("9px 14px").BorderRadius("8px").
		FontSize("12px").FontWeight("bold")
}

func (a *builderApp) badgesCard() *ui.Element {
	return widgetShell("Badges and counters", "A pill badge bound to Go state, with increment, decrement, and reset actions.",
		ui.Div().Display("flex").AlignItems("center").Gap("10px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(
			ui.Span().Text(strconv.Itoa(a.widgets.badges)).
				BackgroundColor("#4677ed").Color("#ffffff").
				Padding("5px 12px").BorderRadius("12px").
				FontSize("12px").FontWeight("bold"),
			widgetsBtn("+ Add", "widgets:badges:inc", false),
			widgetsBtn("- Remove", "widgets:badges:dec", false),
			widgetsBtn("Reset", "widgets:badges:reset", false),
		),
	)
}

func (a *builderApp) toggleCard() *ui.Element {
	label, bg := "Off", "#2b3950"
	if a.widgets.toggled {
		label, bg = "On", "#4677ed"
	}
	return widgetShell("Toggle", "A single on/off switch held in Go state.",
		ui.Div().Display("flex").AlignItems("center").Gap("10px").CSS("margin-top", "12px").Children(
			ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:toggle").
				BackgroundColor(bg).Color("#ffffff").CSS("border", "1px solid #42536b").
				Padding("8px 18px").BorderRadius("14px").FontSize("12px").FontWeight("bold"),
			ui.Text(fmt.Sprintf("State: %s", label)).FontSize("12px").Color("#b5c7df"),
		),
	)
}

func (a *builderApp) widgetsNavCard() *ui.Element {
	return widgetShell("Navigation row", fmt.Sprintf("Selected destination: %s.", a.widgets.picked),
		ui.Nav().Display("flex").Gap("10px").CSS("margin-top", "12px").Children(
			a.navPill("Home", "home"),
			a.navPill("Search", "search"),
			a.navPill("Profile", "profile"),
		),
	)
}

func (a *builderApp) navPill(label, name string) *ui.Element {
	bg := "#2b3950"
	if a.widgets.picked == name {
		bg = "#355596"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:nav:"+name).
		BackgroundColor(bg).Color("#f1f5fd").
		CSS("border", "1px solid #40546f").
		Padding("9px 14px").BorderRadius("14px").FontSize("12px").FontWeight("bold")
}

func (a *builderApp) inputCard() *ui.Element {
	return widgetShell("Text input", fmt.Sprintf("Editable field with id widgets-input (%d chars).", len(a.widgets.inputVal)),
		ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("10px").Children(
			ui.Input().Attr("type", "text").Attr("id", "widgets-input").
				Attr("value", a.widgets.inputVal).Attr("placeholder", "Type here…").Attr("name", "widgets-input").
				BackgroundColor("#111827").Color("#f1f5fd").CSS("border", "1px solid #42536b").
				CSS("box-sizing", "border-box").
				Padding("9px 12px").BorderRadius("8px").FontSize("12px").Width("100%"),
			ui.Div().Display("flex").Gap("10px").Children(
				widgetsBtn("Fill sample", "widgets:input:fill", true),
				widgetsBtn("Clear", "widgets:input:clear", false),
			),
		),
	)
}

func (a *builderApp) searchCard() *ui.Element {
	return widgetShell("Search field", fmt.Sprintf("Query %q filters nothing yet — it only proves state round-trips.", a.widgets.search),
		ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("10px").Children(
			ui.Input().Attr("type", "search").Attr("id", "widgets-search").
				Attr("value", a.widgets.search).Attr("placeholder", "Search widgets…").Attr("name", "widgets-search").
				BackgroundColor("#111827").Color("#f1f5fd").CSS("border", "1px solid #42536b").
				CSS("box-sizing", "border-box").
				Padding("9px 12px").BorderRadius("8px").FontSize("12px").Width("100%"),
			ui.Div().Display("flex").Gap("10px").Children(
				widgetsBtn("Sample", "widgets:search:sample", true),
				widgetsBtn("Clear", "widgets:search:clear", false),
			),
		),
	)
}

func (a *builderApp) checkboxCard() *ui.Element {
	return widgetShell("Checkboxes", "Two independent bools toggled through Go actions.",
		ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("10px").Children(
			a.checkRow("Enable notifications", a.widgets.checkNotify, "widgets:check:notify"),
			a.checkRow("Compact mode", a.widgets.checkCompact, "widgets:check:compact"),
		),
	)
}

func (a *builderApp) checkRow(label string, checked bool, action string) *ui.Element {
	boxBg, mark := "#111827", ""
	if checked {
		boxBg, mark = "#4677ed", "✓"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		Display("flex").AlignItems("center").Gap("10px").
		BackgroundColor("#111827").Color("#dce6f6").CSS("border", "1px solid #42536b").
		CSS("box-sizing", "border-box").
		Padding("9px 12px").BorderRadius("8px").FontSize("12px").CSS("text-align", "left").Width("100%").
		Children(
			ui.Span().Text(mark).Display("flex").AlignItems("center").
				CSS("justify-content", "center").
				BackgroundColor(boxBg).Color("#ffffff").
				Width("18px").Height("18px").CSS("border", "1px solid #5b6f8c").BorderRadius("5px").
				FontSize("12px").FontWeight("bold"),
			ui.Text(label).Color("#dce6f6"),
		)
}

func (a *builderApp) radioCard() *ui.Element {
	return widgetShell("Radio row", fmt.Sprintf("Single choice held in Go: %s.", a.widgets.radio),
		ui.Div().Display("flex").Gap("10px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(
			a.radioPill("Small", "small"),
			a.radioPill("Medium", "medium"),
			a.radioPill("Large", "large"),
		),
	)
}

func (a *builderApp) radioPill(label, value string) *ui.Element {
	selected := a.widgets.radio == value
	dot := ""
	if selected {
		dot = "●"
	}
	bg := "#2b3950"
	if selected {
		bg = "#355596"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:radio:"+value).
		Display("flex").AlignItems("center").Gap("8px").
		BackgroundColor(bg).Color("#f1f5fd").CSS("border", "1px solid #40546f").
		Padding("9px 14px").BorderRadius("14px").FontSize("12px").FontWeight("bold").
		Children(ui.Span().Text(dot).Color("#9db9f5").FontSize("11px"))
}

func (a *builderApp) selectCard() *ui.Element {
	return widgetShell("Select options", fmt.Sprintf("Plan %q picked from a select-style row.", a.widgets.selectOpt),
		ui.Div().Display("flex").Gap("10px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(
			a.selectPill("Starter", "starter"),
			a.selectPill("Pro", "pro"),
			a.selectPill("Team", "team"),
		),
	)
}

func (a *builderApp) selectPill(label, value string) *ui.Element {
	selected := a.widgets.selectOpt == value
	prefix := "○"
	if selected {
		prefix = "◉"
	}
	return widgetsBtn(prefix+" "+label, "widgets:select:"+value, selected)
}

func (a *builderApp) stepperCard() *ui.Element {
	return widgetShell("Stepper", "Slider-like control built from -/value/+ buttons (0–10).",
		ui.Div().Display("flex").AlignItems("center").Gap("10px").CSS("margin-top", "12px").Children(
			widgetsBtn("−", "widgets:steps:dec", false),
			ui.Span().Text(strconv.Itoa(a.widgets.steps)).
				BackgroundColor("#111827").Color("#f4f7ff").
				CSS("border", "1px solid #42536b").
				Padding("8px 16px").BorderRadius("8px").FontSize("13px").FontWeight("bold"),
			widgetsBtn("+", "widgets:steps:inc", false),
			widgetsBtn("Reset", "widgets:steps:reset", false),
		),
	)
}

func (a *builderApp) progressCard() *ui.Element {
	width := fmt.Sprintf("%d%%", a.widgets.progress)
	return widgetShell("Progress bar", fmt.Sprintf("Static width %s from Go state — no animation.", width),
		ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("10px").Children(
			ui.Div().Width("100%").CSS("box-sizing", "border-box").BackgroundColor("#2b3950").BorderRadius("6px").
				Padding("3px").Children(
				ui.Div().CSS("width", width).Height("8px").
					BackgroundColor("#4677ed").BorderRadius("6px"),
			),
			ui.Div().Display("flex").AlignItems("center").Gap("10px").Children(
				widgetsBtn("−10", "widgets:progress:dec", false),
				ui.Text(width).FontSize("12px").FontWeight("bold").Color("#b5c7df"),
				widgetsBtn("+10", "widgets:progress:inc", false),
			),
		),
	)
}

func (a *builderApp) tabsCard() *ui.Element {
	return widgetShell("Tabs row", fmt.Sprintf("Active tab: %s.", a.widgets.tabSel),
		ui.Div().Display("flex").Gap("8px").CSS("margin-top", "12px").
			CSS("border-bottom", "1px solid #34445c").Padding("0 0 10px").
			Children(
				a.tabPill("Overview", "overview"),
				a.tabPill("Activity", "activity"),
				a.tabPill("Settings", "settings"),
			),
	)
}

func (a *builderApp) tabPill(label, value string) *ui.Element {
	return widgetsBtn(label, "widgets:tabs:"+value, a.widgets.tabSel == value)
}

func (a *builderApp) paginationCard() *ui.Element {
	pages := []*ui.Element{widgetsBtn("‹ Prev", "widgets:page:prev", false)}
	for i := 1; i <= 4; i++ {
		pages = append(pages, a.pageButton(i))
	}
	pages = append(pages, widgetsBtn("Next ›", "widgets:page:next", false))
	return widgetShell("Pagination", fmt.Sprintf("Page %d of 4.", a.widgets.pageIdx),
		ui.Div().Display("flex").AlignItems("center").Gap("8px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(pages...),
	)
}

func (a *builderApp) pageButton(n int) *ui.Element {
	return widgetsBtn(strconv.Itoa(n), "widgets:page:"+strconv.Itoa(n), a.widgets.pageIdx == n)
}

func (a *builderApp) breadcrumbCard() *ui.Element {
	return widgetShell("Breadcrumb", fmt.Sprintf("Current trail leaf: %s.", a.widgets.crumb),
		ui.Div().Display("flex").AlignItems("center").Gap("8px").CSS("margin-top", "12px").CSS("flex-wrap", "wrap").Children(
			a.crumbLink("Home", "home"),
			ui.Text("/").Color("#5b6f8c"),
			a.crumbLink("Gallery", "gallery"),
			ui.Text("/").Color("#5b6f8c"),
			a.crumbLink("Widgets", "widgets"),
		),
	)
}

func (a *builderApp) crumbLink(label, value string) *ui.Element {
	color, weight := "#8fa3bd", "normal"
	if a.widgets.crumb == value {
		color, weight = "#ffffff", "bold"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:crumb:"+value).
		BackgroundColor("#111827").Color(color).CSS("border", "1px solid #34445c").
		Padding("7px 12px").BorderRadius("7px").FontSize("12px").FontWeight(weight)
}

func (a *builderApp) alertCard() *ui.Element {
	banners := []*ui.Element{
		a.alertBanner("Info", "Gallery state lives in Go.", "#1e3a5f", "#4677ed"),
		a.alertBanner("Success", "Last action saved.", "#1d3b2a", "#3fa66a"),
		a.alertBanner("Warning", "Tab now scrolls past 760px.", "#3d2f1a", "#d9a13b"),
	}
	body := ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("8px").Children(banners...)
	if a.widgets.alertOff {
		return widgetShell("Alerts and banners", "Demo banner dismissed — restore it below.",
			body,
			ui.Div().CSS("margin-top", "10px").Children(widgetsBtn("Show demo banner", "widgets:alert:show", true)),
		)
	}
	return widgetShell("Alerts and banners", "Info, success, and warning variants plus one dismissible banner.",
		body,
		ui.Div().Display("flex").FlexDirection("column").Gap("8px").CSS("margin-top", "8px").Children(
			a.alertBanner("Heads up (dismissible)", "Click hide to flip Go state.", "#1e3a5f", "#4677ed"),
			ui.Div().Children(widgetsBtn("Hide demo banner", "widgets:alert:hide", false)),
		),
	)
}

func (a *builderApp) alertBanner(title, body, bg, edge string) *ui.Element {
	return ui.Div().CSS("border-left", "3px solid "+edge).
		BackgroundColor(bg).Padding("9px 12px").BorderRadius("7px").Children(
		ui.Text(title).FontSize("12px").FontWeight("bold").Color("#f4f7ff"),
		ui.P().Text(body).Margin("3px 0 0").FontSize("11px").Color("#c6d4e8"),
	)
}

func (a *builderApp) avatarCard() *ui.Element {
	return widgetShell("Avatar stack", fmt.Sprintf("Selected avatar: %s. Click a circle to select it.", a.widgets.avatar),
		ui.Div().Display("flex").AlignItems("center").CSS("margin-top", "12px").Children(
			a.avatarCircle("AK", "#4677ed", false),
			a.avatarCircle("JM", "#6b4fd0", true),
			a.avatarCircle("RS", "#2f7d5b", true),
			a.avatarCircle("+5", "#2b3950", true),
		),
	)
}

func (a *builderApp) avatarCircle(initials, bg string, overlap bool) *ui.Element {
	el := ui.Button(initials).Attr("type", "button").Attr("data-action", "widgets:avatar:"+initials).
		Display("flex").AlignItems("center").CSS("justify-content", "center").
		Width("34px").Height("34px").BorderRadius("17px").
		BackgroundColor(bg).Color("#ffffff").FontSize("11px").FontWeight("bold")
	if a.widgets.avatar == initials {
		el.CSS("border", "2px solid #9db9f5")
	} else {
		el.CSS("border", "2px solid #111827")
	}
	if overlap {
		el.CSS("margin-left", "-10px")
	}
	return el
}

func (a *builderApp) tableCard() *ui.Element {
	rows := [][]string{
		{"Buttons", a.widgets.variant, "Open"},
		{"Badges", strconv.Itoa(a.widgets.badges), "Inspect"},
		{"Toggle", toggledWord(a.widgets.toggled), "View"},
	}
	trs := []*ui.Element{a.tableHead()}
	for i, r := range rows {
		trs = append(trs, a.tableRow(i, r))
	}
	return widgetShell("Data table", fmt.Sprintf("Mini 3×3 table — row %d selected.", a.widgets.tableRow+1),
		ui.Div().CSS("margin-top", "12px").Children(
			ui.Tag("table").CSS("width", "100%").CSS("border-collapse", "collapse").Children(
				ui.Tag("tbody").Children(trs...),
			),
		),
	)
}

func (a *builderApp) tableHead() *ui.Element {
	cell := func(text string) *ui.Element {
		return ui.Tag("th").Text(text).Color("#8fa3bd").FontSize("11px").FontWeight("bold").
			CSS("text-align", "left").Padding("6px 8px").CSS("border-bottom", "1px solid #34445c")
	}
	return ui.Tag("tr").Children(cell("Widget"), cell("State"), cell("Action"))
}

func (a *builderApp) tableRow(idx int, cols []string) *ui.Element {
	bg := "#111827"
	if a.widgets.tableRow == idx {
		bg = "#355596"
	}
	td := func(text string) *ui.Element {
		return ui.Tag("td").Text(text).Color("#dce6f6").FontSize("12px").Padding("7px 8px").
			CSS("border-bottom", "1px solid #2c3a52")
	}
	return ui.Tag("tr").BackgroundColor(bg).Children(
		td(cols[0]),
		td(cols[1]),
		ui.Tag("td").Padding("4px 8px").CSS("border-bottom", "1px solid #2c3a52").Children(
			widgetsBtn(cols[2], "widgets:table:"+strconv.Itoa(idx), false),
		),
	)
}

func toggledWord(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

func (a *builderApp) spotlightCard() *ui.Element {
	return widgetShell("Composed cards", "Cards compose text and one action without extra state.",
		ui.Div().CSS("margin-top", "12px").Display("flex").FlexDirection("column").Gap("8px").Children(
			widgetLabel("Featured"),
			ui.Text("Spotlight card").FontWeight("bold").Color("#f4f7ff"),
			ui.P().Text(fmt.Sprintf("The badge counter reads %d here too.", a.widgets.badges)).
				Margin("0").FontSize("12px").Color("#abbcd2"),
			widgetDivider(),
			widgetLabel("Actions"),
			ui.Div().Display("flex").Gap("10px").CSS("flex-wrap", "wrap").Children(
				widgetsBtn("Open", "widgets:card:spotlight", true),
				widgetsBtn("Inspect", "widgets:card:compact", false),
			),
		),
	)
}

func (a *builderApp) handleWidgets(action string) bool {
	name, ok := strings.CutPrefix(action, "widgets:")
	if !ok {
		return false
	}
	switch name {
	case "variant:primary", "variant:secondary", "variant:danger", "variant:ghost":
		a.widgets.variant = strings.TrimPrefix(name, "variant:")
		a.notice = fmt.Sprintf("Button variant %q selected.", a.widgets.variant)
		return true
	case "badges:inc":
		a.widgets.badges++
		a.notice = fmt.Sprintf("Badge counter raised to %d.", a.widgets.badges)
		return true
	case "badges:dec":
		if a.widgets.badges > 0 {
			a.widgets.badges--
		}
		a.notice = fmt.Sprintf("Badge counter lowered to %d.", a.widgets.badges)
		return true
	case "badges:reset":
		a.widgets.badges = defaultWidgetsState().badges
		a.notice = "Badge counter reset to its default."
		return true
	case "toggle":
		a.widgets.toggled = !a.widgets.toggled
		a.notice = fmt.Sprintf("Toggle switched %s.", toggledWord(a.widgets.toggled))
		return true
	case "input:fill":
		a.widgets.inputVal = defaultWidgetsState().inputVal
		a.notice = "Text input filled with the sample value."
		return true
	case "input:clear":
		a.widgets.inputVal = ""
		a.notice = "Text input cleared."
		return true
	case "search:sample":
		a.widgets.search = defaultWidgetsState().search
		a.notice = fmt.Sprintf("Search filled with %q.", a.widgets.search)
		return true
	case "search:clear":
		a.widgets.search = ""
		a.notice = "Search cleared."
		return true
	case "check:notify":
		a.widgets.checkNotify = !a.widgets.checkNotify
		a.notice = fmt.Sprintf("Notifications checkbox %s.", toggledWord(a.widgets.checkNotify))
		return true
	case "check:compact":
		a.widgets.checkCompact = !a.widgets.checkCompact
		a.notice = fmt.Sprintf("Compact mode checkbox %s.", toggledWord(a.widgets.checkCompact))
		return true
	case "steps:inc":
		if a.widgets.steps < 10 {
			a.widgets.steps++
		}
		a.notice = fmt.Sprintf("Stepper raised to %d.", a.widgets.steps)
		return true
	case "steps:dec":
		if a.widgets.steps > 0 {
			a.widgets.steps--
		}
		a.notice = fmt.Sprintf("Stepper lowered to %d.", a.widgets.steps)
		return true
	case "steps:reset":
		a.widgets.steps = defaultWidgetsState().steps
		a.notice = "Stepper reset to its default."
		return true
	case "progress:inc":
		if a.widgets.progress > 90 {
			a.widgets.progress = 100
		} else {
			a.widgets.progress += 10
		}
		a.notice = fmt.Sprintf("Progress raised to %d%%.", a.widgets.progress)
		return true
	case "progress:dec":
		if a.widgets.progress < 10 {
			a.widgets.progress = 0
		} else {
			a.widgets.progress -= 10
		}
		a.notice = fmt.Sprintf("Progress lowered to %d%%.", a.widgets.progress)
		return true
	case "alert:hide":
		a.widgets.alertOff = true
		a.notice = "Demo banner dismissed."
		return true
	case "alert:show":
		a.widgets.alertOff = false
		a.notice = "Demo banner restored."
		return true
	case "":
		return false
	}
	if rest, ok := strings.CutPrefix(name, "nav:"); ok {
		switch rest {
		case "home", "search", "profile":
			a.widgets.picked = rest
			a.notice = fmt.Sprintf("Navigated to %s (demo selection).", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "radio:"); ok {
		switch rest {
		case "small", "medium", "large":
			a.widgets.radio = rest
			a.notice = fmt.Sprintf("Radio size %q picked.", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "select:"); ok {
		switch rest {
		case "starter", "pro", "team":
			a.widgets.selectOpt = rest
			a.notice = fmt.Sprintf("Plan %q selected.", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "tabs:"); ok {
		switch rest {
		case "overview", "activity", "settings":
			a.widgets.tabSel = rest
			a.notice = fmt.Sprintf("Tab %q active.", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "page:"); ok {
		switch rest {
		case "prev":
			if a.widgets.pageIdx > 1 {
				a.widgets.pageIdx--
			}
			a.notice = fmt.Sprintf("Page %d of 4.", a.widgets.pageIdx)
			return true
		case "next":
			if a.widgets.pageIdx < 4 {
				a.widgets.pageIdx++
			}
			a.notice = fmt.Sprintf("Page %d of 4.", a.widgets.pageIdx)
			return true
		}
		if n, err := strconv.Atoi(rest); err == nil && n >= 1 && n <= 4 {
			a.widgets.pageIdx = n
			a.notice = fmt.Sprintf("Page %d of 4.", n)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "crumb:"); ok {
		switch rest {
		case "home", "gallery", "widgets":
			a.widgets.crumb = rest
			a.notice = fmt.Sprintf("Breadcrumb leaf %q.", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "avatar:"); ok {
		switch rest {
		case "AK", "JM", "RS", "+5":
			a.widgets.avatar = rest
			a.notice = fmt.Sprintf("Avatar %q selected.", rest)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "table:"); ok {
		if n, err := strconv.Atoi(rest); err == nil && n >= 0 && n <= 2 {
			a.widgets.tableRow = n
			a.notice = fmt.Sprintf("Table row %d selected.", n+1)
			return true
		}
		return false
	}
	if rest, ok := strings.CutPrefix(name, "card:"); ok {
		switch rest {
		case "buttons", "badges", "toggle", "spotlight", "compact":
			a.notice = fmt.Sprintf("Card %q opened (demo action).", rest)
			return true
		}
		return false
	}
	return false
}
