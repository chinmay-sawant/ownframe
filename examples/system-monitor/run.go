//go:build !js

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/collector"
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/storage"
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/ui"
)

// run opens storage, starts the collector, and blocks in the window.
func run(data, mode string, stress bool, seed int64) error {
	if mode != "dummy" && mode != "live" {
		return fmt.Errorf("system-monitor: -mode must be dummy or live, got %q", mode)
	}

	live := mode == "live"
	ctx := context.Background()

	dir := data
	if dir == "" {
		var err error

		dir, err = storage.DefaultDir()
		if err != nil {
			return err
		}
	}

	st, err := storage.Open(dir)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := st.Close(); cerr != nil {
			log.Printf("system-monitor: store close: %v", cerr)
		}
	}()

	// Seed the labeled fixture once; a reopen never duplicates it.
	if !live {
		fx := collector.DummyFixture(seed, time.Now(),
			collector.DefaultHistory, collector.DefaultHistoryStep)
		if _, err := st.Seed(ctx, fx); err != nil {
			log.Printf("system-monitor: dummy seeding skipped: %v", err)
		}
	}

	mgr := collector.New(collector.Options{Mode: sourceMode(live), Seed: seed, Stress: stress})
	if err := mgr.Start(ctx); err != nil {
		return err
	}

	app, err := ui.New(ctx, ui.Config{
		Source: ui.NewCollector(mgr),
		Store:  ui.NewCollectorStore(st),
		Live:   live,
	})
	if err != nil {
		_ = mgr.Close(context.Background())

		return err
	}

	defer func() {
		if cerr := app.Close(); cerr != nil {
			log.Printf("system-monitor: close: %v", cerr)
		}
	}()

	return ownframe.Run(ctx, app.Page())
}

// sourceMode maps the flag to the collector mode.
func sourceMode(live bool) collector.Mode {
	if live {
		return collector.ModeLive
	}

	return collector.ModeDummy
}
