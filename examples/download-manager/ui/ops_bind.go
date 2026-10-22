package ui

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe"
)

// bind reacquires every operation the paint uses. Call it after any Redraw
// or resize, both of which replace the display list.
func (a *App) bind() {
	d := a.page.Testing().Display()
	a.bound = a.page.Testing().Generation()
	a.bindings = bindings{}

	if d == nil {
		return
	}

	boxes := a.page.Boxes()
	byID := make(map[string]ownframe.Box, len(boxes))
	for _, box := range boxes {
		byID[box.ID] = box
	}

	for i := range a.view.Active {
		a.bindings.fill = append(a.bindings.fill, nil)
		a.bindings.trackW = append(a.bindings.trackW, 0)

		if track, ok := byID["track-"+strconv.Itoa(i)]; ok {
			a.bindings.fill[i] = ownframe.Fill(d, track, barColor)
			_, _, w, _ := ownframe.BoxUnits(d, track)
			a.bindings.trackW[i] = w
		}

		a.bindings.pct = append(a.bindings.pct, textAt(d, byID, "pct-"+strconv.Itoa(i)))
		a.bindings.speed = append(a.bindings.speed, textAt(d, byID, "spd-"+strconv.Itoa(i)))
		a.bindings.eta = append(a.bindings.eta, textAt(d, byID, "eta-"+strconv.Itoa(i)))
		w, h := a.page.Size()
		a.bindings.row = append(a.bindings.row, rowStrip(byID, i, w, h))
	}

	sum := [...]string{"sum-active", "sum-running", "sum-done", "sum-failed"}
	for i, id := range sum {
		a.bindings.sum[i] = textAt(d, byID, id)
	}

	detail := [...]string{"detail-state", "detail-size", "detail-speed"}
	for i, id := range detail {
		a.bindings.detail[i] = textAt(d, byID, id)
	}

	a.bindings.foot = textAt(d, byID, "foot-stats")
}

// textAt returns the first text operation inside the named box, or nil.
func textAt(d *ownframe.Display, byID map[string]ownframe.Box, id string) *ownframe.DisplayOp {
	box, ok := byID[id]
	if !ok {
		return nil
	}

	return ownframe.Text(d, box)
}
