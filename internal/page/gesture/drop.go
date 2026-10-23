package gesture

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// MaxDropFiles caps one drop. MaxDropBytes caps one file's bytes, matching
// the window's stream limit for callers that skip the window.
const (
	MaxDropFiles = 32
	MaxDropBytes = 16 << 20
)

// Limit caps the file count and wraps each reader at MaxDropBytes.
func Limit(files []host.Drop) []host.Drop {
	return limit(files, MaxDropFiles, MaxDropBytes)
}

func limit(files []host.Drop, maxFiles int, maxBytes int64) []host.Drop {
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	for i := range files {
		files[i] = limitOne(files[i], maxBytes)
	}

	return files
}

func limitOne(d host.Drop, maxBytes int64) host.Drop {
	if d.IsDir || d.Read == nil {
		return d
	}

	next := d.Read
	name := d.Name
	d.Read = func() ([]byte, error) {
		data, err := next()
		if err != nil {
			return nil, err
		}

		if int64(len(data)) > maxBytes {
			return nil, fmt.Errorf("drop: %s exceeds %d bytes", name, maxBytes)
		}

		return data, nil
	}

	return d
}
