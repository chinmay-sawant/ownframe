package atspi

// NewLiveBus returns the production AT-SPI bus.
func NewLiveBus(actions ActionHandler) Bus {
	return newLiveBus(actions)
}
