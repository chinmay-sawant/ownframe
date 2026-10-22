package atspi

// FakeBus records AT-SPI calls. It is a test recorder, not a screen
// reader, and it opens no socket.
type FakeBus struct {
	name    string
	claimed []string
	apps    []Application
	events  []Event
	closed  bool
	live    bool
}

// NewFakeBus returns a connected recorder. Live is true only for the
// seam; it does not mean Orca or Accerciser can see a window.
func NewFakeBus() *FakeBus {
	return &FakeBus{live: true}
}

func (b *FakeBus) Live() bool { return b != nil && b.live && !b.closed }

func (b *FakeBus) Status() Status {
	return Status{Platform: "fake", Live: b.Live(), Reason: "test recorder"}
}

func (b *FakeBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	if b.name != "" && b.name != name {
		return ErrName
	}
	b.name = name
	b.claimed = append(b.claimed, name)
	return nil
}

func (b *FakeBus) Publish(app Application) error {
	b.apps = append(b.apps, app)
	return nil
}

func (b *FakeBus) Emit(ev Event) error {
	b.events = append(b.events, ev)
	return nil
}

func (b *FakeBus) Close() error {
	b.closed = true
	b.live = false
	return nil
}
