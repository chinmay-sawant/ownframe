package csspreview

import "strings"

func animationBehavior(name string) string {
	if strings.HasPrefix(name, "-webkit-") {
		return webkitAnimationBehavior(strings.TrimPrefix(name, "-webkit-"))
	}
	return standardAnimationBehavior(name)
}
