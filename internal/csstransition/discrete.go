package csstransition

// Position changes geometry on matching operations. Visibility and display
// change operation presence and need a frame compositor this package lacks.
func discretePaintProperty(property string) bool {
	return property == "position"
}
