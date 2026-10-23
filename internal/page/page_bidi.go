package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// SanitizeBidi strips bidi controls from untrusted text. Call it on
// chat, notification, and other app-external strings before SetData.
func SanitizeBidi(s string) string { return draw.SanitizeBidi(s) }

// CheckIdentifier enforces the precis nickname profile on handles,
// room names, and other identifiers rendered from untrusted input.
func CheckIdentifier(s string) error { return draw.CheckIdentifier(s) }
