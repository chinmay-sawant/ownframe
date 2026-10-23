package player

type observer struct {
	LastKind, LastID string
	Observed         []string
}

func (o *observer) Observe(id string) {
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

func (o *observer) Notify(kind, id string) {
	if o == nil {
		return
	}
	o.LastKind = kind
	o.LastID = id
}

func (o *observer) Reset() {
	if o == nil {
		return
	}
	o.LastKind = ""
	o.LastID = ""
	o.Observed = nil
}

type visibilityObserver struct{ observer }

type resizeObserver struct{ observer }
