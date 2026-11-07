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
// nil when no trigger was supplied. The panel always has an id, and the
// backdrop id is different, so a click in the body hits the panel and does
// not dismiss. A backdrop click calls OnOpenChange. When that callback is
// set, a click on the grab handle calls it with false. Escape handling and
// focus scope stay with the host.
//
// The backdrop paints at the modal layer: above the app topbar and sidebar,
// below the toast and command surfaces.
func Drawer(options DrawerOptions, children ...*ui.Element) *ui.Element {
	panelID := stablePartID(options.ID, "-panel", "of-drawer-panel")
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
		drawerGrab(panelID, options.OnOpenChange),
		overlayHeader("drawer", overlayHeading("drawer", options.Title, options.Description, titleID, descriptionID), nil),
		overlayBody("drawer", panelID, children),
		overlayStackedFooter("drawer", panelID, options.Footer),
	)

	backdrop := ui.Div().Class("of-drawer-backdrop").Children(options.Overrides.Apply(panel))
	backdrop.ID(stablePartID(options.ID, "-backdrop", "of-drawer-backdrop"))
	if options.OnOpenChange != nil {
		backdrop.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, false) })
	}
	root := ui.Div().Class("of-drawer")
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(trigger, backdrop)
}

// drawerGrab is the painted handle. A click dismisses when the host supplied
// OnOpenChange. The bar itself stays decorative.
func drawerGrab(panelID string, onOpenChange func(context.Context, bool) error) *ui.Element {
	grab := ui.Div().Class("of-drawer-grab").Children(
		ui.Span().Class("of-drawer-handle").Attr("aria-hidden", "true"),
	)
	if onOpenChange == nil {
		return grab
	}
	return grab.ID(partID(panelID, "-handle")).Attr("aria-label", "Close").OnClick(func(ctx context.Context) error {
		return onOpenChange(ctx, false)
	})
}
