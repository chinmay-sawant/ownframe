package main

import (
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
			BackgroundColor(c).CSS("border", border).
			CSS("width", "40px").CSS("height", "26px").BorderRadius("6px"))
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
			BackgroundColor(bg).Color("#f1f5fd").
			CSS("border", "1px solid #40546f").
			Padding("8px 10px").BorderRadius("7px").FontSize("12px").FontWeight("bold"))
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
			BackgroundColor(bg).Color("#f1f5fd").
			CSS("border", "1px solid #40546f").
			Padding("8px 10px").BorderRadius("7px").FontSize("12px").FontWeight("bold"))
	}

	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			ui.Div().Children(
				ui.Text("Theme").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
				ui.P().Text("Pick an accent, corner radius, and spacing. The preview below restyles live from the same state.").
					Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
			),
			ui.Div().Padding("16px").BackgroundColor("#202d42").BorderRadius("11px").
				Display("flex").FlexDirection("column").Gap("12px").
				Children(
					ui.Text("Accent").FontWeight("bold").Color("#e9f0fa"),
					ui.Div().Display("flex").AlignItems("center").Gap("8px").
						Children(accentSwatches...),
					ui.Text("Corner radius").FontWeight("bold").Color("#e9f0fa"),
					ui.Div().Display("flex").AlignItems("center").Gap("8px").
						Children(radiusSwatches...),
					ui.Text("Spacing").FontWeight("bold").Color("#e9f0fa"),
					ui.Div().Display("flex").AlignItems("center").Gap("8px").
						Children(spacingSwatches...),
				),
			ui.Div().Padding(spacing).BackgroundColor("#26364e").BorderRadius(radius).
				Display("flex").FlexDirection("column").Gap(spacing).
				Children(
					ui.Div().Display("flex").AlignItems("center").Gap("8px").
						Children(
							ui.Text("Preview card").FontWeight("bold").Color("#f4f7ff"),
							ui.Text("New").FontSize("11px").FontWeight("bold").
								BackgroundColor(accent).Color("#ffffff").
								Padding("4px 8px").BorderRadius(radius),
						),
					ui.P().Text("Accent "+accent+" · Radius "+radius+" · Spacing "+spacing).
						Margin("0").FontSize("12px").Color("#b5c7df"),
					ui.Button("Themed action").Attr("type", "button").
						Attr("data-action", "theme:accent:"+strings.TrimPrefix(accent, "#")).
						BackgroundColor(accent).Color("#ffffff").
						CSS("border", "1px solid "+accent).
						Padding("9px 12px").BorderRadius(radius).
						FontSize("12px").FontWeight("bold"),
				),
			ui.Div().Padding("12px 15px").BackgroundColor("#202d42").BorderRadius("10px").
				Children(
					ui.P().Text("Convenience methods such as BackgroundColor, BorderRadius, Padding, and Gap call the same CSS(property, value) path that covers the 555 Implemented matrix names. The matrix is property-name coverage backed by Blinkless validation and rendering; the builder adds no new layout.").
						Margin("0").FontSize("12px").Color("#abbcd2"),
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
