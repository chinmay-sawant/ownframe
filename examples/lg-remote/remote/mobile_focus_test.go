package remote

import (
	"context"
	"testing"
)

func TestPhoneKeyboardFocusActivatesButton(t *testing.T) {
	a := newTest(t, WithPhone(true))
	f := &fakeLink{}
	a.SetLink(f)
	ctx := context.Background()
	p := a.Page()
	if err := p.Focus(ctx, "volup"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if f.spec != "vol:Up" {
		t.Fatalf("activation %q", f.spec)
	}
	p.SetFormValue("host", "10.0.0.8")
	if err := p.Focus(ctx, "host"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if f.spec != "pair:" || f.host != "10.0.0.8" {
		t.Fatalf("pair %q %q", f.spec, f.host)
	}
}

func TestPhonePadCaptureExcludesOtherControls(t *testing.T) {
	a := newTest(t, WithPhone(true))
	f := &fakeLink{}
	a.SetLink(f)
	clickID(t, a, "tab-pad")
	ctx := context.Background()
	p := a.Page()
	for _, b := range p.Boxes() {
		if b.ID != "pad" && b.ID != "connect" {
			continue
		}
		claimed, err := p.Testing().BeginDrag(ctx, b.X+1, b.Y+1)
		if err != nil {
			t.Fatal(err)
		}
		if claimed != (b.ID == "pad") {
			t.Fatalf("claim %s=%v", b.ID, claimed)
		}
		if claimed {
			if err = p.Testing().MoveDrag(ctx, b.X+21, b.Y+11); err != nil {
				t.Fatal(err)
			}
			if f.spec != "move:20,10" {
				t.Fatalf("move %s", f.spec)
			}
			if err = p.Testing().EndDrag(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
}
