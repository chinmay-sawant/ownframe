package draw

import (
	"strings"
	"time"
)

// Execute runs run into a builder grown to hint bytes when hint > 0.
// When track is true it times the run and stores the duration in last.
// A nil last skips the store. run receives the builder.
func Execute(track bool, hint int, last *time.Duration, run func(*strings.Builder) error) (string, error) {
	start := Clock(track)

	var body strings.Builder
	if hint > 0 {
		body.Grow(hint)
	}

	err := run(&body)
	if last != nil {
		Note(track, last, start)
	}
	if err != nil {
		return "", err
	}

	return body.String(), nil
}
