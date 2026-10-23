package gesture

import (
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestDropLimitCountAndBytes(t *testing.T) {
	t.Parallel()

	files := make([]host.Drop, MaxDropFiles+1)
	if got := len(Limit(files)); got != MaxDropFiles {
		t.Fatalf("files %d", got)
	}

	dir := host.Drop{Name: "dir", IsDir: true, Read: func() ([]byte, error) {
		return []byte("no"), nil
	}}
	if got := limitOne(dir, 1); got.Read != nil {
		data, err := got.Read()
		if err != nil || string(data) != "no" {
			t.Fatalf("dir read %q %v", data, err)
		}
	}

	plain := host.Drop{Name: "a.txt", Read: func() ([]byte, error) {
		return []byte("hello"), nil
	}}
	wrapped := limitOne(plain, 4)
	_, err := wrapped.Read()
	if err == nil || !strings.Contains(err.Error(), "a.txt exceeds 4 bytes") {
		t.Fatalf("err %v", err)
	}

	plain.Read = func() ([]byte, error) { return []byte("ok"), nil }
	data, err := limitOne(plain, 4).Read()
	if err != nil || string(data) != "ok" {
		t.Fatalf("data %q %v", data, err)
	}

	boom := errors.New("read failed")
	plain.Read = func() ([]byte, error) { return nil, boom }
	if _, err = limitOne(plain, 4).Read(); !errors.Is(err, boom) {
		t.Fatalf("err %v", err)
	}

	if got := limitOne(host.Drop{Name: "bare"}, 4); got.Read != nil {
		t.Fatal("nil reader was wrapped")
	}
}
