//go:build !windows

package window

// detectMonitorHz returns the default refresh rate on non-Windows targets
// without cgo. Ebiten's MonitorType does not expose refresh rate, and browser
// or mobile hosts pace frames natively.
func detectMonitorHz(_, _, _, _ int) int {
	return defaultMonitorHz
}
