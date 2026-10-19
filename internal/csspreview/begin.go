package csspreview

import "time"

func (d *demo) begin() error {
	t, ok := targetByKey(d.pendingKey)
	if !ok {
		return nil
	}
	s := d.samples[t.key]
	box, ok := findBox(d.p.Boxes(), t.id)
	if !ok {
		return nil
	}
	display := d.p.Display()
	s.origin = time.Now()
	plan, err := bind(display, box, t, s.indices, s.base, s.origin)
	if err != nil {
		return err
	}
	s.plan = plan
	s.active = s.plan.duration > 0 && len(s.plan.pairs) > 0
	return nil
}
