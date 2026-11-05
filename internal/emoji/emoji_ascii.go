package emoji

// asciiLimit is the first byte value outside ASCII. UTF-8 lead bytes for
// non-ASCII runes all sit at or above it, so it marks the fast-path split.
const asciiLimit = 0x80

// emojiLeadF0 and emojiLeadE2 are the UTF-8 lead bytes that open every
// supported emoji sequence: F0 starts the four-byte supplementary symbols
// while E2 starts the three-byte miscellaneous and arrow blocks.
const (
	emojiLeadF0 = 0xF0
	emojiLeadE2 = 0xE2
)

// emojiSecond9F, emojiSecond9A, emojiSecond9C, emojiSecond9D, and
// emojiSecondAD are the UTF-8 second bytes that follow the lead above:
// 9F after F0 for the supplementary block, and 9A, 9C, 9D, or AD after
// E2 for the three-byte sequences Replace understands.
const (
	emojiSecond9F = 0x9F
	emojiSecond9A = 0x9A
	emojiSecond9C = 0x9C
	emojiSecond9D = 0x9D
	emojiSecondAD = 0xAD
)

// maybeHasEmoji reports whether s could contain a supported emoji sequence.
// Every supported sequence starts with E2 followed by 9A, 9C, 9D, or AD, or
// with F0 9F, so any page without those lead pairs skips the copy and the
// per-tag match in Replace. False positives only cost the full pass; a
// missing pair must never skip a supported sequence.
func maybeHasEmoji(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case emojiLeadF0:
			if i+1 < len(s) && s[i+1] == emojiSecond9F {
				return true
			}
		case emojiLeadE2:
			if i+1 < len(s) {
				switch s[i+1] {
				case emojiSecond9A, emojiSecond9C, emojiSecond9D, emojiSecondAD:
					return true
				}
			}
		}
	}
	return false
}
