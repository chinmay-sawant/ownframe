package source

// Check is the source and frame fields Validate reads.
type Check struct {
	File      string
	HTML      string
	ThemeFile string
	Theme     string
	Width     int
	Height    int
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
}

// Validate rejects a file and an inline string together, and unusable sizes.
// Zero minimums are 1. A positive maximum below that minimum is badSize.
func Validate(c Check, badSource, badSize error) error {
	if c.File != "" && c.HTML != "" {
		return badSource
	}
	if c.ThemeFile != "" && c.Theme != "" {
		return badSource
	}
	if c.Width <= 0 || c.Height <= 0 {
		return badSize
	}
	minW, minH := Mins(c.MinWidth, c.MinHeight)
	if c.MaxWidth > 0 && minW > c.MaxWidth {
		return badSize
	}
	if c.MaxHeight > 0 && minH > c.MaxHeight {
		return badSize
	}
	return nil
}

// Mins turns a non-positive minimum into 1.
func Mins(w, h int) (int, int) {
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	return w, h
}
