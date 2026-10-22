package music

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Voice is one decoded stream the engine drives. Sound implements it; tests
// use a fake.
type Voice interface {
	Play()
	Pause()
	Playing() bool
	Position() time.Duration
	Duration() time.Duration
	Seek(time.Duration) error
	SetVolume(float64)
	Close() error
}

// Sound plays encoded MP3 or WAV bytes through the Ebiten audio context.
type Sound struct {
	player *audio.Player
	dur    time.Duration
}

// NewSound decodes data and returns a ready player. It returns ErrNoAudio
// when no Ebiten audio context is running, which is the case in -web mode
// and in tests.
func NewSound(data []byte) (*Sound, error) {
	ctx := audio.CurrentContext()
	if ctx == nil {
		return nil, ErrNoAudio
	}

	stream, length, err := decode(data, ctx.SampleRate())
	if err != nil {
		return nil, err
	}

	player, err := ctx.NewPlayer(stream)
	if err != nil {
		return nil, err
	}

	seconds := float64(length) / wavFrameBytes / float64(ctx.SampleRate())

	return &Sound{player: player, dur: time.Duration(seconds * float64(time.Second))}, nil
}

// Play starts or resumes the stream.
func (s *Sound) Play() { s.player.Play() }

// Pause stops the stream at its current position.
func (s *Sound) Pause() { s.player.Pause() }

// Playing reports whether the stream is advancing.
func (s *Sound) Playing() bool { return s.player.IsPlaying() }

// Position is the current playback offset.
func (s *Sound) Position() time.Duration { return s.player.Position() }

// Duration is the decoded length of the stream.
func (s *Sound) Duration() time.Duration { return s.dur }
