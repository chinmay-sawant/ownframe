//go:build !windows && (!linux || !cgo)

package window

type nativeWindowDrag struct{}

func (*nativeWindowDrag) begin()              {}
func (*nativeWindowDrag) move() (bool, error) { return false, nil }

func (*nativeWindowDrag) end() {}

func (*nativeWindowDrag) managed() bool { return false }

func setNativeWindowRect(x, y, w, h int) bool { return false }

func nativeWindowPosition() (int, int, bool) { return 0, 0, false }

func nativeWindowGeometry() (workarea, bool) { return workarea{}, false }

func (*nativeWindowDrag) cursor() (int, int, bool) { return 0, 0, false }

func nativeWindowTarget() (hostWindowTarget, bool) { return hostWindowTarget{}, false }
