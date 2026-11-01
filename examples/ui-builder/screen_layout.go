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
	direction := "Column"
	if st.Direction == "row" {
		direction = "Row"
	}
	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirection%s).Gap(ui.Px(%s)).Padding(ui.Px(%s)).Children(\n",
		direction,
		strings.TrimSuffix(st.Gap, "px"),
		strings.TrimSuffix(st.Padding, "px"),
	)
	for i := 0; i < st.Boxes; i++ {
		fmt.Fprintf(
			&sb,
			"    ui.Div().BackgroundColor(ui.Hex(%q)).Children(ui.Text(\"Box %d\")),\n",
			layoutPalette[i%len(layoutPalette)],
			i+1,
		)
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
		BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
		Border(ui.RawValue("1px solid #40546f")).
		Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) layoutPanel() *ui.Element {
	st := a.layout.normalized()

	controls := ui.Div().Display(ui.DisplayFlex).FlexWrap(ui.RawValue("wrap")).
		AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
		Children(
			ui.Text("Direction:").FontSize(ui.Px(12)).Color(ui.Hex("#9aacbf")),
			layoutToggleButton("Row", "layout:dir:row", st.Direction == "row"),
			layoutToggleButton("Column", "layout:dir:column", st.Direction == "column"),
			ui.Text("Gap:").FontSize(ui.Px(12)).Color(ui.Hex("#9aacbf")),
			layoutToggleButton("8px", "layout:gap:8px", st.Gap == "8px"),
			layoutToggleButton("16px", "layout:gap:16px", st.Gap == "16px"),
			layoutToggleButton("24px", "layout:gap:24px", st.Gap == "24px"),
			ui.Text("Padding:").FontSize(ui.Px(12)).Color(ui.Hex("#9aacbf")),
			layoutToggleButton("8px", "layout:pad:8px", st.Padding == "8px"),
			layoutToggleButton("16px", "layout:pad:16px", st.Padding == "16px"),
			layoutToggleButton("24px", "layout:pad:24px", st.Padding == "24px"),
			ui.Text("Boxes:").FontSize(ui.Px(12)).Color(ui.Hex("#9aacbf")),
			layoutToggleButton("− Fewer", "layout:fewer", false),
			ui.Text(fmt.Sprintf("%d", st.Boxes)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
			layoutToggleButton("+ More", "layout:more", false),
		)

	boxes := make([]*ui.Element, 0, st.Boxes)
	for i := 0; i < st.Boxes; i++ {
		color := layoutPalette[i%len(layoutPalette)]
		boxes = append(boxes, ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).
			Padding(ui.Edges(ui.Px(24), ui.Px(12))).BorderRadius(ui.Px(8)).
			BackgroundColor(ui.RawValue(color)).Color(ui.Hex("#ffffff")).
			TextAlign(ui.TextAlignCenter).FontWeight(ui.FontWeightBold).
			Children(ui.Text(fmt.Sprintf("Box %d", i+1))))
	}

	preview := ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.RawValue(st.Direction)).
		Gap(ui.RawValue(st.Gap)).Padding(ui.RawValue(st.Padding)).
		BackgroundColor(ui.Hex("#182336")).BorderRadius(ui.Px(10)).
		Border(ui.RawValue("1px solid #34445c")).
		Children(boxes...)

	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				ui.Text("Layout playground").FontSize(ui.Px(24)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
				ui.P().Text("Toggle the options to update the preview and its typed Go source.").
					Margin(ui.RawValue("7px 0 0")).Color(ui.Hex("#a9bbd1")).FontSize(ui.Px(12)),
			),
			controls,
			preview,
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
				Children(
					layoutToggleButton("Copy code", "layout:copy", false),
					ui.Text("Go code for the preview above — updates with every toggle.").FontSize(ui.Px(12)).Color(ui.Hex("#9aacbf")),
				),
			ui.P().Text(a.layoutSnippet()).Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(ui.Hex("#9fe8c9")).
				FontFamily(ui.RawValue("monospace")).OverflowWrap(ui.OverflowWrapBreakWord).
				WhiteSpace(ui.WhiteSpacePreWrap).
				BackgroundColor(ui.Hex("#182336")).Padding(ui.Edges(ui.Px(10), ui.Px(12))).BorderRadius(ui.Px(8)),
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
