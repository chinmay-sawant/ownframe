package main

import "time"

var standardTransitionRuntimeCases = []transitionRuntimeCase{
	{"transition", "left", "left 1000ms linear", "left 2000ms ease-in", 500 * time.Millisecond},
	{"transition-delay", "left", "0ms", "300ms", 150 * time.Millisecond},
	{"transition-duration", "left", "1000ms", "2000ms", 500 * time.Millisecond},
	{"transition-property", "", "left", "width", 500 * time.Millisecond},
	{"transition-timing-function", "left", "linear", "steps(4,end)", 400 * time.Millisecond},
}
