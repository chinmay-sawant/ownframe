package replay

import (
	"errors"

	"golang.org/x/image/font/sfnt"
)

// MaxUserFontBytes caps a user-supplied font at 2 MiB.
// MaxUserFontGlyphs caps its glyph count.
const (
	MaxUserFontBytes  = 2 * 1024 * 1024
	MaxUserFontGlyphs = 32768
)

// ErrFontPolicy means a user font breaks the untrusted font policy.
var ErrFontPolicy = errors.New("ownframe: font rejected by policy")

// ValidateUserFont allows only small parseable fonts with a bounded
// glyph count. Bundled and system fonts skip this; call it on any
// font bytes that arrive from untrusted input before parsing.
func ValidateUserFont(data []byte) error {
	if len(data) == 0 || len(data) > MaxUserFontBytes {
		return ErrFontPolicy
	}
	font, err := sfnt.Parse(data)
	if err != nil {
		return ErrFontPolicy
	}
	if font.NumGlyphs() > MaxUserFontGlyphs {
		return ErrFontPolicy
	}
	return nil
}
