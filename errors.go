package ownframe

import (
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

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
