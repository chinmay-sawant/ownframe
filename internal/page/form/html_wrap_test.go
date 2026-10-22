package form

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestFieldValueWraps checks that a long value wraps inside the rewritten
// field and grows the box instead of overflowing on one line.
func TestFieldValueWraps(t *testing.T) {
	t.Parallel()

	value := strings.Repeat("wrap ", 40)
	src := `<input id="e" type="text" style="width:120px">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: value}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "", false)
	_, boxes, err := render.Paint(context.Background(), got, 640, 400)
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range boxes {
		if b.ID != "e" {
			continue
		}

		if b.Text != value {
			t.Fatalf("text = %q", b.Text)
		}

		if b.H < 60 {
			t.Fatalf("H = %v, want a wrapped, grown box", b.H)
		}

		return
	}

	t.Fatal("no box for e")
}
