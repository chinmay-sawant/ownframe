package source

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const (
	wantMaxBytes         = 8 << 20
	wantDataImage        = 1 << 20
	wantDOMDepth         = 256
	wantDOMNodes         = 20000
	testReadLimit        = 4
	testDepthLimit       = 1
	testNodeLimit        = 1
	testFrame            = 10
	testMinWider         = 50
	testImageCap         = 8
	oversizeSize         = 9
	oversizeLimit        = 8
	statLie        int64 = 1
)

func TestSourceCaps(t *testing.T) {
	t.Parallel()
	if MaxBytes != wantMaxBytes || MaxDataImageBytes != wantDataImage {
		t.Fatalf("bytes %d image %d", MaxBytes, MaxDataImageBytes)
	}
	if MaxDOMDepth != wantDOMDepth || MaxDOMNodes != wantDOMNodes {
		t.Fatalf("dom %d %d", MaxDOMDepth, MaxDOMNodes)
	}
	if ErrTooLarge.Error() != "ownframe: source too large" {
		t.Fatal(ErrTooLarge)
	}
}

func TestSourceLoad(t *testing.T) {
	t.Parallel()
	bad := errors.New("ownframe: bad source")
	text, info, err := Load(OS{}, Text{Inline: "<p>x</p>", Limit: testReadLimit, BadSource: bad})
	if err != nil || text != "<p>x</p>" || info != nil {
		t.Fatalf("inline %v %q %v", err, text, info)
	}
	_, _, err = Load(OS{}, Text{File: "a", Inline: "b", Limit: testReadLimit, BadSource: bad})
	if err == nil || err.Error() != bad.Error() {
		t.Fatalf("both = %v", err)
	}
}

func TestSourceReadLimit(t *testing.T) {
	t.Parallel()
	bad := errors.New("ownframe: bad source")
	mem := NewMem()
	mem.Write("a.html", []byte("hi"))
	data, info, err := Read(mem, "a.html", testReadLimit, bad)
	if err != nil || string(data) != "hi" || info.Size() != int64(len("hi")) {
		t.Fatalf("read %v %q", err, data)
	}
	mem.Write("b.html", []byte("abcdef"))
	_, _, err = Read(mem, "b.html", testReadLimit, bad)
	if !errors.Is(err, bad) || !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	_, _, err = Read(mem, "missing.html", testReadLimit, bad)
	if !errors.Is(err, bad) || !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

type lieReader struct{}

func (lieReader) Stat(string) (os.FileInfo, error) { return lieInfo{}, nil }

func (lieReader) ReadFile(string) ([]byte, error) { return []byte("abcdef"), nil }

type lieInfo struct{}

func (lieInfo) Name() string       { return "x" }
func (lieInfo) Size() int64        { return statLie }
func (lieInfo) Mode() os.FileMode  { return 0 }
func (lieInfo) ModTime() time.Time { return time.Time{} }
func (lieInfo) IsDir() bool        { return false }
func (lieInfo) Sys() any           { return nil }

func TestSourceReadLenAfterStat(t *testing.T) {
	t.Parallel()
	bad := errors.New("ownframe: bad source")
	_, info, err := Read(lieReader{}, "x", testReadLimit, bad)
	if !errors.Is(err, bad) || !errors.Is(err, ErrTooLarge) || info == nil || info.Size() != statLie {
		t.Fatal(err)
	}
}

func TestSourceOversizeMessage(t *testing.T) {
	t.Parallel()
	bad := errors.New("ownframe: bad source")
	err := Oversize("theme", oversizeSize, oversizeLimit, bad)
	if !errors.Is(err, bad) || !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	const want = "ownframe: bad source: ownframe: source too large: theme: source too large (9 bytes, limit 8)"
	if err.Error() != want {
		t.Fatalf("%q", err.Error())
	}
}

func TestSourceMemModTimeMoves(t *testing.T) {
	t.Parallel()
	mem := NewMem()
	mem.Write("a", []byte("one"))
	first, err := mem.Stat("a")
	if err != nil {
		t.Fatal(err)
	}
	mem.Write("a", []byte("two!"))
	second, err := mem.Stat("a")
	if err != nil {
		t.Fatal(err)
	}
	if first.ModTime().Equal(second.ModTime()) || second.Size() != int64(len("two!")) {
		t.Fatalf("stat %v %d", second.ModTime(), second.Size())
	}
	got, err := mem.ReadFile("a")
	if err != nil {
		t.Fatal(err)
	}
	got[0] = 'X'
	again, err := mem.ReadFile("a")
	if err != nil || string(again) != "two!" {
		t.Fatalf("copy %q %v", again, err)
	}
}

func TestSourceOSRead(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "a.html")
	if err := os.WriteFile(path, []byte("<p>x</p>"), 0o600); err != nil {
		t.Fatal(err)
	}
	bad := errors.New("ownframe: bad source")
	data, info, err := Read(OS{}, path, MaxBytes, bad)
	if err != nil || string(data) != "<p>x</p>" || info.Size() != int64(len(data)) {
		t.Fatalf("%v %q", err, data)
	}
}

