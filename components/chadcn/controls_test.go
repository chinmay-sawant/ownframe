package chadcn_test

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestButtonSizesChangeRenderedGeometry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 640, Height: 480, Theme: chadcn.Stylesheet}, ui.Div().Children(
		chadcn.Button("Action", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).ID("small"),
		chadcn.Button("Action", chadcn.ButtonOptions{}).ID("medium"),
		chadcn.Button("Action", chadcn.ButtonOptions{Size: chadcn.ButtonLarge}).ID("large"),
		chadcn.Button("Action", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).ID("outlined"),
		chadcn.Button("Action", chadcn.ButtonOptions{Variant: chadcn.ButtonText}).ID("text"),
		chadcn.Button("Action", chadcn.ButtonOptions{Variant: chadcn.ButtonDanger}).ID("danger"),
	))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	small, ok := pagetest.BoxByID(p, "small")
	if !ok {
		t.Fatal("missing small button")
	}
	medium, ok := pagetest.BoxByID(p, "medium")
	if !ok {
		t.Fatal("missing medium button")
	}
	large, ok := pagetest.BoxByID(p, "large")
	if !ok {
		t.Fatal("missing large button")
	}
	if !(small.H < medium.H && medium.H < large.H && small.W < medium.W && medium.W < large.W) {
		t.Fatalf("sizes did not change geometry: small=%.0fx%.0f medium=%.0fx%.0f large=%.0fx%.0f", small.W, small.H, medium.W, medium.H, large.W, large.H)
	}
}

func TestControlledSwitchAndCheckboxActivation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	switched, checked := false, false
	var view *ui.View
	var tree func() *ui.Element
	tree = func() *ui.Element {
		return ui.Div().Children(
			chadcn.Switch("Notifications", switched).ID("switch").OnClick(func(ctx context.Context) error { switched = !switched; return view.Update(ctx, tree()) }),
			chadcn.Checkbox("Compact", checked).ID("check").OnClick(func(ctx context.Context) error { checked = !checked; return view.Update(ctx, tree()) }),
		)
	}
	var err error
	view, err = ui.Mount(ownframe.Config{Width: 480, Height: 320}, tree())
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Form().Focus(ctx, "switch"); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Edit().KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	// Native Enter sends KeyDown followed by Submit.
	if err := view.Page.Edit().Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if !switched || checked {
		t.Fatal("keyboard activation changed the wrong state")
	}
	if err := pagetest.Click(ctx, view.Page, "check"); err != nil {
		t.Fatal(err)
	}
	if !checked || !switched {
		t.Fatal("checkbox activation changed the wrong state")
	}
	if strings.Count(view.Page.HTML(), `aria-checked="true"`) != 2 {
		t.Fatal("checked semantics were not published")
	}
}

func TestProgressBoundsAndAccessibleValue(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		value int
		want  string
	}{{-5, "0"}, {42, "42"}, {150, "100"}} {
		source, err := ui.Render(chadcn.Progress("Build", tt.value))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(source, `role="progressbar"`) || !strings.Contains(source, `aria-valuenow="`+tt.want+`"`) {
			t.Fatalf("incorrect progress semantics for %d", tt.value)
		}
	}
}

func TestSwitchPaintsTrackAndThumb(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320, Theme: `:root{--of-primary:#123456;--of-on-primary:#ffffff}`},
		ui.Div().Children(chadcn.Switch("Notifications", true).ID("switch"), chadcn.TextField("field", "Name", "")))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	boxes := p.Boxes()
	parent := -1
	for index, box := range boxes {
		if box.ID == "switch" {
			parent = index
		}
	}
	found := false
	img, err := png.Decode(bytes.NewReader(p.Testing().PNGWithContext(ctx)))
	if err != nil {
		t.Fatal(err)
	}
	for _, box := range boxes {
		if box.ParentIndex != parent || box.Tag != "div" || box.W != 36 || box.H != 20 {
			continue
		}
		found = true
		track := color.RGBAModel.Convert(img.At(int(box.X)+4, int(box.Y)+10))
		thumb := color.RGBAModel.Convert(img.At(int(box.X)+26, int(box.Y)+10))
		if track != (color.RGBA{R: 18, G: 52, B: 86, A: 255}) || thumb != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
			t.Fatalf("switch paint track=%v thumb=%v", track, thumb)
		}
	}
	if !found {
		for i, b := range boxes {
			t.Logf("box=%d tag=%s id=%s parent=%d size=%.1fx%.1f", i, b.Tag, b.ID, b.ParentIndex, b.W, b.H)
		}
		t.Fatal("switch track has no sized box")
	}
}
