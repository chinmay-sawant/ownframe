package music

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

func TestDecodeWAV(t *testing.T) {
	t.Parallel()

	clip := DemoTune("decode")
	stream, length, err := decode(clip.Data, demoRate)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if length != int64(demoSeconds*demoRate*4) {
		t.Fatalf("length = %d", length)
	}

	buf := make([]byte, 256)
	if _, err := stream.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
}

func TestDecodeRejectsUnknownData(t *testing.T) {
	t.Parallel()

	if _, _, err := decode([]byte("not audio at all"), demoRate); err == nil {
		t.Fatal("decode accepted unknown data")
	}
}

func TestNewSoundWithoutContext(t *testing.T) {
	t.Parallel()

	if audio.CurrentContext() != nil {
		t.Skip("an audio context is already running")
	}

	if _, err := NewSound(DemoTune("quiet").Data); err != ErrNoAudio {
		t.Fatalf("err = %v", err)
	}
}
