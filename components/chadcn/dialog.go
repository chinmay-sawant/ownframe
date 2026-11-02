package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// DialogOptions describes a controlled modal dialog. The host owns Open and
// focus: set Open from OnOpenChange, rebuild the view, and scope focus with
// DialogFocus while the panel is mounted.
//
// ID is the stable host ID. The panel and backdrop derive their IDs from it;
// PanelID and BackdropID override the derived values. A panel ID keeps a click
// inside the panel from reaching the backdrop action.
//
// Title and Description build the standard header. Children fill the body and
// Footer holds the action row. Without Title, Description, or Footer, children
// fill the panel directly, which keeps the pre-Open call shape working. When
// OnOpenChange is set the component binds the trigger, so pass a trigger
// without its own action.
type DialogOptions struct {
	ID           string
	BackdropID   string
	PanelID      string
	Label        string
	Title        string
	Description  string
	CloseAction  string
	Trigger      *ui.Element
	Footer       []*ui.Element
	Open         bool
	Width        ui.Length
	OnOpenChange func(context.Context, bool) error
	Overrides    theme.Overrides
}

// Dialog builds a centered modal surface. When Open is false it returns the
// trigger, or nil when no trigger was supplied. A backdrop click and the header
// close control dispatch CloseAction, or call OnOpenChange when no action is
// set. Click-away, Escape handling, and focus scope stay with the host.
//
// The trigger-based shape mounts the backdrop fixed to the viewport. The
// pre-Open shape, which had no trigger, returns the backdrop itself and fills
// the positioned container the caller mounts it in.
func Dialog(options DialogOptions, children ...*ui.Element) *ui.Element {
	legacy := options.legacyOpen()
	panelID := options.panelID()
	open := options.Open || legacy
	trigger := overlayTrigger(options.Trigger, open, panelID, options.OnOpenChange)
	if !open {
		return trigger
	}
	titleID, descriptionID := overlayPartIDs(panelID, options.Title, options.Description)
	panel := ui.Section().Class("of-dialog-panel")
	if panelID != "" {
		panel.ID(panelID)
	}
	overlayPanelAttrs(panel, "dialog", overlayLabel(options.Label, "Dialog"), titleID, descriptionID)
	if options.Width != (ui.Length{}) {
		panel.Width(options.Width).MaxWidth(ui.Percent(100))
	}

	structured := options.Title != "" || options.Description != "" || len(options.Footer) > 0
	if structured {
		heading := overlayHeading("dialog", options.Title, options.Description, titleID, descriptionID)
		panel.Children(overlayHeader("dialog", heading, overlayClose("dialog", options.CloseAction, options.OnOpenChange)),
			overlayBody("dialog", children), overlayFooter("dialog", options.Footer))
	} else {
		panel.Children(children...)
	}

	backdrop := dialogBackdrop(options, options.Overrides.Apply(panel), legacy)
	if legacy {
		return backdrop
	}
	root := ui.Div().Class("of-dialog")
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(trigger, backdrop)
}

// legacyOpen reports the pre-Open contract: a caller that supplied no trigger
// rendered Dialog only while it was open and identified it by its part IDs.
func (o DialogOptions) legacyOpen() bool {
	return !o.Open && o.Trigger == nil && (o.BackdropID != "" || o.PanelID != "")
}

func (o DialogOptions) panelID() string {
	if o.PanelID != "" {
		return o.PanelID
	}
	return partID(o.ID, "-panel")
}

func (o DialogOptions) backdropID() string {
	if o.BackdropID != "" {
		return o.BackdropID
	}
	return partID(o.ID, "-backdrop")
}

func dialogBackdrop(options DialogOptions, panel *ui.Element, contained bool) *ui.Element {
	backdrop := ui.Div().Class("of-dialog-backdrop").
		ZIndex(ui.Number(50)).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).
		JustifyContent(ui.JustifyContentCenter).Children(panel)
	if contained {
		backdrop.Position(ui.PositionAbsolute).Top(ui.Px(0)).Left(ui.Px(0)).Right(ui.Px(0)).Bottom(ui.Px(0)).
			Width(ui.WidthAuto).Height(ui.HeightAuto)
	}
	if id := options.backdropID(); id != "" {
		backdrop.ID(id)
	}
	switch {
	case options.CloseAction != "":
		backdrop.Attr("data-action", options.CloseAction)
	case options.OnOpenChange != nil:
		backdrop.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, false) })
	}
	return backdrop
}

