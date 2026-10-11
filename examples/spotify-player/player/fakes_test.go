package player

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/music"
)

// fakeVoice is a music.Voice a test can inspect.
type fakeVoice struct {
	playing bool
	pos     time.Duration
	dur     time.Duration
	volume  float64
}

func (f *fakeVoice) Play()                       { f.playing = true }
func (f *fakeVoice) Pause()                      { f.playing = false }
func (f *fakeVoice) Playing() bool               { return f.playing }
func (f *fakeVoice) Position() time.Duration     { return f.pos }
func (f *fakeVoice) Duration() time.Duration     { return f.dur }
func (f *fakeVoice) Seek(at time.Duration) error { f.pos = at; return nil }
func (f *fakeVoice) SetVolume(v float64)         { f.volume = v }
func (f *fakeVoice) Close() error                { return nil }

// fakeResolver answers every query with one clip.
type fakeResolver struct {
	clip music.Clip
}

func (r *fakeResolver) Resolve(context.Context, string, int) (music.Clip, error) {
	return r.clip, nil
}

// newFakeApp builds a page whose engine never opens an audio device.
func newFakeApp(t testing.TB) (*App, *fakeVoice) {
	t.Helper()

	voice := &fakeVoice{dur: 200 * time.Second}
	engine := music.NewEngine(&fakeResolver{clip: music.Clip{
		Title: "Test clip", Creator: "Tester", License: "CC0 1.0", Data: []byte("x"),
	}})
	engine.Open = func([]byte) (music.Voice, error) { return voice, nil }

	app, err := NewWith("", engine)
	if err != nil {
		t.Fatalf("NewWith: %v", err)
	}

	t.Cleanup(app.Close)

	return app, voice
}

// waitAudio ticks until the engine finishes loading.
func waitAudio(t *testing.T, app *App) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for app.audio.Loading() && time.Now().Before(deadline) {
		if err := app.Tick(context.Background()); err != nil {
			t.Fatalf("Tick: %v", err)
		}

		time.Sleep(time.Millisecond)
	}

	if app.audio.Loading() {
		t.Fatal("audio still loading after 2s")
	}
}
