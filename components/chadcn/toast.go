package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// ToastOptions describes one transient notification. Open false renders
// nothing, so the host can drop a closed toast from the view. CloseAction is
// dispatched by the close button when it is not empty.
type ToastOptions struct {
	Action      *ui.Element
	Title       string
	Description string
	CloseAction string
	Open        bool
}

// Toast builds an absolutely positioned notification in the bottom-right
// corner of its containing block. Mount it inside the block that owns that
// corner. The host owns the queue and decides when Open changes.
func Toast(options ToastOptions) *ui.Element {
	if !options.Open {
		return nil
	}
	toast := ui.Div().Class("of-toast").Attr("role", "status").Attr("aria-live", "polite").
		Attr("aria-atomic", "true").Position(ui.PositionAbsolute).
		Right(ui.Px(16)).Bottom(ui.Px(16)).Width(ui.Px(356)).
		BoxSizing(ui.BoxSizingBorderBox).ZIndex(ui.Number(60))
	body := ui.Div().Class("of-toast-body")
	if options.Title != "" {
		body.Children(ui.Div().Class("of-toast-title").Text(options.Title))
	}
	if options.Description != "" {
		body.Children(ui.P().Class("of-toast-description").Text(options.Description))
	}
	toast.Children(body)
	if options.Action != nil {
		toast.Children(ui.Div().Class("of-toast-action").Children(options.Action))
	}
	closeButton := ui.Button("×").Attr("type", "button").Class("of-toast-close").
		Attr("aria-label", "Close").Width(ui.Px(24)).Height(ui.Px(24))
	if options.CloseAction != "" {
		closeButton.Attr("data-action", options.CloseAction)
	}
	return toast.Children(closeButton)
}
