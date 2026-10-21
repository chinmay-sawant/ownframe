package emoji

// hasNonASCII reports whether s contains a byte outside ASCII. Every
// supported emoji starts with a non-ASCII rune, so an ASCII page needs no
// scan or copy in Replace.
func hasNonASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return true
		}
	}
	return false
}
