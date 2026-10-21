package ownframe

import (
	"errors"
	"fmt"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

// ErrPanicked means a panic was recovered.
var ErrPanicked = errors.New("ownframe: panic")

var (
	reportMu   sync.Mutex
	lastReport string
)

// reportError stores err in a crash report file and returns the original
// error without the file path. Run calls it for every startup and window
// failure. The returned error keeps the original for errors.Is. A nil page
// keeps a blank title. Use LastReportPath for the file path.
func reportError(page *Page, err error) error {
	title := ""
	if page != nil {
		title = page.Title()
	}

	path, werr := crash.Write(title, err.Error())
	reportMu.Lock()
	lastReport = path
	reportMu.Unlock()
	if werr != nil {
		return fmt.Errorf("ownframe: %w: %w", err, werr)
	}

	return fmt.Errorf("ownframe: %w", err)
}

// LastReportPath is the crash file from the last reportError call.
func LastReportPath() string {
	reportMu.Lock()
	defer reportMu.Unlock()

	return lastReport
}
