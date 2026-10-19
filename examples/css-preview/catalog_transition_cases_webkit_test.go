package main

import "time"

var webkitTransitionRuntimeCases = []transitionRuntimeCase{
	{"-webkit-transition", "left", "left 1000ms linear", "left 2000ms ease-in", 500 * time.Millisecond},
	{"-webkit-transition-delay", "left", "0ms", "300ms", 150 * time.Millisecond},
	{"-webkit-transition-duration", "left", "1000ms", "2000ms", 500 * time.Millisecond},
	{"-webkit-transition-property", "", "left", "width", 500 * time.Millisecond},
	{"-webkit-transition-timing-function", "left", "linear", "steps(4,end)", 400 * time.Millisecond},
}
