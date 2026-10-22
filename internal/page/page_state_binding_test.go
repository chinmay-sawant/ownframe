package page

import (
	"reflect"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/state"
)

func TestStateRebindDropsFormerStore(t *testing.T) {
	p := &Page{}
	old := state.NewStore(map[string]any{"old": 0})
	current := state.NewStore(map[string]any{"current": 0})
	p.BindState(old)
	old.TryPatchChange("old", 1)
	p.BindState(current)
	if len(old.Listeners()) != 0 {
		t.Fatal("former store still retains the page listener")
	}
	if got := p.ChangedProperties(); len(got) != 0 {
		t.Fatalf("former metadata survived rebinding: %v", got)
	}
	old.TryPatchChange("old", 2)
	current.TryPatchChange("current", 1)
	if got := p.ChangedProperties(); !reflect.DeepEqual(got, []string{"current", "current"}) {
		t.Fatalf("changes after rebinding = %v", got)
	}
}

func TestStateBindingIsIdempotentAndIndependent(t *testing.T) {
	store := state.NewStore(map[string]any{"count": 0})
	first, second := &Page{}, &Page{}
	first.BindState(store)
	first.BindState(store)
	second.BindState(store)
	if len(store.Listeners()) != 2 {
		t.Fatal("binding the same store twice registered another listener")
	}
	second.BindState(nil)
	store.TryPatchChange("count", 1)
	if len(first.ChangedProperties()) == 0 {
		t.Fatal("detaching one page removed the other page's listener")
	}
	if len(second.ChangedProperties()) != 0 {
		t.Fatal("detached page still receives store changes")
	}
}

func TestStateRebindIgnoresAlreadyDispatchedCallback(t *testing.T) {
	old := state.NewStore(map[string]any{"old": 0})
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	old.AddListener(func(state.Change) {
		close(started)
		<-release
	})
	p := &Page{}
	p.BindState(old)
	go func() {
		defer close(done)
		old.TryPatchChange("old", 1)
	}()
	<-started
	p.BindState(state.NewStore(map[string]any{}))
	close(release)
	<-done
	if got := p.ChangedProperties(); len(got) != 0 {
		t.Fatalf("in-flight former-store callback changed the page: %v", got)
	}
}

func TestConcurrentStateRebindingAndMetadata(t *testing.T) {
	p := &Page{}
	stores := []*state.Store{
		state.NewStore(map[string]any{"count": 0}),
		state.NewStore(map[string]any{"count": 0}),
	}
	var wg sync.WaitGroup
	const iterations = 100
	for _, store := range stores {
		wg.Go(func() {
			for i := range iterations {
				store.TryPatchChange("count", i)
				p.ChangedProperties()
				p.TakeChangedProperties()
				p.SyncState()
				p.State()
			}
		})
	}
	for range iterations {
		p.BindState(stores[0])
		p.BindState(stores[1])
		p.BindState(nil)
	}
	wg.Wait()
	p.BindState(nil)
	for _, store := range stores {
		if len(store.Listeners()) != 0 {
			t.Fatal("rebinding leaked registrations")
		}
	}
}
