package page

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFileShowsNoCaret(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="f" type="file" value="a.txt">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "f"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(p.source, `data-ownframe-focus="1"`) {
		t.Fatalf("focus missing: %s", p.source)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file paints a caret: %s", p.source)
	}

	p.formState().blinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file blinked a caret: %s", p.source)
	}

	if err := p.Type(ctx, "b"); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("f"); got != "a.txtb" {
		t.Fatalf("typed file value %q", got)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file paints a caret after type: %s", p.source)
	}
}
