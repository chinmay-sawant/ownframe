package music

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEnginePauseDuringLoad(t *testing.T) {
	t.Parallel()

	engine, voice := newFakeEngine()
	engine.Want(context.Background(), "jazz", 0)
	engine.Pause()

	waitVoice(t, engine)
	if voice.Playing() {
		t.Fatal("a paused engine started the voice")
	}

	engine.Toggle()
	if !voice.Playing() {
		t.Fatal("toggle did not resume")
	}
}

func TestEngineEnded(t *testing.T) {
	t.Parallel()

	engine, voice := newFakeEngine()
	engine.Want(context.Background(), "jazz", 0)
	waitVoice(t, engine)

	voice.finish()

	if !engine.Ended() {
		t.Fatal("Ended = false at the end")
	}

	if engine.Playing() {
		t.Fatal("Playing = true at the end")
	}
}

func TestEngineSeekFraction(t *testing.T) {
	t.Parallel()

	engine, voice := newFakeEngine()
	engine.Want(context.Background(), "jazz", 0)
	waitVoice(t, engine)

	engine.SeekFraction(0.25)
	if voice.Position() != 15*time.Second {
		t.Fatalf("pos = %v", voice.Position())
	}
}

func TestEngineResolveError(t *testing.T) {
	t.Parallel()

	engine := NewEngine(&fakeResolver{err: errors.New("offline")})
	engine.Open = func([]byte) (Voice, error) { return &fakeVoice{}, nil }
	engine.Want(context.Background(), "jazz", 0)

	deadline := time.Now().Add(2 * time.Second)
	for engine.Loading() && time.Now().Before(deadline) {
		engine.Update()
		time.Sleep(time.Millisecond)
	}

	if engine.Err() == nil {
		t.Fatal("no error recorded")
	}
}

func TestEngineResolveTimeout(t *testing.T) {
	t.Parallel()

	engine := NewEngine(hangResolver{})
	engine.Timeout = 20 * time.Millisecond
	engine.Open = func([]byte) (Voice, error) { return &fakeVoice{}, nil }
	engine.Want(context.Background(), "jazz", 0)

	deadline := time.Now().Add(2 * time.Second)
	for engine.Loading() && time.Now().Before(deadline) {
		engine.Update()
		time.Sleep(time.Millisecond)
	}

	if engine.Err() == nil {
		t.Fatal("a hung resolve did not time out")
	}
}
