package dashboard

import "strconv"

func itoa(value int) string { return strconv.Itoa(value) }

func twoDigits(value int) string {
	if value < 10 {
		return "0" + itoa(value)
	}
	return itoa(value)
}
