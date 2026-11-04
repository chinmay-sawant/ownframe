package emoji

// asciiLimit is the first byte value outside ASCII. UTF-8 lead bytes for
// non-ASCII runes all sit at or above it, so it marks the fast-path split.
const asciiLimit = 0x80

// maybeHasEmoji reports whether s could contain a supported emoji sequence.
// Every supported sequence starts with E2 followed by 9A, 9C, 9D, or AD, or
// with F0 9F, so any page without those lead pairs skips the copy and the
// per-tag match in Replace. False positives only cost the full pass; a
// missing pair must never skip a supported sequence.
func maybeHasEmoji(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 0xF0:
			if i+1 < len(s) && s[i+1] == 0x9F {
				return true
			}
		case 0xE2:
			if i+1 < len(s) {
				switch s[i+1] {
				case 0x9A, 0x9C, 0x9D, 0xAD:
					return true
				}
			}
		}
	}
	return false
}
