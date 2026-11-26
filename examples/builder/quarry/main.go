// Command quarry is a local-first SQLite and PostgreSQL database client built
// with the ownframe builder UI. It renders through blinkless with no browser.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chinmay-sawant/ownframe"

	quarryui "github.com/chinmay-sawant/ownframe/examples/builder/quarry/ui"
)

func main() {
	web := flag.Bool("web", false, "serve the UI in a browser preview instead of opening a window")
	headless := flag.Bool("headless", false, "serve without opening a window (same as -web)")
	addr := flag.String("addr", "127.0.0.1:8150", "address for -web or -headless")
	db := flag.String("db", "", "SQLite file path, postgres:// DSN, or mock:postgres (default: the seeded demo database)")
	dataDir := flag.String("data-dir", "", "override the config directory (state.db and demo database)")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	snapshot := flag.String("snapshot", "", "window mode only: also serve a picture of the window as it is now at this loopback address (for example 127.0.0.1:8151); GET /frame.png returns it")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (needs -tags ownframe_debug)")
	flag.Parse()
	if err := run(*web || *headless, *addr, *db, *dataDir, *dev, *perf, *debugAddr, *snapshot); err != nil {
		log.Fatal(err)
	}
}

func run(web bool, addr, db, dataDir string, dev, perf bool, debugAddr, snapshot string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := quarryui.NewApp(ctx, quarryui.LaunchOptions{DBPath: db, DataDir: dataDir, DemoDir: dataDir, Perf: perf})
	if err != nil {
		return err
	}
	defer a.Close()
	a.Page().Window().SetDevTools(dev)
	if err := a.Page().Redraw(ctx); err != nil {
		return err
	}
	if web {
		if snapshot != "" {
			return fmt.Errorf("-snapshot is for the window; with -web the page itself is served")
		}
		return ownframe.Serve(ctx, a.Page(), addr)
	}
	if snapshot != "" {
		hub := &snapshotHub{}
		hub.wake = func() { a.Page().WantFrameAt(time.Now()) }
		a.SetFrameHook(func(ctx context.Context) {
			hub.answer(func() []byte { return a.Page().Probe().PNGWithContext(ctx) })
		})
		go func() {
			if err := serveSnapshots(ctx, snapshot, hub); err != nil {
				fmt.Fprintln(os.Stderr, "quarry:", err)
			}
		}()
		fmt.Fprintf(os.Stderr, "quarry: window snapshot at http://%s/frame.png\n", snapshot)
	}
	// The page draws its own title bar. The drag region yields to any open
	// modal, popover, or menu, and to the window controls on the right.
	return ownframe.RunWithOptions(ctx, a.Page(), ownframe.WindowOptions{
		CustomHeader: true,
		PlacementID:  "quarry",
		DebugAddr:    debugAddr,
		Draggable:    a.Draggable,
	})
}
