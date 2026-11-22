package web

import (
	"bytes"
	"net/http"
	"strings"
)

// imageSource is a page that can hand back one registered picture.
type imageSource interface {
	ImageData(src string) ([]byte, bool)
}

// image serves a picture the page registered with SetImage.
// A stylesheet url("logo") and an img src="logo" both request /logo.
func (s *server) image(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	name, ok := imageName(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	src, ok := s.app.(imageSource)
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.mu.Lock()
	data, ok := src.ImageData(name)
	s.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	ct := imageContentType(data)
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if ct == "image/svg+xml" {
		w.Header().Set("Content-Security-Policy", "sandbox")
	}
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(data); err != nil { //nolint:gosec // reason: data is first-party SetImage bytes keyed by validated imageName, Content-Type is allowlisted.
		return
	}
}

func imageName(path string) (string, bool) {
	name := strings.TrimPrefix(path, "/")
	if name == "" || strings.Contains(name, "\\") || strings.ContainsRune(name, 0) || strings.Contains(name, "://") {
		return "", false
	}

	for part := range strings.SplitSeq(name, "/") {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}

	return name, true
}

// imageContentType labels whatever SetImage stored using an explicit image
// MIME allowlist. Unrecognized payloads fall back to application/octet-stream
// so non-image bytes are never served as text/html.
func imageContentType(data []byte) string {
	if svgBytes(data) {
		return "image/svg+xml"
	}
	if ct, ok := boxedImage(data); ok {
		return ct
	}

	switch sniffed := http.DetectContentType(data); sniffed {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp",
		"image/x-icon", "image/vnd.microsoft.icon":
		return sniffed
	default:
		return "application/octet-stream"
	}
}

func svgBytes(data []byte) bool {
	head := bytes.TrimSpace(data)
	const sniffCap = 512
	if len(head) > sniffCap {
		head = head[:sniffCap]
	}

	return bytes.Contains(bytes.ToLower(head), []byte("<svg"))
}

func boxedImage(data []byte) (string, bool) {
	const mark = "ftyp"
	at := bytes.Index(data, []byte(mark))
	brandLen := len("avif")
	if at < 0 || at > len(mark)*2 || len(data) < at+len(mark)+brandLen {
		return "", false
	}

	switch string(data[at+len(mark) : at+len(mark)+brandLen]) {
	case "avif", "avis":
		return "image/avif", true
	default:
		return "", false
	}
}
