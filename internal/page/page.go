// Package page turns an HTML template into a picture and hit boxes.
package page

import (
	"errors"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// Box is one element a click can land on.
// X, Y, W, H are CSS pixels, origin at the top left of the picture.
type Box = layout.Box

var (
	// ErrEmptyHTML means New was given a blank template.
	ErrEmptyHTML = errors.New("ownframe: empty html")

	// ErrBadSource means a source file could not be read at New, or a
	// Config set both HTML and File, or both Theme and ThemeFile.
	ErrBadSource = errors.New("ownframe: bad source")

	// ErrBadSize means a width or a height is unusable.
	ErrBadSize = errors.New("ownframe: bad size")

	// ErrNilPage means Run, Serve, or BindMobile was called without a page.
	ErrNilPage = errors.New("ownframe: nil page")

	errNilContext = fetch.ErrNilContext

	// errNoImage is the resolver answer for a src SetImage does not hold.
	errNoImage = errors.New("ownframe: no image for src")
)
