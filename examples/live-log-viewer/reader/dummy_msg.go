package reader

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func dummyMessage(key string, seq int64, r uint64) string {
	var msg string

	switch key {
	case "api":
		msg = fmt.Sprintf("GET /v1/users/%d 200 %dms", r%dummyUsers, r%dummyMsA+1)
	case "worker":
		msg = fmt.Sprintf("job %d finished in %dms", seq/dummyMod3, r%dummyMsB+1)
	case "db":
		msg = fmt.Sprintf("checkpoint %d pages wal=%dK", seq/dummyPages, r%dummyWal)
	default:
		msg = fmt.Sprintf("%s record %d", key, seq)
	}

	switch {
	case seq%dummyMod71 < dummySevA:
		msg = "connection refused; retrying"
	case seq%dummyMod73 == 0:
		msg = "panic: runtime error: index out of range"
	case seq%dummyMod67 == 0:
		msg = "stack for request\n\tat net/http.serverHandler\n\tat main.serve"
	}

	if seq%dummyMod89 == 0 {
		msg += " — очередь ✓ 日本語 café 🚀"
	}

	if seq%dummyMod83 == 0 {
		msg += " " + strings.Repeat("payload ", dummyRepeat)
	}

	if seq%dummyMod101 == 0 {
		msg += "\x80\x80"
	}

	return msg
}

func severityFor(r uint64) entry.Severity {
	switch n := r % dummySevMod; {
	case n < dummySevA:
		return entry.Trace
	case n < dummySevB:
		return entry.Debug
	case n < dummySevC:
		return entry.Info
	case n < dummySevD:
		return entry.Warn
	case n < dummySevE:
		return entry.Error
	default:
		return entry.Fatal
	}
}
