// android also matches the linux tag, and ios matches darwin.

//go:build (!linux && !windows && !darwin) || android || ios

package clipboard

func writeOS(string) {}

func readOS() (string, bool) { return "", false }

func clearOS() {}
