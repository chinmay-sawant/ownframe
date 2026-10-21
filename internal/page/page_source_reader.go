package page

import (
	"os"
	"sync"
	"time"
)

type sourceReader interface {
	Stat(string) (os.FileInfo, error)
	ReadFile(string) ([]byte, error)
}

type osSource struct{}

func (osSource) Stat(n string) (os.FileInfo, error) { return os.Stat(n) }
func (osSource) ReadFile(n string) ([]byte, error)  { return os.ReadFile(n) }

var srcFiles sourceReader = osSource{}

func installSourceReader(r sourceReader) {
	if r == nil {
		srcFiles = osSource{}
		return
	}
	srcFiles = r
}

type memInfo struct {
	n string
	s int64
	m time.Time
}

// memFileMode is the fake permission bits the in-memory source reports.
const memFileMode os.FileMode = 0o600

func (m memInfo) Name() string       { return m.n }
func (m memInfo) Size() int64        { return m.s }
func (m memInfo) Mode() os.FileMode  { return memFileMode }
func (m memInfo) ModTime() time.Time { return m.m }
func (m memInfo) IsDir() bool        { return false }
func (m memInfo) Sys() any           { return nil }

type memSource struct {
	mu sync.Mutex
	d  map[string][]byte
	m  map[string]time.Time
	n  int64
}

func newMemSource() *memSource {
	return &memSource{d: map[string][]byte{}, m: map[string]time.Time{}}
}

func (m *memSource) write(p string, b []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.n++
	m.d[p] = append([]byte(nil), b...)
	m.m[p] = time.Unix(m.n, 0)
}

func (m *memSource) Stat(p string) (os.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.d[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	return memInfo{n: p, s: int64(len(b)), m: m.m[p]}, nil
}

func (m *memSource) ReadFile(p string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.d[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), b...), nil
}
