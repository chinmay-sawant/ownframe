package main

// webkitAnimationRuntimeCases exercise the prefixed aliases. They share the
// standard base declaration; the prefixed longhand is the value under test.
var webkitAnimationRuntimeCases = []animationRuntimeCase{
	{"-webkit-animation", "route 1000ms linear -250ms 1 normal both paused", "route 2000ms linear -250ms 1 normal both paused", "", false},
	{"-webkit-animation-delay", "-250ms", "-500ms", "", false},
	{"-webkit-animation-direction", "normal", "reverse", "", false},
	{"-webkit-animation-duration", "1000ms", "2000ms", "", false},
	{"-webkit-animation-fill-mode", "both", "none",
		`animation-name:route;animation-duration:1000ms;animation-delay:-1250ms;` +
			`animation-timing-function:linear;animation-play-state:paused;`, false},
	{"-webkit-animation-iteration-count", "1", "2",
		`animation-name:route;animation-duration:1000ms;animation-delay:-1250ms;` +
			`animation-fill-mode:both;animation-timing-function:linear;animation-play-state:paused;`, false},
	{"-webkit-animation-name", "route", "returnRoute", "", false},
	{"-webkit-animation-play-state", "running", "paused", "", true},
	{"-webkit-animation-timing-function", "linear", "steps(4,end)",
		`animation-name:route;animation-duration:1000ms;animation-delay:-400ms;` +
			`animation-fill-mode:both;animation-play-state:paused;`, false},
}
