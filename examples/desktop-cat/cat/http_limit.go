package cat

import (
	"net"
	"net/http"
	"time"
)

// notifyEvery is one /notify token per 100ms; burst 10. notifyInflight
// caps concurrent /notify handlers; maxNotifyIPs bounds the table.
const (
	notifyEvery    = 100 * time.Millisecond
	notifyBurst    = 10
	notifyInflight = 4
	maxNotifyIPs   = 1024
	notifyKeep     = 5 * time.Minute
)

type notifySeen struct {
	tokens float64
	last   time.Time
}

// allowNotify reports one token for the request source IP. State lives
// on the inbox, so each cat gets its own table.
func (i *Inbox) allowNotify(addr string) bool {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	now := time.Now()
	i.limitMu.Lock()
	defer i.limitMu.Unlock()
	seen, ok := i.throttleHits[ip]
	if !ok {
		i.throttleHits[ip] = &notifySeen{tokens: notifyBurst - 1, last: now}
		return true
	}
	seen.tokens += now.Sub(seen.last).Seconds() / notifyEvery.Seconds()
	seen.last = now
	if len(i.throttleHits) > maxNotifyIPs {
		for key, item := range i.throttleHits {
			if now.Sub(item.last) > notifyKeep {
				delete(i.throttleHits, key)
			}
		}
	}
	if seen.tokens > notifyBurst {
		seen.tokens = notifyBurst
	}
	if seen.tokens < 1 {
		return false
	}
	seen.tokens--
	return true
}

// limitNotify caps /notify floods per IP with 429 and in-flight with 503.
func (i *Inbox) limitNotify(next http.Handler) http.Handler {
	i.limitMu.Lock()
	if i.throttleHits == nil {
		i.throttleHits = map[string]*notifySeen{}
	}
	if i.notifyLive == nil {
		i.notifyLive = make(chan struct{}, notifyInflight)
	}
	i.limitMu.Unlock()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !i.allowNotify(r.RemoteAddr) {
			reply(w, http.StatusTooManyRequests, map[string]string{"error": "slow down"})
			return
		}
		select {
		case i.notifyLive <- struct{}{}:
		default:
			reply(w, http.StatusServiceUnavailable, map[string]string{"error": "busy"})
			return
		}
		defer func() { <-i.notifyLive }()
		next.ServeHTTP(w, r)
	})
}
