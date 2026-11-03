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
type StatCardOptions struct {
	Label string
	Value string
	Delta string
	Tone  StatDeltaTone
}

// StatCard builds a KPI row cell for a dashboard header strip. The delta pill
// only renders when Delta is non-empty.
func StatCard(options StatCardOptions) *ui.Element {
	root := ui.Div().Class("of-stat-card")
	if options.Label != "" {
		root.Children(ui.Span().Class("of-stat-card-label").Text(options.Label))
	}
	if options.Value != "" {
		root.Children(ui.Span().Class("of-stat-card-value").Text(options.Value))
	}
	if options.Delta != "" {
		root.Children(ui.Span().Class(classes("of-stat-card-delta", statDeltaClass(options.Tone))).Text(options.Delta))
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
