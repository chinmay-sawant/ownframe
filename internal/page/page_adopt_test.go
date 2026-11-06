package page

import (
	"context"
	"errors"
	"testing"
)

func TestAdoptPatchedSourceKeepsHTMLInStep(t *testing.T) {
	t.Parallel()

	oldSource := builderDocHead + "Adopt" + builderDocMid +
		`<div id="x">old</div>` + builderDocTail
	newSource := builderDocHead + "Adopt" + builderDocMid +
		`<div id="x">new</div>` + builderDocTail

	p, err := New(Config{HTML: oldSource, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	text := "new"
	if err := p.PatchElement(ctx, "x", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	if err := AdoptPatchedSource(p, newSource); err != nil {
		t.Fatal(err)
	}
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() = %q, want the adopted source", got)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() after redraw = %q, want the adopted source", got)
	}
}

func TestAdoptPatchedSourceRefusesHandWrittenHTML(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<p>plain</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := AdoptPatchedSource(p, `<p>other</p>`); !errors.Is(err, ErrAdoptSource) {
		t.Fatalf("err = %v, want %v", err, ErrAdoptSource)
	}
	if got := p.HTML(); got != `<p>plain</p>` {
		t.Fatalf("HTML() = %q, want the original source", got)
	}
}
