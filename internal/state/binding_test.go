package state_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/state"
)

func TestBindingRebindDropsFormerStore(t *testing.T) {
	b := &state.Binding{}
	old := state.NewStore(map[string]any{"old": 0})
	current := state.NewStore(map[string]any{"current": 0})
	b.Bind(old)
	old.TryPatchChange("old", 1)
	b.Bind(current)
	if len(old.Listeners()) != 0 {
		t.Fatal("former store still retains the binding listener")
	}
	if got := b.ChangedProperties(); len(got) != 0 {
		t.Fatalf("former metadata survived rebinding: %v", got)
	}
	old.TryPatchChange("old", 2)
	current.TryPatchChange("current", 1)
	if got := b.ChangedProperties(); !reflect.DeepEqual(got, []string{"current", "current"}) {
		t.Fatalf("changes after rebinding = %v", got)
	}
}

func TestBindingIsIdempotentAndIndependent(t *testing.T) {
	store := state.NewStore(map[string]any{"count": 0})
	first, second := &state.Binding{}, &state.Binding{}
	first.Bind(store)
	first.Bind(store)
	second.Bind(store)
	if len(store.Listeners()) != 2 {
		t.Fatal("binding the same store twice registered another listener")
	}
	second.Bind(nil)
	store.TryPatchChange("count", 1)
	if len(first.ChangedProperties()) == 0 {
		t.Fatal("detaching one binding removed the other binding's listener")
	}
	if len(second.ChangedProperties()) != 0 {
		t.Fatal("detached binding still receives store changes")
	}
}

func TestBindingRebindIgnoresAlreadyDispatchedCallback(t *testing.T) {
	old := state.NewStore(map[string]any{"old": 0})
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	old.AddListener(func(state.Change) {
		close(started)
		<-release
	})
	b := &state.Binding{}
	b.Bind(old)
	go func() {
		defer close(done)
		old.TryPatchChange("old", 1)
	}()
	<-started
	b.Bind(state.NewStore(map[string]any{}))
	close(release)
	<-done
	if got := b.ChangedProperties(); len(got) != 0 {
		t.Fatalf("in-flight former-store callback changed the binding: %v", got)
	}
}

func TestBindingConcurrentRebindingAndMetadata(t *testing.T) {
	b := &state.Binding{}
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
				b.ChangedProperties()
				b.TakeChangedProperties()
				b.Sync()
				b.Store()
			}
		})
	}
	for range iterations {
		b.Bind(stores[0])
		b.Bind(stores[1])
		b.Bind(nil)
	}
	wg.Wait()
	b.Bind(nil)
	for _, store := range stores {
		if len(store.Listeners()) != 0 {
			t.Fatal("rebinding leaked registrations")
		}
	}
}

func TestBindingMetadataOwnsItsSnapshots(t *testing.T) {
	var b state.Binding
	input := []string{"first"}
	b.SetChangedProperties(input)
	input[0] = "mutated input"
	snapshot := b.ChangedProperties()
	if !reflect.DeepEqual(snapshot, []string{"first"}) {
		t.Fatalf("metadata aliases input: %v", snapshot)
	}
	snapshot[0] = "mutated output"
	b.AddChangedProperties("", "second", "")
	taken := b.TakeChangedProperties()
	if !reflect.DeepEqual(taken, []string{"first", "second"}) {
		t.Fatalf("metadata aliases snapshot or retains empty paths: %v", taken)
	}
	b.AddChangedProperties("replacement")
	if !reflect.DeepEqual(taken, []string{"first", "second"}) {
		t.Fatalf("drained snapshot aliases reused queue: %v", taken)
	}
	b.ClearChangedProperties()
	if got := b.ChangedProperties(); len(got) != 0 {
		t.Fatalf("clear left metadata: %v", got)
	}
}

func TestBindingSameStoreNewGenerationRejectsOldCallback(t *testing.T) {
	store := state.NewStore(map[string]any{"count": 0})
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	cancelBlocker := store.AddListener(func(state.Change) {
		close(started)
		<-release
	})
	defer cancelBlocker()
	var b state.Binding
	b.Bind(store)
	defer b.Bind(nil)
	go func() {
		defer close(done)
		store.TryPatchChange("count", 1)
	}()
	<-started
	b.Bind(nil)
	b.Bind(store)
	store.ClearChangedProperties()
	close(release)
	<-done
	if got := b.ChangedProperties(); len(got) != 0 {
		t.Fatalf("former generation of the same store supplied metadata: %v", got)
	}
}

func TestBindingDrainPreservesSharedStoreMetadata(t *testing.T) {
	store := state.NewStore(map[string]any{"count": 0})
	var first, second state.Binding
	first.Bind(store)
	second.Bind(store)
	defer first.Bind(nil)
	defer second.Bind(nil)
	store.TryPatchChange("count", 1)
	expected := store.ChangedProperties()
	if len(first.TakeChangedProperties()) == 0 {
		t.Fatal("listener supplied no local metadata")
	}
	if got := first.TakeChangedProperties(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("empty queue fallback = %v, want %v", got, expected)
	}
	if len(second.ChangedProperties()) == 0 {
		t.Fatal("one binding drained the other binding")
	}
	first.ClearChangedProperties()
	if got := store.ChangedProperties(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("local clear changed store metadata: %v", got)
	}
	store.ClearChangedProperties()
	if got := first.ChangedProperties(); len(got) != 0 {
		t.Fatalf("cleared store still supplied fallback: %v", got)
	}
}
