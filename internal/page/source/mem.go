package source

import (
	"os"
	"sync"
	"time"
)

// memFileMode is the fake permission bits the in-memory source reports.
const memFileMode os.FileMode = 0o600

type memInfo struct {
	name string
	size int64
	mod  time.Time
}

func (m memInfo) Name() string       { return m.name }
func (m memInfo) Size() int64        { return m.size }
func (m memInfo) Mode() os.FileMode  { return memFileMode }
func (m memInfo) ModTime() time.Time { return m.mod }
func (m memInfo) IsDir() bool        { return false }
func (m memInfo) Sys() any           { return nil }

// Mem is an in-memory Reader. Write copies the bytes and advances ModTime
// so a watch sees each write.
type Mem struct {
	mu   sync.Mutex
	data map[string][]byte
	mod  map[string]time.Time
	seq  int64
}

// NewMem returns an empty memory reader.
func NewMem() *Mem {
	return &Mem{data: map[string][]byte{}, mod: map[string]time.Time{}}
}

// Write stores a copy of b at path.
func (m *Mem) Write(path string, b []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	m.data[path] = append([]byte(nil), b...)
	m.mod[path] = time.Unix(m.seq, 0)
}

// Stat reports the last Write. A missing path is os.ErrNotExist.
func (m *Mem) Stat(path string) (os.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.data[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return memInfo{name: path, size: int64(len(b)), mod: m.mod[path]}, nil
}

// ReadFile returns a copy of the last Write. A missing path is os.ErrNotExist.
func (m *Mem) ReadFile(path string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.data[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), b...), nil
}
