package ui

import (
	"strings"
	"testing"
)

// TestTreeNodeIDsStable pins the sidebar element ids that context menus and
// tests address nodes by.
func TestTreeNodeIDsStable(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	var ids []string
	for _, b := range r.a.Page().Boxes() {
		if strings.HasPrefix(b.ID, "tree-") {
			ids = append(ids, b.ID)
		}
	}
	for _, want := range []string{"tree-schema:main", "tree-group:main:Tables", "tree-table:main.users", "tree-group:main:Views"} {
		found := false
		for _, id := range ids {
			found = found || id == want
		}
		if !found {
			t.Errorf("missing tree id %q in %v", want, ids)
		}
	}
}
