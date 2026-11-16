// Command blinkless-all opens one window that shows, tab by tab, the CSS
// properties blinkless had not yet implemented: navigation and UI, draft
// borders and shapes, motion path, fragmentation and generated content,
// reading order, and the SVG leftovers. The top of the window states the
// catalog counts. Clicking a tab loads a page that holds only that tab.
//
// Flags: -width, -height, -tab ID, -png out.png (one picture, no window).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	width := flag.Int("width", defaultWidth, "window width in CSS pixels")
	height := flag.Int("height", defaultHeight, "window height in CSS pixels")
	tab := flag.String("tab", "", "tab to open (default: the first)")
	pngOut := flag.String("png", "", "write one PNG and exit, without opening a window")
	demo := flag.Bool("demo", false, "play the scripted live session of the tab with a manual clock and no window")
	demoOut := flag.String("demo-out", "", "directory for the -demo frames (default: a temp directory)")
	flag.Parse()

	if *demo {
		if err := runDemoCommand(*tab, *width, *height, *demoOut); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	if err := run(*width, *height, *tab, *pngOut); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(width, height int, tab, pngOut string) error {
	app, err := NewApp(catalogPath())
	if err != nil {
		return err
	}

	if tab == "" {
		tab = app.FirstTab()
	}

	ctx := context.Background()

	page, err := app.NewPage(ctx, tab, width, height)
	if err != nil {
		return err
	}

	if pngOut != "" {
		picture := page.Probe().PNG()
		if len(picture) == 0 {
			return fmt.Errorf("png: empty picture")
		}

		return os.WriteFile(pngOut, picture, pngMode)
	}

	attachLive(page)

	return ownframe.RunWithOptions(ctx, page, ownframe.WindowOptions{NoAudio: true})
}

// runDemoCommand plays the scripted session of a tab headless and writes its
// frames. Only the nav tab has a live session so far.
func runDemoCommand(tab string, width, height int, dir string) error {
	if tab != "" && tab != "nav" {
		return fmt.Errorf("-demo: only the nav tab has a live session, not %q", tab)
	}

	app, err := NewApp(catalogPath())
	if err != nil {
		return err
	}

	if dir == "" {
		if dir, err = os.MkdirTemp("", "blinkless-all-demo-"); err != nil {
			return err
		}
	} else if err = os.MkdirAll(dir, demoDirMode); err != nil {
		return err
	}

	fmt.Println("demo output:", dir)

	n, err := runNavDemo(context.Background(), app, width, height, dir, os.Stdout)
	if err != nil {
		return err
	}

	fmt.Printf("%d frames written to %s\n", n, dir)

	return nil
}

// NewPage builds the window page with tab open and routes every tab button.
func (a *App) NewPage(ctx context.Context, tab string, width, height int) (*ownframe.Page, error) {
	page, err := ownframe.New(ownframe.Config{Title: "blinkless all", HTML: a.Page(tab), Width: width, Height: height, RetainState: true})
	if err != nil {
		return nil, err
	}

	for _, t := range a.tabs {
		if err = page.Route(actionPrefix+t.ID, a.Page(t.ID)); err != nil {
			return nil, err
		}
	}

	if err = page.Redraw(ctx); err != nil {
		return nil, err
	}

	return page, nil
}

// Default window size, in CSS pixels.
const (
	defaultWidth  = 1000
	defaultHeight = 800
)

// pngMode is the file mode of the written picture.
const pngMode = 0o644
