package main

import (
	"context"
	"os"
	"sync"

	"github.com/chinmay-sawant/ownframe"
)

// walkActive reports the env-gated automation used by
// scripts/css-preview-click-test.py. It keeps the walker's tick registered
// and exposes the next Start button's client point over HTTP.
func walkActive() bool { return os.Getenv("CSS_PREVIEW_WALK") != "" }

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
}

func startWalk(p *ownframe.Page, d *demo, v *view, m *catalogMetrics) {
	if !walkActive() {
		return
	}
	w := &walker{p: p, d: d, v: v, m: m, scrollFor: -1, aim: [3]int{-1, 0, 0}}
	p.SetDevTools(true)
	go w.serve()

	keep := os.Getenv("CSS_PREVIEW_DEVTOOLS") != ""
	closed := false
	var tick func(context.Context) error
	tick = func(ctx context.Context) error {
		if !closed && !keep {
			closed = true
			p.SetDevTools(false)
		}
		if err := d.tick(ctx); err != nil {
			return err
		}
		w.step()
		p.SetTick(tick)
		return nil
	}
	p.SetTick(tick)
}
