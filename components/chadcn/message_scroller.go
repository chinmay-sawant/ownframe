package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// MessageScrollerOptions describes a chat viewport. ID is the host's stable
// root ID; the viewport and the translated transcript derive their IDs from
// it. Height zero keeps the stylesheet height. State is the scroll range
// ScrollArea uses; nil leaves the transcript at the start. Children are
// usually Messages.
type MessageScrollerOptions struct {
	ID       string
	Height   ui.Length
	State    *ScrollState
	Children []*ui.Element
}

// MessageScroller builds a clipped message viewport with a top fade. The host
// owns the transcript. A non-zero State offset translates the transcript inside
// the fixed viewport. Overflow does not scroll the pane. Root, viewport, and
// transcript carry stable part ids from ID when set. Offset moves still rewrite
// transform geometry and appends change the child count, so both stay on the
// full-render path on the current engine pin (see probe_patch_fix2_scroll_test.go).
func MessageScroller(options MessageScrollerOptions) *ui.Element {
	content := ui.Div().Class("of-message-scroller-content").Children(options.Children...)
	if contentID := partID(options.ID, "-content"); contentID != "" {
		content.ID(contentID)
	}
	if options.State != nil && options.State.Offset() != 0 {
		content.Transform(ui.TranslateY(ui.Px(-float64(options.State.Offset()))))
	}
	viewport := ui.Div().Class("of-message-scroller-viewport").Attr("role", "log").
		Attr("aria-live", "polite").Attr("aria-label", "Messages").Children(content)
	if viewportID := partID(options.ID, "-viewport"); viewportID != "" {
		viewport.ID(viewportID)
	}
	if options.Height != (ui.Length{}) {
		viewport.Height(options.Height)
	}
	fade := ui.Div().Class("of-message-scroller-fade").Attr("aria-hidden", "true")
	root := ui.Div().Class("of-message-scroller").Children(viewport, fade)
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root
}
