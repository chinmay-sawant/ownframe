package textrun

import (
	"bytes"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// faceKey identifies one shaped face: the engine font value, the pixel size,
// the language tag, and the OpenType feature list. Features change glyph
// advances (tnum, liga), so a face shaped without them would place the caret
// differently from the painted text; internal/replay keys on them too.
type faceKey struct {
	source   any
	size     float64
	language string
	features string
}

// Trim drops every cached face and parsed font source. The caches are bounded
// by maxSources and maxFaces, but a process that goes idle keeps what the last
// page used; the window calls Trim once it parks. Faces rebuild on next use.
func Trim() {
	textFaces.mu.Lock()
	textFaces.sources = nil
	textFaces.faces = nil
	textFaces.mu.Unlock()
}

// textFaces keeps one parsed face source per engine font. Hits resolve
// under a read lock; the parse runs lock-free with a second check that
// keeps one winner under the write lock.
var textFaces = struct {
	mu      sync.RWMutex
	sources map[any]*text.GoTextFaceSource
	faces   map[faceKey]*text.GoTextFace
}{}

// faceFor returns the Ebiten face for a text operation, or nil when the
// font cannot be parsed.
func faceFor(op *layout.DisplayOp, pxPerPt float64) *text.GoTextFace {
	if op.Font == nil {
		return nil
	}

	size := op.Size * pxPerPt
	key := faceKey{source: op.Font, size: size, language: op.TextLanguage(), features: op.FontFeatures()}

	textFaces.mu.RLock()
	source := textFaces.sources[op.Font]
	if face := textFaces.faces[key]; face != nil {
		textFaces.mu.RUnlock()
		return face
	}
	textFaces.mu.RUnlock()

	if source == nil {
		parsed, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
		if err != nil {
			return nil
		}
		source = parsed
	}

	textFaces.mu.Lock()
	defer textFaces.mu.Unlock()

	if face := textFaces.faces[key]; face != nil {
		return face
	}
	if have := textFaces.sources[op.Font]; have != nil {
		source = have
	} else {
		if textFaces.sources == nil || len(textFaces.sources) >= maxSources {
			textFaces.sources = make(map[any]*text.GoTextFaceSource)
			textFaces.faces = make(map[faceKey]*text.GoTextFace)
		}
		textFaces.sources[op.Font] = source
	}

	face := &text.GoTextFace{Source: source, Size: size, Language: masking.ShapingLanguage(key.language)}
	for _, feature := range masking.ParseFontFeatures(key.features) {
		if tag, err := text.ParseTag(feature.Tag); err == nil {
			face.SetFeature(tag, feature.Value)
		}
	}
	if textFaces.faces == nil {
		textFaces.faces = make(map[faceKey]*text.GoTextFace)
	}
	if len(textFaces.faces) >= maxFaces {
		clear(textFaces.faces)
	}
	textFaces.faces[key] = face

	return face
}
