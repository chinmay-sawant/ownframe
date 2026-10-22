//go:build linux && !android

package clipboard

import "os/exec"

func wlCopyArgs() []string {
	return []string{"--type", "text/plain", "--"}
}

func wlClearArgs() []string {
	return []string{"--clear"}
}

func wlPasteArgs() []string {
	return []string{"--no-newline"}
}

func wlWrite(text string) {
	runWL("wl-copy", wlCopyArgs(), text)
}

func wlClear() {
	runWL("wl-copy", wlClearArgs(), "")
}

func wlRead() (string, bool) {
	return outWL("wl-paste", wlPasteArgs())
}

func lookWL(name string) (string, bool) {
	path, err := exec.LookPath(name)

	return path, err == nil
}
