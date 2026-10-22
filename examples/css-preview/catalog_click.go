package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func (d *demo) click(ctx context.Context, b ownframe.Box) error {
	d.dataChanged = false
	if b.ID == playAllID {
		return d.playAll(ctx)
	}
	for _, row := range d.v.Catalog {
		if row.Button != b.ID {
			continue
		}
		d.prevSelected = d.v.Selected
		d.v.Started[row.Button] = true
		d.v.Selected = row.Name
		readout := row.Name + " = " + row.Value + "\n" + row.Action
		if readout != d.readout {
			// The readout lives in the parsed tree, so the click patches it
			// instead of rebuilding the page. The template keeps the
			// initial snapshot, so the executed source does not change.
			d.readout = readout
			if !d.patchReadout(ctx, readout) {
				// The engine refused the readout, so the page has to
				// draw again and the template renders it.
				d.v.Readout = readout
				d.dataChanged = true
			}
		}
		if row.SampleKey == "" {
			d.startedAt = time.Now()
			d.startedSeq++
			d.p.SetData(d.v)
			return nil
		}
		selected, ok := targetByKey(row.SampleKey)
		if !ok {
			return nil
		}
		return d.start(selected, row.Button)
	}
	selected, ok := targetByButton(b.ID)
	if !ok {
		return nil
	}
	return d.start(selected, selected.button)
}

func (d *demo) start(selected target, button string) error {
	box, ok := findBox(d.p.Boxes(), selected.id)
	if !ok {
		return nil
	}
	if d.samples[selected.key] == nil {
		d.samples[selected.key] = &sample{base: capture(d.p.Testing().Display(), box)}
	}
	d.samples[selected.key].source = button
	d.samples[selected.key].origin = time.Now()
	d.v.Started[selected.key] = true
	d.v.Started[button] = true
	d.startedAt = time.Now()
	d.startedSeq++
	d.pending = append(d.pending, selected.key)
	d.p.SetData(d.v)
	if !walkActive() {
		d.p.SetTick(d.tick)
	}
	return nil
}
