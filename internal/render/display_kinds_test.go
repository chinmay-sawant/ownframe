package render_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestDisplayNamesEveryKindItEmits guards the mirror list in display_ops.go. A
// kind the engine emits but this package does not name is one a caller cannot
// switch on, and it would fall through to the zero value.
func TestDisplayNamesEveryKindItEmits(t *testing.T) {
	t.Parallel()

	named := map[render.Kind]bool{
		render.OpNoop:       true,
		render.OpUnknown:    true,
		render.OpFillRect:   true,
		render.OpStrokeRect: true,
		render.OpLine:       true,
		render.OpText:       true,
		render.OpImage:      true,
		render.OpLinkURI:    true,
		render.OpBullet:     true,
		render.OpGridRun:    true,
		render.OpPath:       true,
	}

	sources := []string{
		`<div style="background:#eee;border:1px solid #333;border-radius:8px">box</div>`,
		`<p>text</p><ul><li>bullet</li></ul><a href="/x">link</a>`,
		`<table style="border-collapse:collapse"><tr><td style="border:1px solid #ccc">a</td></tr></table>`,
		`<div style="isolation:isolate;background:#eee">group</div>`,
		`<div style="background:#eee;border:2px solid #333;border-radius:12px;corner-shape:bevel">path</div>`,
	}

	for _, source := range sources {
		display, err := render.DisplayList(context.Background(), source, 320, 200)
		if err != nil {
			t.Fatal(err)
		}

		for index, op := range display.Ops {
			if !named[op.Kind] {
				t.Errorf("source %q op %d carries unnamed kind %d", source, index, op.Kind)
			}
		}
	}
}
