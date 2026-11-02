package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

const (
	// Defaults for the widgets gallery state.
	widgetsDefaultBadges   = 3
	widgetsDefaultSteps    = 4
	widgetsDefaultProgress = 62
	widgetsDefaultPageIdx  = 2

	// Bounds for the stepper, progress bar, pagination, and table demos.
	widgetsStepsMax     = 10
	widgetsProgressMax  = 100
	widgetsProgressStep = 10
	widgetsPageCount    = 4
	widgetsTableMaxRow  = 2
)

// widgetsState holds the small interactive state for the widgets gallery tab.
type widgetsState struct {
	variant      string
	size         components.ButtonSize
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
		badges:       widgetsDefaultBadges,
		toggled:      true,
		picked:       "home",
		inputVal:     "Hello widgets",
		search:       "ownframe",
		checkNotify:  true,
		checkCompact: false,
		radio:        "medium",
		selectOpt:    "pro",
		steps:        widgetsDefaultSteps,
		progress:     widgetsDefaultProgress,
		tabSel:       "overview",
		pageIdx:      widgetsDefaultPageIdx,
		crumb:        "widgets",
		alertOff:     false,
		avatar:       "AK",
		tableRow:     1,
	}
}

func (a *builderApp) widgetsPanel() *ui.Element {
	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
		BackgroundColor(theme.Background()).
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
		ui.Text(title).FontWeight(ui.FontWeightBold).Color(theme.Text()),
		ui.P().Text(desc).Margin(ui.Edges(ui.Px(6), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
	}
	return components.Card(append(head, body...)...).Flex(ui.Number(1)).MinWidth(ui.Px(0))
}

func widgetsBtn(label, action string, primary bool) *ui.Element {
	variant := components.ButtonTonal
	if primary {
		variant = components.ButtonFilled
	}
	return components.Button(label, components.ButtonOptions{Variant: variant}).Attr("data-action", action)
}

func widgetLabel(text string) *ui.Element {
	return ui.Text(strings.ToUpper(text)).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).Color(theme.Muted())
}

func widgetDivider() *ui.Element {
	return ui.Div().Height(ui.Px(1)).Margin(ui.Edges(ui.Px(10), ui.Px(0))).BackgroundColor(theme.Control())
}

func (a *builderApp) widgetsHeader() *ui.Element {
	return ui.Div().Children(
		a.panelHeading("Component library"),
		ui.P().Text("Reusable controls, consistent states, and themes. Select a component to try its behavior.").
			Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
	)
}

func (a *builderApp) buttonsCard() *ui.Element {
	return widgetShell("Buttons", fmt.Sprintf("Selected variant: %s. Click a variant to preview it.", a.widgets.variant),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(
			a.variantButton("Primary", "primary"),
			a.variantButton("Secondary", "secondary"),
			a.variantButton("Danger", "danger"),
			a.variantButton("Text", "ghost"),
			a.variantButton("Outlined", "outlined"),
		),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).FlexWrap(ui.FlexWrapWrap).MarginTop(ui.Px(12)).Children(
			components.Button("Small", components.ButtonOptions{Size: components.ButtonSmall}).Attr("data-action", "widgets:size:small"),
			components.Button("Medium", components.ButtonOptions{}).Attr("data-action", "widgets:size:medium"),
			components.Button("Large", components.ButtonOptions{Size: components.ButtonLarge}).Attr("data-action", "widgets:size:large"),
			components.Button("Disabled", components.ButtonOptions{Disabled: true}),
		),
	)
}

func (a *builderApp) variantButton(label, variant string) *ui.Element {
	style := components.ButtonTonal
	switch variant {
	case "primary":
		style = components.ButtonFilled
	case "danger":
		style = components.ButtonDanger
	case "ghost":
		style = components.ButtonText
	case "outlined":
		style = components.ButtonOutlined
	}
	return components.Button(label, components.ButtonOptions{Variant: style, Size: a.widgets.size}).
		Attr("data-action", "widgets:variant:"+variant).Attr("aria-pressed", strconv.FormatBool(a.widgets.variant == variant))
}

func (a *builderApp) badgesCard() *ui.Element {
	return widgetShell("Badges and counters", "A pill badge bound to Go state, with increment, decrement, and reset actions.",
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(
			components.Badge(strconv.Itoa(a.widgets.badges)),
			widgetsBtn("+ Add", "widgets:badges:inc", false),
			widgetsBtn("- Remove", "widgets:badges:dec", false),
			widgetsBtn("Reset", "widgets:badges:reset", false),
		),
	)
}

