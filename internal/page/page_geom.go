package page

// pageGeom is the frame size in CSS pixels. It is embedded in Page so
// promotion keeps p.width working and page_types.go stays in the file
// size limit.
type pageGeom struct {
	width     int
	height    int
	minWidth  int
	minHeight int
	maxWidth  int
	maxHeight int
}
