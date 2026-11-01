package main

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	layoutMinBoxes = 3
	layoutMaxBoxes = 6
)

// layoutPalette is the box fill cycle for the preview and the snippet.
var layoutPalette = []string{"#4677ed", "#59d4b8", "#e5a54b", "#e56b6f", "#9b7ede", "#4cc38a"}

type layoutState struct {
	Direction string
	Gap       string
	Padding   string
	Boxes     int
}

func defaultLayoutState() layoutState {
	return layoutState{Direction: "row", Gap: "16px", Padding: "16px", Boxes: 4}
}

func (s layoutState) normalized() layoutState {
	if s.Direction != "column" {
		s.Direction = "row"
	}
	switch s.Gap {
	case "8px", "16px", "24px":
	default:
		s.Gap = "16px"
	}
	switch s.Padding {
	case "8px", "16px", "24px":
	default:
		s.Padding = "16px"
	}
	if s.Boxes < layoutMinBoxes {
		s.Boxes = layoutMinBoxes
	}
	if s.Boxes > layoutMaxBoxes {
		s.Boxes = layoutMaxBoxes
	}
	return s
}

// layoutSnippet renders the full Go builder code for the live preview:
// the container call plus one line per box with its palette color.
func (a *builderApp) layoutSnippet() string {
	st := a.layout.normalized()
	var sb strings.Builder
	fmt.Fprintf(&sb, "ui.Div().Display(\"flex\").FlexDirection(%q).Gap(%q).Padding(%q).Children(\n",
		st.Direction, st.Gap, st.Padding)
	for i := 0; i < st.Boxes; i++ {
		fmt.Fprintf(&sb, "    ui.Div().BackgroundColor(%q).Children(ui.Text(\"Box %d\")),\n",
			layoutPalette[i%len(layoutPalette)], i+1)
	}
	sb.WriteString(")")
	return sb.String()
}

func layoutToggleButton(label, action string, active bool) *ui.Element {
	bg := "#2b3950"
	if active {
		bg = "#355596"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		BackgroundColor(bg).Color("#f1f5fd").
		CSS("border", "1px solid #40546f").
		Padding("9px 12px").BorderRadius("8px").
		FontSize("12px").FontWeight("bold")
}

func (a *builderApp) layoutPanel() *ui.Element {
	st := a.layout.normalized()

	controls := ui.Div().Display("flex").CSS("flex-wrap", "wrap").
		AlignItems("center").Gap("8px").
		Children(
			ui.Text("Direction:").FontSize("12px").Color("#9aacbf"),
			layoutToggleButton("Row", "layout:dir:row", st.Direction == "row"),
			layoutToggleButton("Column", "layout:dir:column", st.Direction == "column"),
			ui.Text("Gap:").FontSize("12px").Color("#9aacbf"),
			layoutToggleButton("8px", "layout:gap:8px", st.Gap == "8px"),
			layoutToggleButton("16px", "layout:gap:16px", st.Gap == "16px"),
			layoutToggleButton("24px", "layout:gap:24px", st.Gap == "24px"),
			ui.Text("Padding:").FontSize("12px").Color("#9aacbf"),
			layoutToggleButton("8px", "layout:pad:8px", st.Padding == "8px"),
			layoutToggleButton("16px", "layout:pad:16px", st.Padding == "16px"),
			layoutToggleButton("24px", "layout:pad:24px", st.Padding == "24px"),
			ui.Text("Boxes:").FontSize("12px").Color("#9aacbf"),
			layoutToggleButton("− Fewer", "layout:fewer", false),
			ui.Text(fmt.Sprintf("%d", st.Boxes)).FontWeight("bold").Color("#f4f7ff"),
			layoutToggleButton("+ More", "layout:more", false),
		)

	boxes := make([]*ui.Element, 0, st.Boxes)
	for i := 0; i < st.Boxes; i++ {
		color := layoutPalette[i%len(layoutPalette)]
		boxes = append(boxes, ui.Div().CSS("flex", "1").CSS("min-width", "0px").
			Padding("24px 12px").BorderRadius("8px").
			BackgroundColor(color).Color("#ffffff").
			CSS("text-align", "center").FontWeight("bold").
			Children(ui.Text(fmt.Sprintf("Box %d", i+1))))
	}

	preview := ui.Div().Display("flex").FlexDirection(st.Direction).
		Gap(st.Gap).Padding(st.Padding).
		BackgroundColor("#182336").BorderRadius("10px").
		CSS("border", "1px solid #34445c").
		Children(boxes...)

	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			ui.Div().Children(
				ui.Text("Layout playground").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
				ui.P().Text("Toggle direction, gap, padding, and box count; the preview below is built with the chained call shown.").
					Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
			),
			controls,
			preview,
			ui.Div().Display("flex").AlignItems("center").Gap("8px").
				Children(
					layoutToggleButton("Copy code", "layout:copy", false),
					ui.Text("Go code for the preview above — updates with every toggle.").FontSize("12px").Color("#9aacbf"),
				),
			ui.P().Text(a.layoutSnippet()).Margin("0").FontSize("12px").Color("#9fe8c9").
				CSS("font-family", "monospace").CSS("overflow-wrap", "break-word").
				CSS("white-space", "pre-wrap").
				BackgroundColor("#182336").Padding("10px 12px").BorderRadius("8px"),
		)
}

func (a *builderApp) handleLayout(action string) bool {
	if !strings.HasPrefix(action, "layout:") {
		return false
	}
	if action == "layout:copy" {
		ownframe.Write(a.layoutSnippet())
		a.notice = "Layout code copied to the clipboard."
		return true
	}
	switch {
	case action == "layout:dir:row":
		a.layout.Direction = "row"
		a.notice = "Layout direction set to row."
	case action == "layout:dir:column":
		a.layout.Direction = "column"
		a.notice = "Layout direction set to column."
	case strings.HasPrefix(action, "layout:gap:"):
		value := strings.TrimPrefix(action, "layout:gap:")
		switch value {
		case "8px", "16px", "24px":
			a.layout.Gap = value
			a.notice = "Layout gap set to " + value + "."
		default:
			a.notice = "Unknown layout gap: " + value + "."
		}
	case strings.HasPrefix(action, "layout:pad:"):
		value := strings.TrimPrefix(action, "layout:pad:")
		switch value {
		case "8px", "16px", "24px":
			a.layout.Padding = value
			a.notice = "Layout padding set to " + value + "."
		default:
			a.notice = "Unknown layout padding: " + value + "."
		}
	case action == "layout:more":
		if a.layout.Boxes < layoutMaxBoxes {
			a.layout.Boxes++
			a.notice = fmt.Sprintf("Layout preview now shows %d boxes.", a.layout.Boxes)
		} else {
			a.notice = fmt.Sprintf("Layout preview already shows the maximum of %d boxes.", layoutMaxBoxes)
		}
	case action == "layout:fewer":
		if a.layout.Boxes > layoutMinBoxes {
			a.layout.Boxes--
			a.notice = fmt.Sprintf("Layout preview now shows %d boxes.", a.layout.Boxes)
		} else {
			a.notice = fmt.Sprintf("Layout preview already shows the minimum of %d boxes.", layoutMinBoxes)
		}
	default:
		a.notice = "Unknown layout action: " + action + "."
	}
	return true
}