func (a *builderApp) toggleCard() *ui.Element {
	return widgetShell("Switch", "A controlled, keyboard-accessible switch.",
		components.Switch("Notifications", a.widgets.toggled).ID("widgets-switch").Attr("data-action", "widgets:toggle"),
		ui.Text("State: "+toggledWord(a.widgets.toggled)).FontSize(ui.Px(12)).Color(theme.Muted()),
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
	bg := theme.Control()
	if a.widgets.picked == name {
		bg = theme.Selected()
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:nav:"+name).
		BackgroundColor(bg).Color(theme.Text()).
		Border(theme.Border()).
		Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(14)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) inputCard() *ui.Element {
	return widgetShell("Text input", fmt.Sprintf("Editable field with id widgets-input (%d chars).", len(a.widgets.inputVal)),
		ui.Div().MarginTop(ui.Px(12)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).Children(
			components.TextField("widgets-input", "Message", a.widgets.inputVal),
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
				BackgroundColor(theme.Background()).Color(theme.Text()).Border(theme.Border()).
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
	return components.Checkbox(label, checked).Attr("data-action", action).Width(ui.Percent(100))
}

func (a *builderApp) radioCard() *ui.Element {
	return widgetShell("Radio row", fmt.Sprintf("Single choice held in Go: %s.", a.widgets.radio),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(
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
	bg := theme.Control()
	if selected {
		bg = theme.Selected()
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "widgets:radio:"+value).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
		BackgroundColor(bg).Color(theme.Text()).Border(theme.Border()).
		Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(14)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold).
		Children(ui.Span().Text(dot).Color(theme.Muted()).FontSize(ui.Px(11)))
}

func (a *builderApp) selectCard() *ui.Element {
	return widgetShell("Select options", fmt.Sprintf("Plan %q picked from a select-style row.", a.widgets.selectOpt),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(
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
				BackgroundColor(theme.Background()).Color(theme.Text()).
				Border(theme.Border()).
				Padding(ui.Edges(ui.Px(8), ui.Px(16))).BorderRadius(ui.Px(8)).FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
			widgetsBtn("+", "widgets:steps:inc", false),
			widgetsBtn("Reset", "widgets:steps:reset", false),
		),
	)
}

func (a *builderApp) progressCard() *ui.Element {
	return widgetShell("Progress", "A bounded percentage with progressbar semantics.",
		components.Progress("Upload progress", a.widgets.progress).ID("widgets-progress").MarginTop(ui.Px(12)),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MarginTop(ui.Px(10)).Children(
			widgetsBtn("−10", "widgets:progress:dec", false),
			ui.Text(fmt.Sprintf("%d%%", a.widgets.progress)).Color(theme.Muted()),
			widgetsBtn("+10", "widgets:progress:inc", false),
		),
	)
}

func (a *builderApp) tabsCard() *ui.Element {
	return widgetShell("Tabs row", fmt.Sprintf("Active tab: %s.", a.widgets.tabSel),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(12)).
			BorderBottom(theme.Border()).Padding(ui.Edges(ui.Px(0), ui.Px(0), ui.Px(10))).
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
	for i := 1; i <= widgetsPageCount; i++ {
		pages = append(pages, a.pageButton(i))
	}
	pages = append(pages, widgetsBtn("Next ›", "widgets:page:next", false))
	return widgetShell("Pagination", fmt.Sprintf("Page %d of %d.", a.widgets.pageIdx, widgetsPageCount),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(pages...),
	)
}

func (a *builderApp) pageButton(n int) *ui.Element {
	return widgetsBtn(strconv.Itoa(n), "widgets:page:"+strconv.Itoa(n), a.widgets.pageIdx == n)
}

func (a *builderApp) breadcrumbCard() *ui.Element {
	return widgetShell("Breadcrumb", fmt.Sprintf("Current trail leaf: %s.", a.widgets.crumb),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).MarginTop(ui.Px(12)).FlexWrap(ui.FlexWrapWrap).Children(
			a.crumbLink("Home", "home"),
			ui.Text("/").Color(theme.Text()),
			a.crumbLink("Gallery", "gallery"),
			ui.Text("/").Color(theme.Text()),
			a.crumbLink("Widgets", "widgets"),
		),
	)
}

func (a *builderApp) crumbLink(label, value string) *ui.Element {
	return components.Button(label, components.ButtonOptions{Variant: components.ButtonText, Size: components.ButtonSmall}).
		Attr("data-action", "widgets:crumb:"+value).Attr("aria-current", strconv.FormatBool(a.widgets.crumb == value))
}

func (a *builderApp) alertCard() *ui.Element {
	title, body, action := "Component updated", "A shared notice follows the current theme.", "widgets:alert:hide"
	button := "Dismiss notice"
	if a.widgets.alertOff {
		title, body, action = "Notice dismissed", "Restore it to try the interaction again.", "widgets:alert:show"
		button = "Restore notice"
	}
	return widgetShell("Notices", "Inline feedback composed from the shared Notice control.",
		components.Notice(title, body).MarginTop(ui.Px(12)), widgetsBtn(button, action, false),
	)
}

func (a *builderApp) avatarCard() *ui.Element {
	return widgetShell("Avatar stack", fmt.Sprintf("Selected avatar: %s. Click a circle to select it.", a.widgets.avatar),
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).MarginTop(ui.Px(12)).Children(
			a.avatarCircle("AK", theme.Primary(), false),
			a.avatarCircle("JM", ui.Hex("#6b4fd0"), true),
			a.avatarCircle("RS", ui.Hex("#2f7d5b"), true),
			a.avatarCircle("+5", theme.Control(), true),
		),
	)
}

func (a *builderApp) avatarCircle(initials string, bg ui.Color, overlap bool) *ui.Element {
	el := ui.Button(initials).Attr("type", "button").Attr("data-action", "widgets:avatar:"+initials).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentCenter).
		Width(ui.Px(34)).Height(ui.Px(34)).BorderRadius(ui.Px(17)).
		BackgroundColor(bg).Color(theme.OnPrimary()).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold)
	if a.widgets.avatar == initials {
		el.Border(ui.SolidBorder(ui.Px(2), theme.Primary()))
	} else {
		el.Border(ui.SolidBorder(ui.Px(2), theme.Background()))
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
			ui.Tag("table").Width(ui.Percent(100)).BorderCollapse(ui.BorderCollapseCollapse).Children(
				ui.Tag("tbody").Children(trs...),
			),
		),
	)
}

