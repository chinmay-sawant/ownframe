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
// nil when no trigger was supplied. A cancel or confirm control is omitted
// when its callback is nil and OnOpenChange is nil. A nil callback calls
// OnOpenChange(false) when that hook is set, so the button is not dead. An
// empty title keeps the aria-label. Escape handling and focus scope stay with
// the host. The backdrop does not dismiss.
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
		alertDialogFooter(options),
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

func alertDialogFooter(options AlertDialogOptions) *ui.Element {
	cancel := alertDialogCancel(options)
	confirm := alertDialogConfirm(options)
	if cancel == nil && confirm == nil {
		return nil
	}
	return ui.Tag("footer").Class("of-alert-dialog-footer").Children(cancel, confirm)
}

func alertDialogCancel(options AlertDialogOptions) *ui.Element {
	button := alertDialogButton(
		options.CancelLabel,
		defaultAlertCancel,
		ButtonOutlined,
		options.OnCancel,
		options.OnOpenChange,
	)
	if button == nil {
		return nil
	}
	// A wrapper keeps the hook class without replacing the button's own classes.
	return ui.Div().Class("of-alert-dialog-cancel").Children(button)
}

func alertDialogConfirm(options AlertDialogOptions) *ui.Element {
	variant := ButtonFilled
	if options.Destructive {
		variant = ButtonDanger
	}
	button := alertDialogButton(
		options.ConfirmLabel,
		defaultAlertConfirm,
		variant,
		options.OnConfirm,
		options.OnOpenChange,
	)
	if button == nil {
		return nil
	}
	return ui.Div().Class("of-alert-dialog-confirm").Children(button)
}

// alertDialogButton omits a control whose callback is nil. When the host only
// supplied OnOpenChange, that callback closes the dialog instead of painting
// a button that does nothing.
func alertDialogButton(
	label, fallback string,
	variant ButtonVariant,
	direct func(context.Context) error,
	onOpenChange func(context.Context, bool) error,
) *ui.Element {
	if direct == nil && onOpenChange == nil {
		return nil
	}
	if label == "" {
		label = fallback
	}
	button := Button(label, ButtonOptions{Variant: variant})
	if direct != nil {
		return button.OnClick(direct)
	}
	return button.OnClick(func(ctx context.Context) error { return onOpenChange(ctx, false) })
}
