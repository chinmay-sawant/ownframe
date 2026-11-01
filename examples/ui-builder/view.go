package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *builderApp) screen() *ui.Element {
	var center *ui.Element
	switch a.tab {
	case "settings":
		center = a.settingsPanel()
	case "layout":
		center = a.layoutPanel()
	case "todo":
		center = a.todoPanel()
	case "list":
		center = a.listPanel()
	case "theme":
		center = a.themePanel()
	case "compare":
		center = a.comparePanel()
	case "dialog":
		center = a.dialogPanel()
	default:
		center = a.widgetsPanel()
	}
	return ui.Div().Width(ui.Percent(100)).MinHeight(ui.RawValue("100vh")).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#e9f0fa")).
		FontFamily(ui.RawValue("sans-serif")).FontSize(ui.Px(14)).
		Children(
			a.titlebar(),
			ui.Div().Display(ui.DisplayFlex).Flex(ui.RawValue("1")).
				MinHeight(ui.Px(0)).MinWidth(ui.Px(0)).Gap(ui.Px(builderContentGap)).Padding(ui.Px(builderPanePadding)).
				Children(a.sidebar(), a.scrollViewport(center)),
			ui.Div().Height(ui.Px(builderFooterHeight)).FlexShrink(ui.Number(0)).
				Padding(ui.Edges(ui.Px(9), ui.Px(18))).BackgroundColor(ui.Hex("#182336")).
				Color(ui.Hex("#a4b5ca")).FontSize(ui.Px(12)).
				Children(ui.Text(a.notice)),
			a.dialogOverlay(),
			a.comparePreviewOverlay(),
		)
}

func (a *builderApp) panelHeading(title string) *ui.Element {
	return ui.Div().ID(builderActivePanelPrefix + a.tab).
		FontSize(ui.Px(24)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")).
		Children(ui.Text(title))
}

func (a *builderApp) titlebar() *ui.Element {
	return ui.Header().Height(ui.Px(builderHeaderHeight)).Display(ui.DisplayFlex).
		AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).
		Padding(ui.RawValue("0 18px")).BackgroundColor(ui.Hex("#202d42")).
		BorderBottom(ui.RawValue("1px solid #34445c")).
		Children(
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
				Children(
					ui.Div().Padding(ui.Edges(ui.Px(8), ui.Px(11))).BorderRadius(ui.Px(9)).
						BackgroundColor(ui.Hex("#4677ed")).FontWeight(ui.FontWeightBold).
						Children(ui.Text("O")),
					ui.Div().Children(
						ui.Text("OwnFrame Builder").FontWeight(ui.FontWeightBold).FontSize(ui.Px(17)),
						ui.P().Text("Seven Go-built tabs · No HTML file").Margin(ui.RawValue("3px 0 0")).
							Color(ui.Hex("#a3b5cc")).FontSize(ui.Px(11)),
					),
				),
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
				Children(
					a.action(a.devLabel(), "win-dev", false),
					a.action("−", "win-min", false),
					a.action("□", "win-max", false),
					a.action("×", "win-close", false),
				),
		)
}

func (a *builderApp) devLabel() string {
	if a.page != nil && a.page.Window().DevTools() {
		return "Dev on"
	}
	return "Dev"
}

func (a *builderApp) action(label, command string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", command).
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(fg)).Border(ui.RawValue(border)).
		Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) sidebar() *ui.Element {
	return ui.Nav().Width(ui.Px(builderSidebarWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(ui.Hex("#202d42")).BorderRadius(ui.Px(11)).
		Padding(ui.Px(16)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		Gap(ui.Px(9)).Children(
		ui.Text("GALLERY").FontWeight(ui.FontWeightBold).FontSize(ui.Px(11)).Color(ui.Hex("#9aacbf")),
		a.navButton("◧ Widgets", "widgets"),
		a.navButton("⚙ Settings", "settings"),
		a.navButton("⊞ Layout", "layout"),
		a.navButton("✓ Todo", "todo"),
		a.navButton("≡ List", "list"),
		a.navButton("◐ Theme", "theme"),
		a.navButton("⧉ Compare", "compare"),
		a.navButton("◈ Dialog", "dialog"),
		ui.Div().Margin(ui.RawValue("25px 0 0")).Padding(ui.Px(12)).
			BackgroundColor(ui.Hex("#2b3950")).BorderRadius(ui.Px(8)).
			Children(
				ui.Text("Pure Go UI").FontWeight(ui.FontWeightBold),
				ui.P().Text("Every tab is one screen_*.go file using ui.Div chaining.").
					Margin(ui.RawValue("7px 0 0")).FontSize(ui.Px(11)).Color(ui.Hex("#b2c1d5")),
			),
	)
}

func (a *builderApp) navButton(label, tab string) *ui.Element {
	bg := "#2b3950"
	if a.tab == tab {
		bg = "#355596"
	}
	return ui.Button(label).Attr("data-action", "tab:"+tab).
		AlignSelf(ui.RawValue("stretch")).TextAlign(ui.TextAlignLeft).
		BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
		Border(ui.RawValue("1px solid #40546f")).
		Padding(ui.Px(11)).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold)
}
