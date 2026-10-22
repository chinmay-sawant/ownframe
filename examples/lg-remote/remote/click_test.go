package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

type fakeLink struct {
	spec string
	host string
	wake int
	scan string
}

func (f *fakeLink) Exec(host, spec string) (string, error) {
	f.host = host
	f.spec = spec
	return "ok", nil
}

func (f *fakeLink) Scan() (string, error) { return f.scan, nil }

func (f *fakeLink) Wake() (string, error) {
	f.wake++
	return "Wake sent", nil
}

func (f *fakeLink) SavedHost() string  { return "" }
func (f *fakeLink) SavedModel() string { return "UP7750PTZ" }

func newTest(t *testing.T, opts ...Option) *App {
	t.Helper()

	app, err := New(opts...)
	if err != nil {
		t.Fatal(err)
	}

	useTestIcons(app)
	app.SetAsync(false)
	app.wantSearch = false
	app.view.PowerOn = true
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return app
}

func clickID(t *testing.T, app *App, id string) {
	t.Helper()

	var box ownframe.Box
	found := false
	for _, item := range app.Page().Boxes() {
		if item.ID == id {
			box = item
			found = true
		}
	}

	if !found {
		t.Fatalf("no box %s", id)
	}

	err := app.Page().Testing().Click(context.Background(), box.X+1, box.Y+1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestThemeToggleStartsDark(t *testing.T) {
	app := newTest(t)
	if app.ThemeLabel() != "Light" {
		t.Fatalf("label %s", app.ThemeLabel())
	}

	clickID(t, app, "theme")
	if app.ThemeLabel() != "Dark" || app.Status() != "Light theme" {
		t.Fatalf("theme %s %s", app.ThemeLabel(), app.Status())
	}
}

func TestVolumeUsesWifiSpec(t *testing.T) {
	app := newTest(t)
	fake := &fakeLink{}
	app.SetLink(fake)
	app.Page().SetFormValue("host", "10.0.0.8")
	clickID(t, app, "volup")
	if err := app.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fake.spec != "vol:Up" || fake.host != "10.0.0.8" {
		t.Fatalf("spec %s host %s", fake.spec, fake.host)
	}

	if app.Status() != "ok" {
		t.Fatalf("status %s", app.Status())
	}
}
