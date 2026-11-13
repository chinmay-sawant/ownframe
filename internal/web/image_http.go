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

	w.Header().Set("Content-Type", imageContentType(data))
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(data); err != nil {
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

// imageContentType labels whatever SetImage stored. The standard sniffer
// covers the common bitmap types. SVG and the ISO image brands are named
// here because that sniffer calls them text or octet-stream. Anything else
// is still served with the sniffed type, so a new format is not dropped.
func imageContentType(data []byte) string {
	if svgBytes(data) {
		return "image/svg+xml"
	}
	if ct, ok := boxedImage(data); ok {
		return ct
	}

	sniffed := http.DetectContentType(data)
	if strings.HasPrefix(sniffed, "image/") {
		return sniffed
	}

	return sniffed
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
	case "heic", "heix", "heif", "mif1":
		return "image/heic", true
	default:
		return "", false
	}
}
