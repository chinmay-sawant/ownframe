//go:build darwin && !ios && !cgo

package clipboard

import "context"

func writeOS(context.Context, string) bool { return false }

func readOS(context.Context) (string, bool) { return "", false }

func clearOS() {}

func diagnoseOS() Status {
	return Status{Backend: "none", Available: false, Reason: "macOS clipboard requires cgo"}
}
