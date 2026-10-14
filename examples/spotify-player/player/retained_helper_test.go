package player

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/element"
)

func TestRetainedHelperStableIDs(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	ids := StableIDs("card", 3)
	before := app.RetainedElements(ids...)
	redraw(t, app)
	after := app.RetainedElements(ids...)
	for i := range ids {
		AssertRetainedIdentity(t, before[i], after[i])
		if before[i].ID != ids[i] {
			t.Fatalf("id = %q", before[i].ID)
		}
	}
}

func TestRetainedHelperRegistry(t *testing.T) {
	app := mustApp(t)
	r := RetainedRegistry(app.Page())
	if r == nil {
		t.Fatal("nil registry")
	}
	_ = app.RetainedElement("must-exist")
	n := RetainFromRegistry(r, "helper-node", "button")
	if n == nil || n.ID != "helper-node" {
		t.Fatalf("node = %+v", n)
	}
	n2 := app.RetainElement("helper-node2", "button")
	if n2 == nil {
		t.Fatal("nil retained node")
	}
	if app.RetainedElement("missing") == nil {
		t.Fatal("missing node")
	}
}

func TestRetainedHelperLifecycle(t *testing.T) {
	n := element.NewNode("life", "button")
	element.Mount(n)
	element.Update(n, []string{"label"})
	AssertRetainedLifecycle(t, n, true, true, false)
	n2 := element.NewNode("life2", "button")
	element.Mount(n2)
	_ = n2
}
