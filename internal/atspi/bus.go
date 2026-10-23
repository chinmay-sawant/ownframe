package atspi

// Bus is the AT-SPI transport seam. FakeBus records calls for tests.
// LiveBus never reports Live in this slice.
type Bus interface {
	Claim(name string) error
	Publish(app Application) error
	Emit(ev Event) error
	Close() error
	Live() bool
	Status() Status
}

// Status describes whether a real assistive-technology bus is connected.
type Status struct {
	Platform string
	Live     bool
	Reason   string
}
