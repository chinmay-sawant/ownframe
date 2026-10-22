package music

import "time"

// replace swaps the voice.
func (e *Engine) replace(voice Voice, clip Clip) {
	e.mu.Lock()
	old := e.voice
	vol := e.volume
	want := e.wantPlay
	e.voice = voice
	e.clip = clip
	e.mu.Unlock()

	if old != nil {
		_ = old.Close()
	}

	voice.SetVolume(vol)
	if want {
		voice.Play()
	}
}

// Play resumes the voice.
func (e *Engine) Play() {
	e.mu.Lock()
	e.wantPlay = true
	voice := e.voice
	e.mu.Unlock()

	if voice != nil {
		voice.Play()
	}
}

// Pause holds the voice.
func (e *Engine) Pause() {
	e.mu.Lock()
	e.wantPlay = false
	voice := e.voice
	e.mu.Unlock()

	if voice != nil {
		voice.Pause()
	}
}

// Toggle flips play state.
func (e *Engine) Toggle() {
	if e.Playing() {
		e.Pause()

		return
	}

	e.Play()
}

// Playing reports advancing audio.
func (e *Engine) Playing() bool {
	e.mu.Lock()
	voice, want := e.voice, e.wantPlay
	e.mu.Unlock()

	return voice != nil && want && voice.Playing()
}

// Ended reports a finished voice.
func (e *Engine) Ended() bool {
	e.mu.Lock()
	voice, want := e.voice, e.wantPlay
	e.mu.Unlock()
	if voice == nil {
		return false
	}
	if want && voice.Playing() {
		return false
	}
	dur := voice.Duration()
	if dur <= 0 {
		return false
	}
	return voice.Position() >= dur
}

// Loading reports a resolve.
func (e *Engine) Loading() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.loading
}

// SeekFraction seeks by 0..1.
func (e *Engine) SeekFraction(fraction float64) {
	e.mu.Lock()
	voice := e.voice
	e.mu.Unlock()
	if voice == nil {
		return
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	_ = voice.Seek(time.Duration(fraction * float64(voice.Duration())))
}

// SetVolume sets volume 0..1.
func (e *Engine) SetVolume(volume float64) {
	volume = clampVolume(volume)
	e.mu.Lock()
	voice := e.voice
	e.volume = volume
	e.mu.Unlock()
	if voice != nil {
		voice.SetVolume(volume)
	}
}
