package ownframe

// Option configures a page before New parses it. Options run in order, so
// a later option wins when two set the same field. Every option keeps the
// zero value safe: opting into a risky surface always takes an explicit,
// InsecureAllow-prefixed name where one exists.
type Option func(*Config)

// WithPerf records pipeline timing, dirty counters, and allocs in Stats.
// It is off by default, so end users pay nothing; a developer opts in to
// fill the DevTools performance rows.
func WithPerf(on bool) Option {
	return func(c *Config) { c.Perf = on }
}

// NewWithOptions parses the page like New after applying every option.
func NewWithOptions(cfg Config, opts ...Option) (*Page, error) {
	for _, opt := range opts {
		opt(&cfg)
	}

	return New(cfg)
}
