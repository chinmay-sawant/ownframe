package remote

import "sync"

type jobQueue struct {
	mu      sync.Mutex
	pending []func()
	running bool
}

func (a *App) later(fn func()) bool {
	a.use()
	if !a.async {
		fn()
		return true
	}
	a.jobs.mu.Lock()
	defer a.jobs.mu.Unlock()
	if len(a.jobs.pending) >= maxJobs {
		a.view.Status = "Commands are still sending. Please wait."
		return false
	}
	a.jobs.pending = append(a.jobs.pending, fn)
	if !a.jobs.running {
		a.jobs.running = true
		go a.runJobs()
	}
	return true
}

func (a *App) runJobs() {
	for {
		a.jobs.mu.Lock()
		if len(a.jobs.pending) == 0 {
			a.jobs.running = false
			a.jobs.mu.Unlock()
			return
		}
		fn := a.jobs.pending[0]
		a.jobs.pending[0] = nil
		a.jobs.pending = a.jobs.pending[1:]
		a.jobs.mu.Unlock()
		fn()
	}
}
