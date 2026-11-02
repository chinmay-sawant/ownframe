package components

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// DialogOptions identifies a controlled modal. CloseAction is dispatched when
// the backdrop is clicked. The caller owns open state and Escape handling.
type DialogOptions struct {
	BackdropID  string
	PanelID     string
	Label       string
	CloseAction string
	Width       ui.Length
	Overrides   theme.Overrides
}

// Dialog builds a labeled modal surface. Pair it with DialogFocus and keep all
// focusable descendants in the focus scope while it is visible.
func Dialog(options DialogOptions, children ...*ui.Element) *ui.Element {
	return ui.Div().ID(options.BackdropID).Attr("data-action", options.CloseAction).
		Position(ui.PositionAbsolute).Top(ui.Px(0)).Left(ui.Px(0)).Right(ui.Px(0)).Bottom(ui.Px(0)).
		ZIndex(ui.Number(50)).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).
		JustifyContent(ui.JustifyContentCenter).BackgroundColor(theme.Backdrop()).
		Children(options.Overrides.Apply(Card(children...)).ID(options.PanelID).Attr("role", "dialog").Attr("aria-modal", "true").
			Attr("aria-label", options.Label).Width(options.Width).Padding(ui.Px(0)).Overflow(ui.OverflowHidden))
}

// DialogFocus saves focus for one modal. Nested modals are not supported.
// Activate before updating the view, Deactivate before removing the dialog.
type DialogFocus struct {
	active   bool
	previous string
}

// Activate limits Tab traversal and requests the first control on the next draw.
func (d *DialogFocus) Activate(page *ownframe.Page, ids ...string) {
	if d.active || len(ids) == 0 {
		return
	}
	d.previous = page.Form().FocusID()
	if d.previous == "" {
		d.previous = page.Edit().LastFocusID()
	}
	d.active = true
	page.Form().SetFocusScope(ids...)
	page.Form().RequestFocus(ids[0])
}

// Deactivate restores normal Tab traversal and requests the saved focus.
func (d *DialogFocus) Deactivate(page *ownframe.Page) {
	if !d.active {
		return
	}
	d.active = false
	page.Form().SetFocusScope()
	page.Form().RequestFocus(d.previous)
}
