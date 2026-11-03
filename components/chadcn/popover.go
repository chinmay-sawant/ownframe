package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// PopupSide selects where a popup panel sits relative to its trigger. It is
// shared by Popover, Tooltip, and HoverCard. The zero value is PopupBottom.
type PopupSide uint8

const (
	PopupBottom PopupSide = iota
	PopupTop
	PopupLeft
	PopupRight
)

func popupSideClass(component string, side PopupSide) string {
	switch side {
	case PopupTop:
		return "of-" + component + "--top"
	case PopupLeft:
		return "of-" + component + "--left"
	case PopupRight:
		return "of-" + component + "--right"
	default:
		return "of-" + component + "--bottom"
	}
}

// normalizePopupSide maps an unknown value onto the zero value so the class
// and the anchor order always agree.
func normalizePopupSide(side PopupSide) PopupSide {
	switch side {
	case PopupTop, PopupLeft, PopupRight:
		return side
	default:
		return PopupBottom
	}
}

// PopoverAlign selects where the panel lines up with the trigger along the
// cross axis. The zero value is PopoverAlignStart. Centring uses the panel's
// token width and stays exact while that width is fixed. On the left and right
// sides the panel aligns with the trigger's top edge.
type PopoverAlign uint8

const (
	PopoverAlignStart PopoverAlign = iota
	PopoverAlignCenter
	PopoverAlignEnd
)

func popoverAlignClass(align PopoverAlign) string {
	switch align {
	case PopoverAlignCenter:
		return "of-popover--align-center"
	case PopoverAlignEnd:
		return "of-popover--align-end"
	default:
		return "of-popover--align-start"
	}
}

// PopoverOptions describes a controlled popover. Open comes from the host.
// OnOpenChange receives the next state when the trigger is clicked. ID names
// the root and derives the content ID. Trigger is rendered in both states.
// Side and Align place the panel; their zero values are PopupBottom and
// PopoverAlignStart.
type PopoverOptions struct {
	ID           string
	Open         bool
	Side         PopupSide
	Align        PopoverAlign
	Trigger      *ui.Element
	OnOpenChange func(ctx context.Context, open bool) error
}

// Popover renders a trigger and, when Open is true, an anchored panel holding
// the children. The host owns open state, Escape handling, click-away, and
// focus scope. The trigger receives aria-expanded, aria-haspopup, and, while
// open, aria-controls. When OnOpenChange is set the trigger click is wired to
// it, so pass a trigger without its own action. A nil trigger returns nil.
func Popover(options PopoverOptions, children ...*ui.Element) *ui.Element {
	contentID := partID(options.ID, "-content")
	trigger := overlayTrigger(options.Trigger, options.Open, contentID, options.OnOpenChange)
	if trigger == nil {
		return nil
	}
	side := normalizePopupSide(options.Side)
	root := ui.Div().Class(classes(
		"of-popover",
		popupSideClass("popover", side),
		popoverAlignClass(options.Align),
	))
	if options.ID != "" {
		root.ID(options.ID)
	}
	triggerBox := ui.Span().Class("of-popover-trigger").Children(trigger)
	if !options.Open {
		return root.Children(triggerBox)
	}
	anchor := ui.Span().Class("of-popover-anchor")
	content := ui.Div().Class("of-popover-content").Attr("role", "dialog").Children(children...)
	if contentID != "" {
		content.ID(contentID)
	}
	anchor.Children(content)
	if side == PopupBottom {
		return root.Children(triggerBox, anchor)
	}
	return root.Children(anchor, triggerBox)
}
