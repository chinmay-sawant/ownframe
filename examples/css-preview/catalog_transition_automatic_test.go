package main

import (
	"testing"
	"time"
)

// TestAutomaticTransitionCardsMove covers the transition cards the page
// runtime captures today: colour and opacity. They must not take the manual
// path, and one sample must change the paint.
func TestAutomaticTransitionCardsMove(t *testing.T) {
	for _, name := range []string{
		"transition-duration", "transition-delay", "transition-behavior",
		"-webkit-transition-duration", "-webkit-transition-delay",
	} {
		t.Run(name, func(t *testing.T) {
			p, d, v := fixturePage(t)
			row := catalogRowByName(t, v, name)
			// Settle the paused animation cards once.
			tickPage(t, p)
			base := targetPaints(t, p, row.TargetID)
			clickID(t, p, row.Button)
			if d.manual[row.SampleKey] != nil {
				t.Fatal("a paint-property transition took the manual path")
			}
			time.Sleep(500 * time.Millisecond)
			tickPage(t, p)
			if samePaint(base, targetPaints(t, p, row.TargetID)) {
				t.Fatalf("%s did not move off its declared paint", name)
			}
		})
	}
}
