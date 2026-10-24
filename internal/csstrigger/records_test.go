package csstrigger_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

func TestResolveUsesCommaAlignedAnimationRecords(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 100,
		Height: 80, Theme: `#target{width:20px;height:20px;animation-name:first,second;` +
			`animation-trigger:--open play, --close play reset;}`})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	boxes := page.Boxes()
	for boxIndex, box := range boxes {
		if box.ID != "target" {
			continue
		}
		if len(box.Animations) != 2 || box.Animations[0].Trigger != "--open play" ||
			box.Animations[1].Trigger != "--close play reset" {
			t.Fatalf("animation records = %#v", box.Animations)
		}
		got := csstrigger.Resolve(boxes, csstrigger.Event{Name: "--close", Source: boxIndex, Active: false})
		if len(got) != 1 || got[0].AnimationIndex != 1 || got[0].Value != "reset" {
			t.Fatalf("close actions = %#v, want reset on animation 1", got)
		}
		return
	}
	t.Fatal("Blinkless display has no target box")
}
