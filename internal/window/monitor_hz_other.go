//go:build !windows && !(linux && !android) && !(darwin && !ios)

package window

// detectMonitorHz returns the default refresh rate on targets without a native
// desktop display probe (such as wasm, Android, and iOS), where the host paces
// frames natively.
func detectMonitorHz(_, _, _, _ int) int {
	return defaultMonitorHz
}
