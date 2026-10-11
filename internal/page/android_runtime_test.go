package page_test

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestAndroidLayoutsResizeAndModalInput(t *testing.T) {
	source, err := os.ReadFile("testdata/android-runtime.html")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p, err := page.New(page.Config{HTML: string(source), Width: 360, Height: 640})
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ Email string }{}
	p.SetData(&data)
	clicked := ""
	p.Handle(page.Handlers{Click: func(_ context.Context, b page.Box) error { clicked = b.Action; return nil }})
	for _, size := range [][2]int{{360, 640}, {640, 360}, {412, 915}, {360, 640}} {
		p.SetSize(size[0], size[1])
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if p.Display() == nil {
			t.Fatal("basic layout fell back")
		}
		a, b := boxFor(t, p, "flex-a"), boxFor(t, p, "flex-b")
		if size[0] <= 400 && b.Y <= a.Y || size[0] > 400 && b.X <= a.X {
			t.Fatal("responsive flex direction")
		}
		if boxFor(t, p, "grid-b").X <= boxFor(t, p, "grid-a").X {
			t.Fatal("grid columns")
		}
		if boxFor(t, p, "settings").H < 900 {
			t.Fatal("long content truncated")
		}
		modal := boxFor(t, p, "modal")
		if err := p.Click(ctx, modal.X+modal.W/2, modal.Y+modal.H/2); err != nil {
			t.Fatal(err)
		}
		if clicked != "modal" {
			t.Fatal("overlay did not receive input")
		}
	}
	if err := p.Focus(ctx, "email"); err != nil {
		t.Fatal(err)
	}
	if err := p.Type(ctx, "mobile@example.test"); err != nil {
		t.Fatal(err)
	}
	if data.Email != "mobile@example.test" {
		t.Fatal("form binding")
	}
	p.SetSize(640, 360)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.FocusedField() != "email" || p.FormValue("email") != data.Email {
		t.Fatal("resize lost focus or value")
	}
}
