package atspi

// Publisher maps snapshots onto a Bus. It does not log accessible text.
type Publisher struct {
	bus       Bus
	app       Application
	doc       uint64
	rev       uint64
	published bool
}

// NewPublisher wraps bus. A nil bus is treated as not live.
func NewPublisher(bus Bus) *Publisher {
	return &Publisher{bus: bus}
}

func (p *Publisher) Live() bool {
	return p != nil && p.bus != nil && p.bus.Live()
}

// Ready reports whether the next Publish should walk the tree and try
// the bus. A bus that exposes its own backoff decides even while it is
// live, so a failed Embed cannot be retried on every frame.
func (p *Publisher) Ready() bool {
	if p == nil || p.bus == nil {
		return false
	}
	if r, ok := p.bus.(interface{ Ready() bool }); ok {
		return r.Ready()
	}
	return p.Live()
}

// SetOrigin records the window content origin for screen coordinates.
// It is a no-op for buses that do not implement it.
func (p *Publisher) SetOrigin(x, y int) {
	if p == nil || p.bus == nil {
		return
	}
	if s, ok := p.bus.(interface{ SetOrigin(int, int) }); ok {
		s.SetOrigin(x, y)
	}
}

// SetSize records the window content size used for the root extents.
// It is a no-op for buses that do not implement it.
func (p *Publisher) SetSize(width, height int) {
	if p == nil || p.bus == nil {
		return
	}
	if s, ok := p.bus.(interface{ SetSize(int, int) }); ok {
		s.SetSize(width, height)
	}
}

func (p *Publisher) Status() Status {
	if p == nil || p.bus == nil {
		return Status{Live: false, Reason: "no bus"}
	}
	return p.bus.Status()
}

// Close releases the bus connection. It is safe to call more than once.
func (p *Publisher) Close() error {
	if p == nil || p.bus == nil {
		return nil
	}
	return p.bus.Close()
}

// Mapped returns the last mapped objects. The slice is a copy.
func (p *Publisher) Mapped() []Object {
	out := make([]Object, len(p.app.Objects))
	copy(out, p.app.Objects)
	return out
}
