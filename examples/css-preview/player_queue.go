package main

import (
	"time"
)

// queue adds a key to the pending queue once. The queue holds one entry per
// row, so a second Play all press does not grow it.
func (d *demo) queue(key string) {
	for _, queued := range d.pending {
		if queued == key {
			return
		}
	}
	d.pending = append(d.pending, key)
}

// queueRows queues the samples of the rows that are in the window and have
// not started yet. A row that scrolls into view starts its preview.
func (d *demo) queueRows(rows []catalogRow) {
	for _, row := range rows {
		if row.SampleKey == "" || !d.v.Started[row.Button] {
			continue
		}
		s := d.samples[row.SampleKey]
		if s != nil && !s.origin.IsZero() {
			continue
		}
		d.queue(row.SampleKey)
	}
}

// captureBases takes the starting paint of every queued sample whose row
// has just rendered. The paint is read before the state patch, so it holds
// the value the transition starts from.
func (d *demo) captureBases() {
	boxes := d.p.Boxes()
	display := d.p.Probe().Display()
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
	}
}

// drain starts every queued sample that can start. A row that left the
// window is dropped: the scroll window queues it again when it comes back.
func (d *demo) drain() error {
	pending := d.pending
	d.pending = nil
	boxes := d.p.Boxes()
	origin := time.Now()
	if d.allPlaying && d.playAllAt.IsZero() {
		d.playAllAt = origin
	}
	for _, key := range pending {
		started, err := d.begin(key, boxes, origin)
		if err != nil {
			return err
		}
		if started {
			continue
		}
		s := d.samples[key]
		if s != nil && s.detached {
			continue
		}
		d.queue(key)
	}

	return d.rebind()
}
