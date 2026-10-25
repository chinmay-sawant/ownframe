package main

import (
	"testing"
)

// TestPlayAllStartsEverySample checks the Play all button: one click marks
// every card started and the page binds the ones its window holds.
func TestPlayAllStartsEverySample(t *testing.T) {
	p, d, v := fixturePage(t)
	clickID(t, p, playAllID)
	started := 0
	for _, row := range v.Catalog {
		if v.Started[row.Button] {
			started++
		}
	}
	if started != 43 {
		t.Fatalf("started cards = %d, want the whole catalog", started)
	}
	tickPage(t, p)
	if got := p.Probe().Motion().Bound; got < 29 {
		t.Fatalf("bound players = %d, want the animation cards", got)
	}
	if len(d.manual) != 9 {
		t.Fatalf("manual samples = %d, want the 6 geometry transitions and 3 trigger cards", len(d.manual))
	}
}

// TestRapidClicksStartEverySample checks that clicks that land before the
// next tick all start, instead of only the last one.
func TestRapidClicksStartEverySample(t *testing.T) {
	p, _, v := fixturePage(t)
	for _, row := range v.Catalog[:3] {
		clickID(t, p, row.Button)
	}
	tickPage(t, p)
	for _, row := range v.Catalog[:3] {
		if !v.Started[row.Button] {
			t.Fatalf("%s was not started", row.Name)
		}
	}
}
