package form

import "testing"

func TestOptionHelpers(t *testing.T) {
	t.Parallel()

	opts := []Option{{Value: "one"}, {Value: "two", Selected: true}}
	if PickOption(opts) != "two" || !opts[1].Selected || opts[0].Selected {
		t.Fatal("pick")
	}
	c := NextOption(Control{Options: opts})
	if c.Value != "one" || !c.Options[0].Selected || SelectedIndex(c.Options) != 0 {
		t.Fatal(c)
	}
	if ShownText("password", "ab") != "••" {
		t.Fatal("password")
	}
	if ShownText("file", "") != "No file" {
		t.Fatal("file")
	}
}
