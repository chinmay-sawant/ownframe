package main

import "testing"

func TestCatalogAnimationSamplesAreFinite(t *testing.T) {
	p, _, _ := fixturePage(t)
	for _, target := range catalogTargets() {
		if target.kind != "animation" {
			continue
		}
		box, ok := findBox(p.Boxes(), target.id)
		if !ok {
			t.Fatalf("missing animation target %q", target.id)
		}
		for _, animation := range box.Animations {
			if animation.IterationInfinite {
				t.Errorf("%s unexpectedly loops forever", target.key)
			}
		}
	}
}
