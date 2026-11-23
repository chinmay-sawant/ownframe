package textrun_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

// textOp renders html and returns its first text op.
func textOp(t *testing.T, html string) *layout.DisplayOp {
	t.Helper()
	p, err := page.New(page.Config{HTML: html, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	for i := range d.Ops {
		if d.Ops[i].Text != "" && d.Ops[i].Font != nil {
			return &d.Ops[i]
		}
	}
	t.Fatal("no text op")

	return nil
}

func TestFaceKeyIncludesFontFeatures(t *testing.T) {
	plain := textOp(t, `<body><p style="font-size:20px">1111 iiii</p></body>`)
	tnum := textOp(t, `<body><p style="font-size:20px;font-feature-settings:'tnum' 1">1111 iiii</p></body>`)
	if tnum.FontFeatures() == "" {
		t.Skip("engine did not carry font-feature-settings on the op")
	}

	a, b := textrun.FaceFor(plain, 1), textrun.FaceFor(tnum, 1)
	if a == nil || b == nil {
		t.Fatal("no face")
	}
	if a == b {
		t.Fatal("faces with different OpenType features share one cache entry")
	}
}

func TestTrimDropsAndRebuildsFaces(t *testing.T) {
	op := textOp(t, `<body><p>hello</p></body>`)
	before := textrun.FaceFor(op, 1)
	if before == nil || textrun.FaceFor(op, 1) != before {
		t.Fatal("face not cached")
	}
	textrun.Trim()
	after := textrun.FaceFor(op, 1)
	if after == nil || after == before {
		t.Fatal("Trim kept the cached face")
	}
}
