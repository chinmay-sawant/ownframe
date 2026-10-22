package studio_test

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/theme-studio/studio"
)

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func start(t *testing.T) *studio.App {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	app, err := studio.New()
	must(t, err)
	must(t, app.Redraw(t.Context()))

	return app
}

func clickBox(t *testing.T, app *studio.App, id string) {
	for _, b := range app.Boxes() {
		if b.ID == id {
			must(t, app.Click(t.Context(), b.X+b.W/2, b.Y+b.H/2))

			return
		}
	}

	t.Fatalf("box %q", id)
}

func hasFill(t *testing.T, app *studio.App, c [3]float64) bool {
	d := app.Page().Probe().Display()
	if d == nil {
		t.Fatal("bitmap")
	}

	for _, op := range d.Ops {
		if op.Kind == layout.DisplayOpFillRect && math.Abs(op.R-c[0]) < 0.01 &&
			math.Abs(op.G-c[1]) < 0.01 && math.Abs(op.B-c[2]) < 0.01 {
			return true
		}
	}

	return false
}

func TestPresetDark(t *testing.T) {
	app := start(t)
	clickBox(t, app, "preset-dark")

	v := app.View()
	var acc studio.Token

	for _, tok := range v.Tokens {
		if strings.HasSuffix(tok.Name, "accent") {
			acc = tok
		}
	}

	if v.Preset != "dark" || !acc.OK || !hasFill(t, app, acc.RGB) {
		t.Fatalf("%q %+v", v.Preset, acc)
	}
}

func TestTypedCSS(t *testing.T) {
	app := start(t)
	css := ":root { --accent: #ff0000; }"

	app.Page().SetFormValue("css", css)
	clickBox(t, app, "btn-apply")

	if !hasFill(t, app, [3]float64{1}) {
		t.Fatal("no fill")
	}

	got, err := os.ReadFile(app.ThemePath())
	if err != nil || string(got) != css {
		t.Fatalf("%q %v", got, err)
	}
}

func TestCopyCSS(t *testing.T) {
	ownframe.UseMemory(true)

	app := start(t)
	css := ":root { --accent: #0000ff; }"

	app.Page().SetFormValue("css", css)
	clickBox(t, app, "btn-copy")

	got, err := ownframe.Read()
	if err != nil || got != css {
		t.Fatal(got, err)
	}
}
