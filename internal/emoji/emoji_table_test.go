package emoji

import (
	"bytes"
	"image/png"
	"testing"
)

func TestTableHasTwentyImages(t *testing.T) {
	t.Parallel()

	files := Files()
	if len(files) != 42 {
		t.Fatalf("files = %d, want 42", len(files))
	}

	for _, name := range files {
		data, ok := PNG(name)
		if !ok || len(data) == 0 {
			t.Fatalf("%s missing", name)
		}

		if _, err := png.Decode(bytes.NewReader(data)); err != nil {
			t.Fatalf("%s does not decode: %v", name, err)
		}
	}
}

func TestUnknownNameMisses(t *testing.T) {
	t.Parallel()

	if _, ok := PNG("1fae0"); ok {
		t.Fatal("unsupported emoji resolved")
	}
}
