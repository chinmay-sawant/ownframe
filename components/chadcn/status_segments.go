package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

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
// The host owns the values; zero totals render an empty track. When ID is set,
// each bar, name, and count takes a stable suffixed id, so a same-shape data
// change (same segments, only values and counts) patches instead of taking
// the full-render path. Added or removed segments stay FULL by design.
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
		overdue := ui.Span().Class("of-status-segments-overdue-label").Text(options.Overdue)
		if id := partID(options.ID, "-overdue-label"); id != "" {
			overdue.ID(id)
		}
		root.Children(ui.Div().Class("of-status-segments-overdue").Children(overdue))
	}
	track := ui.Div().Class("of-status-segments-track").Attr("aria-hidden", "true")
	total := statusTotal(options.Segments)
	for index, item := range options.Segments {
		bar := ui.Div().Class("of-status-segments-bar").Attr("data-tone", statusToneSlot(item.Tone))
		if total > 0 && item.Value > 0 {
			bar.FlexGrow(ui.Number(item.Value))
		} else {
			bar.FlexGrow(ui.Number(0))
		}
		if id := partID(options.ID, "-bar-"+strconv.Itoa(index)); id != "" {
			bar.ID(id)
		}
		track.Children(bar)
	}
	root.Children(track)
	legend := ui.Div().Class("of-status-segments-legend")
	for index, item := range options.Segments {
		dot := ui.Span().Class("of-status-segments-dot").Attr("aria-hidden", "true").
			Attr("data-tone", statusToneSlot(item.Tone))
		name := ui.Span().Class("of-status-segments-name").Text(item.Label)
		if id := partID(options.ID, "-name-"+strconv.Itoa(index)); id != "" {
			name.ID(id)
		}
		row := ui.Div().Class("of-status-segments-item").Children(dot, name)
		if item.Count != "" {
			count := ui.Span().Class("of-status-segments-count").Text(item.Count)
			if id := partID(options.ID, "-count-"+strconv.Itoa(index)); id != "" {
				count.ID(id)
			}
			row.Children(count)
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
