package atspi

import (
	"strconv"
	"strings"
)

// DefaultName is the only well-known name this process may claim.
const DefaultName = "org.a11y.atspi.ownframe.app"

const namePrefix = "org.a11y.atspi.ownframe."
const accessiblePath = "/org/a11y/atspi/accessible/"
const nameBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

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

func objectPath(index int) string {
	return accessiblePath + strconv.Itoa(index)
}
