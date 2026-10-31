package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// MeasureScenario runs a scenario and measures cold render, warm redraw, mutation redraw,
// and resize relayout, returning a detailed Result.
func MeasureScenario(s Scenario, warmIterations int) (Result, error) {
	ctx := context.Background()

	// 1. Cold Render: New page + SetData + first Redraw.
	coldStart := time.Now()
	page, err := ownframe.New(ownframe.Config{
		Title:  s.Name,
		HTML:   s.HTML,
		Width:  s.Width,
		Height: s.Height,
		Perf:   true,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to create page for %s: %w", s.ID, err)
	}

	page.SetData(s.Data)

	if err := page.Redraw(ctx); err != nil {
		return Result{}, fmt.Errorf("initial Redraw failed for %s: %w", s.ID, err)
	}
	coldDuration := time.Since(coldStart)

	stats := page.Probe().Stats()
	boxes := len(page.Boxes())
	ops := stats.Ops

	// 2. Warm Redraw: repeated Redraw on clean state.
	if warmIterations < 1 {
		warmIterations = 5
	}
	var warmDurations []time.Duration
	for i := 0; i < warmIterations; i++ {
		t0 := time.Now()
		if err := page.Redraw(ctx); err != nil {
			return Result{}, fmt.Errorf("warm Redraw failed for %s: %w", s.ID, err)
		}
		warmDurations = append(warmDurations, time.Since(t0))
	}
	avgWarm := averageDuration(warmDurations)

	// 3. Mutation Redraw: data modified then redrawn.
	var mutDuration time.Duration
	if s.Mutate != nil {
		mutatedData := s.Mutate(s.Data)
		page.SetData(mutatedData)
		t0 := time.Now()
		if err := page.Redraw(ctx); err != nil {
			return Result{}, fmt.Errorf("mutation Redraw failed for %s: %w", s.ID, err)
		}
		mutDuration = time.Since(t0)
	}

	// 4. Resize Relayout: changing window dimension triggers relayout.
	page.SetSize(s.Width+1, s.Height)
	t0 := time.Now()
	if err := page.Redraw(ctx); err != nil {
		return Result{}, fmt.Errorf("resize Redraw failed for %s: %w", s.ID, err)
	}
	resizeDuration := time.Since(t0)

	return Result{
		ID:             s.ID,
		Name:           s.Name,
		Level:          s.Level,
		ItemCount:      s.ItemCount,
		Boxes:          boxes,
		Ops:            ops,
		ColdRender:     coldDuration,
		WarmRedraw:     avgWarm,
		MutationRedraw: mutDuration,
		ResizeRelayout: resizeDuration,
		TemplateTime:   stats.LastTemplate,
		StyleTime:      stats.StyleTime,
		LayoutTime:     stats.LayoutTime,
		PaintTime:      stats.PaintTime,
		AllocFrame:     stats.AllocFrame,
	}, nil
}

func averageDuration(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range durs {
		total += d
	}
	return total / time.Duration(len(durs))
}
