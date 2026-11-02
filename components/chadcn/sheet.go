package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// SheetSide selects the edge a sheet slides in from. The zero value is right.
type SheetSide uint8

const (
	SheetRight SheetSide = iota
	SheetLeft
	SheetTop
	SheetBottom
)

// SheetOptions describes a controlled side panel. The host owns Open and focus.
// ID is the stable host ID; the panel derives ID+"-panel" and the backdrop
// derives ID+"-backdrop". Children fill the body, and Footer holds the stacked
// action buttons the reference sheet shows. When OnOpenChange is set the
// component binds the trigger, so pass a trigger without its own action.
type SheetOptions struct {
	ID           string
	Open         bool
	Side         SheetSide
	Title        string
	Description  string
	Trigger      *ui.Element
	Footer       []*ui.Element
	OnOpenChange func(context.Context, bool) error
	Overrides    theme.Overrides
}

// Sheet builds a modal panel anchored to one window edge over a backdrop. When
// Open is false it returns the trigger, or nil when no trigger was supplied.
// A backdrop click and the header close control call OnOpenChange.
func Sheet(options SheetOptions, children ...*ui.Element) *ui.Element {
	panelID := partID(options.ID, "-panel")
	trigger := overlayTrigger(options.Trigger, options.Open, panelID, options.OnOpenChange)
	if !options.Open {
		return trigger
	}
	titleID, descriptionID := overlayPartIDs(panelID, options.Title, options.Description)
	panel := ui.Section().Class(classes("of-sheet-panel", options.Side.className()))
	if panelID != "" {
		panel.ID(panelID)
	}
	overlayPanelAttrs(panel, "dialog", "Sheet", titleID, descriptionID)
	sheetGeometry(panel, options.Side)
	panel.Children(
		overlayHeader("sheet", overlayHeading("sheet", options.Title, options.Description, titleID, descriptionID),
			overlayClose("sheet", "", options.OnOpenChange)),
		overlayBody("sheet", children),
		overlayStackedFooter("sheet", options.Footer),
	)

	backdrop := ui.Div().Class("of-sheet-backdrop").
		ZIndex(ui.Number(50)).Children(options.Overrides.Apply(panel))
	if id := partID(options.ID, "-backdrop"); id != "" {
		backdrop.ID(id)
	}
	if options.OnOpenChange != nil {
		backdrop.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, false) })
	}
	root := ui.Div().Class("of-sheet")
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(trigger, backdrop)
}

func (s SheetSide) className() string {
	switch s {
	case SheetLeft:
		return "of-sheet--left"
	case SheetTop:
		return "of-sheet--top"
	case SheetBottom:
		return "of-sheet--bottom"
	default:
		return "of-sheet--right"
	}
}

// sheetGeometry places the absolute panel against its edge. The size comes from
// CSS so a narrow window keeps the panel inside the viewport.
func sheetGeometry(panel *ui.Element, side SheetSide) {
	panel.Position(ui.PositionAbsolute)
	switch side {
	case SheetLeft:
		panel.Top(ui.Px(0)).Bottom(ui.Px(0)).Left(ui.Px(0))
	case SheetTop:
		panel.Top(ui.Px(0)).Left(ui.Px(0)).Right(ui.Px(0))
	case SheetBottom:
		panel.Bottom(ui.Px(0)).Left(ui.Px(0)).Right(ui.Px(0))
	default:
		panel.Top(ui.Px(0)).Bottom(ui.Px(0)).Right(ui.Px(0))
	}
}
