package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// WindowedListOptions describes a windowed list that mounts only the visible
// rows plus overscan. ID is the list's stable host ID; each mounted row takes
// ID+"-row-<index>" so ui.View.Update can address a changed row by ID. Total
// is the full row count and RowHeight is one row's CSS-pixel height. The host
// keeps ScrollOffset (content pixels from the top) and ViewportHeight, and
// rebuilds on scroll rest. Overscan pads the window above and below. RenderRow
// builds the row for an index in [0, Total); a nil row is skipped.
type WindowedListOptions struct {
	ID             string
	Total          int
	RowHeight      float64
	ViewportHeight float64
	ScrollOffset   float64
	Overscan       int
	RenderRow      func(index int) *ui.Element
}

// WindowedRowID derives a row's stable ID from the list host ID. It returns an
// empty string when the host did not supply an ID, so callers never invent one.
func WindowedRowID(id string, index int) string {
	return partID(id, "-row-"+strconv.Itoa(index))
}

// WindowedRange returns the half-open [start, end) row window covering the
// viewport at scrollOffset plus overscan rows on each side. A non-positive
// row height or viewport height returns the full range. Results are clamped
// to [0, total].
func WindowedRange(total int, rowHeight, viewportHeight, scrollOffset float64, overscan int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	if rowHeight <= 0 || viewportHeight <= 0 {
		return 0, total
	}
	if overscan < 0 {
		overscan = 0
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	first := int(scrollOffset / rowHeight)
	visible := int((viewportHeight + rowHeight - 1) / rowHeight)
	if visible < 1 {
		visible = 1
	}
	start = first - overscan
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end = first + visible + overscan
	if end < start {
		end = start
	}
	if end > total {
		end = total
	}
	return start, end
}

// WindowedList builds the visible row window with symmetric spacer blocks
// before and after it, so the list keeps its full scroll height while mounting
// only end-start rows. Rows keep index-derived IDs across scrolls, so a value
// edit inside a stable window takes the ui.View.Update patch path; a scroll
// that shifts the window changes the structure and takes the full render by
// design. A nil RenderRow mounts spacers only.
func WindowedList(options WindowedListOptions) *ui.Element {
	total := options.Total
	if total < 0 {
		total = 0
	}
	start, end := WindowedRange(total, options.RowHeight, options.ViewportHeight, options.ScrollOffset, options.Overscan)
	root := ui.Div().Class("of-windowed-list")
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.RowHeight > 0 {
		if before := float64(start) * options.RowHeight; before > 0 {
			spacer := ui.Div().Class("of-windowed-list-spacer")
			if id := partID(options.ID, "-spacer-before"); id != "" {
				spacer.ID(id)
			}
			root.Children(spacer.Height(ui.Px(before)))
		}
	}
	if options.RenderRow != nil {
		for index := start; index < end; index++ {
			row := options.RenderRow(index)
			if row == nil {
				continue
			}
			if id := WindowedRowID(options.ID, index); id != "" {
				row.ID(id)
			}
			root.Children(row)
		}
	}
	if options.RowHeight > 0 {
		if after := float64(total-end) * options.RowHeight; after > 0 {
			spacer := ui.Div().Class("of-windowed-list-spacer")
			if id := partID(options.ID, "-spacer-after"); id != "" {
				spacer.ID(id)
			}
			root.Children(spacer.Height(ui.Px(after)))
		}
	}
	return root
}
