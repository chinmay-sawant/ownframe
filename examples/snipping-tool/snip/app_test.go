package snip_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/snipping-tool/snip"
)

// TestMain keeps every test off the desktop clipboard.
func TestMain(m *testing.M) {
	ownframe.UseMemory(true)
	os.Exit(m.Run())
}

var pngMagic = []byte{137, 80, 78, 71, 13, 10, 26, 10}

func newTestApp(t *testing.T, c snip.Capturer) (*snip.App, context.Context) {
	t.Helper()

	app, err := snip.New(c)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.SetSize(snip.DefaultWidth, snip.DefaultHeight)

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app, ctx
}

func fakeShot(w, h int) snip.Capturer {
	raw := snip.DemoImage(w, h)
	return func(context.Context) ([]byte, error) { return raw, nil }
}

func TestNewCapturesFake(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, fakeShot(160, 100))
	view := app.View()

	if !view.HasImage || view.Size != "160x100 png" {
		t.Fatalf("view = %+v, want 160x100 png shot", view)
	}

	if !bytes.HasPrefix(app.Preview(), pngMagic) {
		t.Fatal("preview is not a PNG")
	}
}

func TestCaptureFallbackDemo(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	fail := snip.Capturer(func(context.Context) ([]byte, error) { return nil, boom })
	app, _ := newTestApp(t, fail)
	view := app.View()

	if !view.HasImage {
		t.Fatal("fallback shows no image")
	}

	if !strings.HasPrefix(view.Status, "demo image") {
		t.Fatalf("status = %q, want demo image", view.Status)
	}
}
