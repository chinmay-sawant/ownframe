// Package source checks template text, DOM shape, and asset paths.
// It does not import package page. Callers pass bytes, strings, limits,
// and the sentinels the page shows.
package source

import (
	"errors"
	"fmt"
	"os"
)

const (
	// MaxBytes is the template, theme, and file source cap.
	MaxBytes = 8 << 20

	// MaxDOMDepth caps parse-time nesting (GEN-29): layout recursion past
	// it is a fatal stack overflow that recover cannot catch.
	MaxDOMDepth = 256

	// MaxDOMNodes caps the parse-time element count.
	MaxDOMNodes = 20000

	// MaxDataImageBytes caps one data:image/* URL at resolve time (GEN-30).
	// The cap applies to the encoded URL, so decode can never exceed it.
	MaxDataImageBytes = 1 << 20
)

var (
	// ErrTooLarge reports a source larger than the limit passed to Read.
	ErrTooLarge = errors.New("ownframe: source too large")

	// ErrDOMTooDeep means the template nests past the depth limit.
	ErrDOMTooDeep = errors.New("ownframe: template nests too deep")

	// ErrTooManyNodes means the template holds past the node limit.
	ErrTooManyNodes = errors.New("ownframe: template has too many elements")

	// ErrImageTooLarge means a data:image/* URL passed the size cap.
	ErrImageTooLarge = errors.New("ownframe: image data URL too large")

	// ErrBadAsset means an asset name escapes its directory.
	ErrBadAsset = errors.New("ownframe: bad asset name")
)

// Reader stats and reads one template or theme path.
type Reader interface {
	Stat(string) (os.FileInfo, error)
	ReadFile(string) ([]byte, error)
}

// OS reads the real filesystem. Symlinks are followed.
type OS struct{}

func (OS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func (OS) ReadFile(name string) ([]byte, error) {
	//nolint:gosec // reason: template files are trusted build input (documented trust boundary).
	return os.ReadFile(name)
}

// Text is one template or theme: an inline string, or a file, not both.
type Text struct {
	File      string
	Inline    string
	Limit     int64
	BadSource error
}

// Load returns Inline when File is empty. Both set is BadSource.
// A file is read under Limit.
func Load(files Reader, text Text) (string, os.FileInfo, error) {
	if text.File == "" {
		return text.Inline, nil, nil
	}
	if text.Inline != "" {
		return "", nil, text.BadSource
	}
	data, info, err := Read(files, text.File, text.Limit, text.BadSource)
	if err != nil {
		return "", nil, err
	}
	return string(data), info, nil
}

// Read reads one watched file. A missing or unreadable file is badSource
// wrapped with the path and the reason. Trusted-author paths only:
// symlinks are followed, so keep the path author-controlled.
// If untrusted paths are ever accepted, confine reads with os.Root,
// available since go 1.26.10 (this module's toolchain); nothing in
// internal/ uses it today.
func Read(files Reader, path string, limit int64, badSource error) ([]byte, os.FileInfo, error) {
	info, err := files.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %w", badSource, path, err)
	}
	if info.Size() > limit {
		return nil, info, Oversize(path, info.Size(), limit, badSource)
	}
	data, err := files.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %w", badSource, path, err)
	}
	if int64(len(data)) > limit {
		return nil, info, Oversize(path, int64(len(data)), limit, badSource)
	}
	if fresh, err := files.Stat(path); err == nil {
		info = fresh
	}
	return data, info, nil
}

// Oversize builds the wrapped oversize error. It wraps both badSource and
// ErrTooLarge so errors.Is matches either.
func Oversize(path string, size, limit int64, badSource error) error {
	return fmt.Errorf("%w: %w: %s: source too large (%d bytes, limit %d)", badSource, ErrTooLarge, path, size, limit)
}
