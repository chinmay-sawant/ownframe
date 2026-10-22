package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

// Publisher maps snapshots onto a Bus. It does not log accessible text.
type Publisher struct {
	bus   Bus
	app   Application
	doc   uint64
	rev   uint64
	focus host.AccessibilityNodeID
}

// NewPublisher wraps bus. A nil bus is treated as not live.
func NewPublisher(bus Bus) *Publisher {
	return &Publisher{bus: bus}
}

func (p *Publisher) Live() bool {
	return p != nil && p.bus != nil && p.bus.Live()
}

func (p *Publisher) Status() Status {
	if p == nil || p.bus == nil {
		return Status{Live: false, Reason: "no bus"}
	}
	return p.bus.Status()
}

// Mapped returns the last mapped objects. The slice is a copy.
func (p *Publisher) Mapped() []Object {
	out := make([]Object, len(p.app.Objects))
	copy(out, p.app.Objects)
	return out
}
