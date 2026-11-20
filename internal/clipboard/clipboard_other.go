// android also matches the linux tag, and ios matches darwin.

//go:build (!linux && !windows && !darwin) || android || ios

package clipboard

import "context"

func writeOS(context.Context, string) bool { return false }

func readOS(context.Context) (string, bool) { return "", false }

func clearOS() {}

func diagnoseOS() Status {
	return Status{Backend: "none", Available: false, Reason: "no OS clipboard on this platform"}
}
