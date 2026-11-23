package reload

import (
	"crypto/sha256"
	"fmt"
	"os"
)

// Files stats and reads one watched path.
type Files interface {
	Stat(string) (os.FileInfo, error)
	ReadFile(string) ([]byte, error)
}

// Hooks is the reader and the install callbacks for one poll.
// Oversize builds the page's source-too-large error. HTML and Theme
// validate and install the source for the changed file.
type Hooks struct {
	Files    Files
	Limit    int64
	Oversize func(path string, size int64) error
	HTML     func(string) error
	Theme    func(string) error
}

// Poll rereads each changed file. HTML and theme are polled independently, so
// a syntax error in one never starves a reload of the other. A file whose
// install failed keeps its error until its stat changes; unchanged broken
// bytes are not parsed again. It reports whether a template or theme was
// installed, and the first error. The caller redraws only when applied is
// true.
func (s *State) Poll(h Hooks) (bool, error) {
	if s == nil || !s.On() {
		return false, nil
	}
	applied := false
	var first error
	for _, kind := range []Kind{HTML, Theme} {
		ok, err := s.one(kind, h)
		if ok {
			applied = true
		}
		if err != nil && first == nil {
			first = err
		}
	}
	return applied, first
}

// one rereads one path when its stat changed and its hash differs.
func (s *State) one(kind Kind, h Hooks) (bool, error) {
	wp := s.at(kind)
	if wp.path == "" {
		return false, nil
	}
	info, err := h.Files.Stat(wp.path)
	if err != nil {
		return false, noteStat(wp, err)
	}
	wp.statErr = ""
	if pend := s.pending[kind]; pend != nil {
		if sameStat(info, pend.info) {
			return false, pathErr(wp.path, pend.err)
		}
		s.pending[kind] = nil
	}
	if sameStat(info, wp.info) {
		return false, nil
	}
	if info.Size() > h.Limit {
		return false, noteStat(wp, h.Oversize(wp.path, info.Size()))
	}
	data, err := h.Files.ReadFile(wp.path)
	if err != nil {
		return false, noteStat(wp, err)
	}
	if int64(len(data)) > h.Limit {
		return false, noteStat(wp, h.Oversize(wp.path, int64(len(data))))
	}
	if wp.info != nil && wp.sum == sha256.Sum256(data) {
		wp.info = info
		return false, nil
	}
	return s.apply(kind, data, info, h)
}

// apply passes the source to the install callback. An error remembers the
// failed stat and leaves the last accepted file in place.
func (s *State) apply(kind Kind, data []byte, info os.FileInfo, h Hooks) (bool, error) {
	wp := s.at(kind)
	if err := install(kind, data, h); err != nil {
		s.pending[kind] = &edit{info: info, err: err}
		return false, pathErr(wp.path, err)
	}
	s.pending[kind] = nil
	wp.accept(data, info)
	return true, nil
}

func install(kind Kind, data []byte, h Hooks) error {
	if kind != HTML {
		return h.Theme(string(data))
	}
	return h.HTML(string(data))
}

// noteStat reports a stat or read error once per distinct message.
func noteStat(wp *watched, err error) error {
	if wp.statErr == err.Error() {
		return nil
	}
	wp.statErr = err.Error()
	return pathErr(wp.path, err)
}

func pathErr(path string, err error) error {
	return fmt.Errorf("%s: %w", Clean(path), err)
}

func sameStat(a, b os.FileInfo) bool {
	if a == nil || b == nil {
		return false
	}
	return a.ModTime().Equal(b.ModTime()) && a.Size() == b.Size()
}
