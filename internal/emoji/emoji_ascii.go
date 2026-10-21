package emoji

// asciiLimit is the first byte value outside ASCII. UTF-8 lead bytes for
// non-ASCII runes all sit at or above it, so it marks the fast-path split.
const asciiLimit = 0x80

// hasNonASCII reports whether s contains a byte outside ASCII. Every
// supported emoji starts with a non-ASCII rune, so an ASCII page skips the
// copy and the per-byte match in Replace. This helper is that scan.
func hasNonASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= asciiLimit {
			return true
		}
	}
	return false
}
