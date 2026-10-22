package transfer

import (
	"bytes"
	"context"
	"os"
	"testing"
)

// TestFakeRestartsLargerPartial restarts when the partial is too long.
func TestFakeRestartsLargerPartial(t *testing.T) {
	dest, partial := fakePaths(t)
	id := "larger-1"

	if err := os.WriteFile(partial, bodyBytes(0, FakeSize(id)+10), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := fastFake().Download(context.Background(), Request{
		JobID: id, Dest: dest, Partial: partial,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if out.Resumed {
		t.Error("larger partial reported resumed")
	}

	if got := readAll(t, dest); !bytes.Equal(got, bodyBytes(0, FakeSize(id))) {
		t.Error("restarted body mismatch")
	}
}
