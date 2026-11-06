package pagetest_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// TestBoxByActionIndexed checks the page's action index answers the helper:
// the later box wins a repeated data-action, and misses and the empty action
// report false.
func TestBoxByActionIndexed(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML: `<div id="a" data-action="save" style="width:20px;height:20px">a</div>` +
			`<div id="b" data-action="save" style="width:20px;height:20px">b</div>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	box, ok := pagetest.BoxByAction(p, "save")
	if !ok {
		t.Fatal("save was not found")
	}
	if box.ID != "b" {
		t.Fatalf("BoxByAction(save) = %q, want the later box b", box.ID)
	}

	if _, ok := pagetest.BoxByAction(p, "missing"); ok {
		t.Fatal("a missing action must not resolve")
	}
	if _, ok := pagetest.BoxByAction(p, ""); ok {
		t.Fatal("the empty action must not resolve")
	}
}
