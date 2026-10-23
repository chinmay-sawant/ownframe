package atspi

import "errors"

var (
	ErrLiveUnavailable = errors.New("ownframe: atspi live bus unavailable")
	ErrName            = errors.New("ownframe: atspi dbus name rejected")
)
