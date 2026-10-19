package csspreview

import "time"

func (d *demo) begin() {
	t, ok := targetByKey(d.kind, d.pendingKey)
	if !ok {
		return
	}
	s := d.samples[t.key]
	box, ok := findBox(d.p.Boxes(), t.id)
	if !ok {
		return
	}
	display := d.p.Display()
	s.origin = time.Now()
	s.plan = bind(display, box, t, s.base, s.origin)
	s.active = s.plan.duration > 0 && len(s.plan.pairs) > 0
}
