package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// MessageScrollerOptions describes a chat viewport. Height zero keeps the
// stylesheet height. State is the scroll range ScrollArea uses; nil leaves the
// transcript at the start. Children are usually Messages.
type MessageScrollerOptions struct {
	Height   ui.Length
	State    *ScrollState
	Children []*ui.Element
}

// MessageScroller builds a clipped message viewport with a top fade. The host
// owns the transcript. A non-zero State offset translates the transcript inside
// the fixed viewport. Overflow does not scroll the pane.
func MessageScroller(options MessageScrollerOptions) *ui.Element {
	content := ui.Div().Class("of-message-scroller-content").Children(options.Children...)
	if options.State != nil && options.State.Offset() != 0 {
		content.Transform(ui.TranslateY(ui.Px(-float64(options.State.Offset()))))
	}
	viewport := ui.Div().Class("of-message-scroller-viewport").Attr("role", "log").
		Attr("aria-live", "polite").Attr("aria-label", "Messages").Children(content)
	if options.Height != (ui.Length{}) {
		viewport.Height(options.Height)
	}
	fade := ui.Div().Class("of-message-scroller-fade").Attr("aria-hidden", "true")
	return ui.Div().Class("of-message-scroller").Children(viewport, fade)
}
