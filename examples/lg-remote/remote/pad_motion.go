package remote

import (
	"fmt"
	"math"
	"sync"
)

type padMotion struct {
	mu      sync.Mutex
	dx, dy  float64
	pending bool
}

func (a *App) queueMotion(host string, dx, dy float64) {
	m := &a.motion
	m.mu.Lock()
	scale := float64(a.view.Sensitivity) / sensitivityScale
	m.dx += dx * scale
	m.dy += dy * scale
	queue := !m.pending
	m.pending = true
	m.mu.Unlock()
	if queue && !a.later(func() { a.sendMotion(host) }) {
		m.mu.Lock()
		m.pending = false
		m.dx, m.dy = 0, 0
		m.mu.Unlock()
	}
}

func (a *App) sendMotion(host string) {
	m := &a.motion
	m.mu.Lock()
	x, y := int(math.Trunc(m.dx)), int(math.Trunc(m.dy))
	m.dx -= float64(x)
	m.dy -= float64(y)
	m.pending = false
	m.mu.Unlock()
	if x == 0 && y == 0 {
		return
	}
	_, err := a.use().Exec(host, fmt.Sprintf("move:%d,%d", x, y))
	if err != nil {
		a.note(err.Error(), "", "")
	}
}
