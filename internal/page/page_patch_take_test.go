package page

import "testing"

func TestTakeDrainsLocalProps(t *testing.T) {
	p, err := New(Config{HTML: `<div id="a">A</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	p.SetChangedProperties([]string{"a"})
	first := p.TakeChangedProperties()
	if len(first) != 1 || first[0] != "a" {
		t.Fatalf("first = %v", first)
	}
	if second := p.TakeChangedProperties(); len(second) != 0 {
		t.Fatalf("second = %v, want empty", second)
	}
	p.SetChangedProperties([]string{"b"})
	if got := p.TakeChangedProperties(); len(got) != 1 || got[0] != "b" {
		t.Fatalf("second set = %v", got)
	}
	if got := p.TakeChangedProperties(); len(got) != 0 {
		t.Fatalf("second drain = %v", got)
	}
}
