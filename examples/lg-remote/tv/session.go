package tv

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Session keeps one TV connection and the saved key.
type Session struct {
	mu    sync.Mutex
	c     *Client
	path  string
	store Store
	saved atomic.Pointer[Store]
	steps chan struct{}
	cli   atomic.Pointer[Client]
}

// NewSession loads a saved key from dir/lg-remote.json.
func NewSession(dir string) *Session {
	path := filepath.Join(dir, "lg-remote.json")

	s := &Session{path: path, store: Load(path)}
	s.publishStore()
	return s
}

// SavedHost is the last IP that paired.
func (s *Session) SavedHost() string { return s.savedStore().Host }

// SavedModel is the model name from the last pairing.
func (s *Session) SavedModel() string { return s.savedStore().Model }

// Exec connects if needed and runs spec.
func (s *Session) Exec(host, spec string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensure(host); err != nil {
		return "", err
	}

	msg, err := s.c.Exec(spec)
	if err == nil || s.c.Alive() || spec == "power:" {
		return msg, err
	}

	s.c.Close()
	s.c = nil
	if err = s.ensure(host); err != nil {
		return "", err
	}

	return s.c.Exec(spec)
}

// Scan looks for a TV and remembers its address and name.
func (s *Session) Scan() (string, error) {
	found := Discover()
	if len(found) == 0 {
		return "", errors.New("no TV answered on this Wi-Fi")
	}

	hit := choose(found)
	s.mu.Lock()
	s.store.Host = hit.IP
	if hit.Name != "" {
		s.store.Model = hit.Name
	}
	s.publishStore()
	_ = Save(s.path, s.store)
	s.mu.Unlock()

	return hit.IP, nil
}

// Wake sends Wake-on-LAN to every saved MAC.
func (s *Session) Wake() (string, error) {
	s.mu.Lock()
	macs := append([]string{}, s.store.MACs...)
	host := s.store.Host
	s.mu.Unlock()

	if len(macs) == 0 {
		return "", errors.New("connect once over Wi-Fi so wake can learn the TV")
	}

	var last error
	for _, mac := range macs {
		last = sendWake(mac, host)
		if last == nil {
			return "Wake sent", nil
		}
	}

	return "", last
}
