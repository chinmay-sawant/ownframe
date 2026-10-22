package form

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestRewritePaint(t *testing.T) {
	t.Parallel()

	text := `<input type="text" id="email" value="x">`
	area := `<textarea id="bio">no</textarea>`
	tick := `<input type="checkbox" id="ok" checked>`
	html := text + area + tick
	spans := []Span{
		{Control: Control{ID: "email", Tag: "input", Type: "text", Value: "ada"}, Start: 0, End: len(text)},
		{Control: Control{ID: "bio", Tag: "textarea", Value: "ada bio"}, Start: len(text), End: len(text) + len(area)},
		{
			Control: Control{ID: "ok", Tag: "input", Type: "checkbox", Checked: true},
			Start:   len(text) + len(area),
			End:     len(html),
		},
	}
	got := Rewrite(html, spans, nil, "", false)
	_, boxes, err := render.Paint(context.Background(), got, 640, 400)
	if err != nil {
		t.Fatal(err)
	}

	found := map[string]layout.Box{}
	for _, box := range boxes {
		if box.ID != "" {
			found[box.ID] = box
		}
	}
	if found["email"].Text != "ada" || found["bio"].Text != "ada bio" || found["ok"].Tag != "input" {
		t.Fatalf("%+v", found)
	}
}
