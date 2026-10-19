package csspreview

import "testing"

func TestCombinedFixtureExportsVendorAndTransitionMetadata(t *testing.T) {
	p, _, _ := fixturePage(t)
	wkTarget, _ := targetByKey(catalogSampleKey("-webkit-animation"))
	wk, _ := findBox(p.Boxes(), wkTarget.id)
	if len(wk.Animations) != 1 || wk.Animations[0].Name != "webkit-hop" || wk.Animations[0].TimingFunction != "ease-in-out" {
		t.Fatalf("WebKit animation alias metadata: %+v", wk.Animations)
	}
	for _, target := range targets() {
		if target.kind != "transition" {
			continue
		}
		box, _ := findBox(p.Boxes(), target.id)
		found := false
		for _, record := range box.Transitions {
			if record.Property != target.property || record.DurationMS <= 0 || record.TimingFunction == "" || record.Behavior == "" {
				continue
			}
			found = true
			if target.key == catalogSampleKey("transition-delay") && record.Behavior != "allow-discrete" {
				t.Errorf("behavior = %q", record.Behavior)
			}
		}
		if !found {
			t.Errorf("transition metadata incomplete for %q: %+v", target.key, box.Transitions)
		}
	}
}
