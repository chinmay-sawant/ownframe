// Package benchutil shares the three perf benchmark demos: one normal
// desktop form, one large scrolling list, and one flappy-bird-like tick
// animation. The runnable mains stay thin and the Go benchmarks measure
// Redraw and tick cost through these apps.
package benchutil

import (
	"context"
	"flag"

	"github.com/chinmay-sawant/ownframe"
)

// Flags registers -web and -addr and returns them for a benchmark demo.
func Flags(defaultAddr string) (web *bool, addr *string) {
	web = flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr = flag.String("addr", defaultAddr, "listen address for -web")

	return web, addr
}

// ServeOrRun serves the page in a browser with -web or opens the window.
// DevTools stays off unless the app turns it on. A page with Perf on keeps
// its sampling in a window and gains /debug/pprof/* when served.
func ServeOrRun(ctx context.Context, page *ownframe.Page, web bool, addr string) error {
	if web {
		return ownframe.ServeWithOptions(ctx, page, addr, ownframe.ServeOptions{InsecureAllowPprof: page.Testing().Perf(), InsecureAllowDebugState: page.Testing().Perf()})
	}

	return ownframe.Run(ctx, page)
}
