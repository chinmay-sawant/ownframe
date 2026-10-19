package page

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestCutChangeGetsBox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="a" type="text"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	var want Box
	for _, b := range p.Boxes() {
		if b.ID == "a" {
			want = b
		}
	}

	var got []Box
	p.Handle(Handlers{Change: func(_ context.Context, box Box) error {
		got = append(got, box)
		return nil
	}})

	text, ok, err := p.Cut(ctx)
	if err != nil || !ok || text != "ab" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Fatalf("change %v want %v", got, want)
	}
	if p.FormValue("a") != "" {
		t.Fatalf("value %q", p.FormValue("a"))
	}
}

func TestCutWritesField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := &bftView{Email: "ada"}
	p := bftPage(t, bftField)
	bftDraw(t, p, view)
	bftClick(t, p, "email")

	text, ok, err := p.Cut(ctx)
	if err != nil || !ok || text != "ada" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if view.Email != "" || p.FormValue("email") != "" {
		t.Fatalf("view %q form %q", view.Email, p.FormValue("email"))
	}
}

func TestCutErrorStopsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="a" type="text"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("boom")
	p.Handle(Handlers{Change: func(context.Context, Box) error { return boom }})
	gen := p.Generation()

	text, ok, err := p.Cut(ctx)
	if !errors.Is(err, boom) || !ok || text != "ab" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if p.FormValue("a") != "" {
		t.Fatalf("value %q", p.FormValue("a"))
	}
	if p.Generation() != gen {
		t.Fatalf("generation %d -> %d", gen, p.Generation())
	}
}
