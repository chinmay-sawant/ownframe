package ui

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the action of the clicked box. The page draws after this.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch {
	case strings.HasPrefix(box.Action, "cell:"):
		return a.clickCell(ctx, box.Action)
	case strings.HasPrefix(box.Action, "col:"):
		if col, ok := actionIndex(box.Action); ok {
			a.selectRect(0, col, a.sheet().Rows-1, col)
		}

		return nil
	case strings.HasPrefix(box.Action, "row:"):
		if row, ok := actionIndex(box.Action); ok {
			a.selectRect(row, 0, row, a.sheet().Cols-1)
		}

		return nil
	case strings.HasPrefix(box.Action, "sheet:"):
		a.switchSheet(strings.TrimPrefix(box.Action, "sheet:"))

		return nil
	}

	switch box.Action {
	case "formula":
		s := a.selection()
		a.startEdit(s.ActiveR, s.ActiveC, a.editText())
	case "undo":
		a.history(true)
	case "redo":
		a.history(false)
	case "theme":
		a.toggleTheme()
	case "import":
		a.csv = csvState{open: true, replace: true}
		a.status = "choose a CSV file"
	case "export":
		a.csv = csvState{open: true, export: true, path: a.exportPath()}
		a.status = "export " + a.sheet().Name
	case "csv-commit":
		return a.commitCSV(ctx)
	case "csv-write":
		return a.writeCSV(ctx)
	case "csv-replace":
		a.csv.replace = !a.csv.replace
	case "csv-cancel":
		return a.closeCSV(ctx)
	}

	return nil
}
