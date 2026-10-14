package dashboard

import "strconv"

func itoa(value int) string { return strconv.Itoa(value) }

func duration(minutes int) string {
	if minutes < 60 {
		return itoa(minutes) + "m"
	}
	return itoa(minutes/60) + "h " + twoDigits(minutes%60) + "m"
}

func twoDigits(value int) string {
	if value < 10 {
		return "0" + itoa(value)
	}
	return itoa(value)
}
