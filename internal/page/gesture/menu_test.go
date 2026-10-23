package gesture

import "testing"

func TestMenuRows(t *testing.T) {
	t.Parallel()

	rows := ContextMenu(Menu{})
	want := []string{"cut", "copy", "paste", "select-all", "undo", "redo"}
	if len(rows) != len(want) {
		t.Fatalf("rows %d", len(rows))
	}

	for i, id := range want {
		if rows[i].ID != id || rows[i].Label == "" || rows[i].Enabled {
			t.Fatalf("row %d %+v", i, rows[i])
		}
	}

	rows = ContextMenu(Menu{Editable: true, SelectAll: true})
	if rows[0].Enabled || !rows[2].Enabled || !rows[3].Enabled {
		t.Fatalf("editable %+v", rows)
	}

	rows = ContextMenu(Menu{Selection: true, Undo: true, Redo: true})
	if !rows[0].Enabled || !rows[1].Enabled || !rows[4].Enabled || !rows[5].Enabled {
		t.Fatalf("selection %+v", rows)
	}
}
