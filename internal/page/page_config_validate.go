package page

// Validate rejects illegal Config states early: HTML with File, Theme
// with ThemeFile, and unusable sizes. New calls it before reading
// files so conflicts fail at construction instead of first draw.
func (c Config) Validate() error {
	if c.File != "" && c.HTML != "" {
		return ErrBadSource
	}
	if c.ThemeFile != "" && c.Theme != "" {
		return ErrBadSource
	}
	if c.Width <= 0 || c.Height <= 0 {
		return ErrBadSize
	}
	minW, minH := c.mins()
	if c.MaxWidth > 0 && minW > c.MaxWidth {
		return ErrBadSize
	}
	if c.MaxHeight > 0 && minH > c.MaxHeight {
		return ErrBadSize
	}
	return nil
}

// mins returns the effective minimums; zero means 1.
func (c Config) mins() (int, int) {
	minW := c.MinWidth
	if minW <= 0 {
		minW = 1
	}
	minH := c.MinHeight
	if minH <= 0 {
		minH = 1
	}
	return minW, minH
}
