package page

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/emoji"
)

// SetImage registers encoded image bytes (PNG, JPEG, or SVG) for one source
// name. A template image source that spells that name, such as
// background-image: url("logo") or <img src="logo">, resolves to data on the
// next Redraw. The bytes are cloned, so later caller edits do not repaint
// the page. Passing nil data removes the entry.
func (p *Page) SetImage(src string, data []byte) {
	if p.images == nil {
		p.images = map[string][]byte{}
	}

	if data == nil {
		delete(p.images, src)
		p.imagesRev++

		return
	}

	p.images[src] = append([]byte(nil), data...)
	p.imagesRev++
}

// imageBytes resolves one image source for the layout engine. A src the
// page never registered falls back to the bundled emoji pictures, so the
// replacer needs no per-page setup.
func (p *Page) imageBytes(src string) ([]byte, error) {
	if data, ok := p.images[src]; ok {
		return data, nil
	}

	if name, ok := strings.CutPrefix(src, "emoji/"); ok {
		if data, ok := emoji.PNG(name); ok {
			return data, nil
		}
	}

	return nil, errNoImage
}
