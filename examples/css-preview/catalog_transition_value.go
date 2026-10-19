package main

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func transitionValue(b ownframe.Box, name string) string {
	name = strings.TrimPrefix(name, "-webkit-")
	if len(b.Transitions) == 0 {
		return "not present"
	}
	r := b.Transitions[0]
	switch name {
	case "transition":
		return fmt.Sprintf("%s %gms %s %gms", r.Property, r.DurationMS, r.TimingFunction, r.DelayMS)
	case "transition-property":
		return r.Property
	case "transition-duration":
		return fmt.Sprintf("%gms", r.DurationMS)
	case "transition-delay":
		return fmt.Sprintf("%gms", r.DelayMS)
	case "transition-timing-function":
		return r.TimingFunction
	case "transition-behavior":
		return r.Behavior
	}
	return "not present"
}
