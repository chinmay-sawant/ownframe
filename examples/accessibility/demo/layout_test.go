package demo

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestFormControlsFitTheirPanel(t *testing.T) {
	a := renderedApp(t)
	panel := boxByID(t, a.Page().Boxes(), "semantics")
	label := boxByID(t, a.Page().Boxes(), "display-name-label")
	field := boxByID(t, a.Page().Boxes(), "display-name")
	if label.W < panel.W*0.75 || label.H > 70 {
		t.Errorf("display-name label layout is cramped: %.1fx%.1f", label.W, label.H)
	}
	if field.W < panel.W*0.75 {
		t.Errorf("display-name width %.1f is too narrow for panel width %.1f", field.W, panel.W)
	}
	for _, id := range []string{"updates", "contact-email", "contact-phone"} {
		box := boxByID(t, a.Page().Boxes(), id)
		if box.W > 28 || box.H > 28 {
			t.Errorf("%s control is oversized: %.1fx%.1f", id, box.W, box.H)
		}
	}
}

func boxByID(t *testing.T, boxes []ownframe.Box, id string) ownframe.Box {
	t.Helper()
	for _, box := range boxes {
		if box.ID == id {
			return box
		}
	}
	t.Fatalf("box %q missing", id)
	return ownframe.Box{}
}
