package page

import (
	"bytes"
	"testing"
)

// TestAudio stores a clip, checks the read copy, and checks that the
// generation moves only when the bytes change.
func TestAudio(t *testing.T) {
	t.Parallel()

	p := &Page{}
	data, contentType, gen, ok := p.Audio()
	if ok || gen != 0 || data != nil || contentType != "" {
		t.Fatalf("empty clip = %q %q gen=%d ok=%v", data, contentType, gen, ok)
	}

	p.SetAudio(nil, "audio/wav")
	_, _, gen, ok = p.Audio()
	if ok || gen != 0 {
		t.Fatalf("clear of nothing gen=%d ok=%v", gen, ok)
	}

	src := []byte("riff-one")
	p.SetAudio(src, "audio/wav")
	src[0] = 'X'

	got, contentType, gen, ok := p.Audio()
	if !ok || gen != 1 || contentType != "audio/wav" || !bytes.Equal(got, []byte("riff-one")) {
		t.Fatalf("set = %q %q gen=%d ok=%v", got, contentType, gen, ok)
	}

	got[0] = 'Z'
	got, contentType, same, ok := p.Audio()
	if !ok || same != gen || contentType != "audio/wav" || !bytes.Equal(got, []byte("riff-one")) {
		t.Fatalf("after read edit = %q %q gen=%d ok=%v", got, contentType, same, ok)
	}

	p.SetAudio([]byte("riff-one"), "audio/wav")
	_, _, kept, ok := p.Audio()
	if !ok || kept != gen {
		t.Fatalf("same bytes gen=%d, want %d ok=%v", kept, gen, ok)
	}

	p.SetAudio([]byte("riff-two"), "audio/mpeg")
	got, contentType, next, ok := p.Audio()
	if !ok || next != kept+1 || contentType != "audio/mpeg" || !bytes.Equal(got, []byte("riff-two")) {
		t.Fatalf("new bytes = %q %q gen=%d ok=%v", got, contentType, next, ok)
	}

	p.SetAudio([]byte{}, "audio/mpeg")
	got, contentType, cleared, ok := p.Audio()
	if ok || got != nil || contentType != "" || cleared != next+1 {
		t.Fatalf("empty clear = %q %q gen=%d ok=%v", got, contentType, cleared, ok)
	}

	p.SetAudio(nil, "audio/wav")
	_, _, again, ok := p.Audio()
	if ok || again != cleared {
		t.Fatalf("nil after clear gen=%d, want %d ok=%v", again, cleared, ok)
	}
}

func TestAudioPlayback(t *testing.T) {
	t.Parallel()

	p := &Page{}
	playing, volume := p.AudioPlayback()
	if playing || volume != 1 {
		t.Fatalf("default playing=%v volume=%v", playing, volume)
	}

	p.SetAudioPlayback(true, 0.1)
	playing, volume = p.AudioPlayback()
	if !playing || volume != 0.1 {
		t.Fatalf("set playing=%v volume=%v", playing, volume)
	}

	p.SetAudioPlayback(false, 2)
	playing, volume = p.AudioPlayback()
	if playing || volume != 1 {
		t.Fatalf("clamp playing=%v volume=%v", playing, volume)
	}

	p.SetAudioPlayback(false, -1)
	_, volume = p.AudioPlayback()
	if volume != 0 {
		t.Fatalf("low clamp volume=%v", volume)
	}

	var empty *Page
	empty.SetAudioPlayback(true, 0.2)
	playing, volume = empty.AudioPlayback()
	if playing || volume != 1 {
		t.Fatalf("nil page playing=%v volume=%v", playing, volume)
	}
}

func TestAudioNilPage(t *testing.T) {
	t.Parallel()

	var p *Page
	p.SetAudio([]byte("riff-one"), "audio/wav")

	data, contentType, gen, ok := p.Audio()
	if ok || gen != 0 || data != nil || contentType != "" {
		t.Fatalf("nil page = %q %q gen=%d ok=%v", data, contentType, gen, ok)
	}
}
