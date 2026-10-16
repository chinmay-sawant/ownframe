package page

import (
	"context"
	"testing"
)

func TestFocusSetsHostState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	if err := p.FocusNext(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.renderState().Focus; got != "a" {
		t.Fatalf("render focus %q", got)
	}
}

func TestEscapeClearsFocus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	if err := p.Focus(ctx, "a"); err != nil {
		t.Fatal(err)
	}

	gen := p.Generation()
	if err := p.KeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "" || p.Generation() <= gen {
		t.Fatalf("focus %q generation %d", p.FocusID(), p.Generation())
	}
}

func TestCaretKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd ef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.setCaret(2, false)

	var seen []string
	p.Handle(Handlers{KeyDown: func(_ context.Context, key string) error {
		seen = append(seen, key)
		return nil
	}})

	steps := []struct {
		key  string
		want int
	}{
		{"arrowleft", 1},
		{"arrowright", 2},
		{"end", 7},
		{"home", 0},
		{"ctrl+arrowright", 5},
		{"shift+arrowright", 6},
	}

	for _, step := range steps {
		if err := p.KeyDown(ctx, step.key); err != nil {
			t.Fatal(err)
		}

		if p.formState().caret != step.want {
			t.Fatalf("%s caret %d want %d", step.key, p.formState().caret, step.want)
		}
	}

	if p.formState().anchor != 5 {
		t.Fatalf("shift anchor %d", p.formState().anchor)
	}

	if len(seen) != len(steps) || seen[0] != "arrowleft" {
		t.Fatalf("handler saw %q", seen)
	}
}
