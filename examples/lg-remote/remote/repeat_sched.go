package remote

import "time"

type repeatRate struct {
	After    time.Duration
	Interval time.Duration
}

type repeatPolicy struct {
	Delay time.Duration
	Rates []repeatRate
}

// repeatScheduler advances on the page tick; it starts no goroutines.
type repeatScheduler struct {
	policy  repeatPolicy
	started time.Time
	next    time.Time
	active  bool
}

func (s *repeatScheduler) Start(p repeatPolicy, now time.Time) bool {
	s.Stop()
	if p.Delay < 0 || len(p.Rates) == 0 {
		return false
	}
	for i, rate := range p.Rates {
		badAfter := rate.After < 0 || i > 0 && rate.After <= p.Rates[i-1].After
		if rate.Interval <= 0 || badAfter {
			return false
		}
	}
	p.Rates = append([]repeatRate(nil), p.Rates...)
	s.policy, s.started, s.next, s.active = p, now, now.Add(p.Delay), true
	return true
}

func (s *repeatScheduler) Stop() {
	s.policy, s.started, s.next, s.active = repeatPolicy{}, time.Time{}, time.Time{}, false
}

func (s *repeatScheduler) Tick(now time.Time) bool {
	if !s.active || now.Before(s.next) {
		return false
	}
	interval := s.policy.Rates[0].Interval
	held := now.Sub(s.started)
	for _, rate := range s.policy.Rates {
		if held < rate.After {
			break
		}
		interval = rate.Interval
	}
	s.next = now.Add(interval)
	return true
}
