package filepick

import (
	"fmt"
	"strings"
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
