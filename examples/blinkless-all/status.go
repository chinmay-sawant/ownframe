package main

// painterDeclines lists the properties the engine parses and lays out but the
// window painter does not draw, so the page shows the row as partial: the
// catalog says implemented for the engine, and the picture cannot show it. A
// demo whose picture is identical to its control belongs here, not under an
// implemented badge.
var painterDeclines = map[string]bool{
	"visibility": true, "gap": true, "filter": true, "scale": true, "backdrop-filter": true,
	"color-scheme": true, "transform-box": true, "grid-auto-columns": true, "grid-auto-flow": true,
	"writing-mode": true,
}

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
	return noPaint[name] || painterDeclines[name] || painterDeclines[trimWebkit(name)]
}

func trimWebkit(name string) string {
	const prefix = "-webkit-"
	if len(name) > len(prefix) && name[:len(prefix)] == prefix {
		return name[len(prefix):]
	}

	return name
}
