// Package music decodes MP3 and WAV bytes and plays them through the OS audio
// device. It is example-side support: the Spotify player example supplies the
// clips from its bundled CC0 soundtrack, and package ownframe itself does not
// import it.
package music

import (
	"context"
	"errors"
)

// Clip is one piece of playable audio.
type Clip struct {
	Title   string
	Creator string
	License string // "CC0 1.0 Universal", or "generated" for the demo tune
	URL     string
	Data    []byte // encoded MP3 or WAV bytes
}

// Resolver turns a pick number into playable audio. The Spotify player reads
// its bundled catalog; DemoTune stands in for tests.
type Resolver interface {
	Resolve(ctx context.Context, query string, pick int) (Clip, error)
}

// ErrNoAudio means no Ebiten audio context is running.
var ErrNoAudio = errors.New("music: no audio context; run the page through ownframe.Run")
