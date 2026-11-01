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
	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
		BackgroundColor(ui.Hex("#111827")).
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
	return ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(12)).Children(cards...)
}

func widgetShell(title, desc string, body ...*ui.Element) *ui.Element {
	head := []*ui.Element{
		ui.Text(title).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
		ui.P().Text(desc).Margin(ui.RawValue("6px 0 0")).FontSize(ui.Px(12)).Color(ui.Hex("#a9bbd1")),
	}
	return ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).
		Padding(ui.Px(14)).BorderRadius(ui.Px(11)).BackgroundColor(ui.Hex("#202d42")).
		Children(append(head, body...)...)
}

func widgetsBtn(label, action string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(fg)).Border(ui.RawValue(border)).
		FlexShrink(ui.Number(0)).
		Padding(ui.Edges(ui.Px(8), ui.Px(12))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func widgetLabel(text string) *ui.Element {
	return ui.Text(strings.ToUpper(text)).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#8fa3bd"))
}

func widgetDivider() *ui.Element {
	return ui.Div().Height(ui.Px(1)).Margin(ui.RawValue("10px 0")).BackgroundColor(ui.Hex("#34445c"))
}

func (a *builderApp) widgetsHeader() *ui.Element {
	return ui.Div().Children(
		ui.Text("Widgets gallery").FontSize(ui.Px(24)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
		ui.P().Text("Buttons, inputs, toggles, tabs, tables, and cards from the Go builder. Every control below is wired to Go.").
			Margin(ui.RawValue("7px 0 0")).Color(ui.Hex("#a9bbd1")).FontSize(ui.Px(12)),
	)
}

func (a *builderApp) buttonsCard() *ui.Element {
	return widgetShell("Buttons", fmt.Sprintf("Selected variant: %s. Click a variant to preview it.", a.widgets.variant),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(
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
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(fg)).Border(ui.RawValue(border)).
		Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(8)).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) badgesCard() *ui.Element {
	return widgetShell("Badges and counters", "A pill badge bound to Go state, with increment, decrement, and reset actions.",
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(
			ui.Span().Text(strconv.Itoa(a.widgets.badges)).
				BackgroundColor(ui.Hex("#4677ed")).Color(ui.Hex("#ffffff")).
				Padding(ui.Edges(ui.Px(5), ui.Px(12))).BorderRadius(ui.Px(12)).
				FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
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
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MarginTop(ui.Px(12)).Children(
			ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:toggle").
				BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#ffffff")).Border(ui.RawValue("1px solid #42536b")).
				Padding(ui.Edges(ui.Px(8), ui.Px(18))).BorderRadius(ui.Px(14)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
			ui.Text(fmt.Sprintf("State: %s", label)).FontSize(ui.Px(12)).Color(ui.Hex("#b5c7df")),
		),
	)
}

func (a *builderApp) widgetsNavCard() *ui.Element {
	return widgetShell("Navigation row", fmt.Sprintf("Selected destination: %s.", a.widgets.picked),
		ui.Nav().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).Children(
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
		BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
		Border(ui.RawValue("1px solid #40546f")).
		Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(14)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) inputCard() *ui.Element {
	return widgetShell("Text input", fmt.Sprintf("Editable field with id widgets-input (%d chars).", len(a.widgets.inputVal)),
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).Children(
			ui.Input().Attr("type", "text").Attr("id", "widgets-input").
				Attr("value", a.widgets.inputVal).Attr("placeholder", "Type here…").Attr("name", "widgets-input").
				BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#f1f5fd")).Border(ui.RawValue("1px solid #42536b")).
				BoxSizing(ui.BoxSizingBorderBox).
				Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).Width(ui.Percent(100)),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).Children(
				widgetsBtn("Fill sample", "widgets:input:fill", true),
				widgetsBtn("Clear", "widgets:input:clear", false),
			),
		),
	)
}

func (a *builderApp) searchCard() *ui.Element {
	return widgetShell("Search field", fmt.Sprintf("Query %q filters nothing yet — it only proves state round-trips.", a.widgets.search),
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).Children(
			ui.Input().Attr("type", "search").Attr("id", "widgets-search").
				Attr("value", a.widgets.search).Attr("placeholder", "Search widgets…").Attr("name", "widgets-search").
				BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#f1f5fd")).Border(ui.RawValue("1px solid #42536b")).
				BoxSizing(ui.BoxSizingBorderBox).
				Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).Width(ui.Percent(100)),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).Children(
				widgetsBtn("Sample", "widgets:search:sample", true),
				widgetsBtn("Clear", "widgets:search:clear", false),
			),
		),
	)
}

