package cat

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"
)

// Notification is an agent's latest expression and short speech-bubble text.
type Notification struct {
	Expression string `json:"expression"`
	Message    string `json:"message"`
	Source     string `json:"source,omitempty"`
}

// Inbox accepts notifications without touching the page from HTTP goroutines.
type Inbox struct {
	mu      sync.Mutex
	latest  Notification
	version uint64
	Updates chan Notification
	// limitMu guards the /notify throttle table and semaphore.
	limitMu      sync.Mutex
	throttleHits map[string]*notifySeen
	notifyLive   chan struct{}
}

func NewInbox() *Inbox {
	return &Inbox{latest: Notification{Expression: "happy", Message: "Your cat is here. Waiting for a message.", Source: "Desktop cat"}, version: 1, Updates: make(chan Notification, 1), throttleHits: map[string]*notifySeen{}, notifyLive: make(chan struct{}, notifyInflight)}
}

// Submit validates and replaces the latest notification.
func (i *Inbox) Submit(n Notification) error {
	n.Expression = strings.ToLower(strings.TrimSpace(n.Expression))
	n.Message, n.Source = sanitizeBidi(strings.TrimSpace(n.Message)), sanitizeBidi(strings.TrimSpace(n.Source))
	if _, err := expressionIndex(n.Expression); err != nil {
		return err
	}
	if n.Message == "" || utf8.RuneCountInString(n.Message) > 160 || utf8.RuneCountInString(n.Source) > 32 {
		return fmt.Errorf("message must contain 1-160 characters; source at most 32")
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.latest = n
	i.version++
	select {
	case <-i.Updates:
	default:
	}
	i.Updates <- n
	return nil
}

// Latest returns a copy of the accepted notification and its sequence number.
func (i *Inbox) Latest() (Notification, uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.latest, i.version
}
