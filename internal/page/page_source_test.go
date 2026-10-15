package page_test

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// writeSource writes body into a fresh temp file and returns its path.
func writeSource(t *testing.T, name, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func newFilePage(t *testing.T, cfg page.Config) *page.Page {
	t.Helper()

	p, err := page.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestNewReadsFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	html := `<div id="a">from disk</div>`
	fromFile := newFilePage(t, page.Config{
		File:   writeSource(t, "index.html", html),
		Width:  320,
		Height: 200,
	})
	fromString := newFilePage(t, page.Config{HTML: html, Width: 320, Height: 200})

	if fromFile.HTML() != html {
		t.Fatalf("HTML = %q", fromFile.HTML())
	}

	if err := fromFile.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := fromString.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	boxCenter(t, fromFile, "a")

	fileImg, err := png.Decode(bytes.NewReader(fromFile.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	stringImg, err := png.Decode(bytes.NewReader(fromString.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	if !samePicture(fileImg, stringImg) {
		t.Fatal("file page and string page painted differently")
	}
}

func TestNewReadsThemeFile(t *testing.T) {
	t.Parallel()

	p := newFilePage(t, page.Config{
		HTML:      themePage,
		ThemeFile: writeSource(t, "theme.css", `#swatch{background:#ff0000}`),
		Width:     320,
		Height:    200,
	})

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if r, g, b := swatchRGB(t, p); r < 1 || g > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want red", r, g, b)
	}
}
