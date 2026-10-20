package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func bindCatalogTarget(t *testing.T, d *ownframe.Display, id string, start time.Time) (*animation.Player, int, []int, []cssmotion.Paint) {
	t.Helper()
	for index, box := range d.Boxes {
		if box.ID != id {
			continue
		}
		ops, base := cssmotion.OperationIndices(d, box), cssmotion.Capture(d, box)
		p, err := cssanimation.Bind(d, box, ops, base, start)
		if err != nil {
			t.Fatal(err)
		}
		return p, index, ops, base
	}
	t.Fatalf("box %q missing", id)
	return nil, -1, nil, nil
}

func catalogBoxIndex(boxes []ownframe.Box, id string) int {
	for index := range boxes {
		if boxes[index].ID == id {
			return index
		}
	}
	return -1
}
