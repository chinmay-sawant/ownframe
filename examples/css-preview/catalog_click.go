package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func (d *demo) click(_ context.Context, b ownframe.Box) error {
	for _, row := range d.v.Catalog {
		if row.Button != b.ID {
			continue
		}
		d.v.Started[row.Button] = true
		d.v.Selected = row.Name
		d.v.Readout = row.Name + " = " + row.Value + "\n" + row.Action
		if row.SampleKey == "" {
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
		d.samples[selected.key] = &sample{
			base: capture(d.p.Display(), box), indices: cssmotion.OperationIndices(d.p.Display(), box),
		}
	}
	d.samples[selected.key].source = button
	d.v.Started[selected.key] = true
	d.v.Started[button] = true
	d.pendingKey = selected.key
	d.p.SetData(d.v)
	if !walkActive() {
		d.p.SetTick(d.tick)
	}
	return nil
}
