package main

import "context"

// capturePending takes the starting paint of every pending sample whose row
// is on screen and has never played. The row rendered without the started
// class, so this paint is the value the transition starts from. The row then
// renders with the class and binds.
func (d *demo) capturePending(ctx context.Context) error {
	boxes := d.p.Boxes()
	display := d.p.Probe().Display()
	changed := false
	for _, key := range d.pending {
		s := d.samples[key]
		if s == nil || s.base != nil {
			continue
		}
		t, ok := targetByKey(key)
		if !ok {
			continue
		}
		box, ok := findBox(boxes, t.id)
		if !ok {
			continue
		}
		s.base = capture(display, box)
		s.to = nil
		d.v.Started[t.key] = true
		d.v.Started[t.button] = true
		changed = true
	}
	if !changed {
		return nil
	}
	d.p.SetData(d.v)

	if err := d.p.Redraw(ctx); err != nil {
		return err
	}
	d.treeDirty = false

	return nil
}
