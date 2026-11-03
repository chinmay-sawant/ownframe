package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// MarkerState selects the marker accent. The zero value is neutral.
type MarkerState uint8

const (
	// MarkerDefault is a neutral timeline note.
	MarkerDefault MarkerState = iota
	// MarkerSuccess marks a completed step.
	MarkerSuccess
	// MarkerWarning marks a step that needs attention.
	MarkerWarning
	// MarkerDanger marks a failed step.
	MarkerDanger
)

// MarkerOptions describes a timeline or status marker.
type MarkerOptions struct {
	Label  string
	Detail string
	State  MarkerState
}

// Marker builds a small status marker with a ring icon, a label, and optional
// detail text.
func Marker(options MarkerOptions) *ui.Element {
	marker := ui.Div().Class(classes("of-marker", markerStateClass(options.State)))
	marker.Children(ui.Span().Class("of-marker-icon").Attr("aria-hidden", "true"))
	body := ui.Div().Class("of-marker-body")
	if options.Label != "" {
		body.Children(ui.Span().Class("of-marker-label").Text(options.Label))
	}
	if options.Detail != "" {
		body.Children(ui.Span().Class("of-marker-detail").Text(options.Detail))
	}
	return marker.Children(body)
}

func markerStateClass(state MarkerState) string {
	switch state {
	case MarkerSuccess:
		return "of-marker--success"
	case MarkerWarning:
		return "of-marker--warning"
	case MarkerDanger:
		return "of-marker--danger"
	default:
		return "of-marker--default"
	}
}
