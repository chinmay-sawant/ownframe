package reload

import (
	"crypto/sha256"
	"fmt"
	"html/template"
	"os"
)

// Files stats and reads one watched path.
type Files interface {
	Stat(string) (os.FileInfo, error)
	ReadFile(string) ([]byte, error)
}

// Hooks is the reader and the install callbacks for one poll.
// Oversize builds the page's source-too-large error. Template receives
// a parsed page template. Theme receives the stylesheet bytes as text.
type Hooks struct {
	Files    Files
	Limit    int64
	Oversize func(path string, size int64) error
	Template func(*template.Template, string) error
	Theme    func(string) error
}

// Poll rereads a changed file, or retries a parse that failed last time.
// It reports whether a template or theme was installed, and the first
// error. A later successful install in the same poll clears a pending
// edit from the earlier kind. The caller redraws only when installed
// is true.
func (s *State) Poll(h Hooks) (bool, error) {
	if s == nil || !s.On() {
		return false, nil
	}
	if s.pending != nil {
		return s.retry(h)
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

// retry reapplies the bytes a parse error kept. A later stat change
// drops those bytes and reads the path again.
func (s *State) retry(h Hooks) (bool, error) {
	pend := s.pending
	wp := s.at(pend.kind)
	info, err := h.Files.Stat(wp.path)
	if err != nil {
		return false, noteStat(wp, err)
	}
	wp.statErr = ""
	if !sameStat(info, pend.info) {
		s.pending = nil
		return s.one(pend.kind, h)
	}
	return s.apply(pend.kind, pend.data, info, h)
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

// apply parses HTML before the install callback. A parse or install
// error keeps the bytes pending and leaves the last accepted file in place.
func (s *State) apply(kind Kind, data []byte, info os.FileInfo, h Hooks) (bool, error) {
	wp := s.at(kind)
	if err := install(kind, data, h); err != nil {
		s.pending = &edit{kind: kind, data: data, info: info}
		return false, pathErr(wp.path, err)
	}
	s.pending = nil
	wp.accept(data, info)
	return true, nil
}

func install(kind Kind, data []byte, h Hooks) error {
	if kind != HTML {
		return h.Theme(string(data))
	}
	src := string(data)
	tpl, err := template.New("page").Parse(src)
	if err != nil {
		return err
	}
	return h.Template(tpl, src)
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
