package reader

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// dummyLine builds one deterministic record covering the shapes the viewer
// must survive: mixed severity, Unicode, malformed timestamps, long entries,
// repeated messages, embedded newlines, and invalid encoding.
func dummyLine(seed int64, key string, seq int64) []byte {
	if seq%dummyMod73 == 1 {
		return []byte("\tat main.handler(server.go:42)")
	}

	if seq%dummyMod73 == dummySevA {
		return []byte("\t... 3 more")
	}

	r := mix64(uint64(seed), key, uint64(seq)) //nolint:gosec // reason: seed and seq are non-negative dummy counters, conversion cannot overflow
	ts := DummyBase.Add(time.Duration(seq)*250*time.Millisecond +
		time.Duration(r%dummyJitter)*time.Millisecond)

	var b strings.Builder

	switch {
	case seq%dummyMod79 == 0:
		b.WriteString("2026-13-45 99:61:61 ")
	case seq%dummyMod61 == 0:
	case seq%dummyMod31 == 0:
		b.WriteString(strconv.FormatInt(ts.Unix(), 10) + " ")
	default:
		b.WriteString(ts.Format("2006-01-02T15:04:05.000Z") + " ")
	}

	sev := severityFor(r)
	if seq%dummyMod73 == 0 {
		sev = entry.Error
	}

	showSev := seq%dummyMod13 != 0
	json := seq%dummyMod5 == 0 && showSev

	switch {
	case json:
		fmt.Fprintf(&b, `{"level":"%s","msg":"`, strings.ToLower(sev.String()))
	case seq%dummyMod3 == 0 && showSev:
		fmt.Fprintf(&b, "[%s] ", sev.String())
	case showSev:
		fmt.Fprintf(&b, "%s ", sev.String())
	}

	b.WriteString(dummyMessage(key, seq, r))

	if json {
		b.WriteString(`"}`)
	}

	return []byte(b.String())
}
