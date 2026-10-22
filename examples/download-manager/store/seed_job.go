package store

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/transfer"
)

// dummyJob builds row i. The groups cover the demo states.
func dummyJob(i, jobCount int, now time.Time) domain.Job {
	completed, failed, cancelled, queued := dummyCounts(jobCount)

	var state domain.State

	switch {
	case i < completed:
		state = domain.StateCompleted
	case i < completed+failed:
		state = domain.StateFailed
	case i < completed+failed+cancelled:
		state = domain.StateCancelled
	case i < completed+failed+cancelled+queued:
		state = domain.StateQueued
	default:
		state = domain.StatePaused
	}

	id := fmt.Sprintf("dummy-%04d", i)
	name := dummyNames[i%len(dummyNames)]
	size := int64(seedBaseSize) + int64(i*seedSizeStep)%(seedSizeSpan)
	created := now.Add(-time.Duration(jobCount-i) * seedAgeStepMin * time.Minute)

	job := domain.Job{
		ID:          id,
		URL:         "https://example.invalid/dl/" + name + "?token=dummy",
		Destination: filepath.Join(transfer.DummyDir(), id+"-"+name),
		Name:        name,
		State:       state,
		Expected:    size,
		ETag:        fmt.Sprintf(`"dummy-%d"`, i%seedETagMod),
		CreatedAt:   created,
		UpdatedAt:   created.Add(seedTouchMinutes * time.Minute),
	}

	switch state {
	case domain.StateCompleted:
		job.Done = size
		job.Attempts = 1
	case domain.StateFailed:
		job.Done = size / seedFailFrac
		job.Attempts = 2
		job.Error = dummyErrors[i%len(dummyErrors)]
	case domain.StateCancelled:
		job.Done = size / seedCancelFrac
		job.Attempts = 1
	case domain.StateQueued:
		if i%seedUnknownMod == 0 {
			job.Expected = transfer.Unknown
		}
	case domain.StatePaused:
		job.Done = size / seedHalfFrac
		job.Attempts = 1
	}

	return job
}
