package ownframe

import (
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

// ErrPanicked means a panic was recovered.
var ErrPanicked = errors.New("ownframe: panic")

// reportError stores err in a crash report file and returns an error
// naming that file. Run calls it for every startup and window failure,
// so no failure leaves only a terminal line. The returned error keeps
// the original for errors.Is. A nil page keeps a blank title.
func reportError(page *Page, err error) error {
	title := ""
	if page != nil {
		title = page.Title()
	}

	path, werr := crash.Write(title, err.Error())
	if werr != nil {
		return fmt.Errorf("ownframe: %w: %s: %w", err, path, werr)
	}

	return fmt.Errorf("ownframe: %w: %s", err, path)
}
