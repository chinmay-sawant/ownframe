package main

import "time"

var standardAnimationRuntimeCases = []animationRuntimeCase{
	{"animation", "route 1000ms linear 1 normal both running", "route 2000ms linear 1 normal both running", 500 * time.Millisecond, false},
	{"animation-delay", "0ms", "300ms", 150 * time.Millisecond, false},
	{"animation-delay-start", "0ms", "300ms", 150 * time.Millisecond, false},
	{"animation-delay-end", "0ms", "500ms", 1250 * time.Millisecond, true},
	{"animation-direction", "normal", "reverse", 250 * time.Millisecond, false},
	{"animation-duration", "1000ms", "2000ms", 500 * time.Millisecond, false},
	{"animation-fill-mode", "both", "none", 1250 * time.Millisecond, false},
	{"animation-iteration-count", "1", "2", 1250 * time.Millisecond, false},
	{"animation-name", "route", "returnRoute", 500 * time.Millisecond, false},
	{"animation-play-state", "running", "paused", 500 * time.Millisecond, false},
	{"animation-timing-function", "linear", "steps(4,end)", 400 * time.Millisecond, false},
}
