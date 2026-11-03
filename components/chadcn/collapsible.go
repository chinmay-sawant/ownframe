package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// CollapsibleOptions describes a controlled single disclosure. Open comes from
// the host. ID names the root and derives the content ID. Trigger is rendered
// in both states and receives aria-expanded, aria-controls, and, when
// OnOpenChange is set, the toggle click.
type CollapsibleOptions struct {
	Trigger      *ui.Element
	OnOpenChange func(ctx context.Context, open bool) error
	ID           string
	Open         bool
}

// Collapsible renders a host-supplied trigger and, when Open is true, its
// content. Pass a trigger without its own action, since the constructor wires
// the click to OnOpenChange. A nil trigger returns nil.
func Collapsible(options CollapsibleOptions, children ...*ui.Element) *ui.Element {
	if options.Trigger == nil {
		return nil
	}
	contentID := partID(options.ID, "-content")
	trigger := options.Trigger
	trigger.Attr("aria-expanded", ariaBool(options.Open))
	if options.Open && contentID != "" {
		trigger.Attr("aria-controls", contentID)
	}
	if options.OnOpenChange != nil {
		next := !options.Open
		trigger.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, next) })
	}
	root := ui.Div().Class("of-collapsible")
	if options.ID != "" {
		root.ID(options.ID)
	}
	root.Children(trigger)
	if options.Open {
		content := ui.Div().Class("of-collapsible-content")
		if contentID != "" {
			content.ID(contentID)
		}
		root.Children(content.Children(children...))
	}
	return root
}
