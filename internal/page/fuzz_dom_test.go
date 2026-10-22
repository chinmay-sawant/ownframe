package page

import (
	"strings"
	"testing"
)

// FuzzDOMLimits feeds markup to the depth/node guard (GEN-24). Seeds
// cover flat, nested, void, comment, and attribute-quote cases.
func FuzzDOMLimits(f *testing.F) {
	for _, seed := range []string{"<p>x</p>", "<div><div></div></div>", "<br><img src=\"a>b\">", "<!-- <div> -->", "<!DOCTYPE html>", strings.Repeat("<div>", 300), strings.Repeat("<i>x</i>", 10)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = checkDOMLimits(s)
	})
}

// FuzzImageSrc feeds image sources to the resolve-time allowlist
// (GEN-24). Blocked schemes must always error, whatever the case.
func FuzzImageSrc(f *testing.F) {
	for _, seed := range []string{"logo", "emoji/smile", "https://x/y.png", "data:image/png;base64,AA==", "file:///etc/passwd", "javascript:alert(1)", "FILE:///x", "data:text/plain,hi"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		err := checkImageSrc(s)
		low := strings.ToLower(s)
		if strings.HasPrefix(low, "file:") || strings.HasPrefix(low, "javascript:") {
			if err == nil {
				t.Fatalf("blocked scheme passed: %q", s)
			}
		}
	})
}
