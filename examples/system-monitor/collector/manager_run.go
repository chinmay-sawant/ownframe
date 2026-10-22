package collector

import (
	"context"
	"time"
)

// loopCount is the summary, process, and detail loops Start launches.
const loopCount = 3

// Start launches the summary, process, and detail loops and returns.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()

		return ErrStarted
	}
	if m.closed {
		m.mu.Unlock()

		return ErrClosed
	}
	m.started = true
	m.start = time.Now()

	run, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	if m.mode == ModeDummy {
		m.prefillLocked()
	}
	m.mu.Unlock()

	m.wg.Add(loopCount)
	go m.loop(run, m.opts.SummaryInterval, m.collectSummary)
	go m.loop(run, m.opts.ProcessInterval, m.collectProcesses)
	go m.loop(run, m.opts.DetailInterval, m.collectDetail)

	return nil
}

// Close cancels the loops and waits for them and for calls in flight. It
// returns ctx.Err() when the wait outruns ctx; the pool still closes in the
// background when the last call returns.
func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()

		return nil
	}
	m.closed = true
	cancel := m.cancel
	m.cancel = nil
	started := m.started
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})

	go func() {
		if started {
			m.wg.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		m.pool.Close()

		return nil
	case <-ctx.Done():
		go func() {
			<-done
			m.pool.Close()
		}()

		return ctx.Err()
	}
}

// Mode returns the active source mode.
func (m *Manager) Mode() Mode {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.mode
}

// loop ticks every interval until ctx is done.
func (m *Manager) loop(ctx context.Context, every time.Duration, collect func(context.Context, uint64)) {
	defer m.wg.Done()

	t := time.NewTicker(every)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			collect(ctx, m.generation())
		}
	}
}
