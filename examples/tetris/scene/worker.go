package scene

import (
	"context"
	"errors"
	"sync"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
	"github.com/chinmay-sawant/ownframe/examples/tetris/store"
)

// ErrWorkerBusy reports a request refused because the worker queue is
// full. The scene treats it as a failed request.
var ErrWorkerBusy = errors.New("scene: worker queue is full")

// worker bridges the scene to the core store: one goroutine serializes
// the blocking calls, a bounded request queue feeds it, and a bounded
// result queue feeds the tick. The tick never waits on SQL.
type worker struct {
	st   *store.Store
	reqs chan request
	res  chan Result
	quit chan struct{}
	done chan struct{}

	base   context.Context
	cancel context.CancelFunc
	once   sync.Once
	seq    uint64

	set store.Settings
}

// request is one job for the worker goroutine.
type request struct {
	kind  ResultKind
	id    uint64
	page  int
	dark  bool
	clear bool
	res   game.Result
	rep   *game.Replay
	snap  game.Snapshot
}

// NewWorker starts the storage bridge for one open store.
func NewWorker(st *store.Store) Store {
	base, cancel := context.WithCancel(context.Background())
	w := &worker{
		st:     st,
		reqs:   make(chan request, workerQueue),
		res:    make(chan Result, workerResults),
		quit:   make(chan struct{}),
		done:   make(chan struct{}),
		base:   base,
		cancel: cancel,
		set:    store.DefaultSettings(),
	}

	go w.run()

	return w
}

// Poll returns one finished result, or false when none is ready.
func (w *worker) Poll() (Result, bool) {
	select {
	case r := <-w.res:
		return r, true
	default:
		return Result{}, false
	}
}

// Close cancels outstanding work, joins the worker, and closes the store.
func (w *worker) Close() error {
	w.once.Do(func() {
		w.cancel()
		close(w.quit)
	})

	<-w.done

	return w.st.Close()
}
