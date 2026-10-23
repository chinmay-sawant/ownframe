package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// SetImage registers encoded image bytes (PNG, JPEG, or SVG) for one source
// name. A template image source that spells that name, such as
// background-image: url("logo") or <img src="logo">, resolves to data on the
// next Redraw. The bytes are cloned, so later caller edits do not repaint
// the page. Passing nil data removes the entry.
func (p *Page) SetImage(src string, data []byte) {
	p.images = draw.Store(p.images, src, data)
	p.imagesRev++
}

// imageBytes resolves one image source for the layout engine. A src the
// page never registered falls back to the bundled emoji pictures, so the
// replacer needs no per-page setup. Blocked schemes (GEN-30) fail before
// any lookup.
func (p *Page) imageBytes(src string) ([]byte, error) {
	if err := checkImageSrc(src); err != nil {
		return nil, err
	}

	data, ok := draw.Lookup(p.images, src)
	if !ok {
		return nil, errNoImage
	}

	return data, nil
}
