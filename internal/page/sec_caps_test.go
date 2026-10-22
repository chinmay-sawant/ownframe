package page

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestSec_PageFieldCaps pins GEN-28: the author maxlength is honored and
// the default cap bounds fields without one.
func TestSec_PageFieldCaps(t *testing.T) {
	t.Parallel()
	p, err := New(Config{HTML: `<input id="a" maxlength="3" value=""><input id="b" value="">`, Width: 10, Height: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.SetFormValue("a", "abcdef")
	if got := p.FormValue("a"); got != "abc" {
		t.Fatalf("maxlength: %q", got)
	}
	p.SetFormValue("b", strings.Repeat("x", DefaultMaxFieldLen+100))
	if got := p.FormValue("b"); len([]rune(got)) != DefaultMaxFieldLen {
		t.Fatalf("default cap: %d", len([]rune(got)))
	}
	var zero Page
	next, _, _, changed := zero.insertValue(Control{MaxLength: 2}, 0, 0, "abcdef")
	if !changed || next.Value != "ab" {
		t.Fatalf("insert: %q %v", next.Value, changed)
	}
}

// TestSec_PageSrcAllowlist pins GEN-30: blocked schemes fail at resolve
// time, oversized data:image fails, and the allowlist passes.
func TestSec_PageSrcAllowlist(t *testing.T) {
	t.Parallel()
	for _, src := range []string{"file:///etc/passwd", "javascript:alert(1)", "data:text/plain,hi", "vbscript:x"} {
		if err := checkImageSrc(src); !errors.Is(err, ErrUnsupportedLinkScheme) {
			t.Fatalf("%q: %v", src, err)
		}
	}
	if err := checkImageSrc("data:image/png;base64," + strings.Repeat("A", MaxDataImageBytes)); !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("oversize: %v", err)
	}
	for _, src := range []string{"logo", "emoji/smile", "https://x/y.png", "http://x/y.png", "data:image/png;base64,AA=="} {
		if err := checkImageSrc(src); err != nil {
			t.Fatalf("%q: %v", src, err)
		}
	}
	if _, err := parseOutboundLink("javascript:alert(1)"); !errors.Is(err, ErrUnsupportedLinkScheme) {
		t.Fatalf("href: %v", err)
	}
}
