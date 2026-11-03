package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// HoverCardOptions describes a controlled hover card. Open comes from the
// host. CSSHover keeps closed content mounted so the host can toggle the
// of-hover-card--hovered class with a local element patch. ID names the root
// and derives the content ID. Side places the card; its zero value is
// PopupBottom.
type HoverCardOptions struct {
	ID       string
	Open     bool
	CSSHover bool
	Side     PopupSide
}

// HoverCard renders a trigger and, when Open is true, an anchored card holding
// the content. The host owns hover timing. A nil trigger returns nil.
func HoverCard(options HoverCardOptions, trigger *ui.Element, content ...*ui.Element) *ui.Element {
	if trigger == nil {
		return nil
	}
	side := normalizePopupSide(options.Side)
	mode := ""
	if options.CSSHover && !options.Open {
		mode = "of-hover-card--css-hover"
	}
	if options.CSSHover && options.Open {
		mode = "of-hover-card--css-hover of-hover-card--hovered"
	}
	root := ui.Div().Class(classes("of-hover-card", popupSideClass("hover-card", side), mode))
	if options.ID != "" {
		root.ID(options.ID)
	}
	triggerBox := ui.Span().Class("of-hover-card-trigger").Children(trigger)
	if !options.Open && !options.CSSHover {
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