func (a *builderApp) checkboxCard() *ui.Element {
	return widgetShell("Checkboxes", "Two independent bools toggled through Go actions.",
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).Children(
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
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
		BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#dce6f6")).Border(ui.RawValue("1px solid #42536b")).
		BoxSizing(ui.BoxSizingBorderBox).
		Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).TextAlign(ui.TextAlignLeft).Width(ui.Percent(100)).
		Children(
			ui.Span().Text(mark).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).
				JustifyContent(ui.JustifyContentCenter).
				BackgroundColor(ui.RawValue(boxBg)).Color(ui.Hex("#ffffff")).
				Width(ui.Px(18)).Height(ui.Px(18)).Border(ui.RawValue("1px solid #5b6f8c")).BorderRadius(ui.Px(5)).
				FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
			ui.Text(label).Color(ui.Hex("#dce6f6")),
		)
}

func (a *builderApp) radioCard() *ui.Element {
	return widgetShell("Radio row", fmt.Sprintf("Single choice held in Go: %s.", a.widgets.radio),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(
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
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
		BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).Border(ui.RawValue("1px solid #40546f")).
		Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(14)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold).
		Children(ui.Span().Text(dot).Color(ui.Hex("#9db9f5")).FontSize(ui.Px(11)))
}

func (a *builderApp) selectCard() *ui.Element {
	return widgetShell("Select options", fmt.Sprintf("Plan %q picked from a select-style row.", a.widgets.selectOpt),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(
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
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MarginTop(ui.Px(12)).Children(
			widgetsBtn("−", "widgets:steps:dec", false),
			ui.Span().Text(strconv.Itoa(a.widgets.steps)).
				BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#f4f7ff")).
				Border(ui.RawValue("1px solid #42536b")).
				Padding(ui.Edges(ui.Px(8), ui.Px(16))).BorderRadius(ui.Px(8)).FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
			widgetsBtn("+", "widgets:steps:inc", false),
			widgetsBtn("Reset", "widgets:steps:reset", false),
		),
	)
}

func (a *builderApp) progressCard() *ui.Element {
	width := fmt.Sprintf("%d%%", a.widgets.progress)
	return widgetShell("Progress bar", fmt.Sprintf("Static width %s from Go state — no animation.", width),
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).Children(
			ui.Div().Width(ui.Percent(100)).BoxSizing(ui.BoxSizingBorderBox).BackgroundColor(ui.Hex("#2b3950")).BorderRadius(ui.Px(6)).
				Padding(ui.Px(3)).Children(
				ui.Div().Width(ui.RawValue(width)).Height(ui.Px(8)).
					BackgroundColor(ui.Hex("#4677ed")).BorderRadius(ui.Px(6)),
			),
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).Children(
				widgetsBtn("−10", "widgets:progress:dec", false),
				ui.Text(width).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#b5c7df")),
				widgetsBtn("+10", "widgets:progress:inc", false),
			),
		),
	)
}

func (a *builderApp) tabsCard() *ui.Element {
	return widgetShell("Tabs row", fmt.Sprintf("Active tab: %s.", a.widgets.tabSel),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(12)).
			BorderBottom(ui.RawValue("1px solid #34445c")).Padding(ui.RawValue("0 0 10px")).
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
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(pages...),
	)
}

func (a *builderApp) pageButton(n int) *ui.Element {
	return widgetsBtn(strconv.Itoa(n), "widgets:page:"+strconv.Itoa(n), a.widgets.pageIdx == n)
}

