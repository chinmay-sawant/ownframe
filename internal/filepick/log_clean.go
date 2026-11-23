package filepick

import (
	"fmt"
	"os"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/debugtag"
)

// cleanLog strips CR and LF so failure text cannot forge log lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}

// cleanArgs sanitizes debug values so paths and errors stay on one line.
func cleanArgs(args []any) []any {
	for i, a := range args {
		switch v := a.(type) {
		case string:
			args[i] = cleanLog(v)
		case error:
			if v != nil {
				args[i] = cleanLog(v.Error())
			}
		default:
			s := fmt.Sprint(a)
			if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
				args[i] = cleanLog(s)
			}
		}
	}
	return args
}

// debugf writes to stderr when OWNFRAME_FILEPICK_DEBUG is not empty and the
// debug tag is built in. Values are sanitized to one line. Keep this flag off
// in shipped builds.
func debugf(format string, args ...any) {
	if !debugtag.Enabled {
		return
	}
	if os.Getenv("OWNFRAME_FILEPICK_DEBUG") == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "filepick: "+format+"\n", cleanArgs(args)...)
}
