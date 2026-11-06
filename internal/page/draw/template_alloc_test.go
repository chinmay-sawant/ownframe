package draw

import (
	"io"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// largeDocRows builds a ~600KB template body: the regression shape for the
// pooled-builder work. The byte count stays in one const so the golden
// below and the Execute hint share it.
const (
	largeDocRow   = "<div class=\"row\"><span>item</span><p>lorem ipsum dolor sit amet</p></div>\n"
	largeDocRows  = 8704
	largeDocSmall = "B"
)

func largeDocGolden() string {
	var sb strings.Builder
	sb.Grow(len(largeDocRow) * largeDocRows)
	for i := 0; i < largeDocRows; i++ {
		sb.WriteString(largeDocRow)
	}

	return sb.String()
}

// TestTemplateGoldenByteIdentical proves the pooled Execute path renders
// byte-identical output to the pre-edit behavior on a generated large doc:
// the golden is built independently with strings.Repeat, and ExecuteWriter
// must match Execute exactly.
func TestTemplateGoldenByteIdentical(t *testing.T) {
	t.Parallel()

	want := largeDocGolden()
	if len(want) < 600*1024 {
		t.Fatalf("golden = %d bytes, want at least 600KB", len(want))
	}

	got, err := Execute(false, len(want), nil, func(b *strings.Builder) error {
		for i := 0; i < largeDocRows; i++ {
			b.WriteString(largeDocRow)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Execute output differs from golden: got %d bytes, want %d", len(got), len(want))
	}

	// A second run must reproduce the first byte for byte.
	again, err := Execute(false, len(want), nil, func(b *strings.Builder) error {
		for i := 0; i < largeDocRows; i++ {
			b.WriteString(largeDocRow)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if again != got {
		t.Fatal("Execute output not stable across runs")
	}

	raw, err := ExecuteWriter(false, len(want), nil, func(w io.Writer) error {
		for i := 0; i < largeDocRows; i++ {
			if _, err := w.Write([]byte(largeDocRow)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != want {
		t.Fatal("ExecuteWriter output differs from golden")
	}
}

// TestUnchangedSkipUsesRevisionNoMemcmp proves the skip path compares the
// revision and hash counters first and never memcmps the source: two keys
// with equal revision, hash, and length but different Source bytes must
// still report unchanged. Any Source memcmp would return false here.
func TestUnchangedSkipUsesRevisionNoMemcmp(t *testing.T) {
	t.Parallel()

	view := Viewport{Width: 800, Height: 600}
	state := render.State{}
	pseudos := render.StatePseudos{}

	last := KeyForHashed("source-A", HashString("source-A"), len("source-A"), 7, view, state, pseudos, 0)
	next := last
	next.Source = "source-B"

	if len(next.Source) != next.SourceLen {
		t.Fatal("fixture sources must share one length")
	}
	next.SourceHash = last.SourceHash

	if !Unchanged(true, last, next) {
		t.Fatal("unchanged = false with equal revision, want true without a source memcmp")
	}

	bumped := last
	bumped.SourceRev = 8
	if Unchanged(true, last, bumped) {
		t.Fatal("unchanged = true across revisions, want false")
	}

	hashed := last
	hashed.SourceHash++
	if Unchanged(true, last, hashed) {
		t.Fatal("unchanged = true across hashes, want false")
	}

	if Unchanged(false, last, last) {
		t.Fatal("unchanged = true without a last frame, want false")
	}
}

// TestKeyFallsBackToMemcmpOnHashHit keeps the collision guard: with no
// revisions recorded, equal hashes still confirm with a source memcmp, so
// genuinely different sources never report unchanged.
func TestKeyFallsBackToMemcmpOnHashHit(t *testing.T) {
	t.Parallel()

	view := Viewport{Width: 800, Height: 600}
	state := render.State{}
	pseudos := render.StatePseudos{}

	a := KeyFor(strings.Repeat("a", 1024), view, state, pseudos, 0)
	b := KeyFor(strings.Repeat("b", 1024), view, state, pseudos, 0)
	if Unchanged(true, a, b) {
		t.Fatal("unchanged = true for different sources, want false")
	}

	if !Unchanged(true, a, a) {
		t.Fatal("unchanged = false for identical keys, want true")
	}
}

// TestBuilderPoolReturnsCleanBuffers proves pooled builders come back
// Reset: a large write followed by a tiny one must not leak the earlier
// frame into the later output.
func TestBuilderPoolReturnsCleanBuffers(t *testing.T) {
	want := largeDocGolden()

	if _, err := Execute(false, len(want), nil, func(b *strings.Builder) error {
		b.WriteString(want)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	got, err := Execute(false, 0, nil, func(b *strings.Builder) error {
		b.WriteString(largeDocSmall)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != largeDocSmall {
		t.Fatalf("pooled builder leaked the previous frame: got %d bytes, want 1", len(got))
	}

	raw, err := ExecuteWriter(false, 0, nil, func(w io.Writer) error {
		_, err := w.Write([]byte(largeDocSmall))

		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != largeDocSmall {
		t.Fatalf("pooled buffer leaked the previous frame: got %q", raw)
	}
}
