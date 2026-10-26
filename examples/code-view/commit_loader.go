package main

import (
	"context"
	"sync"
)

const (
	// commitPatchQueue is how many requests may wait for the worker. A click
	// handler never blocks on git; a full queue drops the newest request.
	commitPatchQueue = 64
	// commitPatchReady is the finished-result buffer Poll drains.
	commitPatchReady = 64
	// defaultCommitPatchLimit bounds the loader cache like the app-level cache.
	defaultCommitPatchLimit = maxCommitPatchCache
)

// commitPatchRequest names one file patch the history rail wants.
type commitPatchRequest struct {
	Hash string
	Path string
}

// commitPatchResult is one finished patch, or the error that stopped it. Key is
// commitDetailKey(Hash, Path) so the page can match the result to its request.
type commitPatchResult struct {
	Key  string
	File *changedFile
	Err  error
}

// commitPatchJob carries the request context alongside the request. The
// request struct itself stays free of runtime state.
type commitPatchJob struct {
	ctx context.Context
	req commitPatchRequest
}

// commitPatchLoader reads commit patches on one worker goroutine, so a click
// never waits for git. Request queues work, Poll drains finished results on the
// page loop, and the loader keeps the newest capacity patches for cache hits.
type commitPatchLoader struct {
	load     func(ctx context.Context, hash, path string) (*changedFile, error)
	reqs     chan commitPatchJob
	results  chan commitPatchResult
	quit     chan struct{}
	done     chan struct{}
	once     sync.Once
	ctx      context.Context
	cancel   context.CancelFunc
	capacity int
	cache    map[string]*changedFile
	order    []string
}

// NewCommitPatchLoader starts the patch worker for root. A capacity of zero or
// less keeps the same number of patches as the synchronous cache. Close must be
// called to stop the worker.
func NewCommitPatchLoader(root string, capacity int) *commitPatchLoader {
	if capacity <= 0 {
		capacity = defaultCommitPatchLimit
	}

	// The worker outlives one request, so it owns a cancelable context that
	// Close cancels. Every request context is derived from it below.
	ctx, cancel := context.WithCancel(context.Background())

	loader := &commitPatchLoader{
		load: func(ctx context.Context, hash, path string) (*changedFile, error) {
			return loadCommitFile(ctx, root, hash, path)
		},
		reqs:     make(chan commitPatchJob, commitPatchQueue),
		results:  make(chan commitPatchResult, commitPatchReady),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
		capacity: capacity,
		cache:    make(map[string]*changedFile, capacity),
		order:    make([]string, 0, capacity),
	}

	go loader.run()

	return loader
}

// Request queues one patch. It never blocks: a full queue or a closed loader
// drops the request, and the caller can ask again on a later click or tick.
func (l *commitPatchLoader) Request(ctx context.Context, hash, path string) {
	if l == nil || path == "" || ctx == nil {
		return
	}

	select {
	case <-l.quit:
		return
	default:
	}

	select {
	case l.reqs <- commitPatchJob{ctx: ctx, req: commitPatchRequest{Hash: hash, Path: path}}:
	default:
	}
}

// Poll returns one finished patch, or false when none is ready. It never waits.
func (l *commitPatchLoader) Poll() (commitPatchResult, bool) {
	if l == nil {
		return commitPatchResult{}, false
	}

	select {
	case result := <-l.results:
		return result, true
	default:
		return commitPatchResult{}, false
	}
}

// Close cancels the in-flight load, drops queued work, and waits for the worker
// to exit. It is safe to call more than once and after a request.
func (l *commitPatchLoader) Close() {
	if l == nil {
		return
	}

	l.once.Do(func() {
		// Close quit first so the worker drops a result that the cancelled
		// load returns next.
		close(l.quit)
		l.cancel()
	})

	<-l.done

	// No producer remains, so a result finished before the close cannot leak
	// into a Poll after it.
	for {
		select {
		case <-l.results:
		default:
			return
		}
	}
}

// run serializes the blocking patch reads until Close.
func (l *commitPatchLoader) run() {
	defer close(l.done)

	for {
		select {
		case job := <-l.reqs:
			l.loadJob(job)
		case <-l.quit:
			return
		}
	}
}

// loadJob answers from the cache or calls the loader. The caller's context and
// Close both cancel the read, so a shutdown never waits out a git timeout.
func (l *commitPatchLoader) loadJob(job commitPatchJob) {
	key := commitDetailKey(job.req.Hash, job.req.Path)
	if file := l.cache[key]; file != nil {
		l.emit(commitPatchResult{Key: key, File: file})

		return
	}

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	stop := context.AfterFunc(job.ctx, cancel)
	defer stop()

	file, err := l.load(ctx, job.req.Hash, job.req.Path)
	if err != nil {
		l.emit(commitPatchResult{Key: key, Err: err})

		return
	}

	l.store(key, file)
	l.emit(commitPatchResult{Key: key, File: file})
}

// store adds one patch and evicts the oldest once the cache is over capacity.
func (l *commitPatchLoader) store(key string, file *changedFile) {
	if _, seen := l.cache[key]; !seen {
		l.order = append(l.order, key)
	}

	l.cache[key] = file

	for len(l.order) > l.capacity {
		oldest := l.order[0]
		l.order = l.order[1:]
		delete(l.cache, oldest)
	}
}

// emit offers one result without blocking the worker. A result that finishes
// during or after Close is stale and dropped.
func (l *commitPatchLoader) emit(result commitPatchResult) {
	select {
	case <-l.quit:
		return
	default:
	}

	select {
	case l.results <- result:
	default:
	}
}
