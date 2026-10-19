package csspreview

import "testing"

func TestEveryCatalogCardOwnsItsPreviewAndStart(t *testing.T) {
	p, _, v := fixturePage(t, "all")
	sampled := 0
	targets := map[string]bool{}
	for _, row := range v.Catalog {
		card, hasCard := findBox(p.Boxes(), "catalog-row-"+row.Number)
		preview, hasPreview := findBox(p.Boxes(), row.TargetID)
		button, hasButton := findBox(p.Boxes(), row.Button)
		if !hasCard || !hasPreview || !hasButton || !insideBox(card, preview) || !insideBox(card, button) {
			t.Errorf("%s preview and Start are outside their numbered card: card=%+v preview=%+v button=%+v", row.Name, card, preview, button)
		}
		if row.SampleKey == "" {
			t.Errorf("%s has no working preview", row.Name)
			continue
		}
		sampled++
		target, ok := targetByKey("all", row.SampleKey)
		if !ok || target.id != row.TargetID || target.button != row.Button {
			t.Errorf("%s does not own its registered sample", row.Name)
		}
		if targets[row.TargetID] {
			t.Errorf("%s shares target %q", row.Name, row.TargetID)
		}
		targets[row.TargetID] = true
	}
	if sampled != 43 {
		t.Fatalf("working previews = %d, want 43", sampled)
	}
}
