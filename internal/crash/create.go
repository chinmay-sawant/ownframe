package crash

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func create(folder, body string) (string, error) {
	stamp := time.Now().UTC().Format("20060102-150405")
	for n := 0; n < 1000; n++ {
		name := stamp + ".txt"
		if n > 0 {
			name = fmt.Sprintf("%s-%d.txt", stamp, n)
		}

		path := filepath.Join(folder, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return "", err
		}

		if err := f.Chmod(0o600); err != nil {
			cerr := f.Close()

			return "", errors.Join(err, cerr)
		}

		_, werr := f.WriteString(body)
		cerr := f.Close()
		if werr != nil || cerr != nil {
			return "", errors.Join(werr, cerr)
		}

		return path, nil
	}

	return "", fmt.Errorf("crash: no free name")
}
