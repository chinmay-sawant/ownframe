package music

import "context"

// Want starts resolving query in the background. pick chooses among matches,
// and the track starts as soon as it is ready. Timeout caps the resolve.
func (e *Engine) Want(ctx context.Context, query string, pick int) {
	e.mu.Lock()
	e.gen++
	gen := e.gen
	e.loading = true
	e.wantPlay = true
	e.err = nil
	e.mu.Unlock()

	timeout := e.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	go func() {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		clip, err := e.res.Resolve(ctx, query, pick)

		e.latestMu.Lock()
		if gen >= e.latest.Gen {
			e.latest = Result{Gen: gen, Clip: clip, Err: err}
		}
		e.latestMu.Unlock()
	}()
}
