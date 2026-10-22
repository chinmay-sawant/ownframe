package player

import "time"

type RuntimeBehaviorHelper struct {
	Loading    *loadingMarker
	Marker     *loadingMarker
	Observer   *observer
	Visibility *visibilityObserver
	Resize     *resizeObserver
	Watchdog   *watchdog
}

func NewRuntimeBehaviorHelper() *RuntimeBehaviorHelper {
	m := &loadingMarker{}
	return &RuntimeBehaviorHelper{
		Loading: m, Marker: m,
		Observer:   &observer{},
		Visibility: &visibilityObserver{},
		Resize:     &resizeObserver{},
		Watchdog:   &watchdog{},
	}
}

func NewRuntimeHelper() *RuntimeBehaviorHelper { return NewRuntimeBehaviorHelper() }

func (h *RuntimeBehaviorHelper) StartLoading(s string) { h.Loading.Show(s) }
func (h *RuntimeBehaviorHelper) SlowLoading(n time.Time, d time.Duration) bool {
	return h.Loading.MarkSlow(n, d)
}
func (h *RuntimeBehaviorHelper) ResolveLoading() { h.Loading.Hide() }
func (h *RuntimeBehaviorHelper) Observe(id string) {
	h.Observer.Observe(id)
	h.Visibility.Observe(id)
	h.Resize.Observe(id)
}
func (h *RuntimeBehaviorHelper) Notify(k, id string) { h.Observer.Notify(k, id) }
func (h *RuntimeBehaviorHelper) SampleWatchdog(n time.Time, w time.Duration, l int) bool {
	return h.Watchdog.Sample(n, w, l)
}
func (h *RuntimeBehaviorHelper) Reset() {
	h.Loading.Reset()
	h.Observer.Reset()
	h.Visibility.Reset()
	h.Resize.Reset()
	h.Watchdog.Reset()
}
