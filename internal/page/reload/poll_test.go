package reload

import (
	"errors"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"
)

func TestReloadCleanStripsBreaks(t *testing.T) {
	t.Parallel()
	if got := Clean("a\r\nb"); got != "a  b" || strings.Contains(got, "\n") {
		t.Fatalf("clean %q", got)
	}
}

func TestWatchHotAndEmpty(t *testing.T) {
	t.Parallel()
	if Open(Spec{}).On() {
		t.Fatal("empty watch is on")
	}
	var none *State
	none.SetHot(true)
	if none.On() {
		t.Fatal("nil watch is on")
	}
	s := Open(Spec{File: "a.html", Hot: false, HTML: []byte("x")})
	if s.On() {
		t.Fatal("cold watch is on")
	}
	s.SetHot(true)
	if !s.On() {
		t.Fatal("hot watch is off")
	}
	s.SetHot(false)
	if s.On() {
		t.Fatal("watch stayed on")
	}
}

func TestReloadPollSwapHashAndRetry(t *testing.T) {
	t.Parallel()

	fs := newFake()
	fs.put("a.html", []byte("<p>one</p>"), 1)
	info, err := fs.Stat("a.html")
	if err != nil {
		t.Fatal(err)
	}
	s := Open(Spec{File: "a.html", Hot: true, HTML: []byte("<p>one</p>"), HTMLInfo: info})
	hooks, got := fs.hooks(64)

	applied, err := s.Poll(hooks)
	if applied || err != nil || got.templates != 0 || fs.reads != 0 {
		t.Fatalf("clean poll applied %v err %v templates %d reads %d", applied, err, got.templates, fs.reads)
	}

	fs.put("a.html", []byte("<p>one</p>"), 2)
	applied, err = s.Poll(hooks)
	if applied || err != nil || got.templates != 0 || fs.reads != 1 {
		t.Fatalf("same hash applied %v err %v templates %d reads %d", applied, err, got.templates, fs.reads)
	}
	applied, err = s.Poll(hooks)
	if applied || err != nil || fs.reads != 1 {
		t.Fatalf("stable stat applied %v err %v reads %d", applied, err, fs.reads)
	}

	fs.put("a.html", []byte("<p>two</p>"), 3)
	applied, err = s.Poll(hooks)
	if !applied || err != nil || got.templates != 1 || got.src != "<p>two</p>" {
		t.Fatalf("swap applied %v err %v templates %d src %q", applied, err, got.templates, got.src)
	}

	fs.put("a.html", []byte("{{"), 4)
	applied, err = s.Poll(hooks)
	if applied || err == nil || got.templates != 1 {
		t.Fatalf("bad parse applied %v err %v templates %d", applied, err, got.templates)
	}
	reads := fs.reads
	applied, err = s.Poll(hooks)
	if applied || err == nil || fs.reads != reads {
		t.Fatalf("retry applied %v err %v reads %d want %d", applied, err, fs.reads, reads)
	}

	fs.put("a.html", []byte("<p>ok</p>"), 5)
	applied, err = s.Poll(hooks)
	if !applied || err != nil || got.src != "<p>ok</p>" {
		t.Fatalf("fixed applied %v err %v src %q", applied, err, got.src)
	}
}

func TestReloadStatOnceAndOversize(t *testing.T) {
	t.Parallel()

	fs := newFake()
	fs.put("a.html", []byte("ok"), 1)
	info, err := fs.Stat("a.html")
	if err != nil {
		t.Fatal(err)
	}
	s := Open(Spec{File: "a.html", Hot: true, HTML: []byte("ok"), HTMLInfo: info})
	hooks, _ := fs.hooks(8)
	delete(fs.files, "a.html")

	applied, err := s.Poll(hooks)
	if applied || err == nil {
		t.Fatalf("missing applied %v err %v", applied, err)
	}
	applied, err = s.Poll(hooks)
	if applied || err != nil {
		t.Fatalf("repeat applied %v err %v", applied, err)
	}

	fs.put("a.html", []byte("12345"), 2)
	hooks, got := fs.hooks(4)
	s = Open(Spec{File: "a.html", Hot: true, HTML: []byte("ok"), HTMLInfo: info})
	applied, err = s.Poll(hooks)
	if applied || !errors.Is(err, errBig) || got.templates != 0 {
		t.Fatalf("oversize applied %v err %v", applied, err)
	}
	applied, err = s.Poll(hooks)
	if applied || !errors.Is(err, errBig) {
		t.Fatalf("oversize repeat applied %v err %v", applied, err)
	}

	fs.put("a.html", []byte("1234"), 3)
	applied, err = s.Poll(hooks)
	if !applied || err != nil || got.src != "1234" {
		t.Fatalf("limit applied %v err %v src %q", applied, err, got.src)
	}
}

