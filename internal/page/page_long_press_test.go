package page

import (
	"context"
	"testing"
)

func TestLongPressSelectsWord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="hello world"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "hello world")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	claimed, err := p.LongPress(ctx, left+width*0.1, base-4)
	if err != nil {
		t.Fatal(err)
	}

	if !claimed {
		t.Fatal("a text-field hold is unclaimed")
	}

	if start, end := p.formState().bounds(11); start != 0 || end != 5 {
		t.Fatalf("word [%d,%d)", start, end)
	}
}

func TestLongPressCallsHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<div id="card" style="width:120px;height:40px">`+
		`<div style="width:20px;height:20px">hit</div></div>`)
	bftDraw(t, p, nil)

	got := ""
	p.Handle(Handlers{LongPress: func(_ context.Context, box Box) error {
		got = box.ID

		return nil
	}})

	child := Box{}

	for _, b := range p.Boxes() {
		if b.Tag == "div" && b.ID == "" && b.W > 0 {
			child = b
		}
	}

	if child.W <= 0 {
		t.Fatal("no child div box")
	}

	claimed, err := p.LongPress(ctx, child.X+child.W/2, child.Y+child.H/2)
	if err != nil {
		t.Fatal(err)
	}

	if !claimed {
		t.Fatal("a handled hold is unclaimed")
	}

	if got != "card" {
		t.Fatalf("handler box = %q, want card", got)
	}
}

func TestLongPressNilHandlerUnclaimed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<div id="card" style="width:120px;height:40px">tap</div>`)
	bftDraw(t, p, nil)

	box := p.boxByID("card")

	claimed, err := p.LongPress(ctx, box.X+box.W/2, box.Y+box.H/2)
	if err != nil {
		t.Fatal(err)
	}

	if claimed {
		t.Fatal("a nil handler claimed the press")
	}
}
