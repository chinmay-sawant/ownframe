package state

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPatchIdentityConcurrent(t *testing.T) {
	s := NewStore(map[string]any{})
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			id := fmt.Sprintf("id-%d", g)
			path := fmt.Sprintf("k%d", g)
			for i := 0; i < 50; i++ {
				c := s.PatchChangeFor(id, path, i)
				if c.Identity != id {
					t.Errorf("identity = %q want %q", c.Identity, id)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	got := s.Changes()
	if len(got) != 400 {
		t.Fatalf("changes = %d want 400", len(got))
	}
	for _, c := range got {
		if c.Identity != c.ElementID || c.Identity != c.StableID {
			t.Fatalf("mismatch = %+v", c)
		}
	}
}

func TestPatchListenerReentrant(t *testing.T) {
	s := NewStore(map[string]any{"count": 0})
	s.AddListener(func(c Change) {
		if c.Path == "count" {
			_ = s.Data()
			s.PatchChange("other", 2)
		}
	})
	done := make(chan struct{})
	go func() {
		s.PatchChange("count", 1)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock on reentry")
	}
	if got := dataMap(t, s)["other"]; got != 2 {
		t.Fatalf("other = %v", got)
	}
}
