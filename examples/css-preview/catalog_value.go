package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func catalogValue(p *ownframe.Page, name string) string {
	target, sampled := targetByKey(catalogSampleKey(name))
	if !sampled {
		return "not present"
	}
	box, _ := findBox(p.Boxes(), target.id)
	if name == "trigger-scope" {
		card, _ := findBox(p.Boxes(), catalogCardID(name))
		return otherValue(card, name)
	}
	if strings.HasPrefix(name, "view-transition") {
		return otherValue(box, name)
	}
	if strings.Contains(name, "animation") {
		if len(box.Animations) != 0 {
			return animationValue(box.Animations[0], name)
		}
	}
	return transitionValue(box, name)
}
