package ownframe

import "context"

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) { innerPage(p).SetData(data) }

// Handle registers the input functions. A later call replaces the whole set.
func (p *Page) Handle(h Handlers) { innerPage(p).Handle(h) }

// AddHandlers merges only non-nil callbacks into the existing set.
func (p *Page) AddHandlers(h Handlers) { innerPage(p).AddHandlers(h) }

// SetTheme stores an extra stylesheet the next Redraw applies.
func (p *Page) SetTheme(source string) error {
	return innerPage(p).SetTheme(source)
}

// SetTick registers fn to run once per frame before the window draws.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	innerPage(p).SetTick(fn)
}

// SetImage registers encoded image bytes for one source name.
func (p *Page) SetImage(src string, data []byte) {
	innerPage(p).SetImage(src, data)
}

// SetAudio stores one encoded clip for the web preview.
func (p *Page) SetAudio(data []byte, contentType string) {
	innerPage(p).SetAudio(data, contentType)
}

// Audio returns a copy of the current clip, its content type, and the generation.
func (p *Page) Audio() (data []byte, contentType string, gen uint64, ok bool) {
	return innerPage(p).Audio()
}

// SetAudioPlayback publishes whether the preview should play and at what
// volume, from 0 to 1. The browser player follows it.
func (p *Page) SetAudioPlayback(playing bool, volume float64) {
	innerPage(p).SetAudioPlayback(playing, volume)
}

// AudioPlayback reports the published playing flag and volume.
func (p *Page) AudioPlayback() (playing bool, volume float64) {
	return innerPage(p).AudioPlayback()
}
