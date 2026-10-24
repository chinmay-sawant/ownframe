package main

import (
	"testing"
	"time"
)

// TestStartChangesOnlyItsTarget checks that one Start click changes the
// clicked card's preview and leaves the other 42 previews at their declared
// paint.
func TestStartChangesOnlyItsTarget(t *testing.T) {
	for _, name := range []string{"animation", "transition-property"} {
		t.Run(name, func(t *testing.T) {
			p, _, _ := fixturePage(t)
			// Settle the paused cards once; their 0% pose is the declared paint
			// on the first sample and must not count as a sibling change.
			tickPage(t, p)
			selected, ok := targetByKey(catalogSampleKey(name))
			if !ok {
				t.Fatalf("missing sample for %q", name)
			}
			base := make(map[string][]paint, len(targets()))
			for _, item := range targets() {
				base[item.key] = targetPaints(t, p, item.id)
			}
			clickID(t, p, selected.button)
			time.Sleep(300 * time.Millisecond)
			tickPage(t, p)
			changed := !samePaint(base[selected.key], targetPaints(t, p, selected.id))
			if !changed {
				t.Fatalf("target %q did not change", selected.id)
			}
			for _, item := range targets() {
				if item.key == selected.key {
					continue
				}
				if !samePaint(base[item.key], targetPaints(t, p, item.id)) {
					t.Errorf("starting %q changed sibling %q", selected.key, item.key)
				}
			}
		})
	}
}
