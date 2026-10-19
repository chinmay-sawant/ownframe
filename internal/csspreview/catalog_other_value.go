package csspreview

import "github.com/chinmay-sawant/ownframe"

func otherValue(b ownframe.Box, name string) string {
	switch name {
	case "trigger-scope":
		return b.TriggerScope
	case "view-transition-class":
		return b.ViewTransitionClass
	case "view-transition-group":
		return b.ViewTransitionGroup
	case "view-transition-name":
		return b.ViewTransitionName
	case "view-transition-scope":
		return b.ViewTransitionScope
	}
	return "not present"
}
