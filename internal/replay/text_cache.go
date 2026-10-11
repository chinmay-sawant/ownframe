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

// textCache keeps one parsed face source per engine font. Parsing the SFNT
// data is expensive; Ebiten caches glyph images per source and size, so the
// faces are reused too.
type textCache struct {
	mu      sync.Mutex
	sources map[any]*text.GoTextFaceSource
	faces   map[faceKey]*text.GoTextFace
}

var fonts textCache

func textFace(op *layout.DisplayOp) *text.GoTextFace {
	if op.Font == nil {
		return nil
	}

	fonts.mu.Lock()
	defer fonts.mu.Unlock()

	source := fonts.sources[op.Font]
	if source == nil {
		parsed, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
		if err != nil {
			return nil
		}

		if fonts.sources == nil || len(fonts.sources) >= 64 {
			fonts.sources = make(map[any]*text.GoTextFaceSource)
			fonts.faces = make(map[faceKey]*text.GoTextFace)
		}

		fonts.sources[op.Font] = parsed
		source = parsed
	}

	size := op.Size * pxPerPt
	key := faceKey{source: op.Font, size: size, language: op.TextLanguage()}

	if face := fonts.faces[key]; face != nil {
		return face
	}

	face := &text.GoTextFace{Source: source, Size: size, Language: language.Make(key.language)}
	if len(fonts.faces) >= 256 {
		clear(fonts.faces)
	}
	fonts.faces[key] = face

	return face
}
