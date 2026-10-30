//go:build !linux && !windows

package window

// perfRSS reports unknown resident memory on platforms without a reader.
// It returns 0 and the caller renders that as unknown, since genuine
// resident memory is never zero.
func perfRSS() uint64 {
	return 0
}
