package source

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Resolver serves img, font, and CSS bytes only from fsys.
// Names are slash paths rooted at the asset dir; absolute paths,
// backslashes, drive letters, null bytes, and dot-dot escapes fail.
func Resolver(fsys fs.FS) func(src string) ([]byte, error) {
	return func(src string) ([]byte, error) {
		clean, err := Clean(src)
		if err != nil {
			return nil, err
		}
		data, err := fs.ReadFile(fsys, clean)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

// Clean rejects an asset name that escapes its directory.
func Clean(src string) (string, error) {
	if src == "" || strings.ContainsRune(src, 0) {
		return "", ErrBadAsset
	}
	if strings.Contains(src, "\\") || path.IsAbs(src) {
		return "", ErrBadAsset
	}
	if len(src) >= 2 && src[1] == ':' {
		return "", ErrBadAsset
	}
	clean := path.Clean(src)
	if clean == "." || clean == ".." {
		return "", ErrBadAsset
	}
	if strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", ErrBadAsset
	}
	for _, part := range strings.Split(clean, "/") {
		if part == ".." || part == "" {
			return "", ErrBadAsset
		}
	}
	return clean, nil
}

// ImageSrc allowlists one image source at resolve time (GEN-30).
// App keys ("logo", "emoji/x") and http(s) pass; data:image/* passes
// within maxBytes. file:, javascript:, and the rest return unsupported
// before any registry or emoji lookup runs.
func ImageSrc(src string, maxBytes int, unsupported error) error {
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
			return fmt.Errorf("%w: data", unsupported)
		}
		if len(src) > maxBytes {
			return ErrImageTooLarge
		}
		return nil
	default:
		return fmt.Errorf("%w: %s", unsupported, scheme)
	}
}
