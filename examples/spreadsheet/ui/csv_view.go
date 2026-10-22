package ui

import (
	"fmt"
	"strings"
)

// buildDialog lays out the CSV import or export overlay.
func (a *App) buildDialog() *Dialog {
	if !a.csv.open {
		return nil
	}

	d := &Dialog{W: uiDialogW, H: uiDialogH}
	if a.csv.export {
		d.Title = "Export CSV"
		d.Prompt = "Write the active sheet to this path."
		d.Field = "csvsave"
		d.FieldType = "text"
		d.FieldVal = a.csv.path
		d.Buttons = []Box{
			{Action: "csv-write", Text: "Write", W: uiBtnW},
			{Action: "csv-cancel", Text: "Cancel", W: uiBtnW},
		}
	} else {
		mode := "Replace: off"
		if a.csv.replace {
			mode = "Replace: on"
		}

		d.Title = "Import CSV"
		d.Prompt = "Choose a .csv file, check the preview, then commit."
		d.Field = "csvpath"
		d.FieldType = "file"
		d.FieldVal = a.csv.path
		d.Buttons = []Box{
			{Action: "csv-replace", Text: mode, W: uiBtnWReplace},
			{Action: "csv-commit", Text: "Commit", W: uiBtnW},
			{Action: "csv-cancel", Text: "Cancel", W: uiBtnW},
		}
	}

	if a.csv.ready {
		for i, row := range a.csv.prev.Rows {
			if i == uiPreviewMax {
				d.Rows = append(d.Rows, "...")
				break
			}

			d.Rows = append(d.Rows, strings.Join(row, " | "))
		}

		d.Note = fmt.Sprintf("%d rows x %d cols", a.csv.prev.Total, a.csv.prev.Cols)
		d.Warnings = a.csv.prev.Warnings
	} else {
		d.Note = a.csv.note
	}

	d.X = a.scrollX + max(uiDlgMargin, (a.viewW-d.W)/uiCenterDiv)
	d.Y = a.scrollY + max(uiDlgMargin, (a.viewH-d.H)/uiCenterDiv)
	for i := range d.Buttons {
		d.Buttons[i].H = 26
		d.Buttons[i].X = d.W - (len(d.Buttons)-i)*uiBtnStep
		d.Buttons[i].Y = d.H - uiBtnBottom
		d.Buttons[i].Class = "btn"
	}

	return d
}
