package music

import (
	"context"
	"sync"
	"time"
)

// fakeVoice records the calls an engine makes. The engine drives it from a
// background goroutine while the test polls it, so mu guards every field.
type fakeVoice struct {
	mu      sync.Mutex
	playing bool
	pos     time.Duration
	dur     time.Duration
	volume  float64
	closed  bool
}

func (f *fakeVoice) Play() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playing = true
}

func (f *fakeVoice) Pause() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playing = false
}

func (f *fakeVoice) Playing() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.playing
}

func (f *fakeVoice) Position() time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pos
}

func (f *fakeVoice) Duration() time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.dur
}

func (f *fakeVoice) Seek(at time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pos = at
	return nil
}

func (f *fakeVoice) SetVolume(volume float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.volume = volume
}

func (f *fakeVoice) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// Closed reports whether Close ran.
func (f *fakeVoice) Closed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

// Volume reports the last SetVolume value.
func (f *fakeVoice) Volume() float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.volume
}

// finish marks the voice ended at its duration.
func (f *fakeVoice) finish() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playing = false
	f.pos = f.dur
}

// fakeResolver answers with one fixed clip or error.
type fakeResolver struct {
	clip Clip
	err  error
}

func (r *fakeResolver) Resolve(context.Context, string, int) (Clip, error) {
	return r.clip, r.err
}

// hangResolver blocks until the resolve context is done.
type hangResolver struct{}

func (hangResolver) Resolve(ctx context.Context, _ string, _ int) (Clip, error) {
	<-ctx.Done()

	return Clip{}, ctx.Err()
}
