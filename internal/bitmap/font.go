package bitmap

import (
	"unsafe"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type faceKey struct {
	ptr  uintptr
	size float64
}

type faceCache map[faceKey]font.Face

func (c faceCache) face(data []byte, size float64) font.Face {
	key := faceKey{ptr: uintptr(unsafe.Pointer(&data[0])), size: size}
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
		DPI:  96,
	})
	if err != nil {
		c[key] = nil

		return nil
	}

	c[key] = face

	return face
}
