package store

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/reader"
)

func (in *Ingestor) note(b reader.Batch, entries []entry.Entry) {
	in.mu.Lock()
	defer in.mu.Unlock()

	in.stats.Batches++
	in.stats.Records += int64(len(entries))

	for _, e := range entries {
		in.stats.Bytes += int64(e.Bytes)
	}

	in.stats.Lost += b.Lost
	in.stats.Lag = b.Lag
	in.stats.State = b.State
	in.stats.Generation = b.Generation
	in.stats.Position = b.Position
	in.stats.Done = b.Done
}

func (in *Ingestor) failed(err error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	in.stats.Errors++
	in.stats.LastError = err.Error()
}

func (in *Ingestor) poll() time.Duration {
	if in.pol.Poll <= 0 {
		return ingestPollDef
	}

	return in.pol.Poll
}

func (in *Ingestor) nap(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
