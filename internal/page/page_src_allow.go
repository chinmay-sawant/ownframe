package page

import (
	"errors"
	"fmt"
	"strings"
)

// MaxDataImageBytes caps one data:image/* URL at resolve time (GEN-30).
// The cap applies to the encoded URL, so decode can never exceed it.
const MaxDataImageBytes = 1 << 20

// ErrImageTooLarge means a data:image/* URL passed the size cap.
var ErrImageTooLarge = errors.New("ownframe: image data URL too large")

// checkImageSrc allowlists one image source at resolve time (GEN-30).
// App keys ("logo", "emoji/x") and http(s) pass; data:image/* passes
// within the size cap. file:, javascript:, and the rest are rejected
// before any registry or emoji lookup runs.
func checkImageSrc(src string) error {
	i := strings.Index(src, ":")
	if i < 0 {
		return nil
	}
	scheme := strings.ToLower(src[:i])
	switch scheme {
	case "http", "https":
		return nil
	case "data":
		if !strings.HasPrefix(strings.ToLower(src), "data:image/") {
			return fmt.Errorf("%w: data", ErrUnsupportedLinkScheme)
		}
		if len(src) > MaxDataImageBytes {
			return ErrImageTooLarge
		}
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, scheme)
	}
}
