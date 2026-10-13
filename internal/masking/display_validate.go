package masking

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

func ValidateDisplay(d *layout.Display) error {
	seen, boxes := map[*layout.DisplayGroup]bool{}, map[*layout.DisplayGroup]bool{}
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.IsGroupBegin() && op.Group() != nil && op.Group().Mask != nil {
			box := GroupBounds(op, d)
			if box.Dx() < 0 || box.Dy() < 0 ||
				box.Dx() > maxMaskPixels/max(1, box.Dy()) {
				return fmt.Errorf("mask: element box exceeds %d pixels", maxMaskPixels)
			}
			boxes[op.Group()] = true
		}
		for g := op.Group(); g != nil; g = g.Parent {
			if g.Mask != nil && !seen[g] {
				seen[g] = true
				if err := Validate(g.Mask); err != nil {
					return err
				}
				if err := validateSources(g.Mask); err != nil {
					return err
				}
			}
		}
	}
	for g := range seen {
		if !boxes[g] {
			return fmt.Errorf("mask: element border box is missing")
		}
	}
	return nil
}
