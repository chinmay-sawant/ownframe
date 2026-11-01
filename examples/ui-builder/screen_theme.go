package main

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

type themeState struct {
	Accent  string
	Radius  string
	Spacing string
}

func defaultThemeState() themeState {
	return themeState{Accent: "#4677ed", Radius: "10px", Spacing: "16px"}
}

func (a *builderApp) themePanel() *ui.Element {
	accent := strings.TrimSpace(a.theme.Accent)
	radius := strings.TrimSpace(a.theme.Radius)
	spacing := strings.TrimSpace(a.theme.Spacing)
	def := defaultThemeState()
	if accent == "" {
		accent = def.Accent
	}
	if radius == "" {
		radius = def.Radius
	}
	if spacing == "" {
		spacing = def.Spacing
	}

	accents := []string{"#4677ed", "#59d4b8", "#a855f7", "#f59e0b"}
	radii := []string{"4px", "8px", "12px", "16px"}
	spacings := []string{"8px", "16px", "24px"}

	accentSwatches := make([]*ui.Element, 0, len(accents))
	for _, c := range accents {
		c := c
		border := "1px solid #42536b"
		if strings.EqualFold(strings.TrimSpace(c), accent) {
			border = "2px solid #ffffff"
		}
		accentSwatches = append(accentSwatches, ui.Button("").Attr("type", "button").
			Attr("data-action", "theme:accent:"+strings.TrimPrefix(c, "#")).
			BackgroundColor(ui.RawValue(c)).Border(ui.RawValue(border)).
			Width(ui.Px(40)).Height(ui.Px(26)).BorderRadius(ui.Px(6)))
	}

	radiusSwatches := make([]*ui.Element, 0, len(radii))
	for _, r := range radii {
		r := r
		bg := "#2b3950"
		if strings.TrimSpace(r) == radius {
			bg = "#355596"
		}
		radiusSwatches = append(radiusSwatches, ui.Button(r).Attr("type", "button").
			Attr("data-action", "theme:radius:"+r).
			BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
			Border(ui.RawValue("1px solid #40546f")).
			Padding(ui.Edges(ui.Px(8), ui.Px(10))).BorderRadius(ui.Px(7)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold))
	}

	spacingSwatches := make([]*ui.Element, 0, len(spacings))
	for _, s := range spacings {
		s := s
		bg := "#2b3950"
		if strings.TrimSpace(s) == spacing {
			bg = "#355596"
		}
		spacingSwatches = append(spacingSwatches, ui.Button(s).Attr("type", "button").
			Attr("data-action", "theme:spacing:"+s).
			BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
			Border(ui.RawValue("1px solid #40546f")).
			Padding(ui.Edges(ui.Px(8), ui.Px(10))).BorderRadius(ui.Px(7)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold))
	}

	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("Theme"),
				ui.P().Text("Pick an accent, corner radius, and spacing. The preview below restyles live from the same state.").
					Margin(ui.RawValue("7px 0 0")).Color(ui.Hex("#a9bbd1")).FontSize(ui.Px(12)),
			),
			ui.Div().Padding(ui.Px(16)).BackgroundColor(ui.Hex("#202d42")).BorderRadius(ui.Px(11)).
				Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
				Children(
					ui.Text("Accent").FontWeight(ui.FontWeightBold).Color(ui.Hex("#e9f0fa")),
					ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
						Children(accentSwatches...),
					ui.Text("Corner radius").FontWeight(ui.FontWeightBold).Color(ui.Hex("#e9f0fa")),
					ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
						Children(radiusSwatches...),
					ui.Text("Spacing").FontWeight(ui.FontWeightBold).Color(ui.Hex("#e9f0fa")),
					ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
						Children(spacingSwatches...),
				),
			ui.Div().Padding(ui.RawValue(spacing)).BackgroundColor(ui.Hex("#26364e")).BorderRadius(ui.RawValue(radius)).
				Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.RawValue(spacing)).
				Children(
					ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
						Children(
							ui.Text("Preview card").FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
							ui.Text("New").FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).
								BackgroundColor(ui.RawValue(accent)).Color(ui.Hex("#ffffff")).
								Padding(ui.Edges(ui.Px(4), ui.Px(8))).BorderRadius(ui.RawValue(radius)),
						),
					ui.P().Text("Accent "+accent+" · Radius "+radius+" · Spacing "+spacing).
						Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(ui.Hex("#b5c7df")),
					ui.Button("Themed action").Attr("type", "button").
						Attr("data-action", "theme:accent:"+strings.TrimPrefix(accent, "#")).
						CustomProperty("--builder-accent", accent).
						BackgroundColor(ui.Var("--builder-accent")).Color(ui.Hex("#ffffff")).
						Border(ui.RawValue("1px solid "+accent)).
						Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.RawValue(radius)).
						FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
				),
			ui.Div().Padding(ui.Edges(ui.Px(12), ui.Px(15))).BackgroundColor(ui.Hex("#202d42")).BorderRadius(ui.Px(10)).
				Children(
					ui.P().Text(fmt.Sprintf(
						"Generated methods expose all %d implemented CSS properties to Go completion. "+
							"Typed helpers such as Px, Percent, Hex, and FlexDirectionColumn expose common values. "+
							"Use RawValue for complex values.",
						len(ui.SupportedProperties()),
					)).
						Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(ui.Hex("#abbcd2")),
				),
		)
}

func (a *builderApp) handleTheme(action string) bool {
	if name, ok := strings.CutPrefix(action, "theme:accent:"); ok {
		v := strings.TrimSpace(name)
		if v == "" {
			return false
		}
		if !strings.HasPrefix(v, "#") && !strings.HasPrefix(v, "rgb") && !strings.HasPrefix(v, "hsl") {
			v = "#" + v
		}
		a.theme.Accent = v
		a.notice = "Accent set to " + v + "."
		return true
	}
	if name, ok := strings.CutPrefix(action, "theme:radius:"); ok {
		v := strings.TrimSpace(name)
		if v == "" {
			return false
		}
		if v != "" && (v[0] >= '0' && v[0] <= '9') && !strings.HasSuffix(v, "px") && !strings.HasSuffix(v, "%") {
			v += "px"
		}
		a.theme.Radius = v
		a.notice = "Corner radius set to " + v + "."
		return true
	}
	if name, ok := strings.CutPrefix(action, "theme:spacing:"); ok {
		v := strings.TrimSpace(name)
		if v == "" {
			return false
		}
		if v != "" && (v[0] >= '0' && v[0] <= '9') && !strings.HasSuffix(v, "px") && !strings.HasSuffix(v, "%") {
			v += "px"
		}
		a.theme.Spacing = v
		a.notice = "Spacing set to " + v + "."
		return true
	}
	return false
}
