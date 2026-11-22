package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestImageRouteServesRegisteredPNG(t *testing.T) {
	t.Parallel()

	p := imagePage(t)
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'x'}
	p.SetImage("logo", png)
	srv := &server{app: p}

	rec := httptest.NewRecorder()
	srv.image(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/logo", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("type = %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache = %q", rec.Header().Get("Cache-Control"))
	}
	if !bytes.Equal(rec.Body.Bytes(), png) {
		t.Fatalf("body = %q", rec.Body.Bytes())
	}
}

func TestImageRouteServesAnyRegisteredBytes(t *testing.T) {
	t.Parallel()

	p := imagePage(t)
	srv := &server{app: p}
	cases := []struct {
		name string
		body []byte
		kind string
	}{
		{name: "photo", body: []byte{0xff, 0xd8, 0xff, 0xe0}, kind: "image/jpeg"},
		{name: "mark", body: []byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>"), kind: "image/svg+xml"},
		{name: "still", body: []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), kind: "image/webp"},
		{name: "future", body: []byte("not-a-known-picture"), kind: "application/octet-stream"},
		{name: "html", body: []byte("<html><script>alert(1)</script></html>"), kind: "application/octet-stream"},
	}
	for _, tc := range cases {
		p.SetImage(tc.name, tc.body)
		rec := httptest.NewRecorder()
		srv.image(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/"+tc.name, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", tc.name, rec.Code)
		}
		if rec.Header().Get("Content-Type") != tc.kind {
			t.Fatalf("%s type = %q, want %q", tc.name, rec.Header().Get("Content-Type"), tc.kind)
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("%s missing nosniff: %q", tc.name, rec.Header().Get("X-Content-Type-Options"))
		}
		if tc.kind == "image/svg+xml" && rec.Header().Get("Content-Security-Policy") != "sandbox" {
			t.Fatalf("%s svg missing sandbox CSP: %q", tc.name, rec.Header().Get("Content-Security-Policy"))
		}
		if !bytes.Equal(rec.Body.Bytes(), tc.body) {
			t.Fatalf("%s body = %q", tc.name, rec.Body.Bytes())
		}
	}
}

func TestReadRoutesRejectCrossOrigin(t *testing.T) {
	t.Parallel()

	p := imagePage(t)
	p.SetImage("logo", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'x'})
	p.SetAudio([]byte("RIFF\x00\x00\x00\x00WAVE"), "audio/wav")
	srv := &server{app: p}

	routes := []struct {
		name string
		path string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"dom", "/dom", srv.dom},
		{"state", "/state", srv.state},
		{"audio", "/audio", srv.audio},
		{"image", "/logo", srv.image},
	}
	for _, rt := range routes {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:8080"+rt.path, nil)
		req.Header.Set("Origin", "http://evil.example")
		rec := httptest.NewRecorder()
		rt.fn(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s cross-origin status = %d, want 403", rt.name, rec.Code)
		}
	}
}

func TestImageRouteRejectsMissingAndTraversal(t *testing.T) {
	t.Parallel()

	srv := &server{app: imagePage(t)}
	for _, path := range []string{"/missing", "/../logo", "/a/../../logo", "/."} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1"+path, nil)
		req.URL.Path = path
		srv.image(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
	}
}

func imagePage(t *testing.T) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{
		HTML:   `<p>img</p>`,
		Width:  200,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	return p
}
