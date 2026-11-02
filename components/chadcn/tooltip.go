package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// TooltipOptions describes a controlled tooltip. Open comes from the host
// after it reads Handlers.Hover for the trigger. Content is plain text. ID
// names the root and derives the content ID. Side places the label; its zero
// value is PopupBottom. Give the trigger a stable ID so the host can match the
// hover report.
type TooltipOptions struct {
	ID      string
	Open    bool
	Content string
	Side    PopupSide
}

// Tooltip renders a trigger and, when Open is true, a label above, below, or
// beside it. The label carries role="tooltip", and the trigger points at it
// with aria-describedby. The host owns hover and focus timing: read
// Handlers.Hover, set Open, and rebuild the view. A nil trigger returns nil.
func Tooltip(options TooltipOptions, trigger *ui.Element) *ui.Element {
	if trigger == nil {
		return nil
	}
	contentID := partID(options.ID, "-content")
	if options.Open && contentID != "" {
		trigger.Attr("aria-describedby", contentID)
	}
	side := normalizePopupSide(options.Side)
	root := ui.Span().Class(classes("of-tooltip", popupSideClass("tooltip", side)))
	if options.ID != "" {
		root.ID(options.ID)
	}
	triggerBox := ui.Span().Class("of-tooltip-trigger").Children(trigger)
	if !options.Open {
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
