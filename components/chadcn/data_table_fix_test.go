package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestDataTableFixCaptionHeaderCellClasses renders a captioned, selectable
// table and pins the classes that carry the caption, header, and cell spacing.
// Only the column that opts into wrapping gets the wrap modifier.
func TestDataTableFixCaptionHeaderCellClasses(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		ID:      "logs",
		Caption: "Request log entries",
		Columns: []chadcn.DataColumn{
			{Label: "Time", Width: ui.Px(104)},
			{Label: "Message", Wrap: true},
			{Label: "Status", Align: "right", Width: ui.Px(76)},
		},
		Rows:     [][]string{{"14:32:08.412", "upstream timeout after 30s", "504"}},
		OnSelect: func(_ context.Context, _ int) error { return nil },
	}))
	compositeWant(t, markup,
		`<caption class="of-data-table-caption">Request log entries</caption>`,
		`class="of-data-table-header"`,
		`class="of-data-table-header of-data-table--wrap"`,
		`class="of-data-table-cell"`,
		`class="of-data-table-cell of-data-table--wrap"`,
		`class="of-data-table-cell of-data-table--align-right"`,
		`of-data-table-select-cell`,
		`role="checkbox"`,
		`aria-label="Select row 1"`,
	)
	if got := strings.Count(markup, "of-data-table--wrap"); got != 2 {
		t.Fatalf("wrap modifier count=%d want 2 in %s", got, markup)
	}
}

// TestDataTableFixBodyRowIDs pins the hover id on each body row. The id is
// the table id plus the row index. The select control keeps its own id, and
// the header row stays a plain heading so sort stays with the host.
func TestDataTableFixBodyRowIDs(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		ID:      "logs",
		Columns: []chadcn.DataColumn{{Label: "Time"}, {Label: "Status"}},
		Rows: [][]string{
			{"14:32", "504"},
			{"14:33", "200"},
		},
		Selected: map[int]bool{1: true},
		OnSelect: func(_ context.Context, _ int) error { return nil },
	}))
	compositeWant(t, markup,
		`id="logs-row-0"`,
		`id="logs-row-1"`,
		`id="logs-select-0"`,
		`id="logs-select-1"`,
		`class="of-data-table-row of-data-table-row--selected"`,
	)
	row := compositeTag(t, markup, `id="logs-row-0"`)
	if !strings.Contains(row, "<tr") || strings.Contains(row, "of-data-table-select") {
		t.Fatalf("body row id is not on the row: %s", row)
	}
	control := compositeTag(t, markup, `id="logs-select-0"`)
	if !strings.Contains(control, "<button") {
		t.Fatalf("select id left the control: %s", control)
	}
	head := compositeTag(t, markup, "of-data-table-row--head")
	if strings.Contains(head, "id=") {
		t.Fatalf("header row took a hover id: %s", head)
	}
	headStart := strings.Index(markup, "<thead")
	headEnd := strings.Index(markup, "</thead>")
	if headStart < 0 || headEnd < headStart || strings.Contains(markup[headStart:headEnd], "<button") {
		t.Fatalf("header gained a sort button:\n%s", markup)
	}

	empty := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		ID:      "logs",
		Columns: []chadcn.DataColumn{{Label: "Time"}},
		Empty:   "No results.",
	}))
	compositeWant(t, empty, `id="logs-row-0"`, "No results.")
	compositeDeny(t, empty, `id="logs-row-1"`, "<button")

	plain := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		Columns: []chadcn.DataColumn{{Label: "Time"}},
		Rows:    [][]string{{"14:32"}},
	}))
	if strings.Contains(plain, `id=`) {
		t.Fatalf("table without an id invented a row id: %s", plain)
	}
}

// TestDataTableFixSpacingAndWrapRules holds the caption rule, the shared
// header and cell rhythm, and the opt-in wrap rule in the stylesheet.
func TestDataTableFixSpacingAndWrapRules(t *testing.T) {
	t.Parallel()
	caption := dataTableFixRule(t, chadcn.Stylesheet, ".of-data-table-caption")
	if !strings.Contains(caption, "border-bottom: 1px solid var(--of-border, #e5e5e5);") {
		t.Fatalf("caption rule lost its rule under the caption:\n%s", caption)
	}
	// The rendered caption is padded twice, so the declared value is half of
	// the 10px/16px rhythm the header and cells use.
	if !strings.Contains(caption, "padding: 5px 8px;") {
		t.Fatalf("caption padding is not the compensated rhythm:\n%s", caption)
	}
	for _, selector := range []string{".of-data-table-header", ".of-data-table-cell"} {
		rule := dataTableFixRule(t, chadcn.Stylesheet, selector)
		if !strings.Contains(rule, "padding: 10px 16px;") {
			t.Fatalf("%s does not share the table rhythm:\n%s", selector, rule)
		}
	}
	head := dataTableFixRule(t, chadcn.Stylesheet, ".of-data-table-head")
	if !strings.Contains(head, "background-color: var(--of-control, #f4f4f5);") {
		t.Fatalf("head lost its tinted band:\n%s", head)
	}
	cell := dataTableFixRule(t, chadcn.Stylesheet, ".of-data-table-cell")
	if !strings.Contains(cell, "white-space: nowrap;") {
		t.Fatalf("cells no longer stay on one line by default:\n%s", cell)
	}
	wrap := dataTableFixRule(t, chadcn.Stylesheet, ".of-data-table--wrap")
	for _, want := range []string{"white-space: normal;", "overflow-wrap: anywhere;"} {
		if !strings.Contains(wrap, want) {
			t.Fatalf("wrap rule is missing %q:\n%s", want, wrap)
		}
	}
}

// dataTableFixRule returns one rule body from the stylesheet, without its
// braces.
func dataTableFixRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet is missing %s", selector)
	}
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %s is unterminated", selector)
	}
	return sheet[start : start+end]
}
