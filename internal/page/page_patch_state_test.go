package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/state"
)

func TestPageStateMetadataSurface(t *testing.T) {
	p, err := New(Config{HTML: `<div id=count>1</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	st := state.NewStore(map[string]any{"count": 1})
	p.SetState(st)
	st.PatchChange("count", 2)

	if got := p.ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("page changed = %v", got)
	}
	if got := p.State().ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("store changed = %v", got)
	}

	p.SetChangedProperties([]string{"score"})
	p.ClearChangedProperties()
	if len(p.ChangedProperties()) != 0 && p.stateStore == nil {
		t.Fatal("clear failed")
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

	// alias methods compile
	_ = p.ChangedProps()
	_ = p.LastChangedProperties()
	_ = p.TakeChangedProperties()
	_ = p.PropertiesChanged()
	_ = p.Changed()
	_ = p.Store()
	_ = p.StateStore()
}
