package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// PaginationOptions describes a controlled page selector. Page is the current
// one-based page and TotalPages bounds it; OnSelect receives the requested
// page. A TotalPages below one renders nothing.
type PaginationOptions struct {
	OnSelect   func(ctx context.Context, page int) error
	Page       int
	TotalPages int
}

// paginationSlot is either a page number or an ellipsis between page runs.
type paginationSlot struct {
	number int
	gap    bool
}

// Pagination builds previous, next, and numbered page controls. The current
// page carries aria-current, and the previous and next buttons are disabled at
// the bounds. Long ranges collapse to the first page, a window around the
// current page, and the last page.
func Pagination(options PaginationOptions) *ui.Element {
	if options.TotalPages < 1 {
		return nil
	}
	page := options.Page
	if page < 1 {
		page = 1
	}
	if page > options.TotalPages {
		page = options.TotalPages
	}
	previous := ui.Button("").Attr("type", "button").Class("of-pagination-previous").
		Attr("aria-label", "Go to previous page").
		Children(ui.Span().Class("of-pagination-chevron").Attr("aria-hidden", "true").Text("‹"), ui.Text("Previous"))
	setDisabled(previous, page <= 1)
	if options.OnSelect != nil && page > 1 {
		target := page - 1
		previous.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, target) })
	}
	next := ui.Button("").Attr("type", "button").Class("of-pagination-next").
		Attr("aria-label", "Go to next page").
		Children(ui.Text("Next"), ui.Span().Class("of-pagination-chevron").Attr("aria-hidden", "true").Text("›"))
	setDisabled(next, page >= options.TotalPages)
	if options.OnSelect != nil && page < options.TotalPages {
		target := page + 1
		next.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, target) })
	}

	list := ui.Ul().Class("of-pagination-list").
		Children(ui.Li().Class("of-pagination-item").Children(previous))
	for _, slot := range paginationSlots(page, options.TotalPages) {
		if slot.gap {
			list.Children(ui.Li().Class("of-pagination-item").Children(
				ui.Span().Class("of-pagination-ellipsis").Attr("aria-hidden", "true").Text("···")))
			continue
		}
		number := slot.number
		link := ui.Button(strconv.Itoa(number)).Attr("type", "button").Class("of-pagination-link").
			Attr("aria-label", "Page "+strconv.Itoa(number))
		if number == page {
			link.Attr("aria-current", "page")
		}
		if options.OnSelect != nil {
			link.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, number) })
		}
		list.Children(ui.Li().Class("of-pagination-item").Children(link))
	}
	list.Children(ui.Li().Class("of-pagination-item").Children(next))
	return ui.Nav().Class("of-pagination").Attr("aria-label", "Pagination").Children(list)
}

// paginationSlots lists every page up to the full-range limit and otherwise
// keeps the first page, a one-page window around the current page, and the
// last page with ellipses between them.
func paginationSlots(page, total int) []paginationSlot {
	const fullRange = 7
	if total <= fullRange {
		slots := make([]paginationSlot, 0, total)
		for number := 1; number <= total; number++ {
			slots = append(slots, paginationSlot{number: number})
		}
		return slots
	}
	start, end := max(1, page-1), min(total, page+1)
	slots := make([]paginationSlot, 0, fullRange)
	if start > 1 {
		slots = append(slots, paginationSlot{number: 1})
		if start-1 > 1 {
			slots = append(slots, paginationSlot{gap: true})
		}
	}
	for number := start; number <= end; number++ {
		slots = append(slots, paginationSlot{number: number})
	}
	if end < total {
		if total-end > 1 {
			slots = append(slots, paginationSlot{gap: true})
		}
		slots = append(slots, paginationSlot{number: total})
	}
	return slots
}