// overlayTrigger publishes the popup state on a caller-supplied trigger and
// routes activation to OnOpenChange. The component owns the trigger when a
// callback is set, so the trigger must not also carry its own action.
func overlayTrigger(trigger *ui.Element, expanded bool, panelID string, onOpenChange func(context.Context, bool) error) *ui.Element {
	if trigger == nil {
		return nil
	}
	trigger.Attr("aria-haspopup", "dialog").Attr("aria-expanded", ariaBool(expanded))
	if expanded && panelID != "" {
		trigger.Attr("aria-controls", panelID)
	}
	if onOpenChange != nil {
		trigger.OnClick(func(ctx context.Context) error { return onOpenChange(ctx, !expanded) })
	}
	return trigger
}

// overlayPanelAttrs marks a modal surface and labels it from its title, or from
// the fallback label when no title element exists.
func overlayPanelAttrs(panel *ui.Element, role, label, titleID, descriptionID string) {
	panel.Attr("role", role).Attr("aria-modal", "true").Attr("tabindex", "-1")
	if titleID != "" {
		panel.Attr("aria-labelledby", titleID)
	} else {
		panel.Attr("aria-label", label)
	}
	if descriptionID != "" {
		panel.Attr("aria-describedby", descriptionID)
	}
}

func overlayPartIDs(panelID, title, description string) (titleID, descriptionID string) {
	if title != "" {
		titleID = partID(panelID, "-title")
	}
	if description != "" {
		descriptionID = partID(panelID, "-description")
	}
	return titleID, descriptionID
}

func overlayLabel(label, fallback string) string {
	if label != "" {
		return label
	}
	return fallback
}

func overlayHeading(component, title, description, titleID, descriptionID string) *ui.Element {
	if title == "" && description == "" {
		return nil
	}
	heading := ui.Div().Class("of-" + component + "-heading")
	if title != "" {
		titleElement := ui.Tag("h2").Class("of-" + component + "-title").Text(title)
		if titleID != "" {
			titleElement.ID(titleID)
		}
		heading.Children(titleElement)
	}
	if description != "" {
		descriptionElement := ui.P().Class("of-" + component + "-description").Text(description)
		if descriptionID != "" {
			descriptionElement.ID(descriptionID)
		}
		heading.Children(descriptionElement)
	}
	return heading
}

func overlayHeader(component string, heading, close *ui.Element) *ui.Element {
	if heading == nil && close == nil {
		return nil
	}
	return ui.Tag("header").Class("of-"+component+"-header").Children(heading, close)
}

// overlayClose builds the icon-only close control. It disappears when the host
// supplied no way to close, so the header never shows a dead button.
func overlayClose(component, closeAction string, onOpenChange func(context.Context, bool) error) *ui.Element {
	if closeAction == "" && onOpenChange == nil {
		return nil
	}
	closeButton := ui.Button("×").Class("of-"+component+"-close").
		Attr("type", "button").Attr("aria-label", "Close")
	if closeAction != "" {
		return closeButton.Attr("data-action", closeAction)
	}
	return closeButton.OnClick(func(ctx context.Context) error { return onOpenChange(ctx, false) })
}

func overlayBody(component string, children []*ui.Element) *ui.Element {
	if len(children) == 0 {
		return nil
	}
	return ui.Div().Class("of-" + component + "-body").Children(children...)
}

func overlayFooter(component string, children []*ui.Element) *ui.Element {
	if len(children) == 0 {
		return nil
	}
	return ui.Tag("footer").Class("of-" + component + "-footer").Children(children...)
}

// overlayStackedFooter pins one full-width row per child, which is the action
// layout the sheet and drawer references use.
func overlayStackedFooter(component string, children []*ui.Element) *ui.Element {
	slots := make([]*ui.Element, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		slots = append(slots, ui.Div().Class("of-"+component+"-footer-slot").Children(child))
	}
	if len(slots) == 0 {
		return nil
	}
	return ui.Tag("footer").Class("of-" + component + "-footer").Children(slots...)
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