func TestReloadThemeAndHTML(t *testing.T) {
	t.Parallel()

	fs := newFake()
	fs.put("a.html", []byte("<p>one</p>"), 1)
	fs.put("t.css", []byte("a"), 1)
	hi, err := fs.Stat("a.html")
	if err != nil {
		t.Fatal(err)
	}
	ti, err := fs.Stat("t.css")
	if err != nil {
		t.Fatal(err)
	}
	s := Open(Spec{
		File: "a.html", ThemeFile: "t.css", Hot: true,
		HTML: []byte("<p>one</p>"), HTMLInfo: hi,
		CSS: []byte("a"), CSSInfo: ti,
	})
	hooks, got := fs.hooks(8)
	fs.put("a.html", []byte("{{"), 2)
	fs.put("t.css", []byte("b"), 2)
	applied, err := s.Poll(hooks)
	if !applied || err == nil || got.templates != 0 || got.themes != 1 || got.css != "b" {
		t.Fatalf("mixed applied %v err %v tpl %d theme %d", applied, err, got.templates, got.themes)
	}
	fs.reads = 0
	applied, err = s.Poll(hooks)
	if applied || err == nil || fs.reads == 0 {
		t.Fatalf("html retried applied %v err %v reads %d", applied, err, fs.reads)
	}
}

func TestReloadPathNewline(t *testing.T) {
	t.Parallel()
	s := Open(Spec{File: "a\nb", Hot: true, HTML: []byte("x")})
	fs := newFake()
	hooks, _ := fs.hooks(8)
	_, err := s.Poll(hooks)
	if err == nil || strings.Contains(err.Error(), "\n") || !strings.Contains(err.Error(), "a b") {
		t.Fatalf("err %v", err)
	}
}

var errBig = errors.New("big")

type fakeInfo struct {
	name string
	size int64
	mod  time.Time
}

func (f fakeInfo) Name() string       { return f.name }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() os.FileMode  { return 0o600 }
func (f fakeInfo) ModTime() time.Time { return f.mod }
func (f fakeInfo) IsDir() bool        { return false }
func (f fakeInfo) Sys() any           { return nil }

type fakeFS struct {
	files map[string][]byte
	mod   map[string]time.Time
	reads int
}

func newFake() *fakeFS {
	return &fakeFS{files: map[string][]byte{}, mod: map[string]time.Time{}}
}

func (f *fakeFS) put(path string, data []byte, mod int64) {
	f.files[path] = append([]byte(nil), data...)
	f.mod[path] = time.Unix(mod, 0)
}

func (f *fakeFS) Stat(path string) (os.FileInfo, error) {
	b, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return fakeInfo{name: path, size: int64(len(b)), mod: f.mod[path]}, nil
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	f.reads++
	b, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), b...), nil
}

func (f *fakeFS) hooks(limit int64) (Hooks, *installCount) {
	got := &installCount{}
	return Hooks{
		Files: f,
		Limit: limit,
		Oversize: func(string, int64) error {
			return errBig
		},
		Template: func(tpl *template.Template, src string) error {
			if tpl == nil {
				return errors.New("nil template")
			}
			got.templates++
			got.src = src
			return nil
		},
		Theme: func(css string) error {
			got.themes++
			got.css = css
			return nil
		},
	}, got
}

type installCount struct {
	templates int
	themes    int
	src       string
	css       string
}
