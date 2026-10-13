package runtimebehavior

// Observer notifies about visibility and resize changes.
type Observer struct {
	Callback  func(kind string, id string)
	LastKind  string
	LastID    string
	Observed  []string
	Connected bool
}

// NewObserver builds an observer with an optional callback.
func NewObserver(cb func(string, string)) *Observer {
	return &Observer{Callback: cb, Connected: true}
}

// Notify calls the callback if set.
func (o *Observer) Notify(kind, id string) {
	if o == nil {
		return
	}
	o.LastKind = kind
	o.LastID = id
	if o.Callback != nil {
		o.Callback(kind, id)
	}
}

// Observe records a target id for notifications.
func (o *Observer) Observe(id string) {
	if o == nil || id == "" {
		return
	}
	for _, x := range o.Observed {
		if x == id {
			return
		}
	}
	o.Observed = append(o.Observed, id)
}

// Unobserve drops a target id.
func (o *Observer) Unobserve(id string) {
	if o == nil {
		return
	}
	out := o.Observed[:0]
	for _, x := range o.Observed {
		if x != id {
			out = append(out, x)
		}
	}
	o.Observed = out
}

// Disconnect stops future callbacks.
func (o *Observer) Disconnect() {
	if o == nil {
		return
	}
	o.Connected = false
	o.Callback = nil
}

// Reset clears observer state.
func (o *Observer) Reset() {
	if o == nil {
		return
	}
	o.LastKind = ""
	o.LastID = ""
	o.Observed = nil
	o.Connected = true
}

// VisibilityObserver is an observer scoped to visibility changes.
type VisibilityObserver struct {
	Observer
}

// NewVisibilityObserver builds a visibility observer.
func NewVisibilityObserver(cb func(string, string)) *VisibilityObserver {
	return &VisibilityObserver{Observer: *NewObserver(cb)}
}

// ResizeObserver is an observer scoped to resize changes.
type ResizeObserver struct {
	Observer
}

// NewResizeObserver builds a resize observer.
func NewResizeObserver(cb func(string, string)) *ResizeObserver {
	return &ResizeObserver{Observer: *NewObserver(cb)}
}