func (a *builderApp) breadcrumbCard() *ui.Element {
	return widgetShell("Breadcrumb", fmt.Sprintf("Current trail leaf: %s.", a.widgets.crumb),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).MarginTop(ui.Px(12)).FlexWrap(ui.RawValue("wrap")).Children(
			a.crumbLink("Home", "home"),
			ui.Text("/").Color(ui.Hex("#5b6f8c")),
			a.crumbLink("Gallery", "gallery"),
			ui.Text("/").Color(ui.Hex("#5b6f8c")),
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
		BackgroundColor(ui.Hex("#111827")).Color(ui.RawValue(color)).Border(ui.RawValue("1px solid #34445c")).
		Padding(ui.Edges(ui.Px(7), ui.Px(12))).BorderRadius(ui.Px(7)).FontSize(ui.Px(12)).FontWeight(ui.RawValue(weight))
}

func (a *builderApp) alertCard() *ui.Element {
	banners := []*ui.Element{
		a.alertBanner("Info", "Gallery state lives in Go.", "#1e3a5f", "#4677ed"),
		a.alertBanner("Success", "Last action saved.", "#1d3b2a", "#3fa66a"),
		a.alertBanner("Warning", "Tab now scrolls past 760px.", "#3d2f1a", "#d9a13b"),
	}
	body := ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(banners...)
	if a.widgets.alertOff {
		return widgetShell("Alerts and banners", "Demo banner dismissed — restore it below.",
			body,
			ui.Div().MarginTop(ui.Px(10)).Children(widgetsBtn("Show demo banner", "widgets:alert:show", true)),
		)
	}
	return widgetShell("Alerts and banners", "Info, success, and warning variants plus one dismissible banner.",
		body,
		ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).MarginTop(ui.Px(8)).Children(
			a.alertBanner("Heads up (dismissible)", "Click hide to flip Go state.", "#1e3a5f", "#4677ed"),
			ui.Div().Children(widgetsBtn("Hide demo banner", "widgets:alert:hide", false)),
		),
	)
}

func (a *builderApp) alertBanner(title, body, bg, edge string) *ui.Element {
	return ui.Div().BorderLeft(ui.RawValue("3px solid "+edge)).
		BackgroundColor(ui.RawValue(bg)).Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(7)).Children(
		ui.Text(title).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
		ui.P().Text(body).Margin(ui.RawValue("3px 0 0")).FontSize(ui.Px(11)).Color(ui.Hex("#c6d4e8")),
	)
}

func (a *builderApp) avatarCard() *ui.Element {
	return widgetShell("Avatar stack", fmt.Sprintf("Selected avatar: %s. Click a circle to select it.", a.widgets.avatar),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).MarginTop(ui.Px(12)).Children(
			a.avatarCircle("AK", "#4677ed", false),
			a.avatarCircle("JM", "#6b4fd0", true),
			a.avatarCircle("RS", "#2f7d5b", true),
			a.avatarCircle("+5", "#2b3950", true),
		),
	)
}

func (a *builderApp) avatarCircle(initials, bg string, overlap bool) *ui.Element {
	el := ui.Button(initials).Attr("type", "button").Attr("data-action", "widgets:avatar:"+initials).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentCenter).
		Width(ui.Px(34)).Height(ui.Px(34)).BorderRadius(ui.Px(17)).
		BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#ffffff")).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold)
	if a.widgets.avatar == initials {
		el.Border(ui.RawValue("2px solid #9db9f5"))
	} else {
		el.Border(ui.RawValue("2px solid #111827"))
	}
	if overlap {
		el.MarginLeft(ui.Px(-10))
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
		ui.Div().MarginTop(ui.Px(12)).Children(
			ui.Tag("table").Width(ui.Percent(100)).BorderCollapse(ui.RawValue("collapse")).Children(
				ui.Tag("tbody").Children(trs...),
			),
		),
	)
}

func (a *builderApp) tableHead() *ui.Element {
	cell := func(text string) *ui.Element {
		return ui.Tag("th").Text(text).Color(ui.Hex("#8fa3bd")).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).
			TextAlign(ui.TextAlignLeft).Padding(ui.Edges(ui.Px(6), ui.Px(8))).BorderBottom(ui.RawValue("1px solid #34445c"))
	}
	return ui.Tag("tr").Children(cell("Widget"), cell("State"), cell("Action"))
}

func (a *builderApp) tableRow(idx int, cols []string) *ui.Element {
	bg := "#111827"
	if a.widgets.tableRow == idx {
		bg = "#355596"
	}
	td := func(text string) *ui.Element {
		return ui.Tag("td").Text(text).Color(ui.Hex("#dce6f6")).FontSize(ui.Px(12)).Padding(ui.Edges(ui.Px(7), ui.Px(8))).
			BorderBottom(ui.RawValue("1px solid #2c3a52"))
	}
	return ui.Tag("tr").BackgroundColor(ui.RawValue(bg)).Children(
		td(cols[0]),
		td(cols[1]),
		ui.Tag("td").Padding(ui.Edges(ui.Px(4), ui.Px(8))).BorderBottom(ui.RawValue("1px solid #2c3a52")).Children(
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
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
			widgetLabel("Featured"),
			ui.Text("Spotlight card").FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
			ui.P().Text(fmt.Sprintf("The badge counter reads %d here too.", a.widgets.badges)).
				Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(ui.Hex("#abbcd2")),
			widgetDivider(),
			widgetLabel("Actions"),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).FlexWrap(ui.RawValue("wrap")).Children(
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
