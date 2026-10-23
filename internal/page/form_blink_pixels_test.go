package page

import (
	"bytes"
	"context"
	"image/png"
	"testing"
	"time"
)

func TestCaretBlinkDoesNotMoveFieldText(t *testing.T) {
	p, err := New(Config{HTML: `<style>body{margin:0;background:white}</style><input id="e" value="Focused text stays visible">`, Width: 360, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err = p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}
	for _, caret := range []int{0, 13, 27} {
		st := p.formState()
		st.Caret, st.Anchor = caret, caret
		p.resetCaretBlink()
		if err = p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		shown, err := png.Decode(bytes.NewReader(p.PNG()))
		if err != nil {
			t.Fatal(err)
		}
		st = p.formState()
		st.BlinkAt = time.Now().Add(-caretBlinkInterval)
		if err = p.Tick(ctx); err != nil {
			t.Fatal(err)
		}
		hidden, err := png.Decode(bytes.NewReader(p.PNG()))
		if err != nil {
			t.Fatal(err)
		}
		minX, maxX := shown.Bounds().Dx(), -1
		for y := 0; y < shown.Bounds().Dy(); y++ {
			for x := 0; x < shown.Bounds().Dx(); x++ {
				if shown.At(x, y) == hidden.At(x, y) {
					continue
				}
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}
		if maxX < minX {
			t.Fatalf("caret at %d did not change any pixels", caret)
		}
		if maxX-minX > 4 {
			t.Fatalf("caret at %d changed pixels across x=%d..%d; expected only the caret column to change", caret, minX, maxX)
		}
	}
}
