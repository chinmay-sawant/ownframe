package form

import "testing"

func TestOrder(t *testing.T) {
	t.Parallel()

	byID := map[string]Control{
		"a": {ID: "a"},
		"b": {ID: "b", Disabled: true},
		"c": {ID: "c", TabIndex: -1},
		"d": {ID: "d"},
	}
	got := Order([]string{"a", "a", "b", "c", "d", ""}, func(id string) (Control, bool) {
		c, ok := byID[id]
		return c, ok
	})
	if len(got) != 2 || got[0] != "a" || got[1] != "d" {
		t.Fatal(got)
	}
	if Order(nil, nil) == nil {
		t.Fatal("nil")
	}
}
