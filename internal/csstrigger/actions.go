package csstrigger

import "strings"

func triggerActions(value, name string) []string {
	var matched []string
	var currentName string
	var current []string
	for _, token := range strings.Fields(value) {
		if isAction(token) {
			if currentName != "" {
				current = append(current, token)
			}
			continue
		}
		if currentName == name {
			matched = append(matched[:0], current...)
		}
		currentName, current = token, nil
	}
	if currentName == name {
		matched = append(matched[:0], current...)
	}
	return matched
}

func isAction(value string) bool {
	switch value {
	case "none", "play", "play-once", "play-forwards", "play-backwards", "pause", "reset", "replay":
		return true
	default:
		return false
	}
}
