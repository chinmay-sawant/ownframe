package csstrigger

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

func TestAnimationTriggerActionsRespectTriggerScope(t *testing.T) {
	boxes := []ownframe.Box{
		{ID: "scope", ParentIndex: -1, TriggerScope: "--ready"},
		{ID: "source", ParentIndex: 0},
		{ID: "inside", ParentIndex: 0, Animations: []layout.Animation{{Trigger: "--ready play pause"}}},
		{ID: "outside", ParentIndex: -1, Animations: []layout.Animation{{Trigger: "--ready play"}}},
	}
	got := Resolve(boxes, Event{Name: "--ready", Source: 1, Active: true})
	if len(got) != 1 || got[0].BoxIndex != 2 || got[0].Value != "play" {
		t.Fatalf("enter actions = %#v, want scoped play on box 2", got)
	}
	got = Resolve(boxes, Event{Name: "--ready", Source: 1})
	if len(got) != 1 || got[0].BoxIndex != 2 || got[0].Value != "pause" {
		t.Fatalf("exit actions = %#v, want scoped pause on box 2", got)
	}
}
