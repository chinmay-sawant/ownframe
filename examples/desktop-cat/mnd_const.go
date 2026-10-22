//go:build !js && !android && !ios

package main

import "time"

// Notification and media tuning for the desktop cat launcher.
const (
	defaultMargin      = 48
	mediaTextLimit     = 160
	mediaRetryInterval = 5 * time.Second
	postTimeout        = 2 * time.Second
)
