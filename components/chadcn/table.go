package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// Table is the root of a simple data table.
func Table(children ...*ui.Element) *ui.Element {
	return ui.Tag("table").Class("of-table").Children(children...)
}

// TableHeader groups the column heading row.
func TableHeader(children ...*ui.Element) *ui.Element {
	return ui.Tag("thead").Class("of-table-header").Children(children...)
}

// TableBody groups the data rows.
func TableBody(children ...*ui.Element) *ui.Element {
	return ui.Tag("tbody").Class("of-table-body").Children(children...)
}

// TableFooter groups summary rows below the body.
func TableFooter(children ...*ui.Element) *ui.Element {
	return ui.Tag("tfoot").Class("of-table-footer").Children(children...)
}

// TableRow is one table row.
func TableRow(children ...*ui.Element) *ui.Element {
	return ui.Tag("tr").Class("of-table-row").Children(children...)
}

// TableHead is a column heading cell. Extra children follow the text.
func TableHead(text string, children ...*ui.Element) *ui.Element {
	return ui.Tag("th").Class("of-table-head").Attr("scope", "col").
		Text(text).Children(children...)
}

// TableCell is a data cell. Extra children follow the text.
func TableCell(text string, children ...*ui.Element) *ui.Element {
	return ui.Tag("td").Class("of-table-cell").Text(text).Children(children...)
}

// TableCaption is the table summary below the rows.
func TableCaption(text string) *ui.Element {
	return ui.Tag("caption").Class("of-table-caption").Text(text)
}
