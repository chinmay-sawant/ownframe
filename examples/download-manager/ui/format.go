package ui

import (
	"fmt"
	"net/url"
	"time"
)

// formatBytes returns a short binary size such as "12.4 MiB". A negative
// size is unknown and prints "--".
func formatBytes(n int64) string {
	if n < 0 {
		return "--"
	}

	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// formatSpeed returns the transfer rate as a size per second, or "--" when
// the rate is zero or negative.
func formatSpeed(bps float64) string {
	if bps <= 0 {
		return "--"
	}

	return formatBytes(int64(bps+speedRound)) + "/s"
}

// formatETA returns a short remaining time. A negative duration is unknown.
func formatETA(d time.Duration) string {
	if d < 0 {
		return "--"
	}

	d = d.Round(time.Second)
	switch {
	case d < 10*time.Second:
		return "<10s"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm %02ds", int(d.Minutes()), int(d.Seconds())%secsPerMin)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh %02dm", int(d.Hours()), int(d.Minutes())%secsPerMin)
	default:
		return fmt.Sprintf("%dd %02dh", int(d.Hours())/hoursPerDay, int(d.Hours())%hoursPerDay)
	}
}

// redactURL removes the userinfo from a URL before display. An unparsable
// value is returned unchanged.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}

	u.User = nil

	return u.String()
}
