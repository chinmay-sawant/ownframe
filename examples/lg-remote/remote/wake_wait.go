package remote

import "time"

type wakeState struct {
	host    string
	next    time.Time
	attempt int
}

func (a *App) noteWake(host string) {
	if host != "" {
		a.notes <- update{wakeHost: host}
	}
}

func (a *App) armWake(host string) {
	a.wake = wakeState{host: host, next: time.Now().Add(wakeDelay)}
}

func (a *App) pollWake() bool {
	if a.wake.host == "" || time.Now().Before(a.wake.next) {
		return false
	}
	if a.wake.attempt == wakeMaxAttempts {
		a.wake = wakeState{}
		a.view.Status = "Wake sent. The TV is still starting. Tap Connect to retry."
		return true
	}
	host := a.wake.host
	a.wake.attempt++
	a.wake.next = time.Now().Add(wakeRetryDelay)
	a.later(func() { a.doExec(host, "pair:") })
	return false
}
