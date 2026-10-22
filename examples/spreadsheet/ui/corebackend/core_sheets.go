package corebackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/storage"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// pick walks the workbook pages until it finds the name, or the first
// workbook when name is empty.
func (b *Backend) pick(ctx context.Context, name string) (storage.WorkbookInfo, error) {
	var cursor workbook.ID

	for {
		page, err := b.st.ListWorkbooks(ctx, cursor, workbook.PageSize)
		if err != nil {
			return storage.WorkbookInfo{}, err
		}

		for _, it := range page.Items {
			if name == "" || it.Name == name {
				return it, nil
			}
		}

		if !page.More {
			break
		}

		cursor = page.Next
	}

	return storage.WorkbookInfo{}, fmt.Errorf("corebackend: workbook %q not found", name)
}

// sheet looks up a sheet by the string id the UI uses.
func (b *Backend) sheet(id string) (*workbook.Sheet, bool) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, false
	}

	sh := b.wb.Sheet(workbook.SheetID(n))

	return sh, sh != nil
}

// sheetInfo maps one core sheet to the UI's bounded sheet shape.
func sheetInfo(sh *workbook.Sheet) ui.Sheet {
	return ui.Sheet{
		ID:   strconv.FormatInt(int64(sh.ID()), 10),
		Name: sh.Name(),
		Rows: workbook.MaxRows,
		Cols: workbook.MaxCols,
	}
}

// Sheets returns every sheet in tab order.
func (b *Backend) Sheets() []ui.Sheet {
	sheets := b.wb.Sheets()
	out := make([]ui.Sheet, 0, len(sheets))
	for _, sh := range sheets {
		out = append(out, sheetInfo(sh))
	}

	return out
}

// Info returns one sheet's identity and bounds.
func (b *Backend) Info(id string) (ui.Sheet, bool) {
	sh, ok := b.sheet(id)
	if !ok {
		return ui.Sheet{}, false
	}

	return sheetInfo(sh), true
}

// Revision is the workbook content revision.
func (b *Backend) Revision() uint64 { return uint64(b.wb.Rev()) } //nolint:gosec // reason: non-negative revision counter
