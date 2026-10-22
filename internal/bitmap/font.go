package bitmap

import (
	"hash/fnv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type faceKey struct {
	hash uint64
	size float64
	n    int
}

type faceCache map[faceKey]font.Face

func (c faceCache) face(data []byte, size float64) font.Face {
	if len(data) == 0 {
		return nil
	}
	h := fnv.New64a()
	// hash.Hash.Write never returns an error; it always returns len(data), nil.
	_, _ = h.Write(data)
	key := faceKey{hash: h.Sum64(), size: size, n: len(data)}
	if face, ok := c[key]; ok {
		return face
	}

	parsed, err := opentype.Parse(data)
	if err != nil {
		c[key] = nil

		return nil
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: size,
		DPI:  cssDPI,
	})
	if err != nil {
		c[key] = nil

		return nil
	}

	c[key] = face

	return face
}
