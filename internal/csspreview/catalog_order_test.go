package csspreview

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestCombinedPageUsesPerPropertyPreviews(t *testing.T) {
	p, _, v := fixturePage(t)
	catalog, hasCatalog := findBox(p.Boxes(), "catalog")
	metadata, hasMetadata := findBox(p.Boxes(), "metadata")
	if !hasCatalog || !hasMetadata || catalog.Y >= metadata.Y {
		t.Fatalf("catalog/metadata order: %+v, %+v", catalog, metadata)
	}
	for _, id := range []string{"animations", "transitions"} {
		if _, ok := findBox(p.Boxes(), id); ok {
			t.Errorf("shared %s preview section remains", id)
		}
	}
	for _, row := range v.Catalog {
		card, _ := findBox(p.Boxes(), "catalog-row-"+row.Number)
		preview, _ := findBox(p.Boxes(), row.TargetID)
		button, _ := findBox(p.Boxes(), row.Button)
		if !insideBox(card, preview) || !insideBox(card, button) {
			t.Errorf("row %s does not contain its preview and Start control", row.Number)
		}
	}
}

func insideBox(parent, child ownframe.Box) bool {
	return child.X >= parent.X && child.Y >= parent.Y &&
		child.X+child.W <= parent.X+parent.W && child.Y+child.H <= parent.Y+parent.H
}
