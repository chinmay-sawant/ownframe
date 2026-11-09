package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// ScrollAreaOptions describes a clipped viewport and the scrollbar beside it.
// ID is the host's stable root ID; the viewport, bar, and thumb derive their
// IDs from it. Label names the region and defaults to "Scroll area". State owns
// the offset and a thumb drag, and TrackPixels is the measured bar height.
// Height sizes the viewport; zero uses the stylesheet default. OnChange is
// forwarded to the bar after a track click pages the offset. Children sit
// inside the clipped viewport.
type ScrollAreaOptions struct {
	ID          string
	Label       string
	State       *ScrollState
	TrackPixels int
	Height      ui.Length
	OnChange    func(context.Context) error
	Children    []*ui.Element
}

// ScrollArea clips a scrolling child list and overlays a themed scrollbar.
// State.Offset translates the children; overflow does not scroll the pane.
// The host routes wheel, key, and drag input to State, calls State.Track for
// a pointer on the bar, and rebuilds from OnChange. The bar renders nothing
// while the content fits.
func ScrollArea(options ScrollAreaOptions) *ui.Element {
	viewportID := partID(options.ID, "-viewport")
	children := options.Children
	if options.State != nil && options.State.Offset() != 0 {
		children = []*ui.Element{ui.Div().Class("of-scroll-area-content").
			Transform(ui.TranslateY(ui.Px(-float64(options.State.Offset())))).
			Children(options.Children...)}
	}
	viewport := ui.Div().Class("of-scroll-area-viewport").Position(ui.PositionAbsolute).
		Top(ui.Px(0)).Right(ui.Px(0)).Bottom(ui.Px(0)).Left(ui.Px(0)).
		Children(children...)
	if viewportID != "" {
		viewport.ID(viewportID)
	}
	bar := Scrollbar(ScrollbarOptions{
		ID:          partID(options.ID, "-bar"),
		ThumbID:     partID(options.ID, "-thumb"),
		ControlsID:  viewportID,
		Label:       options.Label,
		State:       options.State,
		TrackPixels: options.TrackPixels,
		OnChange:    options.OnChange,
	})
	root := ui.Div().Class(classes("of-scope", "of-scroll-area")).Attr("role", "region")
	if options.ID != "" {
		root.ID(options.ID)
	}
	label := options.Label
	if label == "" {
		label = "Scroll area"
	}
	root.Attr("aria-label", label)
	if options.Height != (ui.Length{}) {
		root.Height(options.Height)
	}
	return root.Children(viewport, bar)
}
