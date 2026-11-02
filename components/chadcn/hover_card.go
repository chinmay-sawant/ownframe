package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// HoverCardOptions describes a controlled hover card. Open comes from the host
// after it reads Handlers.Hover for the trigger. ID names the root and derives
// the content ID. Side places the card; its zero value is PopupBottom. Give the
// trigger a stable ID so the host can match the hover report.
type HoverCardOptions struct {
	ID   string
	Open bool
	Side PopupSide
}

// HoverCard renders a trigger and, when Open is true, an anchored card holding
// the content. The host owns hover timing: read Handlers.Hover, set Open, and
// rebuild the view. Keep a small delay in the host so a card does not flash
// while the pointer crosses the trigger. A nil trigger returns nil.
func HoverCard(options HoverCardOptions, trigger *ui.Element, content ...*ui.Element) *ui.Element {
	if trigger == nil {
		return nil
	}
	side := normalizePopupSide(options.Side)
	root := ui.Div().Class(classes("of-hover-card", popupSideClass("hover-card", side)))
	if options.ID != "" {
		root.ID(options.ID)
	}
	triggerBox := ui.Span().Class("of-hover-card-trigger").Children(trigger)
	if !options.Open {
		return root.Children(triggerBox)
	}
	card := ui.Div().Class("of-hover-card-content").Attr("role", "dialog").Children(content...)
	if contentID := partID(options.ID, "-content"); contentID != "" {
		card.ID(contentID)
	}
	anchor := ui.Span().Class("of-hover-card-anchor").Children(card)
	if side == PopupBottom {
		return root.Children(triggerBox, anchor)
	}
	return root.Children(anchor, triggerBox)
}
