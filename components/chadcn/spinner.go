package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// SpinnerSize selects the ring diameter. The zero value is medium.
type SpinnerSize uint8

const (
	// SpinnerMedium is the default 16px ring.
	SpinnerMedium SpinnerSize = iota
	// SpinnerSmall is the 12px ring.
	SpinnerSmall
	// SpinnerLarge is the 24px ring.
	SpinnerLarge
)

// SpinnerOptions describes the ring size and its accessible label.
type SpinnerOptions struct {
	Size  SpinnerSize
	Label string
}

// Spinner is an indeterminate loading ring driven by a CSS keyframe.
func Spinner(options SpinnerOptions) *ui.Element {
	size := "of-spinner--medium"
	switch options.Size {
	case SpinnerSmall:
		size = "of-spinner--small"
	case SpinnerLarge:
		size = "of-spinner--large"
	}
	label := options.Label
	if label == "" {
		label = "Loading"
	}
	return ui.Span().Class(classes("of-spinner", size)).
		Attr("role", "status").Attr("aria-label", label)
}
