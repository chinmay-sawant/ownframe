package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestStatusTablesHaveNoOverlap(t *testing.T) {
	t.Parallel()

	for name := range painterDeclines {
		if opOnly[name] {
			t.Errorf("%q is listed in both painterDeclines and opOnly", name)
		}
	}
}

func TestNativelyReplayedPropertiesNotDeclined(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	head := genHead()
	mask, err := shapesMask("mask")
	if err != nil {
		t.Fatal(err)
	}

	byName := map[string]genProp{}
	for _, r := range genProps() {
		byName[r.Name] = r
	}

	for _, name := range []string{"writing-mode", "text-autospace", "object-fit", "object-view-box"} {
		if noPaint[name] || painterDeclines[name] || opOnly[name] {
			t.Errorf("%q should not be in noPaint/painterDeclines/opOnly", name)
		}

		r, ok := byName[name]
		if !ok {
			t.Fatalf("%q not found in genProps", name)
		}

		spec := specFor(r)
		page, err := ownframe.New(ownframe.Config{
			HTML:   head + spec.expand(r, "p", true),
			Width:  300,
			Height: 200,
		})
		if err != nil {
			t.Fatalf("%s: New: %v", name, err)
		}

		page.SetImage("mask", mask)

		if err = page.Redraw(ctx); err != nil {
			t.Fatalf("%s: Redraw: %v", name, err)
		}

		if page.Probe().Display() == nil {
			t.Errorf("%s fell back to bitmap instead of replaying natively", name)
		}
	}
}
