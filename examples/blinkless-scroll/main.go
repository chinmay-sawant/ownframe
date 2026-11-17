// Command blinkless-scroll opens the blinkless scroll fixture in an OwnFrame
// window and plays the host side of CSS scroll-driven timelines and triggers.
// The blinkless engine only parses and exports these values. This program owns
// the scroll offset, the pointer, the clock, and the click and key events, and
// writes the results back into the page as style patches.
//
// Page elements it drives (testdata/css/v012/scroll/page.html in blinkless):
//
//	scroll-bar, scroll-bar-longhand   width follows scroll progress
//	view-card-1, -2, -3               fade and slide in with entry progress
//	pointer-box, pointer-box-longhand fill follows the pointer on the axis
//	trigger-box, trigger-longhand     turn green while the timeline trigger is on
//	                                  (view() = own view timeline, scroll() = page scroll)
//	event-box, event-box-longhand     turn green after a click or key press
//	interest-card, interest-card-longhand
//	                                  turn amber after hover delay passes
//	inert-panel                       takes no hit or focus (auto-panel does)
//
// Flags: -width, -height, -png out.png, -scroll N (offset used by -png),
// -demo and -demo-out DIR (scripted headless session, see demo.go).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// Default window size in CSS pixels, and the exit code for a usage error.
const (
	defaultWidthPx  = 880
	defaultHeightPx = 700
	usageExitCode   = 2
)

// File modes for demo output: picture files are owner-only, and output
// directories are owner-writable and group-readable.
const (
	pngFileMode os.FileMode = 0o600
	dirMode     os.FileMode = 0o750
)

func main() {
	width := flag.Int("width", defaultWidthPx, "window width in CSS pixels")
	height := flag.Int("height", defaultHeightPx, "window height in CSS pixels")
	pngOut := flag.String("png", "", "write one PNG and exit, without opening a window")
	scroll := flag.Int("scroll", 0, "scroll offset in CSS pixels used with -png")
	demo := flag.Bool("demo", false, "play a scripted session headlessly, write PNG frames and an mp4 if ffmpeg exists")
	demoOut := flag.String("demo-out", "", "directory for -demo output (default: a new temp directory)")
	demoWin := flag.Bool("demo-window", false, "open the real window and play the scripted session in it, then write frames and an mp4")
	flag.Parse()

	if flag.NArg() > 1 || *width <= 0 || *height <= 0 || *scroll < 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/blinkless-scroll [-width N] [-height N] [-png out.png] [-scroll N] [-demo [-demo-out DIR]] [page.html]")
		os.Exit(usageExitCode)
	}

	pagePath := scrollPage()
	if flag.NArg() == 1 {
		pagePath = flag.Arg(0)
	}

	if *demoWin {
		if err := runDemoWindow(pagePath, *width, *height, *demoOut); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	if *demo {
		if err := runDemoCommand(pagePath, *width, *height, *demoOut); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	if err := run(pagePath, *width, *height, *pngOut, *scroll); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(pagePath string, width, height int, pngOut string, scroll int) error {
	html, err := os.ReadFile(pagePath) //nolint:gosec // G304: pagePath is the page the operator names on the command line, read on purpose
	if err != nil {
		return err
	}

	ctx := context.Background()

	if pngOut != "" {
		picture, err := renderPNG(ctx, string(html), width, height, scroll)
		if err != nil {
			return err
		}

		return os.WriteFile(pngOut, picture, pngFileMode) //nolint:gosec // G703: pngOut is the -png path the operator gives on the command line, written on purpose
	}

	s, err := newSession(ctx, string(html), width, height, nil)
	if err != nil {
		return err
	}

	s.attachWindow()

	return ownframe.RunWithOptions(ctx, s.page, ownframe.WindowOptions{NoAudio: true})
}

// renderPNG lays the page out, applies the timelines at one scroll offset,
// and returns the picture of the whole page.
func renderPNG(ctx context.Context, html string, width, height, scroll int) ([]byte, error) {
	s, err := newSession(ctx, html, width, height, nil)
	if err != nil {
		return nil, err
	}

	s.page.Scroll().SetScrollOffset(0, scroll)

	if _, err = s.frame(ctx, FrameInput{ScrollY: float64(scroll)}); err != nil {
		return nil, err
	}

	return pngOf(ctx, s)
}

func pngOf(ctx context.Context, s *session) ([]byte, error) {
	picture := s.page.Testing().PNGWithContext(ctx)
	if len(picture) == 0 {
		return nil, fmt.Errorf("png: empty picture")
	}

	return picture, nil
}

// newSession loads the page and builds the host. clock may be nil for the
// wall clock; the demo passes a manual one.
func newSession(ctx context.Context, html string, width, height int, clock Clock) (*session, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "blinkless scroll",
		HTML:   html,
		Width:  width,
		Height: height,
		// Style patches need the retained layout state.
		RetainState: true,
	})
	if err != nil {
		return nil, err
	}

	if err = page.Redraw(ctx); err != nil {
		return nil, err
	}

	display := page.Probe().Display()
	if display == nil {
		return nil, fmt.Errorf("page has no display list")
	}

	model := Build(display)
	host := NewHost(model.Nodes, model.Triggers, model.Events, model.Delays, clock)

	return &session{
		page:    page,
		host:    host,
		applier: NewApplier(BaseStyles(html)),
		html:    html,
		width:   width,
		height:  height,
	}, nil
}

// scrollPage is the fixture in the sibling blinkless checkout.
func scrollPage() string {
	candidates := []string{
		filepath.Join("..", "blinkless", "testdata", "css", "v012", "scroll", "page.html"),
		filepath.Join("blinkless", "testdata", "css", "v012", "scroll", "page.html"),
		filepath.Join("..", "..", "..", "blinkless", "testdata", "css", "v012", "scroll", "page.html"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return candidates[0]
}

// runDemoCommand plays the demo, prints one log line per step, and reports
// where the frames and the video went.
func runDemoCommand(pagePath string, width, height int, dir string) error {
	html, err := os.ReadFile(pagePath) //nolint:gosec // G304: pagePath is the page the operator names on the command line, read on purpose
	if err != nil {
		return err
	}

	if dir == "" {
		if dir, err = os.MkdirTemp("", "blinkless-scroll-demo-"); err != nil {
			return err
		}
	} else if err = os.MkdirAll(dir, dirMode); err != nil {
		return err
	}

	fmt.Println("demo output:", dir)

	logs, err := runDemo(context.Background(), string(html), width, height, dir, true, os.Stdout)
	if err != nil {
		return err
	}

	fmt.Printf("%d frames written to %s\n", len(logs), dir)

	mp4, why := stitch(context.Background(), dir)
	if mp4 == "" {
		fmt.Println(why)

		return nil
	}

	fmt.Println("video:", mp4)

	return nil
}
