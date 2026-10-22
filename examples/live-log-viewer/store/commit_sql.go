package store

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

const insertEntryTail = ` ON CONFLICT(source_id, generation, position) DO UPDATE SET
	session_id  = excluded.session_id,
	ts_ns       = excluded.ts_ns,
	ts_raw      = excluded.ts_raw,
	ts_ok       = excluded.ts_ok,
	severity    = excluded.severity,
	message     = excluded.message,
	bytes       = excluded.bytes,
	multiline   = excluded.multiline,
	truncated   = excluded.truncated,
	malformed   = excluded.malformed,
	partial     = excluded.partial,
	received_ns = excluded.received_ns
	WHERE entries.partial = 1`

func insertSQL(rows []entry.Entry, seq *int64) (string, []any) {
	var b strings.Builder

	b.WriteString(`INSERT INTO entries
		(session_id, source_id, source_seq, position, generation, ts_ns,
		 ts_raw, ts_ok, severity, message, bytes, multiline, truncated,
		 malformed, partial, received_ns) VALUES `)

	args := make([]any, 0, len(rows)*insertColsPerRow)

	for i, e := range rows {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		*seq++
		args = append(args,
			int64(e.Session), int64(e.Source), *seq, e.Position, e.Generation,
			tsNanos(e), e.TimeRaw, boolInt(e.TimeOK), int64(e.Severity),
			e.Message, e.Bytes, boolInt(e.Multiline), boolInt(e.Truncated),
			boolInt(e.Malformed), boolInt(e.Partial), e.Received.UnixNano())
	}

	return b.String() + insertEntryTail, args
}

func tsNanos(e entry.Entry) int64 {
	if e.Time.IsZero() {
		return 0
	}

	return e.Time.UnixNano()
}

func boolInt(v bool) int64 {
	if v {
		return 1
	}

	return 0
}
