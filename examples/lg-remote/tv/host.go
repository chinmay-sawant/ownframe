package tv

import (
	"errors"
	"net"
	"strings"
)

func (s *Session) ensure(host string) error {
	host = cleanHost(host)
	if host == "" {
		host = s.store.Host
	}

	if host == "" {
		return errors.New("enter the TV IP")
	}

	if s.c != nil && s.c.Alive() && s.store.Host == host {
		return nil
	}

	if s.c != nil {
		s.c.Close()
		s.c = nil
	}

	c, err := Open(host, s.store.Key)
	if err != nil {
		return err
	}

	s.c = c
	s.store.Host = host
	s.store.Key = c.Key.Reveal()
	if c.Model != "" {
		s.store.Model = c.Model
	}

	if len(c.MACs) > 0 {
		s.store.MACs = c.MACs
	}

	s.publishStore()
	_ = Save(s.path, s.store)

	return nil
}

func cleanHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "ws://")
	host = strings.TrimPrefix(host, "wss://")
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}

	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}

	return host
}
