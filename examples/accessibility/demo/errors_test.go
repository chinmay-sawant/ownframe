package demo

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestAccessibilityUnavailableBeforeFirstRender(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.accessible().AccessibilityTree()
	if !errors.Is(err, ownframe.ErrAccessibilityUnavailable) {
		t.Fatalf("tree error = %v", err)
	}
}

func TestExampleProbesEveryAccessibilityError(t *testing.T) {
	a := renderedApp(t)
	want := map[string]error{
		"probe-stale":    ownframe.ErrAccessibilityStale,
		"probe-node":     ownframe.ErrAccessibilityNode,
		"probe-action":   ownframe.ErrAccessibilityAction,
		"probe-disabled": ownframe.ErrAccessibilityDisabled,
		"probe-range":    ownframe.ErrAccessibilityRange,
	}
	for probe, sentinel := range want {
		t.Run(probe, func(t *testing.T) {
			err := a.runErrorProbe(context.Background(), probe)
			if !errors.Is(err, sentinel) {
				t.Fatalf("error = %v, want %v", err, sentinel)
			}
		})
	}
}
