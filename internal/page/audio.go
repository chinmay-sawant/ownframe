package page

import (
	"bytes"
	"sync"
)

// audioSlot is one encoded clip the web preview can play. SetAudio can run
// on any goroutine while Audio runs on the preview server, so the mutex
// guards the bytes, the content type, and the generation.
type audioSlot struct {
	mu          sync.Mutex
	data        []byte
	contentType string
	gen         uint64
	playing     bool
	volume      float64
	volumeSet   bool
}

// SetAudio stores one encoded clip for the web preview. A later call
// replaces it. Nil or empty data clears the clip. contentType may be
// empty; the web server sniffs it. The generation increments only when
// the bytes change, so a repeat of the same clip does not restart playback.
func (p *Page) SetAudio(data []byte, contentType string) {
	if p == nil {
		return
	}

	p.audio.set(data, contentType)
}

// Audio returns a copy of the current clip, its content type, and the
// generation. Generation 0 and ok false mean there is no clip.
func (p *Page) Audio() (data []byte, contentType string, gen uint64, ok bool) {
	if p == nil {
		return nil, "", 0, false
	}

	return p.audio.get()
}

// AudioGen is the published clip generation. It is 0 when there is no clip.
// The web preview polls this without copying the bytes.
func (p *Page) AudioGen() uint64 {
	if p == nil {
		return 0
	}

	return p.audio.generation()
}

// SetAudioPlayback publishes whether the web preview should be playing and
// how loud, from 0 to 1. The browser player follows this. The clip bytes
// stay on SetAudio. A volume outside 0 to 1 is clamped.
func (p *Page) SetAudioPlayback(playing bool, volume float64) {
	if p == nil {
		return
	}

	p.audio.setPlayback(playing, volume)
}

// AudioPlayback reports the published playing flag and volume.
// Volume is 1 until SetAudioPlayback runs.
func (p *Page) AudioPlayback() (playing bool, volume float64) {
	if p == nil {
		return false, 1
	}

	return p.audio.playback()
}

func (a *audioSlot) set(data []byte, contentType string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if bytes.Equal(a.data, data) {
		if len(data) > 0 {
			a.contentType = contentType
		}

		return
	}

	a.gen++
	if len(data) == 0 {
		a.data = nil
		a.contentType = ""

		return
	}

	a.data = append([]byte(nil), data...)
	a.contentType = contentType
}

func (a *audioSlot) setPlayback(playing bool, volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	a.mu.Lock()
	a.playing = playing
	a.volume = volume
	a.volumeSet = true
	a.mu.Unlock()
}

func (a *audioSlot) playback() (bool, float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.volumeSet {
		return a.playing, 1
	}

	return a.playing, a.volume
}

func (a *audioSlot) generation() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.data) == 0 {
		return 0
	}

	return a.gen
}

func (a *audioSlot) get() ([]byte, string, uint64, bool) {
	a.mu.Lock()
	data := a.data
	contentType := a.contentType
	gen := a.gen
	a.mu.Unlock()

	if len(data) == 0 {
		return nil, "", gen, false
	}

	return append([]byte(nil), data...), contentType, gen, true
}
