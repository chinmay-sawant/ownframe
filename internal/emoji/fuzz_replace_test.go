package emoji

import "testing"

// FuzzReplace feeds HTML and text fragments to the emoji replacer
// (GEN-24). Seeds cover supported runs, tags, comments, and script
// blocks that must pass through untouched.
func FuzzReplace(f *testing.F) {
	for _, seed := range []string{"hello", "❤️", "❤", "⚠️", "<b>❤️</b>", "<!-- ❤️ -->", "<script>❤️</script>", "<img alt=\"❤️\">", "🫠 plain", "❤️❤️❤️"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out, changed := Replace(s)
		again, changedAgain := Replace(s)
		if out != again || changed != changedAgain {
			t.Fatal("nondeterministic replace")
		}
		if changed == (out == s) {
			t.Fatal("changed/out mismatch")
		}
		if asciiOnly(s) && (changed || out != s) {
			t.Fatal("ascii input changed")
		}
	})
}

func asciiOnly(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}

	return true
}
