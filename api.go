// Package ownframe shows an HTML page in a window.
//
// A program passes a template to New, stores the template data with SetData,
// and registers Handlers for clicks and keys. An optional theme stylesheet
// passed to New or SetTheme restyles the page after the template's own
// styles. Run opens the window. Serve shows the same picture in a browser.
// BindMobile registers the page for an Android or iOS bind. Redraw parses
// the HTML, applies the CSS, and lays the page out through blinkless.
// This package has no layout of its own.
package ownframe

import "github.com/chinmay-sawant/ownframe/internal/page"

// Box is one element a click can land on.
// X, Y, W, and H are CSS pixels, origin at the top left of the picture.
type Box = page.Box

// Config is the template and the frame size for a new page.
// Width and Height are the first frame, in CSS pixels.
// MinWidth and MinHeight are the smallest frame. Zero means 1.
// MaxWidth and MaxHeight ask the host to cap the window. Zero means no cap.
// They do not cap the picture, so the layout follows the window.
// The zero Config is safe: developer overlays and pipeline sampling stay
// off until DevTools or Perf opts in.
type Config = page.Config

// Handlers contains optional callbacks for input and host-delivered links.
type Handlers = page.Handlers

// Page is one HTML template and the last picture it produced.
type Page struct {
	*page.Page
}

const MaxSourceSize = page.MaxSourceSize

// New parses html as an html/template page.
func New(cfg Config) (*Page, error) {
	return wrapPage(page.New(cfg))
}
