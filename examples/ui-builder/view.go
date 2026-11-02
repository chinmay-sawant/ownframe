package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
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
	case "custom":
		center = a.customPanel()
	case "dialog":
		center = a.dialogPanel()
	default:
		center = a.widgetsPanel()
	}
	return components.Screen(
		a.titlebar(),
		ui.Div().Display(ui.DisplayFlex).Flex(ui.Number(1)).
			MinHeight(ui.Px(0)).MinWidth(ui.Px(0)).Gap(ui.Px(builderContentGap)).Padding(ui.Px(builderPanePadding)).
			Children(a.sidebar(), a.scrollViewport(center)),
		ui.Div().Height(ui.Px(builderFooterHeight)).FlexShrink(ui.Number(0)).
			Padding(ui.Edges(ui.Px(9), ui.Px(18))).BackgroundColor(theme.Control()).
			Color(theme.Muted()).FontSize(ui.Px(12)).
			Children(ui.Text(a.notice)),
		a.dialogOverlay(),
		a.comparePreviewOverlay(),
	)
}

func (a *builderApp) panelHeading(title string) *ui.Element {
	return ui.Div().ID(builderActivePanelPrefix + a.tab).
		FontSize(ui.Px(24)).FontWeight(ui.FontWeightBold).Color(theme.Text()).
		Children(ui.Text(title))
}

func (a *builderApp) titlebar() *ui.Element {
	return ui.Header().Height(ui.Px(builderHeaderHeight)).Display(ui.DisplayFlex).
		AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).
		Padding(ui.Edges(ui.Px(0), ui.Px(18))).BackgroundColor(theme.Surface()).
		BorderBottom(theme.Border()).
		Children(
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
				Children(
					ui.Div().Padding(ui.Edges(ui.Px(8), ui.Px(11))).BorderRadius(ui.Px(9)).
						BackgroundColor(theme.Primary()).Color(theme.OnPrimary()).FontWeight(ui.FontWeightBold).
						Children(ui.Text("O")),
					ui.Div().Children(
						ui.Text("OwnFrame Components").FontWeight(ui.FontWeightBold).FontSize(ui.Px(17)),
						ui.P().Text("Controls, themes, and your own components").Margin(ui.Edges(ui.Px(3), ui.Px(0), ui.Px(0))).
							Color(theme.Muted()).FontSize(ui.Px(11)),
					),
				),
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
				Children(
					a.themeSwitcher(),
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
	variant := components.ButtonTonal
	if primary {
		variant = components.ButtonFilled
	}
	return components.Button(label, components.ButtonOptions{Variant: variant}).Attr("data-action", command)
}

func (a *builderApp) sidebar() *ui.Element {
	return ui.Nav().Width(ui.Px(builderSidebarWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(theme.Surface()).BorderRadius(ui.Px(11)).
		Padding(ui.Px(16)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		Gap(ui.Px(6)).Children(
		ui.Text("GALLERY").FontWeight(ui.FontWeightBold).FontSize(ui.Px(11)).Color(theme.Muted()),
		a.navButton("Components", "widgets"),
		a.navButton("Forms", "settings"),
		a.navButton("Layout", "layout"),
		a.navButton("State & motion", "todo"),
		a.navButton("Collections", "list"),
		a.navButton("Theme studio", "theme"),
		a.navButton("Source & preview", "compare"),
		a.navButton("Dialogs", "dialog"),
		a.navButton("Custom component", "custom"),
		ui.Div().Margin(ui.Edges(ui.Px(25), ui.Px(0), ui.Px(0))).Padding(ui.Px(12)).
			BackgroundColor(theme.Control()).BorderRadius(ui.Px(8)).
			Children(
				ui.Text("Your library").FontWeight(ui.FontWeightBold),
				ui.P().Text("Compose shared controls. Add application components in their own package.").
					Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).FontSize(ui.Px(11)).Color(theme.Muted()),
			),
	)
}

func (a *builderApp) navButton(label, tab string) *ui.Element {
	button := components.Button(label, components.ButtonOptions{Size: components.ButtonSmall}).
		Attr("data-action", "tab:"+tab).TextAlign(ui.TextAlignLeft).Padding(ui.Px(9))
	if a.tab == tab {
		button.BackgroundColor(theme.Selected()).Attr("aria-current", "page")
	}
	return button
}
