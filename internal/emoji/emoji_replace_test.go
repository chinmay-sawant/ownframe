package emoji

import (
	"strings"
	"testing"
)

func TestReplaceBasic(t *testing.T) {
	t.Parallel()

	got, changed := Replace(`<p>Hi 😂!</p>`)
	want := `<p>Hi <img data-ownframe-emoji="1" src="emoji/1f602" alt="😂">!</p>`
	if !changed || got != want {
		t.Fatalf("got %q changed=%v", got, changed)
	}
}

func TestReplaceASCIIReturnsInput(t *testing.T) {
	t.Parallel()

	src := `<p>plain ascii</p>`
	got, changed := Replace(src)
	if changed || got != src {
		t.Fatalf("got %q changed=%v", got, changed)
	}
}

// TestReplaceTextValue checks that a bare text value turns into the image
// markup. A value the page patches reaches the tree as one text node, so
// the emoji has to be markup before the tree takes it.
func TestReplaceTextValue(t *testing.T) {
	t.Parallel()

	got, changed := Replace(`hi 😂!`)
	want := `hi <img data-ownframe-emoji="1" src="emoji/1f602" alt="😂">!`
	if !changed || got != want {
		t.Fatalf("got %q changed=%v", got, changed)
	}
}

func TestReplaceSkipsMarkup(t *testing.T) {
	t.Parallel()

	src := `<input value="😂"><a title="a>b 😂">x</a><!-- 😂 --><script>var e="😂";</script><style>.a{content:"😂"}</style>1 < 2 😂`
	got, changed := Replace(src)
	if !changed {
		t.Fatal("the trailing emoji did not replace")
	}

	for _, kept := range []string{`value="😂"`, `title="a>b 😂"`, `<!-- 😂 -->`, `var e="😂";`, `content:"😂"`, `1 < 2 `} {
		if !strings.Contains(got, kept) {
			t.Fatalf("lost %q in %q", kept, got)
		}
	}
}

func TestReplaceVariationAndUnknown(t *testing.T) {
	t.Parallel()

	got, _ := Replace(`❤️❤⚠️`)
	if strings.Count(got, "<img") != 3 {
		t.Fatalf("got %q", got)
	}

	got, changed := Replace(`🫠 plain`)
	if changed || got != `🫠 plain` {
		t.Fatalf("unsupported emoji changed: %q", got)
	}

	got, changed = Replace(`plain text`)
	if changed || got != `plain text` {
		t.Fatalf("plain text changed: %q", got)
	}
}

// TestReplaceGateSkipsNonEmoji proves the fast pre-check keeps color-emoji
// treatment for supported sequences while an emoji-free page skips the
// heavy pass without allocating.
func TestReplaceGateSkipsNonEmoji(t *testing.T) {
	got, changed := Replace(`<p>Hi 😂!</p>`)
	if !changed || !strings.Contains(got, `src="emoji/1f602"`) {
		t.Fatalf("emoji lost color treatment: %q changed=%v", got, changed)
	}

	for _, e := range all() {
		if !maybeHasEmoji(e.seq) {
			t.Fatalf("gate misses supported %q", e.seq)
		}
	}

	for _, src := range []string{`<p>plain ascii</p>`, `<p>ownframe · builder → HTML — café</p>`} {
		src := src
		if out, changed := Replace(src); changed || out != src {
			t.Fatalf("non-emoji changed: %q -> %q changed=%v", src, out, changed)
		}
		if allocs := testing.AllocsPerRun(100, func() { Replace(src) }); allocs != 0 {
			t.Fatalf("non-emoji %q allocated %v per run, want 0", src, allocs)
		}
	}
}
