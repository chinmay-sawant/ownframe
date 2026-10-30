//go:build !linux && !windows

package web

func runtimeRSS() (uint64, bool) {
	return 0, false
}
