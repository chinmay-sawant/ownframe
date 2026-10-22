package store

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/parser"
)

// Ingest builds an Ingestor for a stored source. It reads the committed
// checkpoint so a restart resumes where the last process stopped, and
// rewinds over a stored partial record so the completed line replaces it.
func (s *Store) Ingest(ctx context.Context, src entry.SourceID, pol entry.Policy) (*Ingestor, error) {
	source, err := s.Source(ctx, src)
	if err != nil {
		return nil, err
	}

	if pol == (entry.Policy{}) {
		pol = entry.DefaultPolicy()
	}

	if err := pol.Validate(); err != nil {
		return nil, err
	}

	last, err := s.LastPosition(ctx, src)
	if err != nil {
		return nil, err
	}

	r, err := openReader(source, last, pol)
	if err != nil {
		return nil, err
	}

	in := &Ingestor{
		st: s, src: source, pol: pol, r: r,
		g: parser.NewGrouper(parser.Options{
			MaxLines:   pol.MultilineLines,
			MaxBytes:   pol.MultilineBytes,
			FlushAfter: holdFor(pol),
		}),
	}
	in.stats = IngestStats{
		Source: src, State: source.State,
		Generation: source.Generation, Position: source.Position,
	}

	return in, nil
}

// holdFor is the time a pending primary record waits for a continuation
// line before it is stored on its own.
func holdFor(pol entry.Policy) time.Duration {
	if pol.MultilineHold <= 0 {
		return ingestHoldDef
	}

	return pol.MultilineHold
}
