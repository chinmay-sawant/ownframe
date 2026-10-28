package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Scheduler runs every service on one interval and records the results.
type Scheduler struct {
	services []Service
	interval time.Duration
	prober   *Prober
	store    *Store
	log      *slog.Logger
}

// NewScheduler builds the loop. Run owns the goroutines.
func NewScheduler(services []Service, interval time.Duration, prober *Prober, store *Store, log *slog.Logger) *Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	return &Scheduler{
		services: services,
		interval: interval,
		prober:   prober,
		store:    store,
		log:      log,
	}
}

// Run probes once at startup, then on every tick, and stops with ctx.
func (s *Scheduler) Run(ctx context.Context) {
	s.round(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")

			return
		case <-ticker.C:
			s.round(ctx)
		}
	}
}

// round probes every service with bounded concurrency.
func (s *Scheduler) round(ctx context.Context) {
	const workers = 4

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for _, svc := range s.services {
		wg.Add(1)

		go func(svc Service) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.prober.Probe(ctx, svc)
			if err := s.store.Record(ctx, result); err != nil {
				s.log.Warn("record", "service", svc.Name, "err", err)
			}
		}(svc)
	}

	wg.Wait()
}
