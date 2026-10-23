package replay

import (
	"bytes"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/chinmay-sawant/blinkless/layout"
)

// faceKey identifies one shaped face: the engine font value, the pixel size,
// and the language tag handed to the Ebiten face.
type faceKey struct {
	source   any
	size     float64
	language string
}

// faceSourceCap bounds parsed font sources before a reset.
// faceCap bounds shaped faces before a reset.
const (
	faceSourceCap = 64
	faceCap       = 256
)

// textCache keeps one parsed face source per engine font. Hits resolve
// under a read lock; the parse runs lock-free with a second check that
// keeps one winner under the write lock.
type textCache struct {
	mu      sync.RWMutex
	sources map[any]*text.GoTextFaceSource
	faces   map[faceKey]*text.GoTextFace
}

func (renderer *Renderer) textFace(op *layout.DisplayOp) *text.GoTextFace {
	if op.Font == nil {
		return nil
	}

	size := op.Size * pxPerPt
	key := faceKey{source: op.Font, size: size, language: op.TextLanguage()}

	renderer.fonts.mu.RLock()
	source := renderer.fonts.sources[op.Font]
	if face := renderer.fonts.faces[key]; face != nil {
		renderer.fonts.mu.RUnlock()
		return face
	}
	renderer.fonts.mu.RUnlock()

	if source == nil {
		parsed, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
		if err != nil {
			return nil
		}
		source = parsed
	}

	renderer.fonts.mu.Lock()
	defer renderer.fonts.mu.Unlock()

	if face := renderer.fonts.faces[key]; face != nil {
		return face
	}
	if have := renderer.fonts.sources[op.Font]; have != nil {
		source = have
	} else {
		if renderer.fonts.sources == nil || len(renderer.fonts.sources) >= faceSourceCap {
			renderer.fonts.sources = make(map[any]*text.GoTextFaceSource)
			renderer.fonts.faces = make(map[faceKey]*text.GoTextFace)
		}
		renderer.fonts.sources[op.Font] = source
	}

	face := &text.GoTextFace{Source: source, Size: size, Language: language.Make(key.language)}
	if len(renderer.fonts.faces) >= faceCap {
		clear(renderer.fonts.faces)
	}
	renderer.fonts.faces[key] = face

	return face
}
