package main

import "strings"

// painterDeclines lists the properties the engine parses and lays out but the
// window painter does not draw, so the page shows the row as partial: the
// catalog says implemented for the engine, and the picture cannot show it. A
// demo whose picture is identical to its control belongs here, not under an
// implemented badge.
//
//nolint:gochecknoglobals // read-only property lookup table
var painterDeclines = map[string]bool{}

// opOnly lists the properties that change the engine's operation list while
// the window picture stays the same as the control, so the readout says
// "differs" for a picture that does not.
//
//nolint:gochecknoglobals // read-only property lookup table
var opOnly = map[string]bool{}

// rowStatus is the badge of a row: the catalog status, lowered to partial when
// the window painter does not draw the property.
func (a *App) rowStatus(name string) string {
	st, _ := a.cat.Status(name)
	if st == StatusImplemented && declines(name) {
		return StatusPartial
	}

	return st
}

// declines reports whether the window painter leaves the property out.
func declines(name string) bool {
	return noPaint[name] || painterDeclines[name] || painterDeclines[trimWebkit(name)] || opOnly[name] || pictureMatchesControl(name)
}

// pictureMatchesControl reports whether the row's own readout found the demo
// identical to the control, with nothing exported to the host. Such a row
// shows no effect in the picture, so it cannot wear the implemented badge.
func pictureMatchesControl(name string) bool {
	v, ok := readoutCache.Load(name)
	out, _ := v.(string)

	return ok && strings.HasSuffix(out, "=> same layout and paint as the control")
}

func trimWebkit(name string) string {
	const prefix = "-webkit-"
	if len(name) > len(prefix) && name[:len(prefix)] == prefix {
		return name[len(prefix):]
	}

	return name
}
