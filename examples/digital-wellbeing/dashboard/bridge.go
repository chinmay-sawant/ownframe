package dashboard

import (
	"encoding/json"
	"sync"
)

type bridge struct {
	mu                sync.Mutex
	accessSet, access bool
	appsSet           bool
	apps              []AppChoice
	usageSet          bool
	usage             Usage
	restoreSet        bool
	restore           string
	accessRequest     bool
	backRequest       bool
}

// Usage contains daily totals for all apps or one selected package.
type Usage struct {
	Package string `json:"package"`
	Label   string `json:"label"`
	Days    []struct {
		Label   string `json:"label"`
		Minutes int    `json:"minutes"`
	} `json:"days"`
}

// QueueAccess queues the current Android Usage Access state.
func (a *App) QueueAccess(allowed bool) {
	a.mu.Lock()
	a.access, a.accessSet = allowed, true
	a.mu.Unlock()
}

// QueueApps passes app names and package IDs, never their usage totals.
func (a *App) QueueApps(data string) {
	var catalog struct {
		Apps []AppChoice `json:"apps"`
	}
	if json.Unmarshal([]byte(data), &catalog) != nil {
		return
	}
	a.mu.Lock()
	a.apps, a.appsSet = catalog.Apps, true
	a.mu.Unlock()
}

// QueueUsage passes daily usage totals from Android.
func (a *App) QueueUsage(data string) {
	var usage Usage
	if json.Unmarshal([]byte(data), &usage) != nil {
		return
	}
	a.mu.Lock()
	a.usage, a.usageSet = usage, true
	a.mu.Unlock()
}

// QueueRestore restores the last app selected on this device.
func (a *App) QueueRestore(pkg string) {
	a.mu.Lock()
	a.restore, a.restoreSet = pkg, true
	a.mu.Unlock()
}
