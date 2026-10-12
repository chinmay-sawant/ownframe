package snip

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

var pngMagic = []byte{137, 80, 78, 71, 13, 10, 26, 10}

func TestDemoImageValidPNG(t *testing.T) {
	raw := DemoImage(160, 100)
	if !bytes.HasPrefix(raw, pngMagic) {
		t.Fatal("DemoImage is not a PNG")
	}
	img, err := DecodePNG(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds(); got.Dx() != 160 || got.Dy() != 100 {
		t.Fatalf("bounds = %v, want 160x100", got)
	}
}

func TestDecodeEncodeRoundtrip(t *testing.T) {
	img, err := DecodePNG(DemoImage(64, 48))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := EncodePNG(img)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodePNG(raw)
	if err != nil {
		t.Fatal(err)
	}
	if back.Bounds() != img.Bounds() {
		t.Fatalf("bounds = %v, want %v", back.Bounds(), img.Bounds())
	}
}

func TestFullscreenHeadless(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	raw, err := Fullscreen(ctx)
	if err != nil {
		if err.Error() == "" {
			t.Fatal("empty fullscreen error")
		}
		t.Logf("no screenshot tool: %v", err)
		return
	}
	if !bytes.HasPrefix(raw, pngMagic) {
		t.Fatal("Fullscreen bytes are not a PNG")
	}
}

func TestCaptureWithFake(t *testing.T) {
	want := []byte("fake-png")
	fake := Capturer(func(context.Context) ([]byte, error) { return want, nil })
	got, err := CaptureWith(context.Background(), fake)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("fake capturer bytes not returned")
	}
	boom := errors.New("boom")
	fail := Capturer(func(context.Context) ([]byte, error) { return nil, boom })
	if _, err := CaptureWith(context.Background(), fail); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
}
