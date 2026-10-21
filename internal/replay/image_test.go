package replay

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"
	"unsafe"

	"github.com/chinmay-sawant/blinkless/layout"
)

// tinyPNG builds a 2x2 PNG whose first pixel carries shade, enough to
// exercise decode and caching without a fixture file.
func tinyPNG(t *testing.T, shade uint8) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: shade, A: 255})

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatalf("png: %v", err)
	}

	return out.Bytes()
}

// imageOp builds an op carrying an encoded payload. A display-list consumer
// cannot attach one through the public API, so the test reaches the embedded
// payload directly and allocates a fresh extra instead of mutating the
// engine's shared empty extra.
func imageOp(t *testing.T, data []byte, w, h int, isJPEG bool) layout.DisplayOp {
	t.Helper()

	op := layout.DisplayOp{Kind: layout.DisplayOpImage, IsJPEG: isJPEG}
	root := reflect.ValueOf(&op).Elem()
	extra := reflect.New(root.FieldByName("opExtra").Type().Elem())

	fields := extra.Elem()
	set(t, fields.FieldByName("Image"), reflect.ValueOf(data))
	set(t, fields.FieldByName("ImgW"), reflect.ValueOf(w))
	set(t, fields.FieldByName("ImgH"), reflect.ValueOf(h))
	set(t, root.FieldByName("opExtra"), extra)

	return op
}

// set writes one payload field that reflect refuses to set because the payload
// type is unexported.
func set(t *testing.T, field, value reflect.Value) {
	t.Helper()

	if !field.IsValid() {
		t.Fatal("payload field missing")
	}

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(value) //nolint:gosec // reason: test-only write to an unexported field.
}
