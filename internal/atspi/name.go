package atspi

import (
	"encoding/hex"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// DefaultName is the only well-known name this process may claim.
const (
	DefaultName        = "org.a11y.atspi.ownframe.app"
	ApplicationName    = "ownframe"
	accessibleRootPath = "/org/a11y/atspi/accessible/root"
)

const namePrefix = "org.a11y.atspi.ownframe."
const accessiblePath = "/org/a11y/atspi/accessible/"
const nameBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"

// CheckName allows only org.a11y.atspi.ownframe.* with a simple suffix.
// It rejects org.a11y.Bus, the registry, and any other D-Bus name.
func CheckName(name string) error {
	if !strings.HasPrefix(name, namePrefix) {
		return ErrName
	}
	rest := name[len(namePrefix):]
	if rest == "" {
		return ErrName
	}
	for i := 0; i < len(rest); i++ {
		if !strings.ContainsRune(nameBytes, rune(rest[i])) {
			return ErrName
		}
	}
	return nil
}

func objectPath(id host.AccessibilityNodeID) string {
	return accessiblePath + "n" + hex.EncodeToString([]byte(id))
}
