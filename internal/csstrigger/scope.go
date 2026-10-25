package csstrigger

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func inScope(boxes []layout.Box, source, target int, name string) bool {
	for index := source; index >= 0; index = boxes[index].ParentIndex {
		if scopesName(boxes[index].TriggerScope, name) && !descendant(boxes, index, target) {
			return false
		}
	}
	for index := target; index >= 0; index = boxes[index].ParentIndex {
		if scopesName(boxes[index].TriggerScope, name) && !descendant(boxes, index, source) {
			return false
		}
	}
	return true
}

func scopesName(scope, name string) bool {
	for _, token := range strings.Fields(strings.ReplaceAll(scope, ",", " ")) {
		if token == "all" || token == name {
			return true
		}
	}
	return false
}

func descendant(boxes []layout.Box, ancestor, child int) bool {
	for child >= 0 {
		if child == ancestor {
			return true
		}
		child = boxes[child].ParentIndex
	}
	return false
}
