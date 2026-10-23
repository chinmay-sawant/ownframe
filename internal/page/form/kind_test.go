package form

import "testing"

func TestTextLikeAndCanEdit(t *testing.T) {
	t.Parallel()

	if !TextLike("") || !TextLike("file") || TextLike("checkbox") {
		t.Fatal("textLike")
	}
	if !CanEdit(Control{Tag: "textarea"}) {
		t.Fatal("textarea")
	}
	if CanEdit(Control{Tag: "select"}) || CanEdit(Control{Type: "radio"}) {
		t.Fatal("non-edit")
	}
	if !IsFileInput(Control{Type: "file"}) || IsFileInput(Control{Type: "text"}) {
		t.Fatal("file")
	}
	if !WillChange(Control{Type: "checkbox"}) {
		t.Fatal("checkbox")
	}
	if WillChange(Control{Type: "radio", Checked: true}) {
		t.Fatal("checked radio")
	}
	if !WillChange(Control{Tag: "select", Options: []Option{{}, {}}}) {
		t.Fatal("select")
	}
}
