package web

import (
	"bytes"
	"net/http"
	"strings"
	"unicode"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

const (
	wavHeaderLen = 12
	waveOffset   = 8
	mp3SyncLen   = 2
	mp3SyncFF    = 0xFF
	mp3SyncFB    = 0xFB
	mp3SyncF3    = 0xF3
	mp3SyncF2    = 0xF2

	contentWAV  = "audio/wav"
	contentMPEG = "audio/mpeg"
	contentOGG  = "audio/ogg"
)

// audioSource is a page that published one clip. The web package asserts
// it. It does not declare the method on a screen that lacks it.
type audioSource interface {
	Audio() (data []byte, contentType string, gen uint64, ok bool)
}

// audio serves GET /audio. The browser plays whatever clip the page published.
func (s *server) audio(w http.ResponseWriter, r *http.Request) {
	src, ok := s.app.(audioSource)
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.mu.Lock()
	data, contentType, _, ok := src.Audio()
	s.mu.Unlock()
	if !ok || len(data) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	if ct := audioContentType(contentType, data); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if _, err := w.Write(data); err != nil {
		return
	}
}

// clipGeneration is the published clip generation. It is 0 when the screen
// has no Audio method or no clip. The caller holds s.mu.
func clipGeneration(app host.Screen) uint64 {
	if src, ok := app.(interface{ AudioGen() uint64 }); ok {
		return src.AudioGen()
	}
	src, ok := app.(audioSource)
	if !ok {
		return 0
	}
	_, _, gen, ok := src.Audio()
	if !ok {
		return 0
	}

	return gen
}

// audioContentType uses the stored type when it is an audio media type.
// Otherwise it sniffs the common headers. audio/mpeg is returned only for
// an MP3 header. A recognized clip is not left as application/octet-stream,
// which does not play under nosniff.
func audioContentType(stored string, data []byte) string {
	if ct, ok := storedAudioType(stored); ok {
		return ct
	}
	switch {
	case wavBytes(data):
		return contentWAV
	case mp3Bytes(data):
		return contentMPEG
	case oggBytes(data):
		return contentOGG
	default:
		return ""
	}
}

func storedAudioType(stored string) (string, bool) {
	if !strings.HasPrefix(stored, "audio/") || !headerSafe(stored) {
		return "", false
	}

	return stored, true
}

func headerSafe(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}

	return true
}

func wavBytes(data []byte) bool {
	if len(data) < wavHeaderLen {
		return false
	}

	return bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[waveOffset:wavHeaderLen], []byte("WAVE"))
}

func mp3Bytes(data []byte) bool {
	if bytes.HasPrefix(data, []byte("ID3")) {
		return true
	}
	if len(data) < mp3SyncLen || data[0] != mp3SyncFF {
		return false
	}
	switch data[1] {
	case mp3SyncFB, mp3SyncF3, mp3SyncF2:
		return true
	default:
		return false
	}
}

func oggBytes(data []byte) bool {
	return bytes.HasPrefix(data, []byte("OggS"))
}