func TestDOMDepthAndNodes(t *testing.T) {
	t.Parallel()
	if err := CheckDOM("<div><div>", testDepthLimit, testFrame); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatal(err)
	}
	if err := CheckDOM("<i></i><i></i>", testFrame, testNodeLimit); !errors.Is(err, ErrTooManyNodes) {
		t.Fatal(err)
	}
	if err := CheckDOM("<br><br><div></div>", testDepthLimit, testFrame); err != nil {
		t.Fatal(err)
	}
	if err := CheckDOM("<div/><div></div>", testDepthLimit, testFrame); err != nil {
		t.Fatal(err)
	}
	if err := CheckDOM("<!-- <div><div> --><p></p>", testDepthLimit, testFrame); err != nil {
		t.Fatal(err)
	}
	if err := CheckDOM("<div>", 0, testFrame); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatal(err)
	}
}

func TestDOMScanHelpers(t *testing.T) {
	t.Parallel()
	if !Void("br") || Void("div") || Void("IMG") {
		t.Fatal("void")
	}
	if !NameByte('a') || !NameByte('0') || !NameByte('-') || NameByte(' ') {
		t.Fatal("name")
	}
	end, ok := SkipMeta("<!-- <div> -->tail", 0)
	if !ok || end != len("<!-- <div> -->") {
		t.Fatalf("%d %v", end, ok)
	}
	end, ok = SkipMeta("<!DOCTYPE html><p>", 0)
	if !ok || end != len("<!DOCTYPE html>") {
		t.Fatalf("decl %d %v", end, ok)
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()
	badSrc := errors.New("ownframe: bad source")
	badSize := errors.New("ownframe: bad size")
	ok := Check{HTML: "<p></p>", Width: testFrame, Height: testFrame}
	if err := Validate(ok, badSrc, badSize); err != nil {
		t.Fatal(err)
	}
	both := ok
	both.File = "a.html"
	if err := Validate(both, badSrc, badSize); err == nil || err.Error() != badSrc.Error() {
		t.Fatal(err)
	}
	theme := ok
	theme.Theme = "a"
	theme.ThemeFile = "a.css"
	if err := Validate(theme, badSrc, badSize); !errors.Is(err, badSrc) {
		t.Fatal(err)
	}
	wide := ok
	wide.Width = 0
	if err := Validate(wide, badSrc, badSize); !errors.Is(err, badSize) {
		t.Fatal(err)
	}
	capped := ok
	capped.MinWidth = testMinWider
	capped.MaxWidth = testFrame
	if err := Validate(capped, badSrc, badSize); !errors.Is(err, badSize) {
		t.Fatal(err)
	}
	zeroMin := ok
	zeroMin.MinWidth = 0
	zeroMin.MaxWidth = 1
	if err := Validate(zeroMin, badSrc, badSize); err != nil {
		t.Fatal(err)
	}
	if w, h := Mins(0, 0); w != 1 || h != 1 {
		t.Fatalf("mins %d %d", w, h)
	}
}

func TestAssetResolverRejectsEscape(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{"img/a.png": {Data: []byte("png")}}
	get := Resolver(fsys)
	data, err := get("img/a.png")
	if err != nil || string(data) != "png" {
		t.Fatalf("valid %v %q", err, data)
	}
	for _, bad := range []string{"", "../x", "a/../../x", "/abs", "a\\b", "C:x", "a\x00b"} {
		if _, err := get(bad); !errors.Is(err, ErrBadAsset) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
}

func TestAssetImageSrc(t *testing.T) {
	t.Parallel()
	bad := errors.New("ownframe: unsupported link scheme")
	for _, src := range []string{"file:///etc/passwd", "javascript:alert(1)", "data:text/plain,hi", "FILE:///x"} {
		err := ImageSrc(src, MaxDataImageBytes, bad)
		if !errors.Is(err, bad) {
			t.Fatalf("%q: %v", src, err)
		}
	}
	over := "data:image/png;base64," + strings.Repeat("A", testImageCap)
	if err := ImageSrc(over, testImageCap, bad); !errors.Is(err, ErrImageTooLarge) {
		t.Fatal(err)
	}
	for _, src := range []string{"logo", "emoji/smile", "https://x/y.png", "http://x/y.png", "data:image/png;base64,AA=="} {
		if err := ImageSrc(src, MaxDataImageBytes, bad); err != nil {
			t.Fatalf("%q: %v", src, err)
		}
	}
}

func TestTemplateFuncsReturnStrings(t *testing.T) {
	t.Parallel()
	strType := reflect.TypeOf("")
	funcs := SafeFuncs()
	upper, ok := funcs["upper"]
	if !ok {
		t.Fatal("missing upper")
	}
	typ := reflect.TypeOf(upper)
	if typ.NumOut() != 1 || typ.Out(0) != strType {
		t.Fatalf("upper returns %s", typ)
	}
	out := reflect.ValueOf(upper).Call([]reflect.Value{reflect.ValueOf("ab")})
	if out[0].String() != "AB" {
		t.Fatalf("upper = %s", out[0])
	}
}
