package page

import (
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/state"
)

// TestPageChangedPropsConcurrent covers RACE-1: store listeners fire on
// arbitrary goroutines while the page loop drains changedProps metadata.
// Run with -race: every access must hold changedPropsMu.
func TestPageChangedPropsConcurrent(t *testing.T) {
	p, err := New(Config{HTML: `<div>hi</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	st := state.NewStore(map[string]any{"count": 0})
	p.BindState(st)
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				st.TryPatchChange("count", g*1000+i)
				_ = p.ChangedProperties()
				_ = p.TakeChangedProperties()
				p.SyncState()
			}
		}(g)
	}
	wg.Wait()
	if got := p.State().ChangedProperties(); len(got) == 0 {
		t.Fatal("store metadata lost under concurrency")
	}
}
