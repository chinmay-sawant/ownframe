package state

import (
	"sync"
	"testing"
)

func TestPatchSkipsMissingValue(t *testing.T) {
	s := NewStore(map[string]any{"a": 0})
	var order []string
	s.AddListener(func(c Change) { order = append(order, c.Path) })
	out, skip := s.PatchWithSkipped(Patch{
		Origin:  OriginBackground,
		Changed: []string{"b", "a", "missing"},
		Values:  map[string]any{"a": 1, "b": 2},
	})
	if len(out) != 2 || len(skip) != 1 || skip[0] != "missing" {
		t.Fatalf("out=%d skip=%v", len(out), skip)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("order=%v want [a b]", order)
	}
	if got := s.Data().(map[string]any)["a"]; got != 1 {
		t.Fatalf("a=%v", got)
	}
	if s.Origin() != OriginBackground {
		t.Fatalf("origin=%s", s.Origin())
	}
}

func TestPatchOriginConcurrentRace(t *testing.T) {
	s := NewStore(map[string]any{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.Patch(Patch{Origin: OriginUserInput,
					Changed: []string{"k"}, Values: map[string]any{"k": j}})
				_, _ = s.TryPatchChange("k", j)
			}
		}(i)
	}
	wg.Wait()
	_ = s.Origin()
}
