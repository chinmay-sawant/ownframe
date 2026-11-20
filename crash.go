package ownframe

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/crash"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"sync"
)

// Report writes a local crash file and returns its path.
func Report(title, reason string) (string, error) {
	return crash.Write(title, reason)
}

// SetCrashDir sets the directory Report and a recovered panic use.
// It rejects an empty or unsafe path. An empty value used to
// restore the default; callers should keep a previous value instead.
func SetCrashDir(dir string) error {
	return crash.SetDir(dir)
}

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

var (
	// ErrEmptyHTML means New was given a blank template.
	ErrEmptyHTML = page.ErrEmptyHTML

	// ErrBadSource means a source file could not be read at New, or a
	// Config set both HTML and File, or both Theme and ThemeFile.
	ErrBadSource = page.ErrBadSource

	// ErrBadSize means a width or a height is unusable.
	ErrBadSize = page.ErrBadSize

	// ErrNilPage means Run, Serve, or BindMobile was called without a page.
	ErrNilPage = page.ErrNilPage

	// ErrSourceTooLarge means a template, theme, or source string exceeded
	// MaxSourceSize.
	ErrSourceTooLarge = page.ErrSourceTooLarge

	// ErrPatchRefused means PatchElement could not apply the edit.
	ErrPatchRefused = page.ErrPatchRefused

	// ErrNoDialog means this system has no file dialog.
	ErrNoDialog = filepick.ErrNoDialog

	// ErrCanceled means the person closed the file dialog.
	ErrCanceled = filepick.ErrCanceled
)
