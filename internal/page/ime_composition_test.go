package page

import (
	"context"
	"strings"
	"testing"
)

func TestIMECompositionIsTemporary(t *testing.T) {
	p := imeField(t, "hello")
	beforeEdit, changes := 0, 0
	p.handlers.BeforeEdit = func(context.Context, Box) error {
		beforeEdit++
		return nil
	}
	p.handlers.Change = func(context.Context, Box) error {
		changes++
		return nil
	}
	before := p.Generation()
	if err := p.SetIMEComposition(context.Background(), "かな", 0, 6); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("e"); got != "hello" {
		t.Fatalf("committed value changed to %q", got)
	}
	if beforeEdit != 0 || changes != 0 {
		t.Fatalf("preedit called edit handlers: before=%d change=%d", beforeEdit, changes)
	}
	if p.Generation() <= before || !strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("composition was not drawn")
	}
	if err := p.SetIMEComposition(context.Background(), "", 0, 0); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("composition remained after clear")
	}
}
