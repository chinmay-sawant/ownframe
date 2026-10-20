package transitionpreview

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func (d *demo) tick(context.Context) error {
	if d.pending != "" || d.gen != d.page.Generation() {
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.page.Generation()
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
		box, ok := findBox(d.page.Boxes(), target.id)
		if !ok {
			continue
		}
		player, err := csstransition.Bind(d.page.Display(), box, target.property,
			item.indices, item.base, item.start)
		if err != nil {
			return err
		}
		item.player, item.active = player, true
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
