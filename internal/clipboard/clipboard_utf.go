package clipboard

import (
	"math"
	"unicode/utf16"
)

const utf16HiShift = 8

const utf16Unit = 2

// utf16Clipboard is CF_UNICODETEXT: UTF-16LE and a trailing zero word.
func utf16Clipboard(s string) []byte {
	u := utf16.Encode([]rune(s))
	if len(u) > (math.MaxInt-utf16Unit)/utf16Unit {
		return nil
	}
	b := make([]byte, (len(u)+1)*utf16Unit)

	for i, v := range u {
		b[i*utf16Unit] = byte(v) //nolint:gosec // reason: low byte of the UTF-16LE split.
		b[i*utf16Unit+1] = byte(v >> utf16HiShift)
	}

	return b
}
