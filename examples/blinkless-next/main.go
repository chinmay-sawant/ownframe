// Command blinkless-next opens one blinkless HTML fixture in the OwnFrame window.
// The fixture stays in the blinkless repo. This program reads that file and
// paints it with the blinkless checkout named in the parent go.work.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	width := flag.Int("width", 800, "window width in CSS pixels")
	height := flag.Int("height", 600, "window height in CSS pixels")
	pngOut := flag.String("png", "", "write one PNG and exit, without opening a window")
	flag.Parse()
	if flag.NArg() != 1 || *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/blinkless-next [-width N] [-height N] [-png out.png] page.html")
		os.Exit(2)
	}

	html, err := os.ReadFile(flag.Arg(0))
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

	ctx := context.Background()
	if err = page.Redraw(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *pngOut != "" {
		bytes := page.Probe().PNG()
		if len(bytes) == 0 {
			fmt.Fprintln(os.Stderr, "png: empty picture")
			os.Exit(1)
		}
		if err = os.WriteFile(*pngOut, bytes, 0o644); err != nil {
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
