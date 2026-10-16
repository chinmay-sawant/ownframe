package synth

import (
	"bytes"
	"math"
	"testing"
)

func TestAudioFrequencies(t *testing.T) {
	a4 := noteFrequency("A", 4)
	if math.Abs(a4-440.0) > 0.01 {
		t.Errorf("got A4 = %f, want 440.0", a4)
	}

	c4 := noteFrequency("C", 4)
	if math.Abs(c4-261.63) > 0.1 {
		t.Errorf("got C4 = %f, want ~261.63", c4)
	}

	c5 := noteFrequency("C", 5)
	if math.Abs(c5-523.25) > 0.2 {
		t.Errorf("got C5 = %f, want ~523.25", c5)
	}
}

func TestWaveformSynthesis(t *testing.T) {
	waves := []string{"square", "sawtooth", "triangle", "sine"}
	for _, w := range waves {
		pcm := synthPCM(440.0, w, 48000, 0.05)
		if len(pcm) == 0 {
			t.Fatalf("empty PCM for wave %s", w)
		}
		wav := encodeMonoWAV(pcm, 48000)
		if len(wav) < 44 {
			t.Fatalf("WAV output too short: %d bytes", len(wav))
		}
		if !bytes.HasPrefix(wav, []byte("RIFF")) {
			t.Errorf("missing RIFF header in WAV for %s", w)
		}
		if !bytes.Contains(wav, []byte("WAVEfmt ")) {
			t.Errorf("missing WAVEfmt in WAV for %s", w)
		}
	}
}

func TestAudioFallback(t *testing.T) {
	audio := NewAudioHost()
	defer audio.Close()

	audio.PlayNote("C", 4, "square")
	if audio.PlayCount != 1 {
		t.Errorf("got PlayCount %d, want 1", audio.PlayCount)
	}
	if audio.LastNote != "C" {
		t.Errorf("got LastNote %q, want C", audio.LastNote)
	}
}
