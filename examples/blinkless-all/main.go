// Command blinkless-all opens one window that shows, tab by tab, every CSS
// property blinkless implements. Six hand-built tabs hold the 61 once-missing
// ones: navigation and UI, draft borders and shapes, motion path,
// fragmentation and generated content, reading order, and the SVG leftovers.
// Generated tabs (ids "all-...", at most 20 rows each) hold the rest, each
// with a description, a how-to-validate line and a control-versus-sample
// readout. The top of the window states the catalog counts. Clicking a tab loads a page that holds only that tab.
//
// Flags: -width, -height, -tab ID, -png out.png (one picture, no window).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	width := flag.Int("width", defaultWidth, "window width in CSS pixels")
	height := flag.Int("height", defaultHeight, "window height in CSS pixels")
	tab := flag.String("tab", "", "tab to open (default: the first)")
	pngOut := flag.String("png", "", "write one PNG and exit, without opening a window")
	demo := flag.Bool("demo", false, "play the scripted live session of the tab with a manual clock and no window")
	demoOut := flag.String("demo-out", "", "directory for the -demo frames (default: a temp directory)")
	web := flag.String("web", "", "serve the page at this address (for example 127.0.0.1:8099) instead of opening a window; GET /frame.png returns the picture")
	snapshot := flag.String("snapshot", "", "window mode only: also serve a picture of the window as it is now at this loopback address (for example 127.0.0.1:8099); GET /frame.png returns it and / shows it")
	perf := flag.Bool("perf", false, "enable performance counters (and /debug/state + /debug/pprof/* in -web mode)")
	flag.Parse()

	if *demo {
		if err := runDemoCommand(*tab, *width, *height, *demoOut); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	if err := run(*width, *height, *tab, *pngOut, *web, *snapshot, *perf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(width, height int, tab, pngOut, web, snapshot string, perf bool) error {
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
	if perf {
		page.Window().SetPerf(true)
	}

	if pngOut != "" {
		picture := page.Probe().PNG()
		if len(picture) == 0 {
			return fmt.Errorf("png: empty picture")
		}

		return os.WriteFile(pngOut, picture, pngMode)
	}

	if web != "" {
		if snapshot != "" {
			return fmt.Errorf("-snapshot is for the window; with -web the page itself is served")
		}

		attachWebLive(app, page)

		if perf {
			return ownframe.ServeWithOptions(ctx, page, web, ownframe.ServeOptions{
				InsecureAllowDebugState: true,
				InsecureAllowPprof:      true,
			})
		}

		return ownframe.Serve(ctx, page, web)
	}

	var hub *snapshotHub

	if snapshot != "" {
		hub = newSnapshotHub()

		go func() {
			if serveErr := serveSnapshots(ctx, snapshot, hub); serveErr != nil {
				fmt.Fprintln(os.Stderr, "blinkless-all:", serveErr)
			}
		}()

		fmt.Fprintf(os.Stderr, "blinkless-all: window snapshot at http://%s/frame.png\n", snapshot)
	}

	attachLive(page, hub)

	return ownframe.RunWithOptions(ctx, page, ownframe.WindowOptions{NoAudio: true, Perf: perf})
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

// NewPage builds the window page with tab open and handles tab button clicks
// lazily so unvisited tabs are not laid out at startup.
func (a *App) NewPage(ctx context.Context, tab string, width, height int) (*ownframe.Page, error) {
	if _, ok := a.findTab(tab); !ok {
		return nil, fmt.Errorf("unknown tab %q", tab)
	}

	page, err := ownframe.New(ownframe.Config{Title: "blinkless all", HTML: a.Page(tab), Width: width, Height: height, RetainState: true})
	if err != nil {
		return nil, err
	}

	// The shape-image-threshold demo names url(mask); the readout layout
	// answers it through shapesMask, so the painted page must answer it too.
	if mask, maskErr := shapesMask("mask"); maskErr == nil {
		page.SetImage("mask", mask)
	}

	page.AddHandlers(ownframe.Handlers{
		Wheel: sideWheel(page),
		Click: a.tabClick(page),
	})

	if err = page.Redraw(ctx); err != nil {
		return nil, err
	}

	return page, nil
}

func (a *App) tabClick(page *ownframe.Page) func(context.Context, ownframe.Box) error {
	return func(ctx context.Context, b ownframe.Box) error {
		id, ok := strings.CutPrefix(b.Action, actionPrefix)
		if !ok {
			return nil
		}

		if _, found := a.findTab(id); !found {
			return nil
		}

		return page.Load(ctx, a.Page(id))
	}
}

// Default window size, in CSS pixels.
const (
	defaultWidth  = 1000
	defaultHeight = 800
)

// pngMode is the file mode of the written picture.
const pngMode = 0o644
