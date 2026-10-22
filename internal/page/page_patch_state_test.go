package page

import (
	"context"
	"slices"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/state"
)

func TestPageStateMetadataSurface(t *testing.T) {
	p, err := New(Config{HTML: `<div id=count>1</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	st := state.NewStore(map[string]any{"count": 1})
	p.BindState(st)
	if _, ok := st.TryPatchChange("count", 2); !ok {
		t.Fatal("setPath failed for count")
	}

	if got := p.ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("page changed = %v", got)
	}
	if got := p.State().ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("store changed = %v", got)
	}

	p.SetChangedProperties([]string{"score"})
	p.ClearChangedProperties()
	if got := p.ChangedProperties(); !slices.Equal(got, st.ChangedProperties()) {
		t.Fatalf("local clear did not restore the store fallback: %v", got)
	}
	// state store still has count
	if got := p.State().ChangedProperties(); len(got) == 0 {
		t.Fatal("store metadata lost")
	}

	// DisplayPatch surfaces properties.
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()
	p.SetChangedProperties([]string{"count"})
	patch := p.PatchDisplay(func(d *Display) {
		if len(d.Ops) == 0 {
			t.Fatal("no ops")
		}
		d.Ops[0].X += 1
	})
	if len(patch.ChangedProperties) == 0 || patch.ChangedProperties[0] != "count" {
		t.Fatalf("patch props = %+v", patch)
	}

	p.BindState(nil)
	if p.State() != nil {
		t.Fatal("detached page still exposes a store")
	}
	st.TryPatchChange("count", 3)
	if got := p.ChangedProperties(); len(got) != 0 {
		t.Fatalf("detached page receives metadata: %v", got)
	}
}

func TestTakeDrainsLocalProps(t *testing.T) {
	p, err := New(Config{HTML: `<div id="a">A</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	p.SetChangedProperties([]string{"a"})
	first := p.TakeChangedProperties()
	if len(first) != 1 || first[0] != "a" {
		t.Fatalf("first = %v", first)
	}
	if second := p.TakeChangedProperties(); len(second) != 0 {
		t.Fatalf("second = %v, want empty", second)
	}
	p.SetChangedProperties([]string{"b"})
	if got := p.TakeChangedProperties(); len(got) != 1 || got[0] != "b" {
		t.Fatalf("second set = %v", got)
	}
	if got := p.TakeChangedProperties(); len(got) != 0 {
		t.Fatalf("second drain = %v", got)
	}
}
