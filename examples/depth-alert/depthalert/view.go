package depthalert

import "fmt"

type view struct {
	Scanning bool
	State    string
	Distance string
	Coverage string
}

func distance(mm int32) string {
	if mm <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f m", float64(mm)/mmPerM)
}

func coverage(percent int32) string { return fmt.Sprintf("%d%%", percent) }
