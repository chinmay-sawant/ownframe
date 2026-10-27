//go:build !linux || android || ios || js

package atspi

import "runtime"

type liveBus struct {
	reason string
}

func newLiveBus() Bus {
	return &liveBus{reason: liveReason()}
}

func (b *liveBus) Live() bool { return false }

func (b *liveBus) Status() Status {
	return Status{Platform: runtime.GOOS, Live: false, Reason: b.reason}
}

func (b *liveBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	return ErrLiveUnavailable
}

func (b *liveBus) Publish(Application) error { return ErrLiveUnavailable }

func (b *liveBus) Emit(Event) error { return ErrLiveUnavailable }

func (b *liveBus) Close() error { return nil }
