package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/chinmay-sawant/ownframe"
)

// walkActive reports the env-gated automation used by
// scripts/css-preview-click-test.py. It keeps the walker's tick registered
// and exposes the next Start button's client point over HTTP. A false value
// in CSS_PREVIEW_WALK turns the harness off, so an exported variable that
// carries 0 does not start the server.
func walkActive() bool {
	v := strings.TrimSpace(os.Getenv("CSS_PREVIEW_WALK"))
	if v == "" {
		return false
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}

	return true
}

type walker struct {
	p         *ownframe.Page
	d         *demo
	v         *view
	mu        sync.Mutex
	next      int
	scrollFor int
	aim       [3]int // index, client x, client y; index -1 when unset
	trace     walkTrace
	resetReq  bool
	resetDone chan struct{}
	stopReq   bool
	m         *catalogMetrics
	// boxes is the hit-test box snapshot the game loop published. The
	// HTTP /box route reads it, so the route never reads the page while
	// the loop swaps the boxes.
	boxes []ownframe.Box
	// boxGen is the page generation the snapshot came from. The page
	// bumps it on every redraw and on every element patch, which are the
	// two paths that replace the boxes.
	boxGen uint64
}

func startWalk(p *ownframe.Page, d *demo, v *view, m *catalogMetrics) {
	if !walkActive() {
		return
	}
	w := &walker{p: p, d: d, v: v, m: m, scrollFor: -1, aim: [3]int{-1, 0, 0}}
	p.Testing().SetDevTools(true)
	go w.serve()

	keep := os.Getenv("CSS_PREVIEW_DEVTOOLS") != ""
	closed := false
	var tick func(context.Context) error
	tick = func(ctx context.Context) error {
		if !closed && !keep {
			closed = true
			p.Testing().SetDevTools(false)
		}
		if err := d.tick(ctx); err != nil {
			return err
		}
		w.step(ctx)
		p.SetTick(tick)
		return nil
	}
	p.SetTick(tick)
}
