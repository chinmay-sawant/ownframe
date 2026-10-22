package transfer

import (
	"fmt"
	"hash/fnv"
)

// fakeHash is the stable per-job hash behind the dummy transport.
func fakeHash(jobID string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(jobID))

	return h.Sum64()
}

// FakeSize is the deterministic body size for a dummy job: 128 KiB plus up
// to two mebibytes from the job ID hash.
func FakeSize(jobID string) int64 {
	return fakeBaseSize + int64(fakeHash(jobID)%(fakeSizeSpan))
}

// fakeUnknown reports whether this job's server hides Content-Length.
func fakeUnknown(jobID string) bool { return fakeHash(jobID)%fakeUnknownMod == 0 }

// fakeFails reports whether this job drops the connection once, on the
// first attempt, at about 60 percent.
func fakeFails(jobID string) bool { return fakeHash(jobID)%fakeFailMod == 0 }

// bodyBytes builds the deterministic body slice for offset..offset+n. The
// pattern is non-repeating enough to catch a bad resume offset.
func bodyBytes(offset, n int64) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte((offset+int64(i))*bodyHashBase + bodyStep) //nolint:gosec // reason: offset and index bounded, test body fits a byte.
	}

	return out
}

// FakeChecksum returns the "sha256:<hex>" of a dummy job's whole body.
// Dummy data uses it to exercise the supplied-checksum path.
func FakeChecksum(jobID string) string {
	h := newSHA256()
	size := FakeSize(jobID)
	_ = fakeWriteHash(h, size)

	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}
