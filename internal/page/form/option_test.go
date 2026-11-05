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

func TestNextOptionSkipsDisabled(t *testing.T) {
	t.Parallel()

	opts := []Option{
		{Value: "one", Selected: true, Disabled: true},
		{Value: "two", Disabled: true},
		{Value: "three"},
	}
	c := NextOption(Control{Options: opts, Value: "one"})
	if c.Value != "three" || !c.Options[2].Selected || c.Options[0].Selected || c.Options[1].Selected {
		t.Fatalf("skipped to %+v", c)
	}

	wrap := []Option{
		{Value: "one"},
		{Value: "two", Selected: true},
		{Value: "three", Disabled: true},
	}
	wrapped := NextOption(Control{Options: wrap, Value: "two"})
	if wrapped.Value != "one" || !wrapped.Options[0].Selected || wrapped.Options[2].Selected {
		t.Fatalf("wrap landed on %+v", wrapped)
	}

	allOff := []Option{
		{Value: "one", Selected: true, Disabled: true},
		{Value: "two", Disabled: true},
	}
	stuck := NextOption(Control{Options: allOff, Value: "one"})
	if stuck.Value != "one" || !stuck.Options[0].Selected || stuck.Options[1].Selected {
		t.Fatalf("all disabled changed to %+v", stuck)
	}
}
