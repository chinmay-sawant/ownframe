package csstransition

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// twoElementFrame lays out a target element and a second element, so a
// caller can point the captured indices at another element's operation.
func twoElementFrame(t *testing.T) (*ownframe.Display, ownframe.Box, ownframe.Box) {
	t.Helper()

	html := `<style>#target{width:50px;height:20px;background:#f00;
transition-property:background-color;transition-duration:1000ms}</style>
<div id="target"></div><div id="other" style="width:30px;height:10px;background:#0f0"></div>`

	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	var target, other ownframe.Box
	for _, box := range p.Boxes() {
		switch box.ID {
		case "target":
			target = box
		case "other":
			other = box
		}
	}
	if target.ID == "" || other.ID == "" {
		t.Fatal("the display has no target or other box")
	}

	return p.Display(), target, other
}

// TestBindToRefusesOperationsOutsideTheBox checks that a rebind whose
// indices resolve to another element's operations is refused.
func TestBindToRefusesOperationsOutsideTheBox(t *testing.T) {
	display, target, other := twoElementFrame(t)
	indices := cssmotion.OperationIndices(display, other)
	from := cssmotion.Capture(display, other)
	to := cssmotion.Capture(display, other)
	if len(indices) == 0 {
		t.Fatal("the other element has no operation")
	}

	_, err := BindTo(display, target, "background-color", indices, from, to, time.Unix(7, 0))
	if err == nil {
		t.Fatal("a rebind onto another element's operations was accepted")
	}
}
