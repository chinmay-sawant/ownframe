package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// *page.Page publishes the clip. A fake screen must not stand in for it.
var _ audioSource = (*page.Page)(nil)

func TestAudioRouteMissingClip(t *testing.T) {
	t.Parallel()

	srv := &server{app: audioPage(t)}
	rec := httptest.NewRecorder()
	srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.Bytes())
	}
}

func TestAudioRouteNoAudioMethod(t *testing.T) {
	t.Parallel()

	srv := &server{app: &stateScreen{width: 20, height: 10}}
	rec := httptest.NewRecorder()
	srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestAudioRouteServesPublishedClip(t *testing.T) {
	t.Parallel()

	p := audioPage(t)
	wav := []byte("RIFF\x24\x00\x00\x00WAVEfmt ")
	p.SetAudio(wav, "audio/wav")
	srv := &server{app: p}

	rec := httptest.NewRecorder()
	srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.Bytes())
	}
	if rec.Header().Get("Content-Type") != "audio/wav" {
		t.Fatalf("type = %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache = %q", rec.Header().Get("Cache-Control"))
	}
	if !bytes.Equal(rec.Body.Bytes(), wav) {
		t.Fatalf("body = %q", rec.Body.Bytes())
	}
}

func TestAudioRouteSniffsAndTrustsStoredAudioType(t *testing.T) {
	t.Parallel()

	p := audioPage(t)
	srv := &server{app: p}
	cases := []struct {
		name   string
		body   []byte
		stored string
		kind   string
	}{
		{name: "wav", body: []byte("RIFF\x00\x00\x00\x00WAVE"), stored: "", kind: "audio/wav"},
		{name: "wav-octet", body: []byte("RIFF\x00\x00\x00\x00WAVE"), stored: "application/octet-stream", kind: "audio/wav"},
		{name: "id3", body: []byte("ID3\x03\x00"), stored: "", kind: "audio/mpeg"},
		{name: "sync-fb", body: []byte{0xFF, 0xFB, 0x90, 0x00}, stored: "", kind: "audio/mpeg"},
		{name: "sync-f3", body: []byte{0xFF, 0xF3, 0x90, 0x00}, stored: "", kind: "audio/mpeg"},
		{name: "sync-f2", body: []byte{0xFF, 0xF2, 0x90, 0x00}, stored: "", kind: "audio/mpeg"},
		{name: "ogg", body: []byte("OggS\x00more"), stored: "", kind: "audio/ogg"},
		{name: "stored", body: []byte("OggS\x00"), stored: "audio/flac", kind: "audio/flac"},
		{name: "injected", body: []byte("RIFF\x00\x00\x00\x00WAVE"), stored: "audio/mpeg\r\nX-Injected: 1", kind: "audio/wav"},
	}
	for _, tc := range cases {
		p.SetAudio(tc.body, tc.stored)
		rec := httptest.NewRecorder()
		srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", tc.name, rec.Code)
		}
		if rec.Header().Get("Content-Type") != tc.kind {
			t.Fatalf("%s type = %q", tc.name, rec.Header().Get("Content-Type"))
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s cache = %q", tc.name, rec.Header().Get("Cache-Control"))
		}
		if !bytes.Equal(rec.Body.Bytes(), tc.body) {
			t.Fatalf("%s body = %q", tc.name, rec.Body.Bytes())
		}
		if stringsHasHeaderBreak(rec.Header().Get("Content-Type")) {
			t.Fatalf("%s content type has a line break: %q", tc.name, rec.Header().Get("Content-Type"))
		}
	}
}

func TestAudioRouteEmptyClip(t *testing.T) {
	t.Parallel()

	p := audioPage(t)
	p.SetAudio([]byte("RIFF\x00\x00\x00\x00WAVE"), "audio/wav")
	p.SetAudio(nil, "audio/mpeg")
	srv := &server{app: p}
	rec := httptest.NewRecorder()
	srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestAudioRouteUnknownIsNotMPEG(t *testing.T) {
	t.Parallel()

	p := audioPage(t)
	body := []byte("not-a-known-audio-header")
	p.SetAudio(body, "")
	srv := &server{app: p}
	rec := httptest.NewRecorder()
	srv.audio(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audio", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !bytes.Equal(rec.Body.Bytes(), body) {
		t.Fatalf("body = %q", rec.Body.Bytes())
	}
	if got := rec.Header().Get("Content-Type"); got == "audio/mpeg" || got == "audio/wav" || got == "audio/ogg" {
		t.Fatalf("type = %q, unlabeled bytes must not be guessed as wav, mpeg, or ogg", got)
	}
}

func TestStateAudioGen(t *testing.T) {
	t.Parallel()

	p := audioPage(t)
	srv := &server{app: p}
	if got := stateAudioGen(t, srv); got != 0 {
		t.Fatalf("audioGen before clip = %d", got)
	}

	wav := []byte("RIFF\x24\x00\x00\x00WAVEfmt ")
	p.SetAudio(wav, "audio/wav")
	_, _, gen, ok := p.Audio()
	if !ok {
		t.Fatal("Audio ok = false")
	}
	if got := stateAudioGen(t, srv); got != gen {
		t.Fatalf("audioGen = %d, want %d", got, gen)
	}
	p.SetAudio(nil, "")
	if got := stateAudioGen(t, srv); got != 0 {
		t.Fatalf("audioGen after clear = %d", got)
	}
}

func stateAudioGen(t *testing.T, srv *server) uint64 {
	t.Helper()

	rec := httptest.NewRecorder()
	srv.state(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/state", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	raw, ok := body["audioGen"]
	if !ok {
		t.Fatalf("state missing audioGen: %s", rec.Body.Bytes())
	}
	gen, ok := raw.(float64)
	if !ok {
		t.Fatalf("audioGen = %#v", raw)
	}

	return uint64(gen)
}

func audioPage(t *testing.T) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{
		HTML:   `<p>audio</p>`,
		Width:  200,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func stringsHasHeaderBreak(s string) bool {
	return bytes.ContainsAny([]byte(s), "\r\n")
}
