package csstransition

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestVisibilityTransitionHasNoMatchedPaintOperation(t *testing.T) {
	visible := visibilityFrame(t, "visible")
	hidden := visibilityFrame(t, "hidden")
	visibleBox := targetBox(t, visible)
	box := targetBox(t, hidden)
	if len(box.Transitions) != 1 || box.Transitions[0].Behavior != "allow-discrete" {
		t.Fatalf("Blinkless visibility metadata = %#v", box.Transitions)
	}
	from := cssmotion.Capture(visible.Probe().Display(), visibleBox)
	indices := cssmotion.OperationIndices(hidden.Probe().Display(), box)
	if len(from) == 0 || len(indices) != 0 {
		t.Fatalf("visible/hidden paint counts = %d/%d", len(from), len(indices))
	}
	if _, err := Bind(hidden.Probe().Display(), box, "visibility", indices, from, time.Time{}); err == nil {
		t.Fatal("visibility transition bound without a matched paint operation")
	}
}

func visibilityFrame(t *testing.T, visibility string) *ownframe.Page {
	t.Helper()
	html := `<style>#target{width:40px;height:20px;visibility:` + visibility + `;
transition-property:visibility;transition-duration:1s;transition-behavior:allow-discrete}</style><div id="target">visible text</div>`
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 100, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

func targetBox(t *testing.T, p *ownframe.Page) ownframe.Box {
	t.Helper()
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return ownframe.Box{}
}
