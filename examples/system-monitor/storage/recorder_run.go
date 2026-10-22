package storage

import "context"

// run drains the queue on one goroutine, which is the serialized database
// worker for the recording path.
func (r *Recorder) run() {
	defer close(r.done)

	for rows := range r.queue {
		ctx, cancel := context.WithTimeout(context.Background(), defaultOpTimeout)
		n, err := r.store.Append(ctx, r.sessionID, rows)
		cancel()

		r.mu.Lock()
		if err != nil {
			r.failed++
			r.lastError = err.Error()
		} else {
			r.written += int64(n)
		}
		r.mu.Unlock()
	}
}
