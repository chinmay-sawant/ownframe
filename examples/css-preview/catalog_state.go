package main

// applyRows copies the view's click state onto every row. The windowed Rows
// slice shares the Catalog backing array, so one pass covers both lists.
//
// A restart flips the card's animation name between the two identical
// keyframe blocks the stylesheet declares. The commit hides the target for one
// publish, which releases the old player, so the new name binds a fresh clock
// instead of resuming the finished one. That is CSS, and it needs no driver
// call.
func (v *view) applyRows() {
	for i := range v.Catalog {
		row := &v.Catalog[i]
		row.Started = v.Started[row.Button]
		row.Selected = v.Selected == row.Name
		class := row.BaseClass
		if row.Started {
			class += " is-started"
			if v.Restart[row.Button] {
				class += " restart-b"
			}
		}
		if v.Restarting[row.Button] {
			class += " restarting"
		}
		if row.Selected {
			class += " is-selected"
		}
		row.Class = class
	}
}
