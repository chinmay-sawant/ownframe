package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// officeThemeColors are the 10 Office theme base colors, left to right.
func officeThemeColors() []string {
	return []string{
		"#ffffff", "#000000", "#e7e6e6", "#44546a", "#4472c4",
		"#ed7d31", "#a5a5a5", "#ffc000", "#5b9bd5", "#70ad47",
	}
}

// officeStandardColors are the 10 standard colors on the last picker row.
func officeStandardColors() []string {
	return []string{
		"#c00000", "#ff0000", "#ffc000", "#ffff00", "#92d050",
		"#00b050", "#00b0f0", "#0070c0", "#002060", "#7030a0",
	}
}

// officeTints are the four tint/shade steps below the base row. Positive
// values mix toward white, negative toward black.
func officeTints() []float64 {
	return []float64{0.8, 0.6, 0.4, -0.25}
}

const (
	maxRecentColors = 10
	channelMask     = 0xff // one 8-bit color channel
	channelMax      = 255.0
	roundHalf       = 0.5 // rounds a float channel to the nearest int
)

// hexColorLen is the digit count of an rrggbb color.
const hexColorLen = 6

// parseHex reads #rrggbb. ok is false for anything else.
func parseHex(s string) (r, g, b int, ok bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != hexColorLen {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(v >> 16 & channelMask), int(v >> 8 & channelMask), int(v & channelMask), true
}

// shadeColor mixes hex toward white (t > 0) or black (t < 0).
func shadeColor(hex string, t float64) string {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return hex
	}
	mix := func(c int) int {
		if t >= 0 {
			return int(float64(c) + (channelMax-float64(c))*t + roundHalf)
		}
		return int(float64(c)*(1+t) + roundHalf)
	}
	return fmt.Sprintf("#%02x%02x%02x", mix(r), mix(g), mix(b))
}

// themeMatrix returns the 10-column, 5-row theme swatch matrix: the base row
// then four tint/shade rows.
func themeMatrix() [][]string {
	rows := make([][]string, 0, 1+len(officeTints()))
	rows = append(rows, officeThemeColors())
	for _, t := range officeTints() {
		row := make([]string, len(officeThemeColors()))
		for i, c := range officeThemeColors() {
			row[i] = shadeColor(c, t)
		}
		rows = append(rows, row)
	}
	return rows
}

// swatch is one color cell wired to action "<prefix>:<hex>".
func swatch(prefix, hex string) *ui.Element {
	return ui.Button("").Attr("type", "button").Class("rb-sw").
		Attr("data-action", prefix+":"+hex).
		Attr("aria-label", hex).Attr("title", hex).
		BackgroundColor(ui.Hex(hex))
}

// colorPicker builds the Office color popover. prefix is "xl:fillc" or
// "xl:fontc"; reset is the label of the no-color button.
func (a *app) colorPicker(prefix, reset string) *ui.Element {
	panel := ui.Div().Class("rb-dd-panel rb-picker").Attr("role", "dialog").
		Attr("aria-label", "Color picker")
	panel.Children(ui.Button(reset).Attr("type", "button").Class("rb-dd-item").
		Attr("data-action", prefix+":"))
	panel.Children(ui.Div().Class("rb-picker-title").Text("Theme Colors"))
	for _, row := range themeMatrix() {
		r := ui.Div().Class("rb-sw-row")
		for _, c := range row {
			r.Children(swatch(prefix, c))
		}
		panel.Children(r)
	}
	panel.Children(ui.Div().Class("rb-picker-title").Text("Standard Colors"))
	std := ui.Div().Class("rb-sw-row")
	for _, c := range officeStandardColors() {
		std.Children(swatch(prefix, c))
	}
	panel.Children(std)
	if len(a.x.recentColors) > 0 {
		panel.Children(ui.Div().Class("rb-picker-title").Text("Recent Colors"))
		rec := ui.Div().Class("rb-sw-row")
		for _, c := range a.x.recentColors {
			rec.Children(swatch(prefix, c))
		}
		panel.Children(rec)
	}
	return panel
}

// rememberColor pushes hex to the front of the recent-colors list.
func (a *app) rememberColor(hex string) {
	if hex == "" {
		return
	}
	out := []string{hex}
	for _, c := range a.x.recentColors {
		if c != hex && len(out) < maxRecentColors {
			out = append(out, c)
		}
	}
	a.x.recentColors = out
}

// borderMenuItems lists the border presets of the Borders dropdown.
func borderMenuItems() []struct{ Action, Label string } {
	return []struct{ Action, Label string }{
		{BorderBottom, "Bottom Border"},
		{BorderTop, "Top Border"},
		{BorderLeft, "Left Border"},
		{BorderRight, "Right Border"},
		{BorderNone, "No Border"},
		{BorderAll, "All Borders"},
		{BorderOutside, "Outside Borders"},
		{BorderThick, "Thick Outside Borders"},
		{BorderTotal, "Top and Double Bottom Border"},
	}
}

// borderMenu builds the Borders dropdown panel.
func (a *app) borderMenu() *ui.Element {
	panel := ui.Div().Class("rb-dd-panel").Attr("role", "menu")
	for _, it := range borderMenuItems() {
		panel.Children(ddItem("xl:border:"+it.Action, it.Label, ""))
	}
	return panel
}
