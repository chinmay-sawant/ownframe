package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// BubbleVariant selects the bubble surface. The zero value follows the theme
// surface.
type BubbleVariant uint8

const (
	// BubbleDefault is the received surface.
	BubbleDefault BubbleVariant = iota
	// BubblePrimary is the emphasized, sent-side surface.
	BubblePrimary
	// BubbleMuted is a quiet surface for secondary notes.
	BubbleMuted
)

// BubbleOptions describes a bubble's appearance.
type BubbleOptions struct {
	Variant BubbleVariant
}

// Bubble builds a rounded message bubble around its children. Paint and radius
// live in the shared stylesheet; the host supplies the text.
func Bubble(options BubbleOptions, children ...*ui.Element) *ui.Element {
	return ui.Div().Class(classes("of-bubble", bubbleVariantClass(options.Variant))).
		Children(children...)
}

func bubbleVariantClass(variant BubbleVariant) string {
	switch variant {
	case BubblePrimary:
		return "of-bubble--primary"
	case BubbleMuted:
		return "of-bubble--muted"
	default:
		return "of-bubble--default"
	}
}
