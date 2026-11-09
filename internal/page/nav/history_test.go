package nav

import (
	"strings"
	"testing"
)

// navCorpus covers the delimiter shapes the history seam must preserve:
// empty, single braces, split and adjacent pairs, and ordinary documents.
var navCorpus = []string{
	"",
	"{",
	"}}",
	"{{",
	"{{{",
	"a{{b}}c",
	"{x}",
	"<p>plain</p>",
	"{{.X}}",
	"<!doctype html><html><body><p>{x}</p></body></html>",
}

// TestScanMatchesHashAndDelim proves Scan, Hash, and HasDelim report one
// verdict: the caller-hash path (HasDelim) can replace the fused scan.
func TestScanMatchesHashAndDelim(t *testing.T) {
	t.Parallel()

	for _, s := range navCorpus {
		hash, delim := Scan(s)
		if hash != Hash(s) {
			t.Fatalf("Scan(%q) hash = %d, Hash = %d", s, hash, Hash(s))
		}
		if delim != HasDelim(s) {
			t.Fatalf("Scan(%q) delim = %v, HasDelim = %v", s, delim, HasDelim(s))
		}
		if delim != strings.Contains(s, "{{") {
			t.Fatalf("HasDelim(%q) = %v, strings.Contains = %v", s, delim, strings.Contains(s, "{{"))
		}
	}
}

// TestRetargetHashedMatchesRetarget proves the caller-hash retarget stores
// the same source, hash, delimiter verdict, and index as Retarget.
func TestRetargetHashedMatchesRetarget(t *testing.T) {
	t.Parallel()

	const first = "<p>first</p>"
	const second = "<p>second</p>"

	for _, src := range navCorpus {
		want := New(first)
		want.Push(second)
		want.Retarget(src)

		got := New(first)
		got.Push(second)
		got.RetargetHashed(src, Hash(src))

		if got.Current() != want.Current() {
			t.Fatalf("Current(%q) = %q, want %q", src, got.Current(), want.Current())
		}
		if got.CurrentHash() != want.CurrentHash() {
			t.Fatalf("CurrentHash(%q) = %d, want %d", src, got.CurrentHash(), want.CurrentHash())
		}
		if got.CurrentDelim() != want.CurrentDelim() {
			t.Fatalf("CurrentDelim(%q) = %v, want %v", src, got.CurrentDelim(), want.CurrentDelim())
		}
		if got.Doc() != want.Doc() || got.CanBack() != want.CanBack() || got.CanForward() != want.CanForward() {
			t.Fatalf("RetargetHashed(%q) moved history: doc %d/%d back %v/%v forward %v/%v",
				src, got.Doc(), want.Doc(), got.CanBack(), want.CanBack(), got.CanForward(), want.CanForward())
		}
		if got.SourceAt(1) != src || got.HashAt(1) != Hash(src) || got.DelimAt(1) != HasDelim(src) {
			t.Fatalf("RetargetHashed(%q) stored %q hash=%d delim=%v",
				src, got.SourceAt(1), got.HashAt(1), got.DelimAt(1))
		}
	}
}

// TestRetargetHashedStoresCallerHash proves the hash comes from the caller:
// a hash Retarget would never produce is stored verbatim, and the source
// bytes and delimiter verdict are still recorded.
func TestRetargetHashedStoresCallerHash(t *testing.T) {
	t.Parallel()

	const replacement = "<div>{{.Y}}</div>"
	const callerHash = 42

	h := New("<p>start</p>")
	h.RetargetHashed(replacement, callerHash)

	if h.Current() != replacement {
		t.Fatalf("Current() = %q, want %q", h.Current(), replacement)
	}
	if h.CurrentHash() != callerHash {
		t.Fatalf("CurrentHash() = %d, want %d", h.CurrentHash(), callerHash)
	}
	if !h.CurrentDelim() {
		t.Fatal("CurrentDelim() = false, want the probe verdict")
	}
	if h.Doc() != 0 || h.CanBack() {
		t.Fatalf("RetargetHashed moved history: doc %d, canBack %v", h.Doc(), h.CanBack())
	}
}
