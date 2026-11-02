package main

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

const (
	layoutMinBoxes     = 3
	layoutMaxBoxes     = 6
	layoutDefaultBoxes = 4
)

// layoutPalette is the box fill cycle for the preview and the snippet.
func layoutPalette() []string {
	return []string{"#4677ed", "#59d4b8", "#e5a54b", "#e56b6f", "#9b7ede", "#4cc38a"}
}

type layoutState struct {
	Direction string
	Gap       string
	Padding   string
	Boxes     int
}

func defaultLayoutState() layoutState {
	return layoutState{Direction: "row", Gap: "16px", Padding: "16px", Boxes: layoutDefaultBoxes}
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
	palette := layoutPalette()
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
			palette[i%len(palette)],
			i+1,
		)
	}
	sb.WriteString(")")
	return sb.String()
}

func layoutToggleButton(label, action string, active bool) *ui.Element {
	return widgetsBtn(label, action, active)
}

func layoutPixels(value string) ui.Length {
	switch value {
	case "8px":
		return ui.Px(8)
	case "24px":
		return ui.Px(24)
	default:
		return ui.Px(16)
	}
}

func (a *builderApp) layoutPanel() *ui.Element {
	st := a.layout.normalized()
	palette := layoutPalette()

	controls := ui.Div().Display(ui.DisplayFlex).FlexWrap(ui.FlexWrapWrap).
		AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
		Children(
			ui.Text("Direction:").FontSize(ui.Px(12)).Color(theme.Muted()),
			layoutToggleButton("Row", "layout:dir:row", st.Direction == "row"),
			layoutToggleButton("Column", "layout:dir:column", st.Direction == "column"),
			ui.Text("Gap:").FontSize(ui.Px(12)).Color(theme.Muted()),
			layoutToggleButton("8px", "layout:gap:8px", st.Gap == "8px"),
			layoutToggleButton("16px", "layout:gap:16px", st.Gap == "16px"),
			layoutToggleButton("24px", "layout:gap:24px", st.Gap == "24px"),
			ui.Text("Padding:").FontSize(ui.Px(12)).Color(theme.Muted()),
			layoutToggleButton("8px", "layout:pad:8px", st.Padding == "8px"),
			layoutToggleButton("16px", "layout:pad:16px", st.Padding == "16px"),
			layoutToggleButton("24px", "layout:pad:24px", st.Padding == "24px"),
			ui.Text("Boxes:").FontSize(ui.Px(12)).Color(theme.Muted()),
			layoutToggleButton("− Fewer", "layout:fewer", false),
			ui.Text(fmt.Sprintf("%d", st.Boxes)).FontWeight(ui.FontWeightBold).Color(theme.Text()),
			layoutToggleButton("+ More", "layout:more", false),
		)

	boxes := make([]*ui.Element, 0, st.Boxes)
	for i := 0; i < st.Boxes; i++ {
		color := palette[i%len(palette)]
		boxes = append(boxes, ui.Div().Flex(ui.Number(1)).MinWidth(ui.Px(0)).
			Padding(ui.Edges(ui.Px(24), ui.Px(12))).BorderRadius(ui.Px(8)).
			BackgroundColor(ui.Hex(color)).Color(theme.OnPrimary()).
			TextAlign(ui.TextAlignCenter).FontWeight(ui.FontWeightBold).
			Children(ui.Text(fmt.Sprintf("Box %d", i+1))))
	}

	direction := ui.FlexDirectionRow
	if st.Direction == "column" {
		direction = ui.FlexDirectionColumn
	}
	preview := ui.Div().Display(ui.DisplayFlex).FlexDirection(direction).
		Gap(layoutPixels(st.Gap)).Padding(layoutPixels(st.Padding)).
		BackgroundColor(theme.Control()).BorderRadius(ui.Px(10)).
		Border(theme.Border()).
		Children(boxes...)

	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("Layout playground"),
				ui.P().Text("Toggle the options to update the preview and its typed Go source.").
					Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
			),
			controls,
			preview,
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
				Children(
					layoutToggleButton("Copy code", "layout:copy", false),
					ui.Text("Go code for the preview above — updates with every toggle.").FontSize(ui.Px(12)).Color(theme.Muted()),
				),
			ui.P().Text(a.layoutSnippet()).Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(theme.Muted()).
				FontFamily(ui.FontFamilyMonospace).OverflowWrap(ui.OverflowWrapBreakWord).
				WhiteSpace(ui.WhiteSpacePreWrap).
				BackgroundColor(theme.Control()).Padding(ui.Edges(ui.Px(10), ui.Px(12))).BorderRadius(ui.Px(8)),
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
