package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func (a *builderApp) themePanel() *ui.Element {
	colors := []string{"#315fca", "#21694e", "#6241b5", "#a34c20"}
	accentButtons := make([]*ui.Element, 0, len(colors))
	for _, color := range colors {
		accentButtons = append(accentButtons, components.Button(color, components.ButtonOptions{
			Variant: components.ButtonFilled, Overrides: theme.Overrides{"--of-primary": color, "--of-on-primary": "#ffffff"},
		}).Attr("aria-label", "Use accent "+color).Attr("data-action", "theme:accent:"+strings.TrimPrefix(color, "#")))
	}
	radii := []*ui.Element{}
	for _, value := range []string{"4px", "8px", "12px", "16px"} {
		radii = append(radii, themeValueButton(value, "theme:radius:"+value, a.theme.Radius == value))
	}
	spacings := []*ui.Element{}
	for _, value := range []string{"8px", "16px", "24px"} {
		spacings = append(spacings, themeValueButton(value, "theme:spacing:"+value, a.theme.Spacing == value))
	}
	const previewProgress = 65
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(16)).Children(
		a.panelHeading("Theme studio"),
		ui.P().Text("Choose a global palette, then adjust its design tokens. Every library control inherits the result.").
			Margin(ui.Px(0)).Color(theme.Muted()).FontSize(ui.Px(13)),
		components.Card(
			ui.Text("Global palette").FontWeight(ui.FontWeightBold),
			a.themeSwitcher().MarginTop(ui.Px(12)),
			ui.Text("Accent").MarginTop(ui.Px(16)).FontWeight(ui.FontWeightBold),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(8)).Children(accentButtons...),
			ui.Text("Corner radius").MarginTop(ui.Px(16)).FontWeight(ui.FontWeightBold),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(8)).Children(radii...),
			ui.Text("Card padding").MarginTop(ui.Px(16)).FontWeight(ui.FontWeightBold),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(8)).Children(spacings...),
		),
		components.Card(
			components.Badge("Live preview"),
			ui.P().Text("One global theme, shared across your controls.").FontSize(ui.Px(18)).FontWeight(ui.FontWeightBold),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).Children(
				components.Button("Filled", components.ButtonOptions{Variant: components.ButtonFilled}),
				components.Button("Outlined", components.ButtonOptions{Variant: components.ButtonOutlined}),
				components.Button("Disabled", components.ButtonOptions{Disabled: true}),
			),
			components.Progress("Theme preview", previewProgress).MarginTop(ui.Px(18)),
		),
		components.Notice("Component overrides stay local", "Visit Custom component to compare inherited properties with instance-specific values."),
		components.Button("Reset token overrides", components.ButtonOptions{Variant: components.ButtonText}).Attr("data-action", "theme:reset"),
	)
}

func themeValueButton(label, action string, selected bool) *ui.Element {
	variant := components.ButtonTonal
	if selected {
		variant = components.ButtonFilled
	}
	return components.Button(label, components.ButtonOptions{Variant: variant}).Attr("data-action", action)
}

func (a *builderApp) themeSwitcher() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(6)).Children(
		themeValueButton("Light", "theme:mode:light", a.theme.Appearance == galleryLight),
		themeValueButton("Dark", "theme:mode:dark", a.theme.Appearance == galleryDark),
		themeValueButton("Studio", "theme:mode:studio", a.theme.Appearance == galleryStudio),
	)
}

func (a *builderApp) handleTheme(action string) bool {
	if mode, ok := strings.CutPrefix(action, "theme:mode:"); ok {
		switch galleryAppearance(mode) {
		case galleryLight, galleryDark, galleryStudio:
			a.theme.Appearance = galleryAppearance(mode)
			a.notice = "Global palette: " + mode + "."
			return true
		}
		return false
	}
	if action == "theme:reset" {
		a.theme.Accent = ""
		a.theme.Radius = ""
		a.theme.Spacing = ""
		a.notice = "Global token overrides reset."
		return true
	}
	if value, ok := strings.CutPrefix(action, "theme:accent:"); ok {
		for _, allowed := range []string{"315fca", "21694e", "6241b5", "a34c20"} {
			if value == allowed {
				a.theme.Accent = "#" + value
				a.notice = "Accent updated."
				return true
			}
		}
	}
	for _, property := range []string{"radius", "spacing"} {
		if value, ok := strings.CutPrefix(action, "theme:"+property+":"); ok {
			for _, allowed := range []string{"4px", "8px", "12px", "16px", "24px"} {
				if value != allowed {
					continue
				}
				if property == "radius" {
					a.theme.Radius = value
				} else {
					a.theme.Spacing = value
				}
				a.notice = "Theme tokens updated."
				return true
			}
		}
	}
	return false
}
