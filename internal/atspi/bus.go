package atspi

// Bus is the AT-SPI transport seam. FakeBus records calls for tests.
// LiveBus connects to the Linux accessibility bus when it is available.
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
