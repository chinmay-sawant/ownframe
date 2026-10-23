package gesture

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestCursorShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   Cursor
		want host.Shape
	}{
		{name: "plain", in: Cursor{}, want: host.ShapeDefault},
		{name: "text", in: Cursor{HoverID: "t", Control: true, Editable: true}, want: host.ShapeText},
		{name: "button control", in: Cursor{HoverID: "b", Control: true}, want: host.ShapePointer},
		{name: "anchor", in: Cursor{HoverID: "a", Tag: "a"}, want: host.ShapePointer},
		{name: "button tag", in: Cursor{HoverID: "b", Tag: "button"}, want: host.ShapePointer},
		{name: "action", in: Cursor{HoverID: "x", Action: "go"}, want: host.ShapePointer},
		{name: "link run", in: Cursor{Link: true}, want: host.ShapePointer},
		{name: "plain hover", in: Cursor{HoverID: "p", Tag: "div"}, want: host.ShapeDefault},
	}

	for _, tc := range cases {
		if got := CursorShape(tc.in); got != tc.want {
			t.Fatalf("%s shape %v want %v", tc.name, got, tc.want)
		}
	}
}
