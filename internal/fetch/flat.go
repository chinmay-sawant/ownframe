package fetch

import (
	"net/http"
	"strings"
)

func flat(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for key, values := range h {
		out[key] = strings.Join(values, ", ")
	}

	return out
}
