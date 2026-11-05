package chadcn

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// DataColumn describes one table column. Align accepts left, center, or right
// and defaults to left. Width sizes the header and its cells when set. Wrap lets
// a long cell break onto more lines; cells stay on one line by default.
type DataColumn struct {
	Label string
	Align string
	Width ui.Length
	Wrap  bool
}

// DataTableOptions describes a themed, read-only table. Sorting, filtering, and
// paging stay in the host. OnSelect adds a leading selection column and reports
// the row index; Selected marks rows and may be nil. Empty replaces the default
// "No results." row, and Caption names the table for assistive technology.
// When ID is set, each body row id is that value plus the row index. The
// select control keeps its own id.
type DataTableOptions struct {
	ID       string
	Columns  []DataColumn
	Rows     [][]string
	Empty    string
	Caption  string
	Selected map[int]bool
	OnSelect func(context.Context, int) error
}

// DataTable builds a bordered table with a header, rows, and an empty state.
// The host owns sorting, filtering, paging, and the selected set.
func DataTable(options DataTableOptions) *ui.Element {
	selectable := options.OnSelect != nil
	table := ui.Tag("table").Class("of-data-table-table")
	if options.Caption != "" {
		table.Children(ui.Tag("caption").Class("of-data-table-caption").Text(options.Caption))
	}
	table.Children(dataTableHead(options, selectable), dataTableBody(options, selectable))
	root := ui.Div().Class(classes("of-scope", "of-data-table"))
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root.Children(table)
}

func dataTableHead(options DataTableOptions, selectable bool) *ui.Element {
	cells := make([]*ui.Element, 0, len(options.Columns)+1)
	if selectable {
		cells = append(cells, ui.Tag("th").Class(classes("of-data-table-header", "of-data-table-select-cell")).
			Attr("scope", "col").Attr("aria-hidden", "true"))
	}
	for _, column := range options.Columns {
		cellClass := classes("of-data-table-header", dataTableAlignClass(column.Align), dataTableWrapClass(column.Wrap))
		cell := ui.Tag("th").Class(cellClass).Attr("scope", "col").Text(column.Label)
		if column.Width != (ui.Length{}) {
			cell.Width(column.Width)
		}
		cells = append(cells, cell)
	}
	return ui.Tag("thead").Class("of-data-table-head").Children(
		ui.Tag("tr").Class(classes("of-data-table-row", "of-data-table-row--head")).Children(cells...),
	)
}

func dataTableBody(options DataTableOptions, selectable bool) *ui.Element {
	body := ui.Tag("tbody").Class("of-data-table-body")
	for index, row := range options.Rows {
		selected := options.Selected[index]
		cells := make([]*ui.Element, 0, len(options.Columns)+1)
		if selectable {
			cells = append(cells, ui.Tag("td").Class(classes("of-data-table-cell", "of-data-table-select-cell")).
				Children(dataTableSelect(options, index, selected)))
		}
		for columnIndex, column := range options.Columns {
			value := ""
			if columnIndex < len(row) {
				value = row[columnIndex]
			}
			cellClass := classes("of-data-table-cell", dataTableAlignClass(column.Align), dataTableWrapClass(column.Wrap))
			cells = append(cells, ui.Tag("td").Class(cellClass).Text(value))
		}
		body.Children(dataTableBodyRow(options.ID, index, dataTableSelectedClass(selected), cells...))
	}
	if len(options.Rows) == 0 {
		empty := options.Empty
		if empty == "" {
			empty = "No results."
		}
		span := len(options.Columns)
		if selectable {
			span++
		}
		body.Children(dataTableBodyRow(options.ID, 0, "",
			ui.Tag("td").Class(classes("of-data-table-cell", "of-data-table-empty")).
				Attr("colspan", strconv.Itoa(max(span, 1))).Text(empty),
		))
	}
	return body
}

// dataTableBodyRow builds one body row. Its id is the table id plus the row
// index, so a hover on the row can match. The header is not built here, and
// the select control keeps the id assigned in dataTableSelect.
func dataTableBodyRow(tableID string, index int, selectedClass string, cells ...*ui.Element) *ui.Element {
	row := ui.Tag("tr").Class(classes("of-data-table-row", selectedClass))
	if id := partID(tableID, "-row-"+strconv.Itoa(index)); id != "" {
		row.ID(id)
	}
	return row.Children(cells...)
}

func dataTableSelect(options DataTableOptions, row int, selected bool) *ui.Element {
	button := ui.Button("").Attr("type", "button").
		Class(classes("of-data-table-select", dataTableCheckedClass(selected))).
		Attr("role", "checkbox").Attr("aria-checked", ariaBool(selected)).
		Attr("aria-label", "Select row "+strconv.Itoa(row+1)).
		Children(ui.Span().Class("of-data-table-check").Attr("aria-hidden", "true").Text("✓"))
	if id := partID(options.ID, "-select-"+strconv.Itoa(row)); id != "" {
		button.ID(id)
	}
	if options.OnSelect != nil {
		index := row
		button.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, index) })
	}
	return button
}

func dataTableAlignClass(align string) string {
	switch strings.ToLower(strings.TrimSpace(align)) {
	case "center":
		return "of-data-table--align-center"
	case "right", "end":
		return "of-data-table--align-right"
	default:
		return ""
	}
}

func dataTableWrapClass(wrap bool) string {
	if wrap {
		return "of-data-table--wrap"
	}
	return ""
}

func dataTableSelectedClass(selected bool) string {
	if selected {
		return "of-data-table-row--selected"
	}
	return ""
}

func dataTableCheckedClass(checked bool) string {
	if checked {
		return "of-data-table-select--checked"
	}
	return ""
}
