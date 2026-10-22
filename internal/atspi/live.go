package atspi

import "runtime"

type liveBus struct{}

// NewLiveBus returns the production bus. Live is always false in this
// slice: there is no D-Bus AT-SPI client, and no socket is created.
func NewLiveBus() Bus {
	return liveBus{}
}

func (liveBus) Live() bool { return false }

func (liveBus) Status() Status {
	return Status{Platform: runtime.GOOS, Live: false, Reason: liveReason()}
}

func (liveBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	return ErrLiveUnavailable
}

func (liveBus) Publish(Application) error { return ErrLiveUnavailable }

func (liveBus) Emit(Event) error { return ErrLiveUnavailable }

func (liveBus) Close() error { return nil }
