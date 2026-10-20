package viewtransition

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

func snapshot(t *testing.T, source string) *layout.Display {
	t.Helper()
	doc, err := html.Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	styled, err := css.Apply(context.Background(), doc, css.Options{WidthPx: 320, HeightPx: 240, Media: "screen"})
	if err != nil {
		t.Fatal(err)
	}
	display, err := layout.DisplayListOptions(context.Background(), styled, layout.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return display
}

func boxNamed(d *layout.Display, name string) int {
	for i, b := range d.Boxes {
		if b.ID == name {
			return i
		}
	}
	return -1
}
