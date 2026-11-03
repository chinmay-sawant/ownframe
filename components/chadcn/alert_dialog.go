package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// AlertDialogOptions describes a controlled confirmation dialog. The host owns
// Open and focus. CancelLabel and ConfirmLabel default to "Cancel" and
// "Continue"; the confirm control is filled, or danger-filled when Destructive
// is set. The backdrop is inert: the host closes the dialog from OnConfirm or
// OnCancel, not from a click-away. When OnOpenChange is set the component binds
// the trigger, so pass a trigger without its own action.
type AlertDialogOptions struct {
	ID           string
	Open         bool
	Title        string
	Description  string
	ConfirmLabel string
	CancelLabel  string
	Trigger      *ui.Element
	Destructive  bool
	OnConfirm    func(context.Context) error
	OnCancel     func(context.Context) error
	OnOpenChange func(context.Context, bool) error
	Overrides    theme.Overrides
}

const (
	defaultAlertConfirm = "Continue"
	defaultAlertCancel  = "Cancel"
)

// AlertDialog builds a modal confirmation surface with role "alertdialog" and
// its own cancel and confirm controls. It returns the trigger when closed, or
// nil when no trigger was supplied. Escape handling and focus scope stay with
// the host.
func AlertDialog(options AlertDialogOptions) *ui.Element {
	panelID := partID(options.ID, "-panel")
	trigger := overlayTrigger(options.Trigger, options.Open, panelID, options.OnOpenChange)
	if !options.Open {
		return trigger
	}
	titleID, descriptionID := overlayPartIDs(panelID, options.Title, options.Description)
	panel := ui.Section().Class("of-alert-dialog-panel")
	if panelID != "" {
		panel.ID(panelID)
	}
	overlayPanelAttrs(panel, "alertdialog", "Alert dialog", titleID, descriptionID)
	panel.Children(
		overlayHeader("alert-dialog", overlayHeading("alert-dialog", options.Title, options.Description, titleID, descriptionID), nil),
		ui.Tag("footer").Class("of-alert-dialog-footer").Children(
			alertDialogCancel(options),
			alertDialogConfirm(options),
		),
	)

	backdrop := ui.Div().Class("of-alert-dialog-backdrop").
		ZIndex(ui.Number(50)).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).
		JustifyContent(ui.JustifyContentCenter).Children(options.Overrides.Apply(panel))
	if id := partID(options.ID, "-backdrop"); id != "" {
		backdrop.ID(id)
	}
	root := ui.Div().Class("of-alert-dialog")
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(trigger, backdrop)
}

func alertDialogCancel(options AlertDialogOptions) *ui.Element {
	label := options.CancelLabel
	if label == "" {
		label = defaultAlertCancel
	}
	cancel := Button(label, ButtonOptions{Variant: ButtonOutlined})
	if options.OnCancel != nil {
		cancel.OnClick(options.OnCancel)
	}
	// A wrapper keeps the hook class without replacing the button's own classes.
	return ui.Div().Class("of-alert-dialog-cancel").Children(cancel)
}

func alertDialogConfirm(options AlertDialogOptions) *ui.Element {
	label := options.ConfirmLabel
	if label == "" {
		label = defaultAlertConfirm
	}
	variant := ButtonFilled
	if options.Destructive {
		variant = ButtonDanger
	}
	confirm := Button(label, ButtonOptions{Variant: variant})
	if options.OnConfirm != nil {
		confirm.OnClick(options.OnConfirm)
	}
	return ui.Div().Class("of-alert-dialog-confirm").Children(confirm)
}
