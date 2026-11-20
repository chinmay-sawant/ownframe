//go:build darwin && !ios && !cgo

package clipboard

func writeOS(string) {}

func readOS() (string, bool) { return "", false }

func clearOS() {}

func diagnoseOS() Status {
	return Status{Backend: "none", Available: false, Reason: "macOS clipboard requires cgo"}
}
