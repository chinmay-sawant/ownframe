// Package ws is a small WebSocket client for the TV.
// The handshake sends no Origin header. webOS rejects some browser origins,
// and a desktop or phone client is not a page.
// This package has no server side. Any future server must check Origin
// against an allowlist before completing the upgrade.
package ws

import "errors"

// ErrHandshake means the server did not accept the upgrade.
var ErrHandshake = errors.New("ws: handshake was refused")

// ErrTooLarge means a frame or message passed the read limit.
var ErrTooLarge = errors.New("ws: message too large")
