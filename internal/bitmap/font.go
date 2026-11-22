package bitmap

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type fontDataKey struct {
	ptr *byte
	n   int
}

type faceKey struct {
	ptr  *byte
	n    int
	size float64
}

type faceCache struct {
	parsed map[fontDataKey]*opentype.Font
	faces  map[faceKey]font.Face
}

func newFaceCache() *faceCache {
	return &faceCache{
		parsed: map[fontDataKey]*opentype.Font{},
		faces:  map[faceKey]font.Face{},
	}
}

func (c *faceCache) face(data []byte, size float64) font.Face {
	if len(data) == 0 {
		return nil
	}

	ptr := &data[0]
	key := faceKey{ptr: ptr, n: len(data), size: size}
	if face, ok := c.faces[key]; ok {
		return face
	}

	dkey := fontDataKey{ptr: ptr, n: len(data)}
	parsed, ok := c.parsed[dkey]
	if !ok {
		var err error
		parsed, err = opentype.Parse(data)
		if err != nil {
			c.parsed[dkey] = nil
			c.faces[key] = nil

			return nil
		}
		c.parsed[dkey] = parsed
	}
	if parsed == nil {
		c.faces[key] = nil

		return nil
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: size,
		DPI:  cssDPI,
	})
	if err != nil {
		c.faces[key] = nil

		return nil
	}

	c.faces[key] = face

	return face
}

func (c *faceCache) close() {
	if c == nil {
		return
	}
	for _, face := range c.faces {
		if face != nil {
			_ = face.Close()
		}
	}
}
