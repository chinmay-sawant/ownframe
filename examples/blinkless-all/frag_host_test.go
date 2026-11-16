package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestFragOutlineList(t *testing.T) {
	got, err := fragOutlineHTML(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"level 1, open: Guide",
		"level 2, open: Install",
		"level 2, closed: Usage: CLI",
		"level 3, open: Flags",
		"level 1, open: Reference",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("outline lacks %q:\n%s", want, got)
		}
	}

	if strings.Contains(got, "Skipped by level") {
		t.Error("bookmark-level: none must leave the heading out")
	}
}

func TestFragRunningHeader(t *testing.T) {
	got, err := fragRunningHeaderHTML(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"chapter-title = Chapter Seven", "chapter-id = ch-7"} {
		if !strings.Contains(got, want) {
			t.Errorf("header lacks %q: %s", want, got)
		}
	}

	if strings.Contains(got, "No copy") {
		t.Error("copy-into: none copied text")
	}
}

func TestFragFootnoteValues(t *testing.T) {
	got, err := fragFootnoteHTML(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, `fn-note: footnote-display="compact" footnote-policy="block"`) {
		t.Errorf("footnote values missing: %s", got)
	}
}

// TestFragEllipsisInDrawingList lays out the three paragraphs and reads the
// painted text: the marker is in the display list, not only in the page text.
func TestFragEllipsisInDrawingList(t *testing.T) {
	ctx := context.Background()

	for id, want := range map[string]string{"frag-ell-default": "…", "frag-ell-string": " [more]"} {
		doc, err := fragDocument(ctx, `<!DOCTYPE html><html><body>`+fragEllipsisHTML()+`</body></html>`)
		if err != nil {
			t.Fatal(err)
		}

		display, err := layout.DisplayList(ctx, doc)
		if err != nil {
			t.Fatal(err)
		}

		var text strings.Builder

		for i, op := range display.Ops {
			if op.Kind != layout.DisplayOpText || display.OpOwners[i] < 0 {
				continue
			}

			if display.Boxes[display.OpOwners[i]].ID == id {
				text.WriteString(op.Text)
			}
		}

		if !strings.HasSuffix(text.String(), want) {
			t.Errorf("%s text %q does not end with %q", id, text.String(), want)
		}
	}
}

func TestFragBodyHasEverySection(t *testing.T) {
	body := ""
	for _, p := range fragTab().Props {
		body += p.Demo()
	}

	for _, id := range []string{"frag-outline", "frag-running-header", "frag-ell-default", "frag-ell-string", "frag-ell-none", "frag-footnotes"} {
		if !strings.Contains(body, `id="`+id+`"`) {
			t.Errorf("body lacks %s", id)
		}
	}
}
