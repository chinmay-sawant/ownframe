package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// BadgeVariant selects a badge tone. The zero value is the filled default.
type BadgeVariant uint8

const (
	// BadgeDefault is the filled primary badge.
	BadgeDefault BadgeVariant = iota
	// BadgeSecondary is the muted secondary badge.
	BadgeSecondary
	// BadgeOutline is the transparent badge with a border.
	BadgeOutline
	// BadgeDanger marks a destructive or failed state.
	BadgeDanger
)

// BadgeOptions describes a badge tone.
type BadgeOptions struct {
	Variant BadgeVariant
}

// Badge is a short non-interactive label. Callers may omit the options to get
// the filled default.
func Badge(label string, options ...BadgeOptions) *ui.Element {
	variant := "of-badge--default"
	if len(options) > 0 {
		switch options[0].Variant {
		case BadgeSecondary:
			variant = "of-badge--secondary"
		case BadgeOutline:
			variant = "of-badge--outline"
		case BadgeDanger:
			variant = "of-badge--danger"
		}
	}
	return ui.Span().Class(classes("of-badge", variant)).Text(label)
}
