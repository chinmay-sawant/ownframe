package window

import "time"

// Desktop present cap without VSync.
//
// The active drawing loop stays on VSync-off maximum and paces itself with a
// monotonic sleep, so arbitrary caps (60/120/144/240) work on any monitor
// refresh without busy-spinning. Presents (heavy Draw work) are capped to
// MaxFPS by pacing every loop iteration to the present interval: each
// iteration presents once, and the sleep holds the rate. The screen is
// retained (SetScreenClearedEveryFrame(false)), so idle keep-alive frames
// stay cheap. Updates run hotter via TPS headroom (MaxFPS+FPSHeadroom), so
// the FE shows MaxFPS but the app ticks at the loop rate for input
// smoothness.
//
// Idle parking is untouched: keep-alive wakes (125 ms) are far below the
// limiter intervals, so idle windows park efficiently with no extra sleep.
const (
	// DefaultMaxFPS is the suggested present cap (Settings/CLI choice).
	// The default runtime is uncapped: Options.MaxFPS zero means no limiter.
	DefaultMaxFPS = 60
	// FPSHeadroom is how much faster Updates run than presents when capped.
	// A 60 FPS cap shows 60 on the FE while the app ticks at 80.
	FPSHeadroom = 20
	// MinFPS and MaxAllowedFPS bound explicit caps; outside falls back to
	// uncapped (the default).
	MinFPS        = 30
	MaxAllowedFPS = 1000
)

// NormalizeMaxFPS maps an Options.MaxFPS value to its effective cap.
// 0 (the default) means uncapped. Negative means explicit uncapped (0).
// Positive values in [MinFPS, MaxAllowedFPS] are kept; anything else falls
// back to uncapped. The return is 0 for uncapped, else the present cap.
func NormalizeMaxFPS(v int) int {
	if v <= 0 {
		return 0
	}
	if v < MinFPS || v > MaxAllowedFPS {
		return 0
	}
	return v
}

// LoopFPS is the internal update (TPS) rate for a present cap: MaxFPS +
// headroom. Presents run at MaxFPS while Updates average the loop rate, so
// the FE shows MaxFPS FPS but the app ticks hotter for input smoothness.
// It returns 0 for uncapped.
func LoopFPS(maxFPS int) int {
	if maxFPS <= 0 {
		return 0
	}
	return maxFPS + FPSHeadroom
}

// PresentInterval is the minimum time between loop iterations (each does one
// heavy present). The loop is paced to the present cap; Updates run hotter
// via TPS headroom without extra presents.
func PresentInterval(maxFPS int) time.Duration {
	if maxFPS <= 0 {
		return 0
	}
	return time.Second / time.Duration(maxFPS)
}

// LoopInterval paces loop iterations to the present cap (no VSync sleep
// beyond the cap, no busy-spin).
func LoopInterval(maxFPS int) time.Duration {
	return PresentInterval(maxFPS)
}

// loopWait reports how long to sleep before the next loop iteration to hold
// the present cap. It is pure (no sleeping) so tests need no window.
func loopWait(maxFPS int, lastLoop, now time.Time) time.Duration {
	interval := LoopInterval(maxFPS)
	if interval <= 0 || lastLoop.IsZero() {
		return 0
	}
	if d := lastLoop.Add(interval).Sub(now); d > 0 {
		return d
	}
	return 0
}
