package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// StatDeltaTone selects the delta pill tone. The zero value is neutral.
type StatDeltaTone uint8

const (
	// StatDeltaNeutral is the muted pill for plain context.
	StatDeltaNeutral StatDeltaTone = iota
	// StatDeltaUp marks an improving delta.
	StatDeltaUp
	// StatDeltaDown marks a worsening delta.
	StatDeltaDown
	// StatDeltaWarn marks a delta that needs attention.
	StatDeltaWarn
)

// StatCardOptions describes one KPI cell: a label, a headline value, and an
// optional delta pill. The host owns the numbers and passes them in as text
// so currency, counts, and durations all render without formatting hooks.
// When ID is set, the label, value, and delta spans take stable suffixed ids,
// so a value change patches instead of taking the full-render path.
type StatCardOptions struct {
	ID    string
	Label string
	Value string
	Delta string
	Tone  StatDeltaTone
}

// StatCard builds a KPI row cell for a dashboard header strip. The delta pill
// only renders when Delta is non-empty.
func StatCard(options StatCardOptions) *ui.Element {
	root := ui.Div().Class("of-stat-card")
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Label != "" {
		label := ui.Span().Class("of-stat-card-label").Text(options.Label)
		if id := partID(options.ID, "-label"); id != "" {
			label.ID(id)
		}
		root.Children(label)
	}
	if options.Value != "" {
		value := ui.Span().Class("of-stat-card-value").Text(options.Value)
		if id := partID(options.ID, "-value"); id != "" {
			value.ID(id)
		}
		root.Children(value)
	}
	if options.Delta != "" {
		delta := ui.Span().Class(classes("of-stat-card-delta", statDeltaClass(options.Tone))).Text(options.Delta)
		if id := partID(options.ID, "-delta"); id != "" {
			delta.ID(id)
		}
		root.Children(delta)
	}
	return root
}

func statDeltaClass(tone StatDeltaTone) string {
	switch tone {
	case StatDeltaUp:
		return "of-stat-card-delta--up"
	case StatDeltaDown:
		return "of-stat-card-delta--down"
	case StatDeltaWarn:
		return "of-stat-card-delta--warn"
	default:
		return "of-stat-card-delta--neutral"
	}
}
