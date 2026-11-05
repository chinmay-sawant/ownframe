package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// TooltipOptions describes a controlled tooltip. Open comes from the host.
// CSSHover keeps the label mounted while Open is false and shows it from the
// trigger :hover rule, so the host does not set Open for that path. The host
// can still toggle of-tooltip--hovered with a local element patch. Content is
// plain text. ID names the root and derives the content ID. Side places the
// label; its zero value is PopupBottom.
type TooltipOptions struct {
	ID       string
	Open     bool
	CSSHover bool
	Content  string
	Side     PopupSide
}

// Tooltip renders a trigger and, when Open is true, a label above, below, or
// beside it. The label carries role="tooltip", and the trigger points at it
// with aria-describedby. The host owns hover and focus timing. A nil trigger
// returns nil.
func Tooltip(options TooltipOptions, trigger *ui.Element) *ui.Element {
	if trigger == nil {
		return nil
	}
	contentID := partID(options.ID, "-content")
	if (options.Open || options.CSSHover) && contentID != "" {
		trigger.Attr("aria-describedby", contentID)
	}
	side := normalizePopupSide(options.Side)
	mode := ""
	if options.CSSHover && !options.Open {
		mode = "of-tooltip--css-hover"
	}
	if options.CSSHover && options.Open {
		mode = "of-tooltip--css-hover of-tooltip--hovered"
	}
	root := ui.Span().Class(classes("of-tooltip", popupSideClass("tooltip", side), mode))
	if options.ID != "" {
		root.ID(options.ID)
	}
	triggerBox := ui.Span().Class("of-tooltip-trigger").Children(trigger)
	if options.CSSHover {
		// :hover matches only the element with the hover id. The id has to sit
		// on the trigger, not only on the root.
		triggerBox.ID(cssHoverTriggerID(options.ID, "tooltip"))
	}
	if !options.Open && !options.CSSHover {
		return root.Children(triggerBox)
	}
	content := ui.Span().Class("of-tooltip-content").Attr("role", "tooltip").Text(options.Content)
	if contentID != "" {
		content.ID(contentID)
	}
	anchor := ui.Span().Class("of-tooltip-anchor").Children(content)
	if side == PopupBottom {
		return root.Children(triggerBox, anchor)
	}
	return root.Children(anchor, triggerBox)
}

// cssHoverTriggerID is the id the pointer hits. An empty host id still gets a
// stable trigger id, because an element without an id never matches :hover.
func cssHoverTriggerID(id, component string) string {
	if part := partID(id, "-trigger"); part != "" {
		return part
	}
	return "of-" + component + "-trigger"
}
