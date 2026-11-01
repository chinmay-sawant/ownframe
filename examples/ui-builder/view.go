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
	return ui.Div().Width("100%").CSS("min-height", "100vh").
		Display("flex").FlexDirection("column").
		BackgroundColor("#111827").Color("#e9f0fa").
		CSS("font-family", "sans-serif").FontSize("14px").
		Children(
			a.titlebar(),
			ui.Div().Display("flex").CSS("flex", "1").
				CSS("min-height", "0px").CSS("min-width", "0px").Gap("16px").Padding("16px").
				Children(a.sidebar(), center),
			ui.Div().Padding("9px 18px").BackgroundColor("#182336").
				Color("#a4b5ca").FontSize("12px").
				Children(ui.Text(a.notice)),
			a.dialogOverlay(),
			a.comparePreviewOverlay(),
		)
}

func (a *builderApp) titlebar() *ui.Element {
	return ui.Header().Height("68px").Display("flex").
		AlignItems("center").CSS("justify-content", "space-between").
		Padding("0 18px").BackgroundColor("#202d42").
		CSS("border-bottom", "1px solid #34445c").
		Children(
			ui.Div().Display("flex").AlignItems("center").Gap("10px").
				Children(
					ui.Div().Padding("8px 11px").BorderRadius("9px").
						BackgroundColor("#4677ed").FontWeight("bold").
						Children(ui.Text("O")),
					ui.Div().Children(
						ui.Text("OwnFrame Builder").FontWeight("bold").FontSize("17px"),
						ui.P().Text("Seven Go-built tabs · No HTML file").Margin("3px 0 0").
							Color("#a3b5cc").FontSize("11px"),
					),
				),
			ui.Div().Display("flex").AlignItems("center").Gap("8px").
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
		BackgroundColor(bg).Color(fg).CSS("border", border).
		Padding("9px 12px").BorderRadius("8px").
		FontSize("12px").FontWeight("bold")
}

func (a *builderApp) sidebar() *ui.Element {
	return ui.Nav().Width("205px").CSS("flex-shrink", "0").
		BackgroundColor("#202d42").BorderRadius("11px").
		Padding("16px").Display("flex").FlexDirection("column").
		Gap("9px").Children(
		ui.Text("GALLERY").FontWeight("bold").FontSize("11px").Color("#9aacbf"),
		a.navButton("◧ Widgets", "widgets"),
		a.navButton("⚙ Settings", "settings"),
		a.navButton("⊞ Layout", "layout"),
		a.navButton("✓ Todo", "todo"),
		a.navButton("≡ List", "list"),
		a.navButton("◐ Theme", "theme"),
		a.navButton("⧉ Compare", "compare"),
		a.navButton("◈ Dialog", "dialog"),
		ui.Div().Margin("25px 0 0").Padding("12px").
			BackgroundColor("#2b3950").BorderRadius("8px").
			Children(
				ui.Text("Pure Go UI").FontWeight("bold"),
				ui.P().Text("Every tab is one screen_*.go file using ui.Div chaining.").
					Margin("7px 0 0").FontSize("11px").Color("#b2c1d5"),
			),
	)
}

func (a *builderApp) navButton(label, tab string) *ui.Element {
	bg := "#2b3950"
	if a.tab == tab {
		bg = "#355596"
	}
	return ui.Button(label).Attr("data-action", "tab:"+tab).
		CSS("align-self", "stretch").CSS("text-align", "left").
		BackgroundColor(bg).Color("#f1f5fd").
		CSS("border", "1px solid #40546f").
		Padding("11px").BorderRadius("8px").FontWeight("bold")
}
