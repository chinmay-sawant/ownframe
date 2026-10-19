package main

import "time"

var webkitAnimationRuntimeCases = []animationRuntimeCase{
	{"-webkit-animation", "route 1000ms linear 1 normal both running", "route 2000ms linear 1 normal both running", 500 * time.Millisecond, false},
	{"-webkit-animation-delay", "0ms", "300ms", 150 * time.Millisecond, false},
	{"-webkit-animation-direction", "normal", "reverse", 250 * time.Millisecond, false},
	{"-webkit-animation-duration", "1000ms", "2000ms", 500 * time.Millisecond, false},
	{"-webkit-animation-fill-mode", "both", "none", 1250 * time.Millisecond, false},
	{"-webkit-animation-iteration-count", "1", "2", 1250 * time.Millisecond, false},
	{"-webkit-animation-name", "route", "returnRoute", 500 * time.Millisecond, false},
	{"-webkit-animation-play-state", "running", "paused", 500 * time.Millisecond, false},
	{"-webkit-animation-timing-function", "linear", "steps(4,end)", 400 * time.Millisecond, false},
}
