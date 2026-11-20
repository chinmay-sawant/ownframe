package animationpreview

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

func (d *demo) tick(context.Context) error {
	if d.pending != "" || d.gen != d.page.Probe().Generation() {
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.page.Probe().Generation()
		d.pending = ""
	}
	active := false
	for _, item := range d.samples {
		if item.active && !item.player.Step(time.Now()) {
			item.active = false
		}
		active = active || item.active
	}
	if !active {
		d.page.SetTick(nil)
	}
	return nil
}

func (d *demo) rebind() error {
	for key, item := range d.samples {
		if !item.active && key != d.pending {
			continue
		}
		target, ok := byKey(key)
		if !ok {
			continue
		}
		boxes := d.page.Boxes()
		boxIndex, ok := findBoxIndex(boxes, target.id)
		if !ok {
			continue
		}
		box := boxes[boxIndex]
		player, err := cssanimation.Bind(d.page.Probe().Display(), box, item.indices, item.base, item.start)
		if err != nil {
			return err
		}
		item.player, item.active = player, true
		if key == d.pending && target.trigger != "" {
			source, found := findBoxIndex(boxes, target.button)
			if !found {
				continue
			}
			err := csstrigger.Dispatch(boxes, csstrigger.Event{Name: target.trigger, Source: source, Active: true}, map[int]*animation.Player{boxIndex: player}, item.start)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}
	return ownframe.Box{}, false
}

func findBoxIndex(boxes []ownframe.Box, id string) (int, bool) {
	for i, box := range boxes {
		if box.ID == id {
			return i, true
		}
	}
	return -1, false
}
