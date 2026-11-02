package chadcn

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// tracePercent is the full length of a trace track.
const tracePercent = 100

// TraceRow is one timed step. Start and Width are percentages of the track.
// Status is a free-form word such as "ok", "warning", or "failed".
type TraceRow struct {
	Label  string
	Detail string
	Start  float64
	Width  float64
	Status string
}

// TraceOptions describes a trace list.
type TraceOptions struct {
	Rows []TraceRow
}

// Trace builds a list of timed steps with a status ring and a CSS timing bar.
// The host owns the rows and the clock.
func Trace(options TraceOptions) *ui.Element {
	root := ui.Div().Class("of-trace").Attr("role", "list")
	for _, row := range options.Rows {
		state := traceState(row.Status)
		item := ui.Div().Class(classes("of-trace-row", "of-trace-row--"+state)).Attr("role", "listitem")
		item.Children(ui.Span().Class("of-trace-row-icon").Attr("aria-hidden", "true").Text(traceGlyph(state)))

		body := ui.Div().Class("of-trace-row-body")
		header := ui.Div().Class("of-trace-row-header")
		if row.Label != "" {
			header.Children(ui.Span().Class("of-trace-row-label").Text(row.Label))
		}
		if row.Detail != "" {
			header.Children(ui.Span().Class("of-trace-row-detail").Text(row.Detail))
		}
		body.Children(header)

		start, width := traceSpan(row.Start, row.Width)
		if width > 0 {
			bar := ui.Div().Class("of-trace-bar").Width(ui.Percent(width)).MarginLeft(ui.Percent(start))
			body.Children(ui.Div().Class("of-trace-track").Attr("aria-hidden", "true").Children(bar))
		}
		item.Children(body)
		root.Children(item)
	}
	return root
}

// traceSpan clamps a row inside the 0..100 track.
func traceSpan(start, width float64) (float64, float64) {
	if !finite(start) || start < 0 {
		start = 0
	}
	if start > tracePercent {
		start = tracePercent
	}
	if !finite(width) || width < 0 {
		width = 0
	}
	return start, min(width, tracePercent-start)
}

func traceState(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success", "ok", "passed", "done", "complete", "completed":
		return "success"
	case "warning", "warn", "pending", "running", "retry":
		return "warning"
	case "danger", "error", "errored", "failed", "failure":
		return "danger"
	default:
		return "default"
	}
}

func traceGlyph(state string) string {
	switch state {
	case "success":
		return "✓"
	case "warning", "danger":
		return "!"
	default:
		return ""
	}
}
