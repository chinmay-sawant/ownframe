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
// derives ID+"-backdrop". Children fill the scrolling body, and Footer holds
// the stacked action buttons pinned under that body. A left or right panel is
// 400px wide on desktop and 90% of the window at 520px and below; top and
// bottom sheets stay full width. When OnOpenChange is set the component binds
// the trigger, so pass a trigger without its own action.
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
// The panel always has an id, and the backdrop id is different, so a click in
// the body hits the panel and does not dismiss. A backdrop click and the
// header close control call OnOpenChange when that callback is set.
func Sheet(options SheetOptions, children ...*ui.Element) *ui.Element {
	panelID := stablePartID(options.ID, "-panel", "of-sheet-panel")
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
	closeControl := overlayClose("sheet", "", options.OnOpenChange)
	if closeControl != nil {
		closeControl.ID(partID(panelID, "-close"))
	}
	panel.Children(
		overlayHeader("sheet", overlayHeading("sheet", options.Title, options.Description, titleID, descriptionID), closeControl),
		overlayBody("sheet", panelID, children),
		overlayStackedFooter("sheet", panelID, options.Footer),
	)

	backdrop := ui.Div().Class("of-sheet-backdrop").Children(options.Overrides.Apply(panel))
	backdrop.ID(stablePartID(options.ID, "-backdrop", "of-sheet-backdrop"))
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

// sheetGeometry places the absolute panel against its edge. A side width comes
// from sheet.css and never from an inline style: the stylesheet sets 400px on
// desktop and a 90% fallback at 520px and below, and an inline width would win
// over that media query.
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
