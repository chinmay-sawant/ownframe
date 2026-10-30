package music

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// newFakeEngine returns an engine whose Open hands back one fake voice.
func newFakeEngine() (*Engine, *fakeVoice) {
	voice := &fakeVoice{dur: time.Minute}
	engine := NewEngine(&fakeResolver{clip: Clip{Title: "Free", Data: []byte("x")}})
	engine.Open = func([]byte) (Voice, error) { return voice, nil }

	return engine, voice
}

// waitVoice ticks the engine until its background resolve lands.
func waitVoice(t *testing.T, engine *Engine) *fakeVoice {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		engine.Update()
		if !engine.Loading() {
			engine.mu.Lock()
			voice, ok := engine.voice.(*fakeVoice)
			engine.mu.Unlock()
			if !ok {
				t.Fatalf("no voice: %v", engine.Err())
			}

			return voice
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("the engine never finished loading")

	return nil
}

func TestEngineStartsOnUpdate(t *testing.T) {
	t.Parallel()

	engine, voice := newFakeEngine()
	engine.SetVolume(0.5)
	engine.Want(context.Background(), "jazz", 2)

	if got := waitVoice(t, engine); got != voice {
		t.Fatal("wrong voice installed")
	}

	if !voice.Playing() {
		t.Fatal("the voice did not start")
	}

	if voice.Volume() != 0.5 {
		t.Fatalf("volume = %v", voice.Volume())
	}

	if engine.Credit().Title != "Free" {
		t.Fatalf("credit = %+v", engine.Credit())
	}
}

func TestEngineAsyncDecode(t *testing.T) {
	t.Parallel()

	opened := make(chan struct{})
	var opens atomic.Int32
	engine, voice := newFakeEngine()
	engine.Open = func([]byte) (Voice, error) {
		if opens.Add(1) == 1 {
			close(opened)
		}
		return voice, nil
	}
	engine.Want(context.Background(), "jazz", 1)

	select {
	case <-opened:
		// Decode was triggered asynchronously without Update!
	case <-time.After(time.Second):
		t.Fatal("e.Open was not called in background")
	}

	// The background goroutine publishes the decoded voice after Open
	// returns, so poll Update until it lands instead of assuming one
	// Update is enough. Update applies the already-decoded voice without
	// decoding on the UI thread.
	deadline := time.Now().Add(2 * time.Second)
	for engine.Loading() && time.Now().Before(deadline) {
		engine.Update()
		time.Sleep(time.Millisecond)
	}
	if engine.Loading() {
		t.Fatal("engine still loading after Update with decoded voice")
	}
	if got := opens.Load(); got != 1 {
		t.Fatalf("Open called %d times, want 1 (decode must stay off the UI thread)", got)
	}
}

func TestEngineNewerGenClosesVoice(t *testing.T) {
	t.Parallel()

	voice1 := &fakeVoice{dur: time.Minute}
	voice2 := &fakeVoice{dur: time.Minute}

	blockDecode := make(chan struct{})
	engine := NewEngine(&fakeResolver{})
	engine.Open = func(data []byte) (Voice, error) {
		if string(data) == "first" {
			<-blockDecode
			return voice1, nil
		}
		return voice2, nil
	}

	engine.res = resolverFunc(func(ctx context.Context, query string, pick int) (Clip, error) {
		return Clip{Title: query, Data: []byte(query)}, nil
	})

	engine.Want(context.Background(), "first", 0)

	// Issue a newer Want
	engine.Want(context.Background(), "second", 0)

	// Unblock decode for the first track
	close(blockDecode)

	// Wait for first voice to be closed
	deadline := time.Now().Add(2 * time.Second)
	for !voice1.Closed() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	if !voice1.Closed() {
		t.Fatal("superseded voice was not closed")
	}

	// Wait for engine to resolve track 2
	for engine.Loading() && time.Now().Before(deadline) {
		engine.Update()
		time.Sleep(time.Millisecond)
	}

	if engine.Credit().Title != "second" {
		t.Fatalf("credit = %v, want second", engine.Credit().Title)
	}
}

type resolverFunc func(ctx context.Context, query string, pick int) (Clip, error)

func (f resolverFunc) Resolve(ctx context.Context, query string, pick int) (Clip, error) {
	return f(ctx, query, pick)
}
