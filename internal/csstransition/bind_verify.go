package csstransition

import (
	"errors"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// transitionRecord is the timing a box published for a property.
type transitionRecord struct {
	DurationMS, DelayMS float64
	TimingFunction      string
	Behavior            string
}

// transitionRecordOf returns the box's transition record for the property.
// A record for "all" covers every property.
func transitionRecordOf(box ownframe.Box, property string) (transitionRecord, bool) {
	for i := range box.Transitions {
		item := box.Transitions[i]
		if item.Property == property || item.Property == "all" {
			return transitionRecord{DurationMS: item.DurationMS, DelayMS: item.DelayMS,
				TimingFunction: item.TimingFunction, Behavior: item.Behavior}, true
		}
	}

	return transitionRecord{}, false
}

// ownedByBox refuses when an operation left the box. The captured paints
// name the operations of the box, so an operation that left it means the
// indices resolve to other operations, and the rebind refuses instead of
// writing a captured paint onto them.
func ownedByBox(d *ownframe.Display, box ownframe.Box, ops []*ownframe.DisplayOp) error {
	for _, op := range ops {
		if !boxOwnsOp(d, box, op) {
			return errors.New("csstransition: captured operations no longer belong to the box")
		}
	}

	return nil
}

// boxOwnsOp reports whether the box owns the operation. A box with
// ownership IDs names them; an older box falls back to the geometric test
// the capture used.
func boxOwnsOp(d *ownframe.Display, box ownframe.Box, op *ownframe.DisplayOp) bool {
	if op == nil {
		return false
	}

	if box.OperationIDs == nil {
		return cssmotion.Contains(d, box, *op)
	}

	for _, id := range box.OperationIDs {
		if id == op.ID {
			return true
		}
	}

	return false
}
