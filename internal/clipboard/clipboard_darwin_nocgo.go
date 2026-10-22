//go:build darwin && !ios && !cgo

package clipboard

func writeOS(string) {}

func readOS() (string, bool) { return "", false }

func clearOS() {}
