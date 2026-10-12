package snip_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

func TestSaveWritesTempFile(t *testing.T) {
	t.Parallel()

	app, ctx := newTestApp(t, fakeShot(160, 100))
	dir := t.TempDir()
	app.SetSaveDir(dir)

	if err := app.Handle(ctx, "save"); err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "snip-*.png"))
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %v, err = %v, want one snip png", files, err)
	}

	raw, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(raw, pngMagic) {
		t.Fatal("saved file is not a PNG")
	}

	if !strings.HasPrefix(app.View().Status, "saved ") {
		t.Fatalf("status = %q, want saved path", app.View().Status)
	}
}

func TestCopySetsStatus(t *testing.T) {
	t.Parallel()

	app, ctx := newTestApp(t, fakeShot(160, 100))

	if err := app.Handle(ctx, "copy"); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Status; got != "copied 160x100 png" {
		t.Fatalf("status = %q", got)
	}

	if !strings.HasPrefix(osclip.Read(), string(pngMagic)) {
		t.Fatal("clipboard does not hold the PNG")
	}
}
