package ui

import (
	"strconv"
	"strings"
)

// parseCellAction reads "cell:row:col".
func parseCellAction(action string) (r, c int, ok bool) {
	rest := strings.TrimPrefix(action, "cell:")
	parts := strings.Split(rest, ":")
	if len(parts) != uiParts {
		return 0, 0, false
	}

	r, err1 := strconv.Atoi(parts[0])
	c, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || r < 0 || c < 0 {
		return 0, 0, false
	}

	return r, c, true
}

// actionIndex reads the number in one "name:number" action.
func actionIndex(action string) (int, bool) {
	cut := strings.LastIndexByte(action, ':')
	if cut < 0 {
		return 0, false
	}

	n, err := strconv.Atoi(action[cut+1:])
	if err != nil || n < 0 {
		return 0, false
	}

	return n, true
}
