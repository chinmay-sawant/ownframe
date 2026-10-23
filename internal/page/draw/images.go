package draw

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/emoji"
)

// Store clones data into images under src. nil data removes src.
// A nil map is allocated, including for a delete, so the caller can keep
// the revision bump on every call.
func Store(images map[string][]byte, src string, data []byte) map[string][]byte {
	if images == nil {
		images = map[string][]byte{}
	}
	if data == nil {
		delete(images, src)

		return images
	}

	images[src] = append([]byte(nil), data...)

	return images
}

// Lookup resolves one source. A registered entry wins. A src the page
// never registered falls back to the bundled emoji pictures, so the
// replacer needs no per-page setup. The bool is false when nothing matches.
func Lookup(images map[string][]byte, src string) ([]byte, bool) {
	if data, ok := images[src]; ok {
		return data, true
	}

	name, ok := strings.CutPrefix(src, "emoji/")
	if !ok {
		return nil, false
	}

	return emoji.PNG(name)
}
