package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// live is the in-memory mirror of one job.
type live struct {
	job        domain.Job
	cancel     context.CancelFunc
	queued     bool
	running    bool
	pausing    bool
	cancelling bool
}

// Engine is the bounded worker pool.
type Engine struct {
	opts     Options
	out      *Outbox
	mu       sync.Mutex
	jobs     map[string]*live
	queue    []string
	reserved map[string]bool
	wake     chan struct{}
	wg       sync.WaitGroup

	cancel  context.CancelFunc
	started bool
	closed  bool
}

// New builds an engine. Transport and Store are required.
func New(opts Options) (*Engine, error) {
	if opts.Transport == nil || opts.Store == nil {
		return nil, errors.New("scheduler: Transport and Store are required")
	}

	if opts.Workers <= 0 {
		opts.Workers = 3
	}

	if opts.Queue <= 0 {
		opts.Queue = 64
	}

	if opts.Now == nil {
		opts.Now = time.Now
	}

	if opts.ProgressEvery <= 0 {
		opts.ProgressEvery = defaultProgressEvery
	}

	if opts.ShutdownBudget <= 0 {
		opts.ShutdownBudget = defaultShutdownBudget
	}

	return &Engine{
		opts:     opts,
		out:      NewOutbox(),
		jobs:     map[string]*live{},
		reserved: map[string]bool{},
		wake:     make(chan struct{}, opts.Workers),
	}, nil
}
