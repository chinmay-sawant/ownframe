package reload

import (
	"crypto/sha256"
	"os"
)

// Kind selects the watched HTML file or the theme file.
type Kind int

const (
	// HTML is the page template.
	HTML Kind = iota
	// Theme is the extra stylesheet.
	Theme
)

// Spec is the config and the bytes New already read.
type Spec struct {
	File      string
	ThemeFile string
	Hot       bool
	HTML      []byte
	HTMLInfo  os.FileInfo
	CSS       []byte
	CSSInfo   os.FileInfo
}

// State is one page's file watch. A nil State is off.
type State struct {
	enabled bool
	html    watched
	theme   watched
	pending *edit
}

type watched struct {
	path    string
	info    os.FileInfo
	data    []byte
	sum     [sha256.Size]byte
	statErr string
}

type edit struct {
	kind Kind
	data []byte
	info os.FileInfo
}

// Open builds a watch. It returns nil when neither path is set.
// Hot is ignored on wasm and mobile, where hosts is false.
func Open(spec Spec) *State {
	if spec.File == "" && spec.ThemeFile == "" {
		return nil
	}
	w := &State{enabled: spec.Hot && hosts}
	if spec.File != "" {
		w.html.path = spec.File
		w.html.accept(spec.HTML, spec.HTMLInfo)
	}
	if spec.ThemeFile != "" {
		w.theme.path = spec.ThemeFile
		w.theme.accept(spec.CSS, spec.CSSInfo)
	}
	return w
}

// SetHot turns the watch on or off. A nil State and a wasm or mobile
// host stay off.
func (s *State) SetHot(on bool) {
	if s == nil {
		return
	}
	s.enabled = on && hosts
}

// On reports whether a poll can reread a source file.
func (s *State) On() bool {
	if s == nil || !s.enabled {
		return false
	}
	return s.html.path != "" || s.theme.path != ""
}

func (s *State) at(kind Kind) *watched {
	if kind == Theme {
		return &s.theme
	}
	return &s.html
}

func (w *watched) accept(data []byte, info os.FileInfo) {
	w.data = data
	w.sum = sha256.Sum256(data)
	w.info = info
}
