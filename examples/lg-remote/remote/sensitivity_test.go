package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"
)

func TestSensitivityDragKeyboardAndAccessibility(t *testing.T) {
	a := newTest(t, WithPhone(true))
	clickID(t, a, "tab-pad")
	b := remoteBox(t, a, "sensitivity")
	ctx := context.Background()
	if claimed, err := a.page.Testing().BeginDrag(ctx, b.X, b.Y+1); err != nil || !claimed {
		t.Fatalf("capture %v %v", claimed, err)
	}
	if a.view.Sensitivity != 25 {
		t.Fatal(a.view.Sensitivity)
	}
	for n := 1; n <= 12; n++ {
		if err := a.page.Testing().MoveDrag(ctx, b.X+float64(n), b.Y+1); err != nil {
			t.Fatal(err)
		}
	}
	if a.view.Sensitivity <= 25 {
		t.Fatal("small slider deltas were lost")
	}
	if err := a.page.Testing().MoveDrag(ctx, b.X+b.W, b.Y+1); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.view.Sensitivity != 300 {
		t.Fatal(a.view.Sensitivity)
	}
	if err := a.page.Focus(ctx, "sensitivity"); err != nil {
		t.Fatal(err)
	}
	if err := a.page.KeyDown(ctx, "arrowleft"); err != nil {
		t.Fatal(err)
	}
	if a.view.Sensitivity != 295 {
		t.Fatal(a.view.Sensitivity)
	}
	bridge.QueueAction("sensitivity:150")
	if err := a.accessibilityActions(ctx); err != nil {
		t.Fatal(err)
	}
	if a.view.Sensitivity != 150 {
		t.Fatal(a.view.Sensitivity)
	}
}
