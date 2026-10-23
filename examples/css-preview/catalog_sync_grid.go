package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// stateEditsPerRow is the edits one row carries: the class and the text of
// its state span, and the class of its card.
const stateEditsPerRow = 3

// patchGrid puts the state of every started or selected row into the parsed
// tree in one patch. The engine lays the grid out once, so a page with many
// started rows does not pay a whole document layout per row.
func (d *demo) patchGrid(ctx context.Context) bool {
	edits := d.stateEdits()
	if len(edits) == 0 {
		return true
	}

	return d.p.PatchElement(ctx, catalogGridID, edits...) == nil
}

// stateEdits collects the class and the text every started or selected row
// carries in the parsed tree. Only the rows the tree holds are edited, so a
// patch does not carry the state of a row the window left out.
func (d *demo) stateEdits() []ownframe.ElementEdit {
	rows := d.v.Rows
	if len(rows) == 0 {
		rows = d.v.Catalog
	}
	edits := make([]ownframe.ElementEdit, 0, len(rows)*stateEditsPerRow)
	for _, row := range rows {
		if !d.v.Started[row.Button] && d.v.Selected != row.Name {
			continue
		}
		text := d.spanText(row)
		edits = append(edits,
			ownframe.ElementEdit{ID: row.StateID, Attribute: "class", Value: d.spanClass(row)},
			ownframe.ElementEdit{ID: row.StateID, Text: &text},
			ownframe.ElementEdit{ID: row.CardID, Attribute: "class", Value: d.cardClass(row)})
	}

	return edits
}
