package atspi

// Claimed returns well-known names this recorder accepted.
func (b *FakeBus) Claimed() []string {
	out := make([]string, len(b.claimed))
	copy(out, b.claimed)
	return out
}

// Apps returns published trees in order.
func (b *FakeBus) Apps() []Application {
	out := make([]Application, len(b.apps))
	copy(out, b.apps)
	return out
}

// Events returns notifications with no accessible text payload.
func (b *FakeBus) Events() []Event {
	out := make([]Event, len(b.events))
	copy(out, b.events)
	return out
}
