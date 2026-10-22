package page

import (
	"errors"
	"io/fs"
	"path"
	"strings"
)

// ErrBadAsset means an asset name escapes its directory.
var ErrBadAsset = errors.New("ownframe: bad asset name")

// AssetResolver serves img, font, and CSS bytes only from fsys.
// Names are slash paths rooted at the asset dir; absolute paths,
// backslashes, drive letters, null bytes, and dot-dot escapes fail.
func AssetResolver(fsys fs.FS) func(src string) ([]byte, error) {
	return func(src string) ([]byte, error) {
		clean, err := cleanAsset(src)
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

func cleanAsset(src string) (string, error) {
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
