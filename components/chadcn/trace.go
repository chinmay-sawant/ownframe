package chadcn

import (
	"strconv"
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

// TraceOptions describes a trace list. ID is an optional stable host id;
// each row and its label and detail parts derive their ids from it so a
// same-row detail edit patches instead of reparsing.
type TraceOptions struct {
	ID   string
	Rows []TraceRow
}

// Trace builds a list of timed steps with a status ring and a CSS timing bar.
// The host owns the rows and the clock.
func Trace(options TraceOptions) *ui.Element {
	root := ui.Div().Class("of-trace").Attr("role", "list")
	if options.ID != "" {
		root.ID(options.ID)
	}
	for index, row := range options.Rows {
		state := traceState(row.Status)
		rowID := partID(options.ID, "-row-"+strconv.Itoa(index))
		item := ui.Div().Class(classes("of-trace-row", "of-trace-row--"+state)).Attr("role", "listitem")
		if rowID != "" {
			item.ID(rowID)
		}
		item.Children(ui.Span().Class("of-trace-row-icon").Attr("aria-hidden", "true").Text(traceGlyph(state)))

		body := ui.Div().Class("of-trace-row-body")
		header := ui.Div().Class("of-trace-row-header")
		if row.Label != "" {
			label := ui.Span().Class("of-trace-row-label").Text(row.Label)
			if id := partID(rowID, "-label"); id != "" {
				label.ID(id)
			}
			header.Children(label)
		}
		if row.Detail != "" {
			detail := ui.Span().Class("of-trace-row-detail").Text(row.Detail)
			if id := partID(rowID, "-detail"); id != "" {
				detail.ID(id)
			}
			header.Children(detail)
		}
		body.Children(header)

		start, width := traceSpan(row.Start, row.Width)
		if width > 0 {
			track := ui.Div().Class("of-trace-track").Attr("aria-hidden", "true")
			if start > 0 {
				track.Children(ui.Div().Class("of-trace-gap").FlexGrow(ui.Number(start)))
			}
			track.Children(ui.Div().Class("of-trace-bar").FlexGrow(ui.Number(width)))
			if rest := tracePercent - start - width; rest > 0 {
				track.Children(ui.Div().Class("of-trace-gap").FlexGrow(ui.Number(rest)))
			}
			body.Children(track)
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
