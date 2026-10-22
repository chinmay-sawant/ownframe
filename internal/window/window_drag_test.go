package window

import "testing"

func TestWindowDragNeedsUserPress(t *testing.T) {
	var d windowDrag
	handled, move, _, _, _ := d.step(200, 100, false, true)
	if handled || move {
		t.Fatal("hover moved window")
	}
	handled, move, _, _, _ = d.step(200, 100, true, false)
	if handled || move {
		t.Fatal("transparent pixels began drag")
	}
	d.step(200, 100, false, false)
	d.step(200, 100, true, true)
	_, move, _, dx, dy := d.step(230, 120, true, false)
	if !move || dx != 30 || dy != 20 {
		t.Fatal("drag delta incorrect")
	}
	_, move, _, _, _ = d.step(200, 100, true, true)
	if move {
		t.Fatal("settled grab point moved again")
	}
	_, _, click, _, _ := d.step(200, 100, false, true)
	if click || d.active {
		t.Fatal("drag released as click")
	}
	d.step(200, 100, true, true)
	_, _, click, _, _ = d.step(201, 100, false, true)
	if !click {
		t.Fatal("small gesture did not click")
	}
}
