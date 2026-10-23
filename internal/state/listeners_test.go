package state

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func TestListenerCancellationUsesRegistrationIdentity(t *testing.T) {
	store := NewStore(map[string]any{"count": 0})
	var got []int
	register := func(id int) func() {
		return store.AddListener(func(Change) { got = append(got, id) })
	}
	cancelFirst, cancelSecond := register(1), register(2)
	cancelSecond()
	cancelSecond()
	store.TryPatchChange("count", 1)
	if !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("cancelling the second closure affected another registration: %v", got)
	}
	cancelFirst()
	if len(store.Listeners()) != 0 {
		t.Fatal("cancelled registrations still retained")
	}
}

func TestListenerCancellationAllowsReentry(t *testing.T) {
	store := NewStore(map[string]any{"count": 0, "other": 0})
	var cancel func()
	calls := 0
	cancel = store.AddListener(func(Change) {
		calls++
		cancel()
		store.TryPatchChange("other", 1)
	})
	store.TryPatchChange("count", 1)
	if calls != 1 {
		t.Fatalf("cancelled callback recurred: %d", calls)
	}
}

func TestConcurrentListenerRegistrationAndCancellation(t *testing.T) {
	store := NewStore(map[string]any{"count": 0})
	var calls atomic.Int64
	var wg sync.WaitGroup
	const workers, iterations = 4, 100
	for range workers {
		wg.Go(func() {
			for i := range iterations {
				cancel := store.AddListener(func(Change) { calls.Add(1) })
				store.TryPatchChange("count", i)
				cancel()
				cancel()
			}
		})
	}
	wg.Wait()
	before := calls.Load()
	store.TryPatchChange("count", 0)
	if calls.Load() != before || len(store.Listeners()) != 0 {
		t.Fatal("callbacks survived cancellation")
	}
}
