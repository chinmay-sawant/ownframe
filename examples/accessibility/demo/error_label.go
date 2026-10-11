package demo

import (
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

func errorLabel(err error) string {
	labels := []struct {
		err   error
		label string
	}{
		{ownframe.ErrAccessibilityStale, "ErrAccessibilityStale"},
		{ownframe.ErrAccessibilityNode, "ErrAccessibilityNode"},
		{ownframe.ErrAccessibilityAction, "ErrAccessibilityAction"},
		{ownframe.ErrAccessibilityDisabled, "ErrAccessibilityDisabled"},
		{ownframe.ErrAccessibilityRange, "ErrAccessibilityRange"},
		{ownframe.ErrAccessibilityUnavailable, "ErrAccessibilityUnavailable"},
	}
	for _, item := range labels {
		if errors.Is(err, item.err) {
			return item.label
		}
	}
	if err == nil {
		return "no error"
	}
	return err.Error()
}
