package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// StatusSegment is one bar in a StatusSegments row. Value sets the bar width
// relative to the other segments. Tone selects the palette slot, 1-based.
type StatusSegment struct {
	Label string
	Value float64
	Count string
	Tone  int
}

// StatusSegmentsOptions describes a segmented status row, such as work order
// counts by state. The bars share one track and the legend lists each count.
// The host owns the values; zero totals render an empty track.
type StatusSegmentsOptions struct {
	ID       string
	Label    string
	Overdue  string
	Segments []StatusSegment
}

// StatusSegments builds proportional status bars with a count legend. Bars are
// div geometry, so the row needs no canvas or SVG.
func StatusSegments(options StatusSegmentsOptions) *ui.Element {
	label := options.Label
	if label == "" {
		label = "Status"
	}
	root := ui.Div().Class("of-status-segments").Attr("role", "group").Attr("aria-label", label)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Overdue != "" {
		root.Children(ui.Div().Class("of-status-segments-overdue").Children(
			ui.Span().Class("of-status-segments-overdue-label").Text(options.Overdue),
		))
	}
	track := ui.Div().Class("of-status-segments-track").Attr("aria-hidden", "true")
	total := statusTotal(options.Segments)
	for _, item := range options.Segments {
		bar := ui.Div().Class("of-status-segments-bar").Attr("data-tone", statusToneSlot(item.Tone))
		if total > 0 && item.Value > 0 {
			bar.FlexGrow(ui.Number(item.Value))
		} else {
			bar.FlexGrow(ui.Number(0))
		}
		track.Children(bar)
	}
	root.Children(track)
	legend := ui.Div().Class("of-status-segments-legend")
	for _, item := range options.Segments {
		dot := ui.Span().Class("of-status-segments-dot").Attr("aria-hidden", "true").
			Attr("data-tone", statusToneSlot(item.Tone))
		row := ui.Div().Class("of-status-segments-item").Children(dot,
			ui.Span().Class("of-status-segments-name").Text(item.Label))
		if item.Count != "" {
			row.Children(ui.Span().Class("of-status-segments-count").Text(item.Count))
		}
		legend.Children(row)
	}
	return root.Children(legend)
}

func statusTotal(segments []StatusSegment) float64 {
	total := 0.0
	for _, item := range segments {
		if finite(item.Value) && item.Value > 0 {
			total += item.Value
		}
	}
	return total
}

func statusToneSlot(tone int) string {
	if tone < 1 {
		return "1"
	}
	return chartSeriesSlot(tone - 1)
}