func (a *builderApp) tableHead() *ui.Element {
	cell := func(text string) *ui.Element {
		return ui.Tag("th").Text(text).Color(theme.Muted()).FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).
			TextAlign(ui.TextAlignLeft).Padding(ui.Edges(ui.Px(6), ui.Px(8))).BorderBottom(theme.Border())
	}
	return ui.Tag("tr").Children(cell("Widget"), cell("State"), cell("Action"))
}

func (a *builderApp) tableRow(idx int, cols []string) *ui.Element {
	bg := theme.Background()
	if a.widgets.tableRow == idx {
		bg = theme.Selected()
	}
	td := func(text string) *ui.Element {
		return ui.Tag("td").Text(text).Color(theme.Text()).FontSize(ui.Px(12)).Padding(ui.Edges(ui.Px(7), ui.Px(8))).
			BorderBottom(theme.Border())
	}
	return ui.Tag("tr").BackgroundColor(bg).Children(
		td(cols[0]),
		td(cols[1]),
		ui.Tag("td").Padding(ui.Edges(ui.Px(4), ui.Px(8))).BorderBottom(theme.Border()).Children(
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
			ui.Text("Spotlight card").FontWeight(ui.FontWeightBold).Color(theme.Text()),
			ui.P().Text(fmt.Sprintf("The badge counter reads %d here too.", a.widgets.badges)).
				Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(theme.Muted()),
			widgetDivider(),
			widgetLabel("Actions"),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).FlexWrap(ui.FlexWrapWrap).Children(
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
	case "size:small", "size:medium", "size:large":
		switch name {
		case "size:small":
			a.widgets.size = components.ButtonSmall
		case "size:large":
			a.widgets.size = components.ButtonLarge
		default:
			a.widgets.size = components.ButtonMedium
		}
		a.notice = "Button size updated."
		return true
	case "variant:primary", "variant:secondary", "variant:danger", "variant:ghost", "variant:outlined":
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
		a.page.Form().SetFormValue("widgets-input", a.widgets.inputVal)
		return true
	case "input:clear":
		a.widgets.inputVal = ""
		a.notice = "Text input cleared."
		a.page.Form().SetFormValue("widgets-input", a.widgets.inputVal)
		return true
	case "search:sample":
		a.widgets.search = defaultWidgetsState().search
		a.notice = fmt.Sprintf("Search filled with %q.", a.widgets.search)
		a.page.Form().SetFormValue("widgets-search", a.widgets.search)
		return true
	case "search:clear":
		a.widgets.search = ""
		a.notice = "Search cleared."
		a.page.Form().SetFormValue("widgets-search", a.widgets.search)
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
		if a.widgets.steps < widgetsStepsMax {
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
		if a.widgets.progress > widgetsProgressMax-widgetsProgressStep {
			a.widgets.progress = widgetsProgressMax
		} else {
			a.widgets.progress += widgetsProgressStep
		}
		a.notice = fmt.Sprintf("Progress raised to %d%%.", a.widgets.progress)
		return true
	case "progress:dec":
		if a.widgets.progress < widgetsProgressStep {
			a.widgets.progress = 0
		} else {
			a.widgets.progress -= widgetsProgressStep
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
			a.notice = fmt.Sprintf("Page %d of %d.", a.widgets.pageIdx, widgetsPageCount)
			return true
		case "next":
			if a.widgets.pageIdx < widgetsPageCount {
				a.widgets.pageIdx++
			}
			a.notice = fmt.Sprintf("Page %d of %d.", a.widgets.pageIdx, widgetsPageCount)
			return true
		}
		if n, err := strconv.Atoi(rest); err == nil && n >= 1 && n <= widgetsPageCount {
			a.widgets.pageIdx = n
			a.notice = fmt.Sprintf("Page %d of %d.", n, widgetsPageCount)
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
		if n, err := strconv.Atoi(rest); err == nil && n >= 0 && n <= widgetsTableMaxRow {
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
