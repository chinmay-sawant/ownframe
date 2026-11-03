package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// MessageScrollerOptions describes a chat viewport. Height zero keeps the
// stylesheet height. Children are usually Messages.
type MessageScrollerOptions struct {
	Height   ui.Length
	Children []*ui.Element
}

// MessageScroller builds a scrollable message viewport with a top fade. The
// host owns the transcript and the scroll position.
func MessageScroller(options MessageScrollerOptions) *ui.Element {
	viewport := ui.Div().Class("of-message-scroller-viewport").Attr("role", "log").
		Attr("aria-live", "polite").Attr("aria-label", "Messages").Children(options.Children...)
	if options.Height != (ui.Length{}) {
		viewport.Height(options.Height)
	}
	fade := ui.Div().Class("of-message-scroller-fade").Attr("aria-hidden", "true")
	return ui.Div().Class("of-message-scroller").Children(viewport, fade)
}
