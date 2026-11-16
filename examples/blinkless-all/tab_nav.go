package main

// navTab is the "Navigation and UI" tab: 20 properties in 14 examples. The
// engine stores each value (layout.Box.Declared); nav_*.go and the rows built
// by navBody decide what the host does with it.
func navTab() Tab {
	return Tab{ID: "nav", Title: "Navigation and UI", Props: []Prop{
		{Name: "caret-shape", Covers: []string{"caret", "caret-animation"},
			Note: "The host draws a bar, block or underscore caret from the declared shape and blinks it unless the animation is manual.", Demo: navDemoCaret},
		{Name: "nav-up", Covers: []string{"nav-down", "nav-left", "nav-right"},
			Note: "An arrow key jumps to the #id the element names for that direction, and falls back to the nearest button when it says auto.", Demo: navDemoFocus},
		{Name: "spatial-navigation-function",
			Note: "Normal reaches any button below; grid only reaches buttons that share a column with the current one.", Demo: navDemoGrid},
		{Name: "spatial-navigation-contain",
			Note: "Focus movement stays inside the container and does not reach buttons outside it.", Demo: navDemoContain},
		{Name: "spatial-navigation-action",
			Note: "Scroll turns an arrow key into a 40px scroll step instead of moving focus.", Demo: navDemoAction},
		{Name: "field-sizing",
			Note: "Content grows the field with its text; fixed keeps the width the page gave it.", Demo: navDemoField},
		{Name: "input-security",
			Note: "A password field shows bullets unless input-security is none.", Demo: navDemoSecurity},
		{Name: "slider-orientation",
			Note: "The host reads a pointer position on a slider track as a value, in the direction the property names.", Demo: navDemoSlider},
		{Name: "window-drag",
			Note: "A press on a move region starts a window drag, and a child marked none does not.", Demo: navDemoDrag},
		{Name: "will-change",
			Note: "The host reads the hints and keeps a cached layer for transform and opacity.", Demo: navDemoLayer},
		{Name: "-webkit-appearance",
			Note: "The host draws a plain box when the value is none and a button when it is button.", Demo: navDemoAppear},
		{Name: "-webkit-user-select",
			Note: "A click selects the word under it, everything, or nothing, as the value says.", Demo: navDemoSelect},
		{Name: "text-size-adjust", Covers: []string{"-webkit-text-size-adjust"},
			Note: "None keeps the font size, a percentage scales it, and the standard name wins over the -webkit- one.", Demo: navDemoAdjust},
		{Name: "image-animation",
			Note: "An animated image plays, freezes where it was paused, or stays on its first frame.", Demo: navDemoImage},
	}, Head: func() string { return navHeadCSS }}
}
