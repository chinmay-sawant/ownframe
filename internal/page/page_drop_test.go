package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDropHandlerReceivesNameAndBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<p id="line">{{.}}</p>`,
		Width:  240,
		Height: 160,
	})
	if err != nil {
		t.Fatal(err)
	}

	var name string
	var data []byte

	screen.Handle(page.Handlers{
		Drop: func(_ context.Context, files []page.Drop) error {
			if len(files) != 1 {
				t.Fatalf("files = %d, want 1", len(files))
			}

			name = files[0].Name
			data, err = files[0].Read()

			return err
		},
	})

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	files := []page.Drop{{
		Name: "note.txt",
		Size: 5,
		Read: func() ([]byte, error) { return []byte("hello"), nil },
	}}

	if err := screen.Drop(ctx, files); err != nil {
		t.Fatal(err)
	}

	if name != "note.txt" || string(data) != "hello" {
		t.Fatalf("name = %q, data = %q", name, data)
	}

	// The handler changed no page input, so the redraw runs but commits an
	// identical frame: one more redraw, no new generation.
	if got := screen.Stats().Redraws; got != 2 {
		t.Fatalf("redraws = %d, want 2", got)
	}
	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}
