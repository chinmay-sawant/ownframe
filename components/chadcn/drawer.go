package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// DrawerOptions describes a controlled bottom drawer. The host owns Open and
// focus. ID is the stable host ID; the panel derives ID+"-panel" and the
// backdrop derives ID+"-backdrop". Children fill the body, and Footer holds the
// stacked action buttons below the grab handle. The body carries the height cap
// and scrolls, so the footer stays on screen. When OnOpenChange is set the
// component binds the trigger, so pass a trigger without its own action.
//
// Mount the drawer at the screen root, as a sibling of the page content. The
// backdrop is fixed, so a drawer mounted inside the page panel lets its
// background compete with the page instead of covering it as one modal layer.
type DrawerOptions struct {
	ID           string
	Open         bool
	Title        string
	Description  string
	Trigger      *ui.Element
	Footer       []*ui.Element
	OnOpenChange func(context.Context, bool) error
	Overrides    theme.Overrides
}

// Drawer builds a modal panel anchored to the bottom edge, with a grab handle
// and a stacked action footer. When Open is false it returns the trigger, or
// nil when no trigger was supplied. A backdrop click calls OnOpenChange;
// Escape handling and focus scope stay with the host.
//
// The backdrop paints at the modal layer: above the app topbar and sidebar,
// below the toast and command surfaces.
func Drawer(options DrawerOptions, children ...*ui.Element) *ui.Element {
	panelID := partID(options.ID, "-panel")
	trigger := overlayTrigger(options.Trigger, options.Open, panelID, options.OnOpenChange)
	if !options.Open {
		return trigger
	}
	titleID, descriptionID := overlayPartIDs(panelID, options.Title, options.Description)
	panel := ui.Section().Class("of-drawer-panel")
	if panelID != "" {
		panel.ID(panelID)
	}
	overlayPanelAttrs(panel, "dialog", "Drawer", titleID, descriptionID)
	panel.Children(
		ui.Div().Class("of-drawer-grab").Attr("aria-hidden", "true").Children(ui.Span().Class("of-drawer-handle")),
		overlayHeader("drawer", overlayHeading("drawer", options.Title, options.Description, titleID, descriptionID), nil),
		overlayBody("drawer", children),
		overlayStackedFooter("drawer", options.Footer),
	)

	backdrop := ui.Div().Class("of-drawer-backdrop").Children(options.Overrides.Apply(panel))
	if id := partID(options.ID, "-backdrop"); id != "" {
		backdrop.ID(id)
	}
	if options.OnOpenChange != nil {
		backdrop.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, false) })
	}
	root := ui.Div().Class("of-drawer")
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(trigger, backdrop)
}
