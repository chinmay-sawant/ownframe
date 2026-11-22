//go:build !windows

package window

// nativeFitProbe has no implementation off Windows: the window managers there
// keep a window on screen themselves, and WSLg has its own placement.
func nativeFitProbe() (fitProbe, bool) { return fitProbe{}, false }

func nativeFitApply(fitRect) bool { return false }
