package page

import (
	"context"
	"testing"
)

func TestCaretEditsAtOffset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.setCaret(2, false)

	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "abXcd" || p.formState().caret != 3 {
		t.Fatalf("insert value %q caret %d", p.FormValue("t"), p.formState().caret)
	}

	if err := p.Backspace(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "abcd" || p.formState().caret != 2 {
		t.Fatalf("backspace value %q caret %d", p.FormValue("t"), p.formState().caret)
	}

	if err := p.DeleteWord(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "cd" || p.formState().caret != 0 {
		t.Fatalf("delete word value %q caret %d", p.FormValue("t"), p.formState().caret)
	}
}

func TestCaretMoveDoesNotEdit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)

	before, after := 0, 0
	p.Handle(Handlers{
		BeforeEdit: func(context.Context, Box) error { before++; return nil },
		Change:     func(context.Context, Box) error { after++; return nil },
	})

	bftClick(t, p, "t")
	p.setCaret(2, false)

	if err := p.KeyDown(ctx, "arrowleft"); err != nil {
		t.Fatal(err)
	}

	if before != 0 || after != 0 {
		t.Fatalf("caret move fired before %d change %d", before, after)
	}

	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}

	if before != 1 || after != 1 {
		t.Fatalf("edit fired before %d change %d", before, after)
	}
}
