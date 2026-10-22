package storage

import (
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// cellColumns maps a cell to its column values.
func cellColumns(c workbook.Cell) (kind int, number float64, text, source string) {
	switch c.Kind {
	case workbook.Number:
		return int(workbook.Number), c.Number, "", ""
	case workbook.Text:
		return int(workbook.Text), 0, c.Text, ""
	case workbook.Formula:
		return int(workbook.Formula), 0, "", c.Source
	}

	return 0, 0, "", ""
}

// columnCell maps stored columns back to a cell.
func columnCell(kind int, number float64, text, source string) (workbook.Cell, bool) {
	switch workbook.Kind(kind) { //nolint:gosec // reason: kind is a small workbook enum 0-3, always fits in uint8
	case workbook.Number:
		return workbook.Cell{Kind: workbook.Number, Number: number}, true
	case workbook.Text:
		return workbook.Cell{Kind: workbook.Text, Text: text}, true
	case workbook.Formula:
		return workbook.Cell{Kind: workbook.Formula, Source: source}, true
	}

	return workbook.Cell{}, false
}
