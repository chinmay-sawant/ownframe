package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

type countedLink struct {
	fakeLink
	calls int
}

func (f *countedLink) Exec(host, spec string) (string, error) {
	f.calls++
	return f.fakeLink.Exec(host, spec)
}

func remoteBox(t *testing.T, a *App, id string) ownframe.Box {
	t.Helper()
	for _, b := range a.page.Boxes() {
		if b.ID == id {
			return b
		}
	}
	t.Fatalf("missing %s", id)
	return ownframe.Box{}
}

func TestCommandsStartOnFingerDown(t *testing.T) {
	for _, id := range []string{"left", "right", "volup", "voldn"} {
		a := newTest(t, WithPhone(true))
		f := &countedLink{}
		a.SetLink(f)
		b := remoteBox(t, a, id)
		ctx := context.Background()
		if err := a.page.Press(ctx, b.X+1, b.Y+1); err != nil {
			t.Fatal(err)
		}
		if f.calls != 1 {
			t.Fatalf("%s waits for finger lift", id)
		}
		if err := a.page.Click(ctx, b.X+1, b.Y+1); err != nil {
			t.Fatal(err)
		}
		if err := a.page.Release(ctx); err != nil {
			t.Fatal(err)
		}
		if f.calls != 1 {
			t.Fatalf("%s sent twice", id)
		}
	}
}

func TestStatusDoesNotMoveControls(t *testing.T) {
	for _, phone := range []bool{false, true} {
		a := newTest(t, WithPhone(phone))
		a.page.SetSize(360, 640)
		a.view.Status = "Ready"
		a.setData()
		if err := a.page.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		y := remoteBox(t, a, "connect").Y
		a.view.Status = "Connecting. Accept the pairing prompt on the television to continue."
		a.setData()
		if err := a.page.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		if got := remoteBox(t, a, "connect").Y; got != y {
			t.Fatalf("phone %v: control shifted %.1f", phone, got-y)
		}
	}
}
