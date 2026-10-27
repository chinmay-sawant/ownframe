package music

import "time"

// Position is the current voice position, or zero. The voice pointer is
// copied under lock; Position itself is thread-safe.
func (e *Engine) Position() time.Duration {
	e.mu.Lock()
	voice := e.voice
	e.mu.Unlock()

	if voice == nil {
		return 0
	}

	return voice.Position()
}

// Duration is the current voice length, or zero.
func (e *Engine) Duration() time.Duration {
	e.mu.Lock()
	voice := e.voice
	e.mu.Unlock()

	if voice == nil {
		return 0
	}

	return voice.Duration()
}

// Credit is the free clip behind the current voice.
func (e *Engine) Credit() Clip {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.clip
}

// Err is the last resolve or decode failure.
func (e *Engine) Err() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.err
}

// Close releases the current voice.
func (e *Engine) Close() {
	e.mu.Lock()
	voice := e.voice
	e.voice = nil
	e.mu.Unlock()

	if voice != nil {
		_ = voice.Close()
	}

	e.latestMu.Lock()
	if e.latest.Voice != nil {
		_ = e.latest.Voice.Close()
		e.latest.Voice = nil
	}
	e.latestMu.Unlock()
}
