package page

import "fmt"

// sourceTooLarge builds the wrapped oversize error. It wraps both
// ErrBadSource and ErrSourceTooLarge so errors.Is matches either.
func sourceTooLarge(path string, size int64) error {
	return fmt.Errorf("%w: %w: %s: source too large (%d bytes, limit %d)", ErrBadSource, ErrSourceTooLarge, path, size, maxSourceSize)
}
