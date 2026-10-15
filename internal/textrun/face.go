package textrun

import (
	"bytes"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/chinmay-sawant/blinkless/layout"
)

// faceKey identifies one shaped face: the engine font value, the pixel size,
// and the language tag.
type faceKey struct {
	source   any
	size     float64
	language string
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
	key := faceKey{source: op.Font, size: size, language: op.TextLanguage()}

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
		if textFaces.sources == nil || len(textFaces.sources) >= 64 {
			textFaces.sources = make(map[any]*text.GoTextFaceSource)
			textFaces.faces = make(map[faceKey]*text.GoTextFace)
		}
		textFaces.sources[op.Font] = source
	}

	face := &text.GoTextFace{Source: source, Size: size, Language: language.Make(key.language)}
	if len(textFaces.faces) >= 256 {
		clear(textFaces.faces)
	}
	textFaces.faces[key] = face

	return face
}
