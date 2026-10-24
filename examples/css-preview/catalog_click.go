package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// click applies a Start or Play all click. The card's state is template data,
// so the class that starts a CSS animation reaches the layout through one
// SetData and Redraw. No driver call is made here.
func (d *demo) click(ctx context.Context, b ownframe.Box) error {
	if b.ID == playAllID {
		return d.playAll(ctx)
	}
	row, ok := d.rowForButton(b.ID)
	if !ok {
		return nil
	}
	d.selectRow(row)
	d.prepareManual(row)
	return d.commit(ctx)
}

// selectRow records one row's click state. A second Start on the same row
// flips the keyframe-name variant, which restarts the animation in CSS.
func (d *demo) selectRow(row catalogRow) {
	if d.v.Started[row.Button] {
		d.v.Restart[row.Button] = !d.v.Restart[row.Button]
		d.v.Restarting[row.Button] = true
	}
	d.v.Started[row.Button] = true
	d.v.Selected = row.Name
	d.readout = row.Name + " = " + row.Value + "\n" + row.Action
	d.v.Readout = d.readout
	d.startedAt = time.Now()
	d.startedSeq++
}

// commit renders the view and starts the previews the runtime does not drive.
// A restart first publishes with the target hidden, so the player the old
// keyframe name bound is released; the next publish brings the target back
// with the other keyframe name and a fresh clock.
func (d *demo) commit(ctx context.Context) error {
	if d.restarting() {
		d.v.applyRows()
		d.p.SetData(d.v)
		if err := d.p.Redraw(ctx); err != nil {
			return err
		}
		d.clearRestarting()
	}
	d.v.applyRows()
	d.p.SetData(d.v)
	if err := d.p.Redraw(ctx); err != nil {
		return err
	}
	if err := d.bindManual(d.v.Rows); err != nil {
		return err
	}
	d.wake()
	return nil
}

// restarting reports whether a publish has to hide a target first.
func (d *demo) restarting() bool {
	for _, on := range d.v.Restarting {
		if on {
			return true
		}
	}
	return false
}

// clearRestarting removes the hide class before the card is published again.
func (d *demo) clearRestarting() {
	for key := range d.v.Restarting {
		delete(d.v.Restarting, key)
	}
}
