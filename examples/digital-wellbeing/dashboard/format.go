package dashboard

import "strconv"

func itoa(value int) string { return strconv.Itoa(value) }

func duration(minutes int) string {
	if minutes < minutesPerHour {
		return itoa(minutes) + "m"
	}
	return itoa(minutes/minutesPerHour) + "h " + twoDigits(minutes%minutesPerHour) + "m"
}

func twoDigits(value int) string {
	if value < padLimit {
		return "0" + itoa(value)
	}
	return itoa(value)
}
