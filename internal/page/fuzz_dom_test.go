package page

import (
	"errors"
	"strings"
	"testing"
)

// FuzzDOMLimits feeds markup to the depth/node guard (GEN-24). Seeds
// cover flat, nested, void, comment, and attribute-quote cases.
// Markup the guard documents as over-depth or over-size must error.
func FuzzDOMLimits(f *testing.F) {
	for _, seed := range []string{"<p>x</p>", "<div><div></div></div>", "<br><img src=\"a>b\">", "<!-- <div> -->", "<!DOCTYPE html>", strings.Repeat("<div>", 300), strings.Repeat("<i>x</i>", 10)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		err := checkDOMLimits(s)
		if want := domCapError(s); want != nil && !errors.Is(err, want) {
			t.Fatalf("len %d: got %v, want %v", len(s), err, want)
		}
	})
}

// domCapError reports the cap checkDOMLimits documents once a prefix of
// attribute-free tags passes MaxDOMDepth or MaxDOMNodes. It returns nil
// at the first attribute or unfinished tag, so it under-counts rather
// than failing a short benign string. Reaching either cap takes at least
// MaxDOMDepth "<a>" tags plus a final "<a".
func domCapError(s string) error {
	if len(s) < MaxDOMDepth*len("<a>")+len("<a") {
		return nil
	}
	depth, nodes := 0, 0
	for i := 0; i < len(s); {
		if s[i] != '<' || i+1 >= len(s) {
			i++
			continue
		}
		if s[i+1] == '!' || s[i+1] == '?' {
			end, ok := domComment(s, i)
			if !ok {
				return nil
			}
			i = end
			continue
		}
		name, closing, self, end, ok := simpleTag(s, i)
		if !ok {
			return nil
		}
		i = end
		if name == "" {
			continue
		}
		if closing {
			if depth > 0 {
				depth--
			}
			continue
		}
		nodes++
		if nodes > MaxDOMNodes {
			return ErrTooManyNodes
		}
		if self || isVoidTag(name) {
			continue
		}
		depth++
		if depth > MaxDOMDepth {
			return ErrDOMTooDeep
		}
	}

	return nil
}

// simpleTag parses one attribute-free tag: <name>, </name>, or <name/>.
// ok is false when the tag is unfinished or has attributes, so the caller
// does not read a later byte as a new tag.
func simpleTag(s string, i int) (name string, closing, self bool, end int, ok bool) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		closing = true
		j++
	}
	from := j
	for j < len(s) && domNameByte(s[j]) {
		j++
	}
	nameEnd := j
	if j < len(s) && s[j] == '/' {
		self = true
		j++
	}
	if j >= len(s) || s[j] != '>' {
		return "", false, false, 0, false
	}

	return strings.ToLower(s[from:nameEnd]), closing, self, j + 1, true
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
