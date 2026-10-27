package atspi

// NewLiveBus returns the production AT-SPI bus.
func NewLiveBus() Bus {
	return newLiveBus()
}

