// Command blinkless-next opens one blinkless HTML fixture in the OwnFrame window.
// The fixture stays in the blinkless repo. This program reads that file and
// paints it with the blinkless checkout named in the parent go.work.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	width := flag.Int("width", 880, "window width in CSS pixels")
	height := flag.Int("height", 900, "window height in CSS pixels")
	pngOut := flag.String("png", "", "write one PNG and exit, without opening a window")
	flag.Parse()
	if flag.NArg() > 1 || *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/blinkless-next [-width N] [-height N] [-png out.png] [page.html]")
		os.Exit(2)
	}

	pagePath := reviewPage()
	if flag.NArg() == 1 {
		pagePath = flag.Arg(0)
	}

	html, err := os.ReadFile(pagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	page, err := ownframe.New(ownframe.Config{
		Title:  "blinkless next",
		HTML:   string(html),
		Width:  *width,
		Height: *height,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if swatch, err := swatchPNG(); err == nil {
		page.SetImage("swatch.png", swatch)
	}

	// env-swatch.svg reads env(--color, blue). The link-parameters sample
	// fills it red for one image and leaves the other on the fallback.
	page.SetImage("env-swatch.svg", []byte(envSwatchSVG))

	ctx := context.Background()
	if err = page.Redraw(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *pngOut != "" {
		picture := page.Probe().PNG()
		if len(picture) == 0 {
			fmt.Fprintln(os.Stderr, "png: empty picture")
			os.Exit(1)
		}
		if err = os.WriteFile(*pngOut, picture, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	// This preview does not play sound. Opening the audio device on a
	// machine with no card closes the window before the first paint.
	if err = ownframe.RunWithOptions(ctx, page, ownframe.WindowOptions{NoAudio: true}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// reviewPage is the ten-style gallery in the sibling blinkless checkout.
func reviewPage() string {
	candidates := []string{
		filepath.Join("..", "blinkless", "testdata", "css", "v012", "review", "page.html"),
		filepath.Join("blinkless", "testdata", "css", "v012", "review", "page.html"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return candidates[0]
}

const envSwatchSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="48" viewBox="0 0 96 48">` +
	`<rect width="96" height="48" fill="env(--color, #1d4ed8)"/></svg>`

// swatchPNG is the blue and orange picture used by the background samples.
func swatchPNG() ([]byte, error) {
	const w, h = 48, 24

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			c := color.RGBA{R: 29, G: 78, B: 216, A: 255}
			if x >= w/2 {
				c = color.RGBA{R: 245, G: 158, B: 11, A: 255}
			}
			img.SetRGBA(x, y, c)
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
