// Package bridge is the queue the Android activity polls.
// The page enqueues a Bluetooth command. Java takes it on the UI thread
// and sends the HID report. Desktop builds leave the queue unused.
package bridge

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	mu           sync.Mutex
	q            = []string{}
	dir          string
	bt           string
	btRevision   uint64
	commandReady func()
)

// SetDir selects the folder for the saved TV key.
// Android passes the app files directory. Desktop leaves it empty.
func SetDir(path string) {
	mu.Lock()
	dir = path
	mu.Unlock()
}

// Dir is the folder for lg-remote.json.
func Dir() string {
	mu.Lock()
	path := dir
	mu.Unlock()

	if path != "" {
		return path
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "."
	}

	return filepath.Join(base, "ownframe")
}

// Enqueue adds one command unless the 32-entry queue is full.
func Enqueue(cmd string) bool {
	mu.Lock()
	if len(q) >= maxQueue {
		mu.Unlock()
		return false
	}
	q = append(q, cmd)
	notify := commandReady
	mu.Unlock()
	if notify != nil {
		notify()
	}
	return true
}

// Take removes the next command, or returns "" when the queue is empty.
func Take() string {
	mu.Lock()
	defer mu.Unlock()

	if len(q) == 0 {
		return ""
	}

	cmd := q[0]
	q = q[1:]

	return cmd
}

// SetBluetooth records a status line from the phone radio.
func SetBluetooth(state string) {
	mu.Lock()
	bt = state
	btRevision++
	mu.Unlock()
}

// Bluetooth is the latest radio status. Empty until the phone reports one.
func Bluetooth() string {
	mu.Lock()
	defer mu.Unlock()

	return bt
}
