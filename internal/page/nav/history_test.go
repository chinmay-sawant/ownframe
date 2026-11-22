package nav

import (
	"hash/fnv"
	"strings"
	"testing"
)

func TestSourceRecordsHashAndTemplateMode(t *testing.T) {
	t.Parallel()

	for _, html := range []string{"", "{", "}}", "{{", "{{{", "a{{b}}c", "<p>plain</p>", "<p>{{.X}}</p>"} {
		h := fnv.New64a()
		_, _ = h.Write([]byte(html))
		template := Source(html, false)
		if template.HTML != html || template.Hash != h.Sum64() || template.Static ||
			template.Execute != strings.Contains(html, "{{") {
			t.Fatalf("template entry for %q = %+v", html, template)
		}
		static := Source(html, true)
		if static.HTML != html || static.Hash != template.Hash || !static.Static || static.Execute {
			t.Fatalf("static entry for %q = %+v", html, static)
		}
		if static.SameSource(template) || !static.SameSource(Source(html, true)) {
			t.Fatalf("source comparison ignored mode for %q", html)
		}
	}
}

func TestRetargetKeepsBothHistoryDirections(t *testing.T) {
	t.Parallel()

	first := Source("<p>first</p>", false)
	second := Source("<p>second</p>", false)
	third := Source("<p>third</p>", false)
	h := New(first)
	h.Push(second)
	h.Push(third)
	h.GoTo(1)

	// Adoption can supply the hash it already computed.
	replacement := Entry{HTML: "<p>{{.X}}</p>", Hash: 42, Static: true}
	h.Retarget(replacement)
	if h.Current() != replacement || h.Doc() != 1 || !h.CanBack() || !h.CanForward() {
		t.Fatalf("Retarget changed position or discarded the supplied entry: %+v", h.Current())
	}
	if h.At(h.BackIndex()) != first || h.At(h.ForwardIndex()) != third {
		t.Fatal("Retarget changed adjacent entries")
	}

	h.Push(second)
	if h.Current() != second || h.CanForward() || h.At(h.BackIndex()) != replacement {
		t.Fatal("Push did not replace the forward entry")
	}
}
