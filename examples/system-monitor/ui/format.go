package ui

import "fmt"

// units are the binary size suffixes formatBytes picks from.
var units = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}

// unavailable is the placeholder shown for a missing or denied value.
const unavailable = "n/a"

// formatBytes renders "12.3 MiB" and promotes the unit when rounding would
// read as 1024.0. ok=false means the sample is unavailable.
func formatBytes(v float64, ok bool) string {
	if !ok {
		return unavailable
	}

	if v < 0 {
		v = 0
	}

	i := 0
	for i < len(units)-1 && v >= 1023.95 {
		v /= 1024
		i++
	}

	if i == 0 {
		return fmt.Sprintf("%.0f %s", v, units[i])
	}

	return fmt.Sprintf("%.1f %s", v, units[i])
}

// formatRate renders a byte rate per second.
func formatRate(v float64, ok bool) string {
	if !ok {
		return unavailable
	}

	return formatBytes(v, true) + "/s"
}

// formatPercent renders a 0..1 fraction as a percentage.
func formatPercent(v float64, ok bool) string {
	if !ok {
		return unavailable
	}

	return fmt.Sprintf("%.1f%%", v*percentScale)
}

// cpuText renders a process CPU value, already a percent of one core, so a
// process busy on several cores can read above 100%.
func cpuText(v float64, ok bool) string {
	if !ok {
		return unavailable
	}

	return fmt.Sprintf("%.1f%%", v)
}
