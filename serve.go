package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/web"
)

// Serve listens on addr and blocks. Cancel ends the listener.
// The page at / shows the latest picture. Clicks and the type form call the
// same handlers as Run. Serve draws the page first when it has not been drawn yet.
// Developer endpoints stay off; ServeWithOptions opts into them.
// A Page has one owner: never share it between Run and Serve. Draw and tick
// stay on the page loop.
func Serve(ctx context.Context, page *Page, addr string) error {
	if err := prepare(ctx, page); err != nil {
		return err
	}

	return web.ServeWithOptionsContext(ctx, page, addr, web.Options{})
}

// ServeOptions tunes ServeWithOptions. It re-exports the web options, so
// callers stay on package ownframe. The zero value is safe: no developer
// endpoints and loopback binds only. Risky opt-ins use an InsecureAllow
// prefix; see web.Options.
type ServeOptions = web.Options

// ServeWithOptions listens on addr like Serve, with optional developer
// endpoints. ServeOptions{InsecureAllowPprof: true} serves /debug/pprof/*,
// and ServeOptions{InsecureAllowDebugState: true} serves /debug/state.
// A non-loopback addr needs ServeOptions{InsecureAllowRemoteBind: true}.
// Cancel ends the listener.
func ServeWithOptions(ctx context.Context, page *Page, addr string, opts ServeOptions) error {
	if err := prepare(ctx, page); err != nil {
		return err
	}

	return web.ServeWithOptionsContext(ctx, page, addr, opts)
}
