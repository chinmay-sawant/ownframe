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
		var voice Voice
		if err == nil {
			voice, err = e.Open(clip.Data)
		}

		e.mu.Lock()
		cancelled := gen < e.gen
		e.mu.Unlock()
		if cancelled {
			if voice != nil {
				_ = voice.Close()
			}

			return
		}

		e.latestMu.Lock()
		if gen < e.latest.Gen {
			e.latestMu.Unlock()
			if voice != nil {
				_ = voice.Close()
			}

			return
		}
		if e.latest.Voice != nil {
			_ = e.latest.Voice.Close()
		}
		e.latest = Result{Gen: gen, Clip: clip, Voice: voice, Err: err}
		e.latestMu.Unlock()
	}()
}
