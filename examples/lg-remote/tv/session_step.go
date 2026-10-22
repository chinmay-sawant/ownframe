package tv

import (
	"errors"
	"strings"
)

// ErrStepsBusy means the TV has too many unacknowledged step commands.
var ErrStepsBusy = errors.New("the TV is still responding to earlier presses")

// StartStep sends a volume or channel step before waiting for its reply.
// The caller must call the returned function once to collect the result.
func (s *Session) StartStep(host, spec string) (func() (string, error), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.steps == nil {
		s.steps = make(chan struct{}, stepsChanSize)
	}
	select {
	case s.steps <- struct{}{}:
	default:
		return nil, ErrStepsBusy
	}
	if err := s.ensure(host); err != nil {
		<-s.steps
		return nil, err
	}
	c := s.c
	wait, err := c.startStep(spec)
	if err != nil {
		<-s.steps
		return nil, err
	}
	return func() (string, error) {
		defer func() { <-s.steps }()
		msg, rejected, err := wait()
		if !rejected {
			return msg, err
		}
		// Retry through the pointer only after an explicit rejection.
		// A lost reply may belong to a step the TV already applied.
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.c != c || !c.Alive() {
			return "", err
		}
		kind, name, _ := strings.Cut(spec, ":")
		key, label := "VOLUME", "Volume "+strings.ToLower(name)
		if kind == "ch" {
			key, label = "CHANNEL", "Channel"
		}
		if c.Button(key+strings.ToUpper(name)) != nil {
			return "", err
		}
		return label, nil
	}, nil
}
