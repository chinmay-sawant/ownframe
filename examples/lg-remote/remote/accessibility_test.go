package remote

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"
)

func TestNativeAccessibilityTracksViewportAndActions(t *testing.T) {
	a := newTest(t, WithPhone(true))
	f := &fakeLink{}
	a.SetLink(f)
	a.Page().SetSize(360, 640)
	if err := a.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.publishAccessibility()
	var state struct{ Nodes []accessibleNode }
	if err := json.Unmarshal([]byte(bridge.Accessibility()), &state); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range state.Nodes {
		if n.ID == "volup" {
			found = n.Name == "Volume up" && n.Enabled
		}
	}
	if !found {
		t.Fatal("volume is missing or unnamed")
	}
	if !bridge.QueueAction("click:volup") {
		t.Fatal("queue rejected")
	}
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.spec != "vol:Up" {
		t.Fatal(f.spec)
	}
	bridge.QueueAction("text:10.0.0.8")
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.Page().Form().FormValue("host") != "10.0.0.8" {
		t.Fatal("host edit lost")
	}
	a.Page().Scroll().SetScrollOffset(0, 600)
	a.publishAccessibility()
	if err := json.Unmarshal([]byte(bridge.Accessibility()), &state); err != nil {
		t.Fatal(err)
	}
	for _, n := range state.Nodes {
		if n.X < 0 || n.Y < 0 || n.X+n.W > 360.1 || n.Y+n.H > 640.1 {
			t.Fatalf("fitted node %s is offscreen", n.ID)
		}
	}
	bridge.QueueAction("scroll:forward")
	if err := a.accessibilityActions(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Page().Testing().TakeScroll(); ok {
		t.Fatal("fitted accessibility action scrolled the page")
	}
}
