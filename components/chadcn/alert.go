package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// AlertVariant selects the tone of an Alert. The zero value is info.
type AlertVariant uint8

const (
	// AlertInfo is a neutral informational message.
	AlertInfo AlertVariant = iota
	// AlertSuccess confirms a completed action.
	AlertSuccess
	// AlertWarning calls out a condition that needs attention.
	AlertWarning
	// AlertDanger reports a failure or a destructive outcome.
	AlertDanger
)

// AlertOptions describes an inline status message. Icon, Action, Description,
// and Title are optional; a nil node or an empty string omits that slot.
type AlertOptions struct {
	Icon        *ui.Element
	Action      *ui.Element
	Title       string
	Description string
	Variant     AlertVariant
}

// Alert builds a bordered status panel with role="alert". It is static: the
// caller shows or removes it as part of the view, and there is no dismiss
// state to keep.
func Alert(options AlertOptions) *ui.Element {
	alert := ui.Div().Attr("role", "alert").
		Class(classes("of-alert", "of-alert--"+options.Variant.className()))
	if options.Icon != nil {
		alert.Children(ui.Div().Class("of-alert-icon").Attr("aria-hidden", "true").Children(options.Icon))
	}
	body := ui.Div().Class("of-alert-body")
	if options.Title != "" {
		body.Children(ui.Div().Class("of-alert-title").Text(options.Title))
	}
	if options.Description != "" {
		body.Children(ui.P().Class("of-alert-description").Text(options.Description))
	}
	alert.Children(body)
	if options.Action != nil {
		alert.Children(ui.Div().Class("of-alert-action").Children(options.Action))
	}
	return alert
}

func (v AlertVariant) className() string {
	switch v {
	case AlertSuccess:
		return "success"
	case AlertWarning:
		return "warning"
	case AlertDanger:
		return "danger"
	default:
		return "info"
	}
}
