package entry

import (
	"errors"
	"time"
)

// Overflow picks what happens when a source produces records faster than the
// store consumes them. A replayable file may lag with no loss; a live stream
// must report every skipped record.
type Overflow int

const (
	// OverflowPause lets a replayable source lag behind; nothing is lost.
	OverflowPause Overflow = iota
	// OverflowCountLoss keeps reading a live source and reports each skipped
	// record through reader.Batch.Lost and store.IngestStats.Lost.
	OverflowCountLoss
)

// Policy bounds every piece of ingestion work. The zero value is invalid;
// call DefaultPolicy and adjust fields.
type Policy struct {
	BatchRecords   int
	BatchBytes     int
	MaxRecord      int
	Poll           time.Duration
	MultilineLines int
	MultilineBytes int
	MultilineHold  time.Duration
	Overflow       Overflow
	QueryTimeout   time.Duration
	PruneBatch     int
	PageLimit      int
}

// DefaultPolicy returns the shipped limits.
func DefaultPolicy() Policy {
	return Policy{
		BatchRecords:   policyBatchRecords,
		BatchBytes:     policyBatchBytes,
		MaxRecord:      policyMaxRecord,
		Poll:           policyPoll,
		MultilineLines: policyMultiLines,
		MultilineBytes: policyMultiBytes,
		MultilineHold:  policyMultiHold,
		Overflow:       OverflowPause,
		QueryTimeout:   policyQueryTimeout,
		PruneBatch:     policyPruneBatch,
		PageLimit:      policyPageLimit,
	}
}

// Validate rejects limits that would break bounded ingestion.
func (p Policy) Validate() error {
	if p.BatchRecords < 1 || p.BatchBytes < 1 || p.MaxRecord < 1 {
		return errors.New("entry: batch and record limits must be positive")
	}

	if p.Poll < 0 || p.MultilineLines < 1 || p.MultilineBytes < 1 || p.MultilineHold < 0 {
		return errors.New("entry: poll and multiline limits are invalid")
	}

	if p.Overflow != OverflowPause && p.Overflow != OverflowCountLoss {
		return errors.New("entry: unknown overflow mode")
	}

	return nil
}
