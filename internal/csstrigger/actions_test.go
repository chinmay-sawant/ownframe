package csstrigger

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

func TestResolveUsesLastMatchingTriggerRecord(t *testing.T) {
	boxes := []ownframe.Box{{ParentIndex: -1, Animations: []layout.Animation{{
		Trigger: "--ready play --other pause --ready reset",
	}}}}
	got := Resolve(boxes, Event{Name: "--ready", Source: 0, Active: true})
	if len(got) != 1 || got[0].Value != "reset" {
		t.Fatalf("entry actions = %#v, want last matching reset", got)
	}
	got = Resolve(boxes, Event{Name: "--ready", Source: 0})
	if len(got) != 0 {
		t.Fatalf("exit actions = %#v, want no action", got)
	}
}

func TestResolveIgnoresInvalidEventActionCount(t *testing.T) {
	boxes := []ownframe.Box{{ParentIndex: -1, Animations: []layout.Animation{{
		Trigger: "--ready play pause reset",
	}}}}
	if got := Resolve(boxes, Event{Name: "--ready", Source: 0, Active: true}); len(got) != 0 {
		t.Fatalf("actions = %#v, want no action for three event actions", got)
	}
}
