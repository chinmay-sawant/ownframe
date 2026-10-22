package reader

import (
	"encoding/binary"
	"hash/fnv"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// DummyBase is the timestamp the demo generator starts from.
var DummyBase = time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)

// DummyRecords returns n reproducible records starting at seq. The same
// seed, key, and sequence always produce the same bytes; nothing uses the
// process random source, so a fixture replays after a restart.
func DummyRecords(seed int64, key string, from int64, n, maxBytes int) []entry.RawRecord {
	if n <= 0 {
		return nil
	}

	out := make([]entry.RawRecord, 0, n)

	for i := 0; i < n; i++ {
		seq := from + int64(i)
		data := dummyLine(seed, key, seq)
		rec := entry.RawRecord{Path: key, Generation: 1, Offset: seq, Bytes: 1}

		if maxBytes > 0 && len(data) > maxBytes {
			rec.Truncated = true
			data = data[:maxBytes]
		}

		rec.Data = data
		out = append(out, rec)
	}

	return out
}

// mix64 hashes seed, key, and sequence into the record's field choices. It
// is deterministic across processes and Go versions.
func mix64(seed uint64, key string, seq uint64) uint64 {
	h := fnv.New64a()

	var buf [8]byte

	binary.LittleEndian.PutUint64(buf[:], seed)
	_, _ = h.Write(buf[:])
	_, _ = h.Write([]byte(key))
	binary.LittleEndian.PutUint64(buf[:], seq)
	_, _ = h.Write(buf[:])

	x := h.Sum64()
	x ^= x >> mixShiftA
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> mixShiftB
	x *= 0x94d049bb133111eb

	return x ^ (x >> mixShiftC)
}
