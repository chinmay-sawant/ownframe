package main

// standardAnimationRuntimeCases pairs two values of one property. Each pair
// draws a different frame from a paused animation at a negative delay, so the
// assertion does not depend on when the test's tick lands.
var standardAnimationRuntimeCases = []animationRuntimeCase{
	{"animation", "route 1000ms linear -250ms 1 normal both paused", "route 2000ms linear -250ms 1 normal both paused", "", false},
	{"animation-delay", "-250ms", "-500ms", "", false},
	{"animation-delay-start", "-250ms", "-500ms", "", false},
	{"animation-delay-end", "0ms", "500ms",
		`animation-name:route;animation-duration:1000ms;animation-delay:-1250ms;` +
			`animation-fill-mode:none;animation-timing-function:linear;animation-play-state:paused;`, false},
	{"animation-direction", "normal", "reverse", "", false},
	{"animation-duration", "1000ms", "2000ms", "", false},
	{"animation-fill-mode", "both", "none",
		`animation-name:route;animation-duration:1000ms;animation-delay:-1250ms;` +
			`animation-timing-function:linear;animation-play-state:paused;`, false},
	{"animation-iteration-count", "1", "2",
		`animation-name:route;animation-duration:1000ms;animation-delay:-1250ms;` +
			`animation-fill-mode:both;animation-timing-function:linear;animation-play-state:paused;`, false},
	{"animation-name", "route", "returnRoute", "", false},
	{"animation-play-state", "running", "paused", "", true},
	{"animation-timing-function", "linear", "steps(4,end)",
		`animation-name:route;animation-duration:1000ms;animation-delay:-400ms;` +
			`animation-fill-mode:both;animation-play-state:paused;`, false},
}
