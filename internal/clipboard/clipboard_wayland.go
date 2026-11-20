//go:build linux && !android

// Wayland clipboard uses wl-copy and wl-paste rather than a pure-Go Wayland
// wire client: core wl_data_device requires a focused surface and input serial
// owned by the windowing layer, while zwlr_data_control_v1 / ext_data_control_v1
// is not supported by GNOME Mutter. Delegating to wl-clipboard handles
// compositor quirks with zero extra Go module dependencies, and Diagnose /
// ErrWaylandToolsMissing surface missing helpers explicitly.
package clipboard

import (
	"context"
	"os/exec"
)

func wlCopyArgs() []string {
	return []string{"--type", "text/plain", "--"}
}

func wlClearArgs() []string {
	return []string{"--clear"}
}

func wlPasteArgs() []string {
	return []string{"--no-newline"}
}

func wlWrite(ctx context.Context, text string) bool {
	return runWL(ctx, "wl-copy", wlCopyArgs(), text)
}

func wlClear() {
	runWL(context.Background(), "wl-copy", wlClearArgs(), "")
}

func wlRead(ctx context.Context) (string, bool) {
	return outWL(ctx, "wl-paste", wlPasteArgs())
}

func lookWL(name string) (string, bool) {
	path, err := exec.LookPath(name)

	return path, err == nil
}
