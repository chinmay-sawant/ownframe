package csspreview

import (
	"fmt"
	"testing"
)

func TestCatalogHasOneStartButtonForEveryProperty(t *testing.T) {
	p, _, v := fixturePage(t)
	if len(v.Catalog) != 43 {
		t.Fatalf("catalog rows = %d, want 43", len(v.Catalog))
	}
	names, buttons := map[string]bool{}, map[string]bool{}
	for i, row := range v.Catalog {
		if row.Number != fmt.Sprintf("%02d", i+1) {
			t.Errorf("%s has number %q, want %02d", row.Name, row.Number, i+1)
		}
		if names[row.Name] || buttons[row.Button] {
			t.Fatalf("duplicate catalog row or button: %+v", row)
		}
		names[row.Name], buttons[row.Button] = true, true
		button, ok := findBox(p.Boxes(), row.Button)
		if !ok || button.Tag != "button" {
			t.Errorf("%s has no rendered Start button", row.Name)
		}
		if row.Value == "" {
			t.Errorf("%s has no display-list value", row.Name)
		}
		if row.Name == "transition-behavior" && row.Value != "allow-discrete" {
			t.Errorf("transition-behavior value = %q, want allow-discrete", row.Value)
		}
	}
}
